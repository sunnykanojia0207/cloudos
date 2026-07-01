package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/resource"
	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/kernel/safe"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/packages/logging"
)

// Engine is the top-level coordinator for the Workflow system.
//
// Responsibilities:
//   - Register WorkflowDefinition blueprints
//   - Submit a definition → creates WorkflowRun + WorkflowExecution Resource, enqueues
//   - Run the scheduler loop (dequeue → schedule → execute → persist → repeat)
//   - Pause, resume, cancel runs
//   - Publish lifecycle events
//   - Persist execution state as CloudOS Resources
//
// Architecture:
//
//	Submit → Queue → Scheduler → Executor → Persist → (repeat) → Complete
//	                  ↑                                        ↓
//	               WorkflowRun                          WorkflowExecution Resource
type Engine struct {
	mu          sync.RWMutex
	defs        map[string]*WorkflowDefinition
	runs        map[string]*WorkflowRun
	cancels     map[string]context.CancelFunc // per-run cancellation
	queue       *Queue
	scheduler   *Scheduler
	exec        *Executor
	events      *EventPublisher
	retry       *RetryEvaluator
	resRegistry *resource.Registry
	log         *logging.Logger
	counter     int
	kindEnsured bool // whether WorkflowExecution kind has been registered
}

// EngineDeps holds the dependencies for creating a new Engine.
type EngineDeps struct {
	ResourceRegistry  *resource.Registry
	ControllerManager *controller.Manager
	HealthManager     *health.Manager
	EventBus          interface{} // *events.Bus, passed as interface for nil-safety
	SourceCloner      *source.GitCloner
	RuntimeManager    cr.Runtime
	Logger            *logging.Logger
}

// NewEngine creates a new Workflow Engine.
func NewEngine(deps EngineDeps) *Engine {
	execDeps := ExecutorDeps{
		ResourceRegistry:  deps.ResourceRegistry,
		ControllerManager: deps.ControllerManager,
		HealthManager:     deps.HealthManager,
		SourceCloner:      deps.SourceCloner,
		RuntimeManager:    deps.RuntimeManager,
		Logger:            deps.Logger,
	}

	eng := &Engine{
		defs:        make(map[string]*WorkflowDefinition),
		runs:        make(map[string]*WorkflowRun),
		cancels:     make(map[string]context.CancelFunc),
		queue:       NewQueue(0), // unlimited capacity
		scheduler:   NewScheduler(),
		exec:        NewExecutor(execDeps),
		events:      &EventPublisher{}, // will be set if bus available
		retry:       NewRetryEvaluator(),
		resRegistry: deps.ResourceRegistry,
		log:         deps.Logger,
	}
	// Auto-start the queue so Submit works without calling Start() first.
	eng.queue.Start()
	return eng
}

// SetEventBus configures the event bus for publishing workflow events.
// Called after construction since the bus may not be available at init time.
func (eng *Engine) SetEventBus(bus interface {
	Publish(ctx context.Context, event interface{})
}) {
	if bus != nil {
		eng.events = &EventPublisher{}
	}
}

// ── Definition Management ───────────────────────────────────────────────

// RegisterDefinition stores a workflow definition for later execution.
func (eng *Engine) RegisterDefinition(def *WorkflowDefinition) error {
	if err := ValidateDefinition(def); err != nil {
		return fmt.Errorf("register definition: %w", err)
	}

	eng.mu.Lock()
	defer eng.mu.Unlock()

	if _, exists := eng.defs[def.ID]; exists {
		return fmt.Errorf("definition %q already registered", def.ID)
	}
	eng.defs[def.ID] = def
	eng.log.Debug("workflow definition registered", "id", def.ID, "name", def.Name)
	return nil
}

// GetDefinition returns a workflow definition by ID.
func (eng *Engine) GetDefinition(id string) (*WorkflowDefinition, bool) {
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	def, ok := eng.defs[id]
	return def, ok
}

// ListDefinitions returns all registered definitions.
func (eng *Engine) ListDefinitions() []*WorkflowDefinition {
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	result := make([]*WorkflowDefinition, 0, len(eng.defs))
	for _, def := range eng.defs {
		result = append(result, def)
	}
	return result
}

// ── Run Management ──────────────────────────────────────────────────────

// Submit creates a WorkflowRun from a Definition and enqueues it for execution.
// Returns the created run.
func (eng *Engine) Submit(def *WorkflowDefinition) (*WorkflowRun, error) {
	runID := eng.nextID()

	// Deep-copy nodes for the run
	nodes := copyNodes(def.Nodes)

	run := &WorkflowRun{
		ID:           runID,
		DefinitionID: def.ID,
		Status:       WorkflowPending,
		Nodes:        nodes,
		CreatedAt:    NowUTC(),
		UpdatedAt:    NowUTC(),
	}

	eng.mu.Lock()
	eng.runs[runID] = run
	eng.mu.Unlock()

	eng.log.Info("workflow submitted", "run_id", runID, "definition", def.ID, "nodes", len(nodes))
	eng.events.PublishWorkflowEvent(EventWorkflowSubmitted, run)

	// Create the WorkflowExecution Resource for persistence
	execSpec := WorkflowExecutionSpec{
		WorkflowID: def.ID,
	}
	if err := eng.createExecutionResource(run, execSpec); err != nil {
		eng.log.Warn("failed to persist execution resource", "run_id", runID, "error", err.Error())
		// Non-fatal — execution continues even if persistence fails
	}

	// Enqueue for execution
	ctx := context.Background()
	if err := eng.queue.Enqueue(ctx, QueueItem{
		WorkflowID: runID,
		NodeID:     "", // signal to start scheduling
	}); err != nil {
		return nil, fmt.Errorf("enqueue workflow: %w", err)
	}

	return run, nil
}

// GetRun returns a workflow run by ID.
func (eng *Engine) GetRun(id string) (*WorkflowRun, bool) {
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	run, ok := eng.runs[id]
	return run, ok
}

// ListRuns returns all workflow runs.
func (eng *Engine) ListRuns() []*WorkflowRun {
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	result := make([]*WorkflowRun, 0, len(eng.runs))
	for _, run := range eng.runs {
		result = append(result, run)
	}
	return result
}

// Cancel marks a workflow run as cancelled and cancels any running nodes.
func (eng *Engine) Cancel(runID string) error {
	eng.mu.Lock()
	run, ok := eng.runs[runID]
	if !ok {
		eng.mu.Unlock()
		return fmt.Errorf("run %q not found", runID)
	}

	if run.Status == WorkflowCompleted || run.Status == WorkflowFailed || run.Status == WorkflowCancelled {
		eng.mu.Unlock()
		return fmt.Errorf("run %q already in terminal state %q", runID, run.Status)
	}

	run.Status = WorkflowCancelled
	run.UpdatedAt = NowUTC()

	// Cancel any running node via its context
	if cancel, exists := eng.cancels[runID]; exists {
		cancel()
		delete(eng.cancels, runID)
	}

	// Mark all pending/running nodes as cancelled
	for _, n := range run.Nodes {
		if !n.Status().IsTerminal() {
			n.SetStatus(NodeCancelled)
		}
	}

	// Persist cancelled state
	if err := eng.updateExecutionResource(run); err != nil {
		eng.log.Warn("failed to persist cancelled state", "run_id", runID, "error", err.Error())
	}

	eng.mu.Unlock()

	eng.events.PublishWorkflowEvent(EventWorkflowCancelled, run)
	eng.log.Info("workflow cancelled", "run_id", runID)

	// Cleanup: stop any running processes started by this run.
	eng.exec.CleanupRun(context.Background(), runID)

	return nil
}

// Pause marks a workflow as paused. Currently running nodes complete;
// pending nodes will not start until Resume is called.
func (eng *Engine) Pause(runID string) error {
	eng.mu.Lock()
	defer eng.mu.Unlock()

	run, ok := eng.runs[runID]
	if !ok {
		return fmt.Errorf("run %q not found", runID)
	}
	if run.Status != WorkflowRunning && run.Status != WorkflowPending {
		return fmt.Errorf("run %q cannot be paused (status: %q)", runID, run.Status)
	}

	run.Status = WorkflowPaused
	run.UpdatedAt = NowUTC()

	eng.events.PublishWorkflowEvent(EventWorkflowPaused, run)
	return nil
}

// Resume resumes a paused workflow.
func (eng *Engine) Resume(runID string) error {
	eng.mu.Lock()
	run, ok := eng.runs[runID]
	if !ok {
		eng.mu.Unlock()
		return fmt.Errorf("run %q not found", runID)
	}
	if run.Status != WorkflowPaused {
		eng.mu.Unlock()
		return fmt.Errorf("run %q is not paused (status: %q)", runID, run.Status)
	}

	run.Status = WorkflowRunning
	run.UpdatedAt = NowUTC()
	eng.mu.Unlock()

	eng.events.PublishWorkflowEvent(EventWorkflowResumed, run)

	// Re-enqueue to resume scheduling
	return eng.queue.Enqueue(context.Background(), QueueItem{
		WorkflowID: runID,
		NodeID:     "",
	})
}

// ── Execution Resource Persistence ───────────────────────────────────────

// ensureKind registers the WorkflowExecution kind with the Resource Engine
// if it hasn't been registered yet. Safe to call multiple times.
func (eng *Engine) ensureKind() error {
	if eng.kindEnsured || eng.resRegistry == nil {
		return nil
	}
	// Check if already registered
	if _, exists := eng.resRegistry.GetKind(WorkflowExecutionKind); exists {
		eng.kindEnsured = true
		return nil
	}
	if err := eng.resRegistry.RegisterKind(resource.Kind{
		Name:       WorkflowExecutionKind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		return fmt.Errorf("register workflow execution kind: %w", err)
	}
	eng.kindEnsured = true
	eng.log.Info("workflow execution resource kind registered")
	return nil
}

// createExecutionResource creates a WorkflowExecution Resource from a run.
func (eng *Engine) createExecutionResource(run *WorkflowRun, spec WorkflowExecutionSpec) error {
	if eng.resRegistry == nil {
		return nil // no registry available — skip persistence
	}
	if err := eng.ensureKind(); err != nil {
		return err
	}

	exec := NewWorkflowExecution(run, spec)
	if err := eng.resRegistry.Create(context.Background(), exec); err != nil {
		return fmt.Errorf("create execution resource: %w", err)
	}
	eng.log.Debug("execution resource created", "run_id", run.ID)
	return nil
}

// updateExecutionResource syncs the WorkflowExecution Resource status from a run.
func (eng *Engine) updateExecutionResource(run *WorkflowRun) error {
	if eng.resRegistry == nil {
		return nil
	}

	// Get the existing resource
	obj, err := eng.resRegistry.Get(WorkflowExecutionKind, run.ID)
	if err != nil {
		return fmt.Errorf("get execution resource: %w", err)
	}

	exec, ok := obj.(*WorkflowExecution)
	if !ok {
		return fmt.Errorf("execution resource has unexpected type %T", obj)
	}

	// Sync status from run
	exec.status.SyncFromRun(run)
	exec.metadata.UpdatedAt = time.Now()

	// Update labels
	if exec.metadata.Labels == nil {
		exec.metadata.Labels = make(map[string]string)
	}
	exec.metadata.Labels["workflow.cloudos.io/status"] = string(run.Status)

	if err := eng.resRegistry.Update(context.Background(), exec); err != nil {
		return fmt.Errorf("update execution resource: %w", err)
	}
	return nil
}

// ── Scheduler Loop ──────────────────────────────────────────────────────

// Start begins the scheduler loop. It runs until ctx is cancelled.
// Typically called in a goroutine.
func (eng *Engine) Start(ctx context.Context) {
	eng.log.Info("workflow engine starting")
	eng.queue.Start()

	for {
		item, done, err := eng.queue.Dequeue(ctx)
		if err != nil {
			eng.log.Info("workflow engine stopped")
			return
		}

		err = eng.processItem(ctx, item)
		done()

		if err != nil {
			eng.log.Warn("workflow processing error", "workflow_id", item.WorkflowID, "error", err.Error())
		}
	}
}

func (eng *Engine) processItem(ctx context.Context, item QueueItem) error {
	eng.mu.Lock()
	run, ok := eng.runs[item.WorkflowID]
	if !ok {
		eng.mu.Unlock()
		return fmt.Errorf("run %q not found", item.WorkflowID)
	}

	if run.Status == WorkflowPaused || run.Status == WorkflowCancelled {
		eng.mu.Unlock()
		return nil // paused/cancelled runs are skipped
	}

	// Mark as running on first processing
	if run.Status == WorkflowPending {
		run.Status = WorkflowRunning
		eng.events.PublishWorkflowEvent(EventWorkflowStarted, run)
	}

	// Check if complete
	if eng.scheduler.IsComplete(run) {
		hasFailures := eng.scheduler.HasFailures(run)
		if hasFailures {
			run.Status = WorkflowFailed
		} else {
			run.Status = WorkflowCompleted
		}
		run.CompletedAt = NowUTC()
		run.UpdatedAt = run.CompletedAt

		result := &WorkflowResult{
			Success: !hasFailures,
		}
		if hasFailures {
			result.Summary = fmt.Sprintf("Completed with failures (%d/%d nodes succeeded)", run.CompletedCount(), len(run.Nodes))
		} else {
			result.Summary = fmt.Sprintf("Completed successfully — %d nodes", len(run.Nodes))
		}

		// Collect details from nodes
		for _, n := range run.Nodes {
			if tn, ok := n.(*TaskNode); ok {
				result.Details = append(result.Details, FormatNodeResult(tn))
			}
		}
		run.Result = result

		eventType := EventWorkflowCompleted
		if hasFailures {
			eventType = EventWorkflowFailed
		}
		eng.events.PublishWorkflowEvent(eventType, run)

		// Persist final state
		if err := eng.updateExecutionResource(run); err != nil {
			eng.log.Warn("failed to persist final execution state", "run_id", run.ID, "error", err.Error())
		}

		eng.mu.Unlock()

		if cancel, exists := eng.cancels[run.ID]; exists {
			cancel()
			delete(eng.cancels, run.ID)
		}

		if hasFailures {
			// Cleanup runtime resources only on failure (stop processes, release ports).
			// On success, the deployed process keeps running.
			eng.exec.CleanupRun(context.Background(), run.ID)
		}

		return nil
	}

	// Persist intermediate state before releasing the lock
	if err := eng.updateExecutionResource(run); err != nil {
		eng.log.Warn("failed to persist execution state", "run_id", run.ID, "error", err.Error())
	}

	// Get ready nodes
	ready := eng.scheduler.Ready(run)
	eng.mu.Unlock()

	// Execute each ready node
	for _, node := range ready {
		if err := eng.executeNode(ctx, run, node); err != nil {
			eng.log.Warn("node execution error", "node", node.ID(), "error", err.Error())
		}
	}

	// Re-enqueue to process next batch
	return eng.queue.Enqueue(context.Background(), QueueItem{
		WorkflowID: run.ID,
		NodeID:     "",
	})
}

func (eng *Engine) executeNode(ctx context.Context, run *WorkflowRun, node Node) error {
	taskNode, ok := node.(*TaskNode)
	if !ok {
		// Non-task nodes (e.g., EndNode) succeed immediately
		node.SetStatus(NodeSucceeded)
		return nil
	}

	// Mark as running
	node.SetStatus(NodeRunning)
	eng.events.PublishNodeEvent(EventNodeStarted, run, node)

	// Create cancellable context for this node, with the run ID embedded
	// for cross-node state sharing (e.g., work directory from source.clone).
	runCtx := WithRunID(ctx, run.ID)
	nodeCtx, cancel := context.WithCancel(runCtx)
	eng.mu.Lock()
	eng.cancels[run.ID] = cancel
	eng.mu.Unlock()

	err := eng.exec.Execute(nodeCtx, taskNode)

	if err != nil {
		taskNode.ErrorVal = err.Error()

		// Check retry policy
		if delay, retryErr := eng.retry.ShouldRetry(taskNode); retryErr == nil {
			taskNode.RetryCount++
			taskNode.SetStatus(NodePending) // reset for retry
			eng.events.PublishNodeEvent(EventNodeRetrying, run, node)
			eng.log.Debug("retrying node", "node", node.ID(), "attempt", taskNode.RetryCount, "delay", delay)

			// Re-enqueue after backoff (with panic recovery).
			safe.Go(func() {
				time.Sleep(delay)
				_ = eng.queue.Enqueue(context.Background(), QueueItem{
					WorkflowID: run.ID,
					NodeID:     node.ID(),
				})
			})
			return nil
		}

		node.SetStatus(NodeFailed)
		eng.events.PublishNodeEvent(EventNodeFailed, run, node)
		return err
	}

	node.SetStatus(NodeSucceeded)
	eng.events.PublishNodeEvent(EventNodeSucceeded, run, node)

	// Release cancel func
	eng.mu.Lock()
	delete(eng.cancels, run.ID)
	eng.mu.Unlock()
	cancel()

	return nil
}

// ── Builders ────────────────────────────────────────────────────────────

// CreateDefinition creates a WorkflowDefinition from nodes and registers it.
func (eng *Engine) CreateDefinition(id, name string, nodes []Node) (*WorkflowDefinition, error) {
	def := &WorkflowDefinition{
		ID:        id,
		Name:      name,
		Nodes:     nodes,
		CreatedAt: NowUTC(),
	}
	if err := eng.RegisterDefinition(def); err != nil {
		return nil, err
	}
	return def, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────

func (eng *Engine) nextID() string {
	eng.mu.Lock()
	defer eng.mu.Unlock()
	eng.counter++
	return fmt.Sprintf("wf_%d", eng.counter)
}

func copyNodes(original []Node) []Node {
	result := make([]Node, len(original))
	for i, n := range original {
		switch src := n.(type) {
		case *TaskNode:
			deps := make([]string, len(src.deps))
			copy(deps, src.deps)
			result[i] = &TaskNode{
				id:          src.id,
				name:        src.name,
				status:      NodePending,
				deps:        deps,
				Action:      src.Action,
				Target:      src.Target,
				RetryPolicy: src.RetryPolicy,
				TimeoutVal:  src.TimeoutVal,
			}
		case *EndNode:
			deps := make([]string, len(src.deps))
			copy(deps, src.deps)
			result[i] = &EndNode{
				id:     src.id,
				name:   src.name,
				status: NodePending,
				deps:   deps,
			}
		default:
			// Fallback — shallow copy
			result[i] = n
		}
	}
	return result
}
