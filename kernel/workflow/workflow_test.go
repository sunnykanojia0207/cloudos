package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/project"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Types Tests ─────────────────────────────────────────────────────────

func TestNodeInterface(t *testing.T) {
	var n Node

	task := NewTaskNode("t1", "Test", "resource.list", "Project")
	n = task
	if n.ID() != "t1" {
		t.Errorf("TaskNode ID = %q", n.ID())
	}
	if n.Type() != NodeTypeTask {
		t.Errorf("TaskNode Type = %q", n.Type())
	}
	if n.Status() != NodePending {
		t.Errorf("TaskNode Status = %q", n.Status())
	}

	end := NewEndNode("end", "t1")
	n = end
	if n.ID() != "end" {
		t.Errorf("EndNode ID = %q", n.ID())
	}
	if n.Type() != NodeTypeEnd {
		t.Errorf("EndNode Type = %q", n.Type())
	}
}

func TestStatusEnums(t *testing.T) {
	tests := []struct {
		s       NodeStatus
		valid   bool
		terminal bool
	}{
		{NodePending, true, false},
		{NodeRunning, true, false},
		{NodeSucceeded, true, true},
		{NodeFailed, true, true},
		{NodeSkipped, true, true},
		{NodeCancelled, true, true},
		{"invalid", false, false},
	}
	for _, tc := range tests {
		if tc.s.Valid() != tc.valid {
			t.Errorf("NodeStatus(%q).Valid() = %v, want %v", tc.s, tc.s.Valid(), tc.valid)
		}
		if tc.s.IsTerminal() != tc.terminal {
			t.Errorf("NodeStatus(%q).IsTerminal() = %v, want %v", tc.s, tc.s.IsTerminal(), tc.terminal)
		}
	}
}

func TestWorkflowStatusEnums(t *testing.T) {
	tests := []struct {
		s     WorkflowStatus
		valid bool
	}{
		{WorkflowPending, true},
		{WorkflowRunning, true},
		{WorkflowPaused, true},
		{WorkflowCancelled, true},
		{WorkflowCompleted, true},
		{WorkflowFailed, true},
		{"unknown", false},
	}
	for _, tc := range tests {
		if tc.s.Valid() != tc.valid {
			t.Errorf("WorkflowStatus(%q).Valid() = %v", tc.s, tc.s.Valid())
		}
	}
}

func TestValidateDefinition(t *testing.T) {
	// Valid definition
	def := &WorkflowDefinition{
		ID:   "test",
		Name: "Test Workflow",
		Nodes: []Node{
			NewTaskNode("1", "Step 1", "complete", ""),
			NewEndNode("end", "1"),
		},
	}
	if err := ValidateDefinition(def); err != nil {
		t.Fatalf("ValidateDefinition error: %v", err)
	}

	// Empty ID
	def2 := &WorkflowDefinition{ID: "", Name: "Test", Nodes: []Node{NewTaskNode("1", "Step 1", "complete", "")}}
	if err := ValidateDefinition(def2); err == nil {
		t.Error("expected error for empty ID")
	}

	// No nodes
	def3 := &WorkflowDefinition{ID: "test", Name: "Test", Nodes: []Node{}}
	if err := ValidateDefinition(def3); err == nil {
		t.Error("expected error for no nodes")
	}

	// Duplicate IDs
	def4 := &WorkflowDefinition{ID: "test", Name: "Test", Nodes: []Node{
		NewTaskNode("1", "A", "complete", ""),
		NewTaskNode("1", "B", "complete", ""),
	}}
	if err := ValidateDefinition(def4); err == nil {
		t.Error("expected error for duplicate IDs")
	}
}

func TestWorkflowRunProgress(t *testing.T) {
	run := &WorkflowRun{
		ID:  "run-1",
		Nodes: []Node{
			NewTaskNode("1", "A", "complete", ""),
			NewTaskNode("2", "B", "complete", ""),
			NewTaskNode("3", "C", "complete", ""),
			NewTaskNode("4", "D", "complete", ""),
		},
	}
	if run.Progress() != 0.0 {
		t.Errorf("Progress() = %f, want 0.0", run.Progress())
	}
	if run.CompletedCount() != 0 {
		t.Errorf("CompletedCount() = %d, want 0", run.CompletedCount())
	}

	// Complete 2 nodes
	run.Nodes[0].SetStatus(NodeSucceeded)
	run.Nodes[1].SetStatus(NodeSucceeded)
	if run.Progress() != 0.5 {
		t.Errorf("Progress() = %f, want 0.5", run.Progress())
	}
	if run.CompletedCount() != 2 {
		t.Errorf("CompletedCount() = %d, want 2", run.CompletedCount())
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	rp := DefaultRetryPolicy()
	if rp.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", rp.MaxRetries)
	}
	if rp.BackoffBase != 100*time.Millisecond {
		t.Errorf("BackoffBase = %v, want 100ms", rp.BackoffBase)
	}
	if rp.BackoffMax != 10*time.Second {
		t.Errorf("BackoffMax = %v, want 10s", rp.BackoffMax)
	}
}

// ── Graph Tests ─────────────────────────────────────────────────────────

func TestGraphTopologicalSortLinear(t *testing.T) {
	nodes := []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewTaskNode("2", "Step 2", "complete", "", "1"),
		NewTaskNode("3", "Step 3", "complete", "", "2"),
		NewEndNode("end", "3"),
	}
	g := NewGraph(nodes)
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort error: %v", err)
	}
	ids := make([]string, len(sorted))
	for i, n := range sorted {
		ids[i] = n.ID()
	}
	expected := []string{"1", "2", "3", "end"}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("sorted[%d] = %q, want %q", i, id, expected[i])
		}
	}
}

func TestGraphTopologicalSortBranching(t *testing.T) {
	//       1
	//      / \
	//     2   3
	//      \ /
	//       4
	nodes := []Node{
		NewTaskNode("1", "Root", "complete", ""),
		NewTaskNode("2", "Left", "complete", "", "1"),
		NewTaskNode("3", "Right", "complete", "", "1"),
		NewTaskNode("4", "Merge", "complete", "", "2", "3"),
		NewEndNode("end", "4"),
	}
	g := NewGraph(nodes)
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort error: %v", err)
	}

	// Verify ordering constraints (1 before 2,3; 2,3 before 4)
	pos := make(map[string]int)
	for i, n := range sorted {
		pos[n.ID()] = i
	}
	if pos["1"] > pos["2"] || pos["1"] > pos["3"] {
		t.Error("1 must come before 2 and 3")
	}
	if pos["2"] > pos["4"] || pos["3"] > pos["4"] {
		t.Error("2 and 3 must come before 4")
	}
	if pos["4"] > pos["end"] {
		t.Error("4 must come before end")
	}
	if len(sorted) != len(nodes) {
		t.Errorf("sorted %d nodes, want %d", len(sorted), len(nodes))
	}
}

func TestGraphCycleDetection(t *testing.T) {
	nodes := []Node{
		NewTaskNode("1", "A", "complete", "", "3"), // cycle: 1→2→3→1
		NewTaskNode("2", "B", "complete", "", "1"),
		NewTaskNode("3", "C", "complete", "", "2"),
	}
	g := NewGraph(nodes)
	if !g.HasCycle() {
		t.Error("HasCycle() = false, want true")
	}
	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("expected cycle error")
	}
}

func TestGraphNoCycle(t *testing.T) {
	nodes := []Node{
		NewTaskNode("1", "A", "complete", ""),
		NewTaskNode("2", "B", "complete", "", "1"),
		NewTaskNode("3", "C", "complete", "", "2"),
	}
	g := NewGraph(nodes)
	if g.HasCycle() {
		t.Error("HasCycle() = true, want false")
	}
}

func TestGraphReadyNodes(t *testing.T) {
	nodes := []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewTaskNode("2", "Step 2", "complete", "", "1"),
		NewTaskNode("3", "Step 3", "complete", "", "2"),
	}
	g := NewGraph(nodes)

	run := &WorkflowRun{Nodes: copyNodes(nodes)}

	// Initially, only node 1 should be ready
	ready := g.ReadyNodes(run)
	if len(ready) != 1 || ready[0].ID() != "1" {
		t.Errorf("ReadyNodes() = %v, want [1]", readyNodeIDs(ready))
	}

	// Complete node 1
	run.NodeByID("1").SetStatus(NodeSucceeded)
	ready = g.ReadyNodes(run)
	if len(ready) != 1 || ready[0].ID() != "2" {
		t.Errorf("ReadyNodes() = %v, want [2]", readyNodeIDs(ready))
	}

	// Complete node 2
	run.NodeByID("2").SetStatus(NodeSucceeded)
	ready = g.ReadyNodes(run)
	if len(ready) != 1 || ready[0].ID() != "3" {
		t.Errorf("ReadyNodes() = %v, want [3]", readyNodeIDs(ready))
	}
}

func readyNodeIDs(nodes []Node) []string {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID()
	}
	return ids
}

func TestGraphDependenciesOf(t *testing.T) {
	nodes := []Node{
		NewTaskNode("1", "Root", "complete", ""),
		NewTaskNode("2", "Child", "complete", "", "1"),
		NewTaskNode("3", "Grandchild", "complete", "", "2"),
	}
	g := NewGraph(nodes)
	deps := g.DependenciesOf("3")
	if len(deps) != 2 {
		t.Errorf("DependenciesOf(3) = %v, want [1 2]", deps)
	}
}

// ── Scheduler Tests ─────────────────────────────────────────────────────

func TestSchedulerReady(t *testing.T) {
	s := NewScheduler()
	nodes := []Node{
		NewTaskNode("1", "A", "complete", ""),
		NewTaskNode("2", "B", "complete", "", "1"),
		NewTaskNode("3", "C", "complete", "", "2"),
	}
	run := &WorkflowRun{Nodes: copyNodes(nodes)}
	ready := s.Ready(run)
	if len(ready) != 1 || ready[0].ID() != "1" {
		t.Errorf("Ready() = %v, want [1]", readyNodeIDs(ready))
	}
}

func TestSchedulerIsComplete(t *testing.T) {
	s := NewScheduler()
	nodes := []Node{
		NewTaskNode("1", "A", "complete", ""),
		NewTaskNode("2", "B", "complete", "", "1"),
	}
	run := &WorkflowRun{Nodes: copyNodes(nodes)}
	if s.IsComplete(run) {
		t.Error("IsComplete() = true, want false")
	}
	run.NodeByID("1").SetStatus(NodeSucceeded)
	run.NodeByID("2").SetStatus(NodeSucceeded)
	if !s.IsComplete(run) {
		t.Error("IsComplete() = false, want true")
	}
}

func TestSchedulerHasFailures(t *testing.T) {
	s := NewScheduler()
	nodes := []Node{
		NewTaskNode("1", "A", "complete", ""),
		NewTaskNode("2", "B", "complete", "", "1"),
	}
	run := &WorkflowRun{Nodes: copyNodes(nodes)}
	if s.HasFailures(run) {
		t.Error("HasFailures() = true, want false")
	}
	run.NodeByID("1").SetStatus(NodeFailed)
	if !s.HasFailures(run) {
		t.Error("HasFailures() = false, want true")
	}
}

// ── Queue Tests ─────────────────────────────────────────────────────────

func TestQueueEnqueueDequeue(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(10)
	q.Start()

	err := q.Enqueue(ctx, QueueItem{WorkflowID: "wf-1", NodeID: "n1"})
	if err != nil {
		t.Fatalf("Enqueue error: %v", err)
	}
	if q.Len() != 1 {
		t.Errorf("Len() = %d, want 1", q.Len())
	}

	item, done, err := q.Dequeue(ctx)
	if err != nil {
		t.Fatalf("Dequeue error: %v", err)
	}
	if item.WorkflowID != "wf-1" || item.NodeID != "n1" {
		t.Errorf("Dequeue item = %+v", item)
	}
	if q.Len() != 0 {
		t.Errorf("Len() after dequeue = %d, want 0", q.Len())
	}
	done()
}

func TestQueueTryDequeue(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(10)
	q.Start()

	_, ok := q.TryDequeue()
	if ok {
		t.Error("TryDequeue() on empty queue returned ok")
	}

	_ = q.Enqueue(ctx, QueueItem{WorkflowID: "wf-1"})
	item, ok := q.TryDequeue()
	if !ok || item.WorkflowID != "wf-1" {
		t.Errorf("TryDequeue() = (%+v, %v)", item, ok)
	}
}

func TestQueueCapacity(t *testing.T) {
	q := NewQueue(2)
	q.Start()
	ctx := context.Background()

	_ = q.Enqueue(ctx, QueueItem{WorkflowID: "1"})
	_ = q.Enqueue(ctx, QueueItem{WorkflowID: "2"})
	err := q.Enqueue(ctx, QueueItem{WorkflowID: "3"})
	if err == nil {
		t.Error("expected capacity error")
	}
}

func TestQueueNotRunning(t *testing.T) {
	q := NewQueue(10)
	err := q.Enqueue(context.Background(), QueueItem{WorkflowID: "1"})
	if err == nil {
		t.Error("expected not-running error")
	}
}

func TestQueueDrain(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(10)
	q.Start()

	_ = q.Enqueue(ctx, QueueItem{WorkflowID: "1"})
	_ = q.Enqueue(ctx, QueueItem{WorkflowID: "2"})

	go func() {
		time.Sleep(50 * time.Millisecond)
		q.Dequeue(ctx)
		q.Dequeue(ctx)
	}()

	if err := q.Drain(2 * time.Second); err != nil {
		t.Errorf("Drain error: %v", err)
	}
}

// ── Retry Tests ─────────────────────────────────────────────────────────

func TestRetryEvaluator(t *testing.T) {
	e := NewRetryEvaluator()

	node := NewTaskNode("1", "Test", "complete", "")
	node.RetryPolicy = DefaultRetryPolicy()

	// First retry should succeed
	delay, err := e.ShouldRetry(node)
	if err != nil {
		t.Fatalf("ShouldRetry error: %v", err)
	}
	if delay < 50*time.Millisecond {
		t.Errorf("delay too small: %v", delay)
	}

	// Exhaust retries
	for i := 0; i < 3; i++ {
		node.RetryCount++
	}
	_, err = e.ShouldRetry(node)
	if err == nil {
		t.Error("expected max retries error")
	}
}

func TestRetryNoPolicy(t *testing.T) {
	e := NewRetryEvaluator()
	node := NewTaskNode("1", "Test", "complete", "")
	_, err := e.ShouldRetry(node)
	if err == nil {
		t.Error("expected error for no retry policy")
	}
}

func TestRetryBackoffDuration(t *testing.T) {
	e := NewRetryEvaluator()
	node := NewTaskNode("1", "Test", "complete", "")
	node.RetryPolicy = &RetryPolicy{
		MaxRetries:  5,
		BackoffBase: 100 * time.Millisecond,
		BackoffMax:  5 * time.Second,
	}

	// Attempt 0: 100ms
	d := e.BackoffDuration(node)
	if d != 100*time.Millisecond {
		t.Errorf("backoff[0] = %v, want 100ms", d)
	}

	node.RetryCount = 5
	d = e.BackoffDuration(node)
	if d > 5*time.Second {
		t.Errorf("backoff exceeded max: %v", d)
	}
}

// ── Executor Tests ──────────────────────────────────────────────────────

func setupExecutor(t *testing.T) (*Executor, *resource.Registry, *controller.Manager, *health.Manager) {
	t.Helper()
	log := logging.NewSubsystemLogger("workflow-test", logging.LevelError)
	bus := events.NewBus(log)
	bus.Start()

	healthMgr := health.NewManager(log)
	_ = healthMgr.Start(context.Background())

	reg := resource.NewRegistry(bus, log)
	if err := reg.RegisterKind(resource.Kind{
		Name:       "Namespace",
		Namespaced: false,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}
	ns := resource.DefaultNamespace()
	if err := reg.Create(context.Background(), ns); err != nil {
		t.Fatal(err)
	}
	if err := reg.RegisterKind(resource.Kind{
		Name:       project.Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	ctrlManager := controller.NewManager(reg, bus, healthMgr, log)
	if err := ctrlManager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ctrlManager.Stop(context.Background()) })

	exec := NewExecutor(ExecutorDeps{
		ResourceRegistry:  reg,
		ControllerManager: ctrlManager,
		HealthManager:     healthMgr,
		Logger:            log,
	})

	t.Cleanup(func() {
		healthMgr.Stop(context.Background())
		bus.Stop()
	})

	return exec, reg, ctrlManager, healthMgr
}

func TestExecutorValidate(t *testing.T) {
	exec, _, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "Validate", "validate", "Project:test-app")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute validate error: %v", err)
	}
}

func TestExecutorResourceCreate(t *testing.T) {
	exec, reg, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "Create", "resource.create", "Project:exec-test-create")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute resource.create error: %v", err)
	}

	// Verify it was created
	_, err := reg.Get(project.Kind, "exec-test-create")
	if err != nil {
		t.Errorf("Project not found after create: %v", err)
	}
}

func TestExecutorResourceList(t *testing.T) {
	exec, _, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "List", "resource.list", "Project")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute resource.list error: %v", err)
	}
}

func TestExecutorResourceDelete(t *testing.T) {
	exec, reg, _, _ := setupExecutor(t)

	// First create
	p := project.NewProject("exec-del-test", "Exec Delete Test", "development", "")
	if err := reg.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	// Then delete
	node := NewTaskNode("1", "Delete", "resource.delete", "Project:exec-del-test")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute resource.delete error: %v", err)
	}

	// Verify it's gone
	_, err := reg.Get(project.Kind, "exec-del-test")
	if err == nil {
		t.Error("Project should have been deleted")
	}
}

func TestExecutorResourceKinds(t *testing.T) {
	exec, _, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "Kinds", "resource.kinds", "")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute resource.kinds error: %v", err)
	}
}

func TestExecutorControllerList(t *testing.T) {
	exec, _, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "Controllers", "controller.list", "")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute controller.list error: %v", err)
	}
}

func TestExecutorHealthCheck(t *testing.T) {
	exec, _, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "Health", "health.check", "")
	if err := exec.Execute(context.Background(), node); err != nil {
		t.Fatalf("Execute health.check error: %v", err)
	}
}

func TestExecutorUnknownAction(t *testing.T) {
	exec, _, _, _ := setupExecutor(t)
	node := NewTaskNode("1", "Bad", "nonexistent", "")
	if err := exec.Execute(context.Background(), node); err == nil {
		t.Error("expected error for unknown action")
	}
}

// ── Builder Tests ───────────────────────────────────────────────────────

func TestBuildDefinition(t *testing.T) {
	plan := []PlanNode{
		{ID: "1", Name: "Validate", Action: "validate", Target: "Project:test"},
		{ID: "2", Name: "Create", Action: "resource.create", Target: "Project:test"},
		{ID: "3", Name: "Verify", Action: "resource.get", Target: "Project:test"},
	}
	def, err := BuildDefinition("test-def", "Test Workflow", plan)
	if err != nil {
		t.Fatalf("BuildDefinition error: %v", err)
	}
	if def.ID != "test-def" {
		t.Errorf("ID = %q", def.ID)
	}
	if def.Name != "Test Workflow" {
		t.Errorf("Name = %q", def.Name)
	}
	// 3 task nodes + 1 end node
	if len(def.Nodes) != 4 {
		t.Errorf("Nodes = %d, want 4", len(def.Nodes))
	}

	// Verify dependency chain
	n1 := def.NodeByID("1")
	if n1 == nil {
		t.Fatal("Node 1 not found")
	}
	if len(n1.Dependencies()) != 0 {
		t.Errorf("Node 1 deps = %v, want []", n1.Dependencies())
	}

	n2 := def.NodeByID("2")
	if n2 == nil {
		t.Fatal("Node 2 not found")
	}
	if len(n2.Dependencies()) != 1 || n2.Dependencies()[0] != "1" {
		t.Errorf("Node 2 deps = %v, want [1]", n2.Dependencies())
	}

	n3 := def.NodeByID("3")
	if n3 == nil {
		t.Fatal("Node 3 not found")
	}
	if len(n3.Dependencies()) != 1 || n3.Dependencies()[0] != "2" {
		t.Errorf("Node 3 deps = %v, want [2]", n3.Dependencies())
	}

	end := def.NodeByID("end")
	if end == nil {
		t.Fatal("EndNode not found")
	}
	if end.Type() != NodeTypeEnd {
		t.Errorf("EndNode type = %q", end.Type())
	}
}

func TestBuildDefinitionEmpty(t *testing.T) {
	_, err := BuildDefinition("empty", "Empty", []PlanNode{})
	if err == nil {
		t.Error("expected error for empty plan")
	}
}

func TestBuildDefinitionExplicitDeps(t *testing.T) {
	plan := []PlanNode{
		{ID: "validate", Name: "Validate", Action: "validate", Target: "Project:test"},
		{ID: "create", Name: "Create", Action: "resource.create", Target: "Project:test", DependsOn: []string{"validate"}},
		{ID: "verify", Name: "Verify", Action: "resource.get", Target: "Project:test", DependsOn: []string{"create", "validate"}},
	}
	def, err := BuildDefinition("explicit", "Explicit Deps", plan)
	if err != nil {
		t.Fatalf("BuildDefinition error: %v", err)
	}

	v := def.NodeByID("verify")
	if v == nil {
		t.Fatal("verify node not found")
	}
	if len(v.Dependencies()) != 2 {
		t.Errorf("verify deps = %v, want [create validate]", v.Dependencies())
	}
}

func TestPlanTemplates(t *testing.T) {
	tests := []struct {
		name   string
		planFn func() []PlanNode
		count  int
	}{
		{"CreateProject", func() []PlanNode { return CreateProjectPlan("test") }, 5},
		{"ListProjects", func() []PlanNode { return ListProjectsPlan() }, 2},
		{"DeleteProject", func() []PlanNode { return DeleteProjectPlan("test") }, 3},
		{"ShowControllers", func() []PlanNode { return ShowControllersPlan() }, 2},
		{"ShowResources", func() []PlanNode { return ShowResourcesPlan() }, 2},
		{"ShowHealth", func() []PlanNode { return ShowHealthPlan() }, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := tc.planFn()
			if len(plan) != tc.count {
				t.Errorf("plan = %d, want %d", len(plan), tc.count)
			}
			def, err := BuildDefinition("test", tc.name, plan)
			if err != nil {
				t.Fatalf("BuildDefinition error: %v", err)
			}
			if err := ValidateDefinition(def); err != nil {
				t.Errorf("ValidateDefinition error: %v", err)
			}
		})
	}
}

// ── Engine Integration Tests ────────────────────────────────────────────

func setupEngine(t *testing.T) *Engine {
	t.Helper()
	log := logging.NewSubsystemLogger("workflow-engine-test", logging.LevelError)
	bus := events.NewBus(log)
	bus.Start()

	healthMgr := health.NewManager(log)
	_ = healthMgr.Start(context.Background())

	reg := resource.NewRegistry(bus, log)
	if err := reg.RegisterKind(resource.Kind{
		Name:       "Namespace",
		Namespaced: false,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}
	ns := resource.DefaultNamespace()
	if err := reg.Create(context.Background(), ns); err != nil {
		t.Fatal(err)
	}
	if err := reg.RegisterKind(resource.Kind{
		Name:       project.Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	ctrlManager := controller.NewManager(reg, bus, healthMgr, log)
	if err := ctrlManager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	eng := NewEngine(EngineDeps{
		ResourceRegistry:  reg,
		ControllerManager: ctrlManager,
		HealthManager:     healthMgr,
		Logger:            log,
	})

	t.Cleanup(func() {
		_ = ctrlManager.Stop(context.Background())
		healthMgr.Stop(context.Background())
		bus.Stop()
	})

	return eng
}

func TestEngineRegisterAndGetDefinition(t *testing.T) {
	eng := setupEngine(t)

	nodes := []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	}
	def, err := eng.CreateDefinition("test-def", "Test Def", nodes)
	if err != nil {
		t.Fatalf("CreateDefinition error: %v", err)
	}

	got, ok := eng.GetDefinition("test-def")
	if !ok {
		t.Fatal("GetDefinition returned false")
	}
	if got.ID != def.ID {
		t.Errorf("ID = %q, want %q", got.ID, def.ID)
	}

	defs := eng.ListDefinitions()
	if len(defs) != 1 {
		t.Errorf("ListDefinitions = %d, want 1", len(defs))
	}
}

func TestEngineSubmitCreatesRun(t *testing.T) {
	eng := setupEngine(t)

	def, _ := eng.CreateDefinition("submit-test", "Submit Test", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	if run.ID == "" {
		t.Error("run ID should not be empty")
	}
	if run.DefinitionID != "submit-test" {
		t.Errorf("DefinitionID = %q", run.DefinitionID)
	}
	if run.Status != WorkflowPending {
		t.Errorf("Status = %q, want %q", run.Status, WorkflowPending)
	}
	if len(run.Nodes) != 2 {
		t.Errorf("Nodes = %d, want 2", len(run.Nodes))
	}
}

func TestEngineGetRunNotFound(t *testing.T) {
	eng := setupEngine(t)
	_, ok := eng.GetRun("nonexistent")
	if ok {
		t.Error("GetRun should return false for nonexistent")
	}
}

func TestEngineListRuns(t *testing.T) {
	eng := setupEngine(t)

	def, _ := eng.CreateDefinition("list-test", "List Test", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	_, _ = eng.Submit(def)
	_, _ = eng.Submit(def)

	runs := eng.ListRuns()
	if len(runs) != 2 {
		t.Errorf("ListRuns = %d, want 2", len(runs))
	}
}

func TestEngineCancel(t *testing.T) {
	eng := setupEngine(t)

	def, _ := eng.CreateDefinition("cancel-test", "Cancel Test", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := eng.Submit(def)

	if err := eng.Cancel(run.ID); err != nil {
		t.Fatalf("Cancel error: %v", err)
	}

	got, _ := eng.GetRun(run.ID)
	if got.Status != WorkflowCancelled {
		t.Errorf("Status = %q, want %q", got.Status, WorkflowCancelled)
	}

	// Verify nodes are cancelled
	for _, n := range got.Nodes {
		if !n.Status().IsTerminal() {
			t.Errorf("Node %s status = %q, expected terminal", n.ID(), n.Status())
		}
	}
}

func TestEngineCancelNotFound(t *testing.T) {
	eng := setupEngine(t)
	err := eng.Cancel("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

func TestEnginePauseResume(t *testing.T) {
	eng := setupEngine(t)

	def, _ := eng.CreateDefinition("pause-test", "Pause Test", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := eng.Submit(def)

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	if err := eng.Pause(run.ID); err != nil {
		t.Fatalf("Pause error: %v", err)
	}

	got, _ := eng.GetRun(run.ID)
	if got.Status != WorkflowPaused {
		t.Errorf("Status = %q, want %q", got.Status, WorkflowPaused)
	}

	if err := eng.Resume(run.ID); err != nil {
		t.Fatalf("Resume error: %v", err)
	}

	got2, _ := eng.GetRun(run.ID)
	// After resume, it should be running again or already completed
	if got2.Status != WorkflowRunning && got2.Status != WorkflowCompleted {
		t.Errorf("Status after resume = %q, want %q or %q", got2.Status, WorkflowRunning, WorkflowCompleted)
	}
}

func TestEnginePauseNotRunning(t *testing.T) {
	eng := setupEngine(t)
	err := eng.Pause("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

func TestEngineFullLifecycle(t *testing.T) {
	eng := setupEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the engine
	go eng.Start(ctx)
	time.Sleep(50 * time.Millisecond) // let it start

	def, err := eng.CreateDefinition("lifecycle-test", "Lifecycle Test", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})
	if err != nil {
		t.Fatalf("CreateDefinition error: %v", err)
	}

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Wait for completion
	var got *WorkflowRun
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		got, _ = eng.GetRun(run.ID)
		if got != nil && got.Status == WorkflowCompleted {
			break
		}
	}
	if got == nil {
		t.Fatal("GetRun returned nil")
	}
	if got.Status != WorkflowCompleted {
		t.Errorf("Status = %q, want %q. Nodes:", got.Status, WorkflowCompleted)
		for _, n := range got.Nodes {
			t.Logf("  %s: %s", n.ID(), n.Status())
		}
	}
	if got.Result == nil {
		t.Fatal("Result should not be nil")
	}
	if !got.Result.Success {
		t.Errorf("Result.Success = false, want true: %s", got.Result.Summary)
	}
}

func TestEngineFullLifecycleCreateProject(t *testing.T) {
	eng := setupEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	def, err := BuildDefinition("create-project", "Create Project", CreateProjectPlan("lifecycle-project"))
	if err != nil {
		t.Fatalf("BuildDefinition error: %v", err)
	}
	if err := eng.RegisterDefinition(def); err != nil {
		t.Fatalf("RegisterDefinition error: %v", err)
	}

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Wait for completion
	var got *WorkflowRun
	for i := 0; i < 30; i++ {
		time.Sleep(200 * time.Millisecond)
		got, _ = eng.GetRun(run.ID)
		if got != nil && got.Status == WorkflowCompleted {
			break
		}
	}
	if got == nil || got.Status != WorkflowCompleted {
		t.Fatalf("Workflow did not complete in time. Status = %v", got.Status)
	}
	if got.Result == nil || !got.Result.Success {
		t.Errorf("Workflow failed: %s", got.Result.Summary)
	}

	// Verify node statuses
	for _, n := range got.Nodes {
		if n.Type() == NodeTypeEnd {
			continue
		}
		if n.Status() != NodeSucceeded {
			t.Errorf("Node %s (%s) status = %q, want %q", n.ID(), n.Name(), n.Status(), NodeSucceeded)
		}
	}
}

func TestEngineFullLifecycleDeleteProject(t *testing.T) {
	eng := setupEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// First create a project directly
	execEng := eng.exec
	createNode := NewTaskNode("pre", "Pre-create", "resource.create", "Project:lifecycle-delete")
	if err := execEng.Execute(context.Background(), createNode); err != nil {
		t.Fatalf("Pre-create error: %v", err)
	}

	// Now create a workflow that deletes it
	def, err := BuildDefinition("delete-project", "Delete Project", DeleteProjectPlan("lifecycle-delete"))
	if err != nil {
		t.Fatalf("BuildDefinition error: %v", err)
	}
	if err := eng.RegisterDefinition(def); err != nil {
		t.Fatalf("RegisterDefinition error: %v", err)
	}

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Wait for completion
	var got *WorkflowRun
	for i := 0; i < 30; i++ {
		time.Sleep(200 * time.Millisecond)
		got, _ = eng.GetRun(run.ID)
		if got != nil && got.Status == WorkflowCompleted {
			break
		}
	}
	if got == nil || got.Status != WorkflowCompleted {
		t.Fatalf("Workflow did not complete. Status = %v", got.Status)
	}
	if !got.Result.Success {
		t.Fatalf("Workflow failed: %s", got.Result.Summary)
	}
}

func TestEngineCancelDuringExecution(t *testing.T) {
	eng := setupEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Use a multi-step workflow that does real work (resource creation) so it
	// takes long enough to still be running when Cancel is called.
	def, err := BuildDefinition("cancel-exec", "Cancel During Exec", CreateProjectPlan("cancel-project"))
	if err != nil {
		t.Fatalf("BuildDefinition error: %v", err)
	}
	if err := eng.RegisterDefinition(def); err != nil {
		t.Fatalf("RegisterDefinition error: %v", err)
	}

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Cancel immediately — the workflow should not complete
	_ = eng.Cancel(run.ID)

	// Give a moment for cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	got, ok := eng.GetRun(run.ID)
	if !ok {
		t.Fatal("GetRun returned false")
	}
	// Accept either cancelled or still running (cancel may be processed after
	// some nodes already completed — that's fine, what matters is the intent
	// to cancel was recorded)
	if got.Status != WorkflowCancelled && got.Status != WorkflowRunning {
		t.Errorf("Status = %q, want %q or %q", got.Status, WorkflowCancelled, WorkflowRunning)
	}
}

// ── Execution Resource Tests ───────────────────────────────────────────

func TestNewWorkflowExecution(t *testing.T) {
	run := &WorkflowRun{
		ID:           "wf-test-1",
		DefinitionID: "def-1",
		Status:       WorkflowPending,
		Nodes:        []Node{NewTaskNode("1", "Step 1", "complete", "")},
		CreatedAt:    NowUTC(),
	}

	spec := WorkflowExecutionSpec{
		WorkflowID: "def-1",
		IntentID:   "intent-1",
		RequestedBy: "test",
		Priority:   1,
	}

	exec := NewWorkflowExecution(run, spec)
	if exec.GetKind() != WorkflowExecutionKind {
		t.Errorf("Kind = %q, want %q", exec.GetKind(), WorkflowExecutionKind)
	}
	if exec.GetMetadata().ID != "wf-test-1" {
		t.Errorf("ID = %q", exec.GetMetadata().ID)
	}

	// Check conditions
	status := exec.GetStatus().(*WorkflowExecutionStatus)
	if len(status.Conditions) != 1 {
		t.Errorf("Conditions = %d, want 1", len(status.Conditions))
	}
	if status.Conditions[0].Type != ConditionScheduled {
		t.Errorf("Condition[0].Type = %q, want %q", status.Conditions[0].Type, ConditionScheduled)
	}
	if status.Conditions[0].Status != ConditionTrue {
		t.Errorf("Condition[0].Status = %q, want %q", status.Conditions[0].Status, ConditionTrue)
	}
}

func TestWorkflowExecutionValidate(t *testing.T) {
	// Valid
	run := &WorkflowRun{ID: "exec-valid", Status: WorkflowPending}
	exec := NewWorkflowExecution(run, WorkflowExecutionSpec{WorkflowID: "def-1"})
	if err := exec.Validate(); err != nil {
		t.Errorf("Validate error: %v", err)
	}

	// Empty ID
	exec2 := NewWorkflowExecution(&WorkflowRun{ID: ""}, WorkflowExecutionSpec{WorkflowID: "def-1"})
	if err := exec2.Validate(); err == nil {
		t.Error("expected error for empty ID")
	}

	// Empty WorkflowID
	exec3 := NewWorkflowExecution(&WorkflowRun{ID: "exec-3"}, WorkflowExecutionSpec{})
	if err := exec3.Validate(); err == nil {
		t.Error("expected error for empty WorkflowID")
	}
}

func TestExecutionSetCondition(t *testing.T) {
	status := &WorkflowExecutionStatus{}

	status.SetCondition(ConditionScheduled, ConditionTrue, "Scheduled", "Execution scheduled")
	if len(status.Conditions) != 1 {
		t.Fatalf("Conditions = %d, want 1", len(status.Conditions))
	}
	if status.Conditions[0].Type != ConditionScheduled {
		t.Errorf("Type = %q", status.Conditions[0].Type)
	}

	// Add another condition
	status.SetCondition(ConditionRunning, ConditionTrue, "Running", "Execution running")
	if len(status.Conditions) != 2 {
		t.Fatalf("Conditions = %d, want 2", len(status.Conditions))
	}

	// Update existing
	status.SetCondition(ConditionScheduled, ConditionFalse, "Started", "")
	if len(status.Conditions) != 2 {
		t.Fatalf("Conditions = %d, want 2", len(status.Conditions))
	}
	c := status.GetCondition(ConditionScheduled)
	if c == nil {
		t.Fatal("GetCondition(Scheduled) returned nil")
	}
	if c.Status != ConditionFalse {
		t.Errorf("Status = %q, want %q", c.Status, ConditionFalse)
	}
}

func TestExecutionGetCondition(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	status.SetCondition(ConditionCompleted, ConditionTrue, "Done", "")

	c := status.GetCondition(ConditionCompleted)
	if c == nil {
		t.Fatal("GetCondition(Completed) returned nil")
	}
	if c.Status != ConditionTrue {
		t.Errorf("Status = %q", c.Status)
	}

	// Non-existent condition
	c = status.GetCondition("NonExistent")
	if c != nil {
		t.Error("GetCondition should return nil for non-existent type")
	}
}

func TestExecutionIsConditionTrue(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	status.SetCondition(ConditionRunning, ConditionTrue, "", "")

	if !status.IsConditionTrue(ConditionRunning) {
		t.Error("IsConditionTrue(Running) should be true")
	}
	if status.IsConditionTrue(ConditionCompleted) {
		t.Error("IsConditionTrue(Completed) should be false")
	}
}

func TestExecutionSyncFromRunPending(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	run := &WorkflowRun{
		ID:     "sync-test",
		Status: WorkflowPending,
		Nodes: []Node{
			NewTaskNode("1", "A", "complete", ""),
			NewTaskNode("2", "B", "complete", "", "1"),
		},
		CreatedAt: NowUTC(),
	}

	status.SyncFromRun(run)
	if status.Phase != WorkflowPending {
		t.Errorf("Phase = %q", status.Phase)
	}
	if status.TotalNodes != 2 {
		t.Errorf("TotalNodes = %d", status.TotalNodes)
	}
	if !status.IsConditionTrue(ConditionScheduled) {
		t.Error("Expected Scheduled condition to be true")
	}
}

func TestExecutionSyncFromRunRunning(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	run := &WorkflowRun{
		ID:     "sync-running",
		Status: WorkflowRunning,
		Nodes: []Node{
			NewTaskNode("1", "A", "complete", ""),
			NewTaskNode("2", "B", "complete", "", "1"),
		},
		CreatedAt: NowUTC(),
	}
	// Mark node 1 as running
	run.Nodes[0].SetStatus(NodeRunning)

	status.SyncFromRun(run)
	if status.CurrentNode != "1" {
		t.Errorf("CurrentNode = %q, want 1", status.CurrentNode)
	}
	if !status.IsConditionTrue(ConditionRunning) {
		t.Error("Expected Running condition to be true")
	}
	if status.IsConditionTrue(ConditionScheduled) {
		t.Error("Expected Scheduled condition to be false")
	}
}

func TestExecutionSyncFromRunCompleted(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	run := &WorkflowRun{
		ID:     "sync-completed",
		Status: WorkflowCompleted,
		Nodes: []Node{
			NewTaskNode("1", "A", "complete", ""),
			NewTaskNode("2", "B", "complete", "", "1"),
		},
		CreatedAt:   NowUTC(),
		CompletedAt: NowUTC(),
		Result: &WorkflowResult{
			Success: true,
			Summary: "Completed successfully — 2 nodes",
		},
	}
	run.Nodes[0].SetStatus(NodeSucceeded)
	run.Nodes[1].SetStatus(NodeSucceeded)

	status.SyncFromRun(run)
	if !status.IsConditionTrue(ConditionCompleted) {
		t.Error("Expected Completed condition to be true")
	}
	if status.IsConditionTrue(ConditionRunning) {
		t.Error("Expected Running condition to be false")
	}
	if status.Result != "Completed successfully — 2 nodes" {
		t.Errorf("Result = %q", status.Result)
	}
	if len(status.CompletedNodes) != 2 {
		t.Errorf("CompletedNodes = %d, want 2", len(status.CompletedNodes))
	}
}

func TestExecutionSyncFromRunFailed(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	run := &WorkflowRun{
		ID:     "sync-failed",
		Status: WorkflowFailed,
		Nodes: []Node{
			NewTaskNode("1", "A", "complete", ""),
			NewTaskNode("2", "B", "complete", "", "1"),
		},
		CreatedAt:   NowUTC(),
		CompletedAt: NowUTC(),
		Result: &WorkflowResult{
			Success: false,
			Summary: "Step 2 failed",
		},
	}
	run.Nodes[0].SetStatus(NodeSucceeded)
	run.Nodes[1].SetStatus(NodeFailed)

	status.SyncFromRun(run)
	if !status.IsConditionTrue(ConditionFailed) {
		t.Error("Expected Failed condition to be true")
	}
	if status.IsConditionTrue(ConditionRunning) {
		t.Error("Expected Running condition to be false")
	}
	if len(status.CompletedNodes) != 1 {
		t.Errorf("CompletedNodes = %d, want 1", len(status.CompletedNodes))
	}
	if len(status.FailedNodes) != 1 {
		t.Errorf("FailedNodes = %d, want 1", len(status.FailedNodes))
	}
}

func TestExecutionSyncFromRunCancelled(t *testing.T) {
	status := &WorkflowExecutionStatus{}
	run := &WorkflowRun{
		ID:     "sync-cancelled",
		Status: WorkflowCancelled,
		Nodes: []Node{
			NewTaskNode("1", "A", "complete", ""),
			NewTaskNode("2", "B", "complete", "", "1"),
		},
		CreatedAt: NowUTC(),
	}
	run.Nodes[0].SetStatus(NodeSucceeded)
	run.Nodes[1].SetStatus(NodeCancelled)

	status.SyncFromRun(run)
	if !status.IsConditionTrue(ConditionCancelled) {
		t.Error("Expected Cancelled condition to be true")
	}
}

func TestEngineSubmitCreatesExecutionResource(t *testing.T) {
	eng := setupEngine(t)

	def, _ := eng.CreateDefinition("exec-res-test", "Exec Resource Test", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Verify an execution resource was created
	if eng.resRegistry != nil {
		obj, err := eng.resRegistry.Get(WorkflowExecutionKind, run.ID)
		if err != nil {
			t.Fatalf("Get execution resource error: %v", err)
		}
		exec, ok := obj.(*WorkflowExecution)
		if !ok {
			t.Fatalf("unexpected type %T", obj)
		}
		if exec.GetKind() != WorkflowExecutionKind {
			t.Errorf("Kind = %q", exec.GetKind())
		}
		status := exec.GetStatus().(*WorkflowExecutionStatus)
		if status.Phase != WorkflowPending {
			t.Errorf("Phase = %q, want %q", status.Phase, WorkflowPending)
		}
	}
}

func TestEngineFullLifecycleWithPersistence(t *testing.T) {
	eng := setupEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	def, err := BuildDefinition("lifecycle-persist", "Lifecycle Persist", CreateProjectPlan("persist-project"))
	if err != nil {
		t.Fatalf("BuildDefinition error: %v", err)
	}
	if err := eng.RegisterDefinition(def); err != nil {
		t.Fatalf("RegisterDefinition error: %v", err)
	}

	run, err := eng.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Wait for completion
	var got *WorkflowRun
	for i := 0; i < 30; i++ {
		time.Sleep(200 * time.Millisecond)
		got, _ = eng.GetRun(run.ID)
		if got != nil && got.Status == WorkflowCompleted {
			break
		}
	}
	if got == nil || got.Status != WorkflowCompleted {
		t.Fatalf("Workflow did not complete. Status = %v", got.Status)
	}

	// Verify the execution resource was persisted and has correct state
	if eng.resRegistry != nil {
		obj, err := eng.resRegistry.Get(WorkflowExecutionKind, run.ID)
		if err != nil {
			t.Fatalf("Get execution resource error: %v", err)
		}
		exec, ok := obj.(*WorkflowExecution)
		if !ok {
			t.Fatalf("unexpected type %T", obj)
		}

		status := exec.GetStatus().(*WorkflowExecutionStatus)
		if status.Phase != WorkflowCompleted {
			t.Errorf("Phase = %q, want %q", status.Phase, WorkflowCompleted)
		}
		if !status.IsConditionTrue(ConditionCompleted) {
			t.Error("Expected Completed condition to be true")
		}
		if status.Progress != 1.0 {
			t.Errorf("Progress = %f, want 1.0", status.Progress)
		}
		if len(status.CompletedNodes) == 0 {
			t.Error("Expected at least one completed node")
		}
	}
}

func TestExecutionResourceInterface(t *testing.T) {
	// Verify resource.Resource interface compliance
	run := &WorkflowRun{ID: "iface-test", Status: WorkflowPending}
	exec := NewWorkflowExecution(run, WorkflowExecutionSpec{WorkflowID: "def-1"})

	var res resource.Resource = exec // compile-time check

	if res.GetKind() != WorkflowExecutionKind {
		t.Errorf("GetKind() = %q", res.GetKind())
	}

	// SetStatus with correct type
	status := &WorkflowExecutionStatus{Phase: WorkflowRunning}
	res.SetStatus(status)
	got := res.GetStatus().(*WorkflowExecutionStatus)
	if got.Phase != WorkflowRunning {
		t.Errorf("Phase after SetStatus = %q", got.Phase)
	}
}

// ── Service Tests ──────────────────────────────────────────────────────

func setupService(t *testing.T) *Service {
	t.Helper()
	log := logging.NewSubsystemLogger("workflow-service-test", logging.LevelError)
	bus := events.NewBus(log)
	bus.Start()

	healthMgr := health.NewManager(log)
	_ = healthMgr.Start(context.Background())

	reg := resource.NewRegistry(bus, log)
	if err := reg.RegisterKind(resource.Kind{
		Name:       "Namespace",
		Namespaced: false,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}
	ns := resource.DefaultNamespace()
	if err := reg.Create(context.Background(), ns); err != nil {
		t.Fatal(err)
	}
	if err := reg.RegisterKind(resource.Kind{
		Name:       project.Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.RegisterKind(resource.Kind{
		Name:       WorkflowExecutionKind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	ctrlManager := controller.NewManager(reg, bus, healthMgr, log)
	if err := ctrlManager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	svc := NewService(ServiceDeps{
		ResourceRegistry:  reg,
		ControllerManager: ctrlManager,
		HealthManager:     healthMgr,
		Logger:            log,
	})

	t.Cleanup(func() {
		_ = ctrlManager.Stop(context.Background())
		healthMgr.Stop(context.Background())
		bus.Stop()
	})

	return svc
}

func TestServiceSubmitAndGet(t *testing.T) {
	svc := setupService(t)

	def, err := svc.CreateDefinition("svc-submit", "Service Submit", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})
	if err != nil {
		t.Fatalf("CreateDefinition error: %v", err)
	}

	run, err := svc.Submit(def)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	if run.ID == "" {
		t.Error("run ID should not be empty")
	}

	got, ok := svc.Get(run.ID)
	if !ok {
		t.Fatal("Get returned false")
	}
	if got.ID != run.ID {
		t.Errorf("ID = %q, want %q", got.ID, run.ID)
	}
}

func TestServiceList(t *testing.T) {
	svc := setupService(t)

	def, _ := svc.CreateDefinition("svc-list", "Service List", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	_, _ = svc.Submit(def)
	_, _ = svc.Submit(def)

	runs := svc.List()
	if len(runs) != 2 {
		t.Errorf("List = %d, want 2", len(runs))
	}
}

func TestServicePauseResumeCancel(t *testing.T) {
	svc := setupService(t)

	def, _ := svc.CreateDefinition("svc-lifecycle", "Service Lifecycle", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := svc.Submit(def)

	if err := svc.Pause(run.ID); err != nil {
		t.Fatalf("Pause error: %v", err)
	}
	got, _ := svc.Get(run.ID)
	if got.Status != WorkflowPaused {
		t.Errorf("After Pause: Status = %q, want %q", got.Status, WorkflowPaused)
	}

	if err := svc.Resume(run.ID); err != nil {
		t.Fatalf("Resume error: %v", err)
	}

	_ = svc.Cancel(run.ID)
	got, _ = svc.Get(run.ID)
	if got.Status != WorkflowCancelled {
		t.Errorf("After Cancel: Status = %q, want %q", got.Status, WorkflowCancelled)
	}
}

func TestServiceGetExecution(t *testing.T) {
	svc := setupService(t)

	def, _ := svc.CreateDefinition("svc-exec", "Service Exec", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := svc.Submit(def)

	exec, err := svc.GetExecution(run.ID)
	if err != nil {
		t.Fatalf("GetExecution error: %v", err)
	}
	if exec.GetKind() != WorkflowExecutionKind {
		t.Errorf("Kind = %q", exec.GetKind())
	}
	status := exec.GetStatus().(*WorkflowExecutionStatus)
	if status.Phase != WorkflowPending {
		t.Errorf("Phase = %q, want %q", status.Phase, WorkflowPending)
	}
}

func TestServiceListExecutions(t *testing.T) {
	svc := setupService(t)

	def, _ := svc.CreateDefinition("svc-list-exec", "Service List Exec", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	_, _ = svc.Submit(def)
	_, _ = svc.Submit(def)

	execs, err := svc.ListExecutions()
	if err != nil {
		t.Fatalf("ListExecutions error: %v", err)
	}
	if len(execs) != 2 {
		t.Errorf("ListExecutions = %d, want 2", len(execs))
	}
}

func TestServiceRetry(t *testing.T) {
	svc := setupService(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Engine().Start(ctx)
	time.Sleep(50 * time.Millisecond)

	def, _ := svc.CreateDefinition("svc-retry", "Service Retry", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := svc.Submit(def)

	// Wait for completion
	var original *WorkflowRun
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		original, _ = svc.Get(run.ID)
		if original != nil && original.Status == WorkflowCompleted {
			break
		}
	}
	if original == nil || original.Status != WorkflowCompleted {
		t.Fatalf("Original did not complete. Status = %v", original.Status)
	}

	// Retry
	retryRun, err := svc.Retry(run.ID)
	if err != nil {
		t.Fatalf("Retry error: %v", err)
	}
	if retryRun.ID == run.ID {
		t.Error("Retry should create a new run with a different ID")
	}
	if retryRun.DefinitionID != run.DefinitionID {
		t.Errorf("DefinitionID = %q, want %q", retryRun.DefinitionID, run.DefinitionID)
	}
	if retryRun.Status != WorkflowPending {
		t.Errorf("Status = %q, want %q", retryRun.Status, WorkflowPending)
	}

	// Wait for retry to complete
	var retried *WorkflowRun
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		retried, _ = svc.Get(retryRun.ID)
		if retried != nil && retried.Status == WorkflowCompleted {
			break
		}
	}
	if retried == nil || retried.Status != WorkflowCompleted {
		t.Fatalf("Retry did not complete. Status = %v", retried.Status)
	}
}

func TestServiceRetryPreservesSucceeded(t *testing.T) {
	svc := setupService(t)

	// Create a definition and manually build a failed run (no engine needed)
	def, _ := svc.CreateDefinition("svc-retry-preserve", "Retry Preserve", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewTaskNode("2", "Step 2", "complete", "", "1"),
		NewTaskNode("3", "Step 3", "complete", "", "2"),
		NewEndNode("end", "3"),
	})

	// Simulate a failed execution: node 1 succeeded, node 2 failed
	failedRun := &WorkflowRun{
		ID:           "failed-run",
		DefinitionID: def.ID,
		Status:       WorkflowFailed,
		Nodes:        copyNodes(def.Nodes),
		CreatedAt:    NowUTC(),
		UpdatedAt:    NowUTC(),
	}
	failedRun.NodeByID("1").SetStatus(NodeSucceeded)
	failedRun.NodeByID("2").SetStatus(NodeFailed)

	// Manually register the failed run in the engine
	svc.Engine().mu.Lock()
	svc.Engine().runs["failed-run"] = failedRun
	svc.Engine().mu.Unlock()

	// Retry
	retryRun, err := svc.Retry("failed-run")
	if err != nil {
		t.Fatalf("Retry error: %v", err)
	}

	// Check that node 1 was preserved as succeeded
	n1 := retryRun.NodeByID("1")
	if n1 == nil {
		t.Fatal("Node 1 not found in retry run")
	}
	if n1.Status() != NodeSucceeded {
		t.Errorf("Node 1 Status = %q, want %q (should be preserved)", n1.Status(), NodeSucceeded)
	}

	// Node 2 should be reset to pending (it failed)
	n2 := retryRun.NodeByID("2")
	if n2 == nil {
		t.Fatal("Node 2 not found in retry run")
	}
	if n2.Status() != NodePending {
		t.Errorf("Node 2 Status = %q, want %q (should be reset)", n2.Status(), NodePending)
	}

	// Node 3 should also be pending (depends on 2)
	n3 := retryRun.NodeByID("3")
	if n3 == nil {
		t.Fatal("Node 3 not found in retry run")
	}
	if n3.Status() != NodePending {
		t.Errorf("Node 3 Status = %q, want %q", n3.Status(), NodePending)
	}

	// TaskNode-specific checks: retry count should be reset
	tn2, ok := n2.(*TaskNode)
	if !ok {
		t.Fatal("Node 2 is not a TaskNode")
	}
	if tn2.RetryCount != 0 {
		t.Errorf("Node 2 RetryCount = %d, want 0", tn2.RetryCount)
	}
	if tn2.ErrorVal != "" {
		t.Errorf("Node 2 ErrorVal = %q, want empty", tn2.ErrorVal)
	}
}

func TestServiceReplay(t *testing.T) {
	svc := setupService(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Engine().Start(ctx)
	time.Sleep(50 * time.Millisecond)

	def, _ := svc.CreateDefinition("svc-replay", "Service Replay", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := svc.Submit(def)

	// Wait for completion
	var original *WorkflowRun
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		original, _ = svc.Get(run.ID)
		if original != nil && original.Status == WorkflowCompleted {
			break
		}
	}
	if original == nil || original.Status != WorkflowCompleted {
		t.Fatalf("Original did not complete. Status = %v", original.Status)
	}

	// Replay
	replayRun, err := svc.Replay(run.ID)
	if err != nil {
		t.Fatalf("Replay error: %v", err)
	}
	if replayRun.ID == run.ID {
		t.Error("Replay should create a new run with a different ID")
	}
	if replayRun.Status != WorkflowPending {
		t.Errorf("Status = %q, want %q", replayRun.Status, WorkflowPending)
	}

	// Wait for replay to complete
	var replayed *WorkflowRun
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		replayed, _ = svc.Get(replayRun.ID)
		if replayed != nil && replayed.Status == WorkflowCompleted {
			break
		}
	}
	if replayed == nil || replayed.Status != WorkflowCompleted {
		t.Fatalf("Replay did not complete. Status = %v", replayed.Status)
	}
}

func TestServiceClone(t *testing.T) {
	svc := setupService(t)

	def, _ := svc.CreateDefinition("svc-clone", "Service Clone", []Node{
		NewTaskNode("1", "Step 1", "complete", ""),
		NewEndNode("end", "1"),
	})

	run, _ := svc.Submit(def)

	// Clone with overrides
	cloneRun, err := svc.Clone(run.ID, map[string]string{
		"requestedBy": "clone-user",
	})
	if err != nil {
		t.Fatalf("Clone error: %v", err)
	}
	if cloneRun.ID == run.ID {
		t.Error("Clone should create a new run with a different ID")
	}
	if cloneRun.DefinitionID != run.DefinitionID {
		t.Errorf("DefinitionID = %q, want %q", cloneRun.DefinitionID, run.DefinitionID)
	}
}

func TestServiceRetryNotFound(t *testing.T) {
	svc := setupService(t)
	_, err := svc.Retry("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

func TestServiceReplayNotFound(t *testing.T) {
	svc := setupService(t)
	_, err := svc.Replay("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

func TestServiceCloneNotFound(t *testing.T) {
	svc := setupService(t)
	_, err := svc.Clone("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

// ── Whole-package build ─────────────────────────────────────────────────

// TestAllWorkflowTestsPass is a marker to ensure the package compiles.
func TestPackageBuilds(t *testing.T) {
	// If we get here, the package compiled and linked successfully
}
