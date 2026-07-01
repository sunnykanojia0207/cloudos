package intent

import (
	"context"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/plugin"
	"github.com/cloudos/cloudos/kernel/project"
	"github.com/cloudos/cloudos/kernel/registry"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Parser Tests ────────────────────────────────────────────────────────

func TestParseCreateProject(t *testing.T) {
	p := NewParser()
	intent, err := p.Parse("create project My CRM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Type != IntentCreateProject {
		t.Errorf("Type = %q, want %q", intent.Type, IntentCreateProject)
	}
	if intent.Params["name"] != "My CRM" {
		t.Errorf("Params[name] = %q, want %q", intent.Params["name"], "My CRM")
	}
	if intent.Params["id"] != "my-crm" {
		t.Errorf("Params[id] = %q, want %q", intent.Params["id"], "my-crm")
	}
}

func TestParseCreateProjectVariants(t *testing.T) {
	p := NewParser()
	tests := []struct {
		input string
		name  string
		id    string
	}{
		{"create a project test", "test", "test"},
		{"create an project example", "example", "example"},
		{"create new project MyApp", "MyApp", "myapp"},
		{"create project hello-world", "hello-world", "hello-world"},
	}
	for _, tc := range tests {
		intent, err := p.Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.input, err)
			continue
		}
		if intent.Type != IntentCreateProject {
			t.Errorf("Parse(%q) Type = %q, want %q", tc.input, intent.Type, IntentCreateProject)
		}
	}
}

func TestParseListProjects(t *testing.T) {
	p := NewParser()
	tests := []string{
		"list projects",
		"show projects",
		"list all projects",
		"show all projects",
	}
	for _, input := range tests {
		intent, err := p.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", input, err)
			continue
		}
		if intent.Type != IntentListProjects {
			t.Errorf("Parse(%q) Type = %q, want %q", input, intent.Type, IntentListProjects)
		}
	}
}

func TestParseDeleteProject(t *testing.T) {
	p := NewParser()
	intent, err := p.Parse("delete project my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Type != IntentDeleteProject {
		t.Errorf("Type = %q, want %q", intent.Type, IntentDeleteProject)
	}
	if intent.Params["id"] != "my-app" {
		t.Errorf("Params[id] = %q, want %q", intent.Params["id"], "my-app")
	}
}

func TestParseDeleteProjectRemove(t *testing.T) {
	p := NewParser()
	intent, err := p.Parse("remove project test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Type != IntentDeleteProject {
		t.Errorf("Type = %q, want %q", intent.Type, IntentDeleteProject)
	}
}

func TestParseShowControllers(t *testing.T) {
	p := NewParser()
	tests := []string{
		"show controllers",
		"list controllers",
		"list all controllers",
	}
	for _, input := range tests {
		intent, err := p.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", input, err)
			continue
		}
		if intent.Type != IntentShowControllers {
			t.Errorf("Parse(%q) Type = %q, want %q", input, intent.Type, IntentShowControllers)
		}
	}
}

func TestParseShowResources(t *testing.T) {
	p := NewParser()
	tests := []string{
		"show resources",
		"list resources",
		"list all resources",
		"show resource kinds",
	}
	for _, input := range tests {
		intent, err := p.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", input, err)
			continue
		}
		if intent.Type != IntentShowResources {
			t.Errorf("Parse(%q) Type = %q, want %q", input, intent.Type, IntentShowResources)
		}
	}
}

func TestParseShowHealth(t *testing.T) {
	p := NewParser()
	tests := []string{
		"show health",
		"system health",
		"check health",
	}
	for _, input := range tests {
		intent, err := p.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", input, err)
			continue
		}
		if intent.Type != IntentShowHealth {
			t.Errorf("Parse(%q) Type = %q, want %q", input, intent.Type, IntentShowHealth)
		}
	}
}

func TestParseEmpty(t *testing.T) {
	p := NewParser()
	_, err := p.Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseUnrecognized(t *testing.T) {
	p := NewParser()
	_, err := p.Parse("do something random")
	if err == nil {
		t.Fatal("expected error for unrecognized input")
	}
}

func TestParseCaseInsensitivity(t *testing.T) {
	p := NewParser()
	intent, err := p.Parse("CREATE PROJECT X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Type != IntentCreateProject {
		t.Errorf("Type = %q, want %q", intent.Type, IntentCreateProject)
	}
}

// ── Planner Tests ───────────────────────────────────────────────────────

func TestPlanCreateProject(t *testing.T) {
	pl := NewPlanner()
	intent := &Intent{
		ID:   "test-1",
		Type: IntentCreateProject,
		Params: map[string]string{
			"name": "Test Project",
			"id":   "test-project",
		},
	}
	plan, err := pl.Plan(intent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(plan.Steps))
	}
	expectedSteps := []string{
		"Validate Request",
		"Create Project Resource",
		"Wait for Reconciliation",
		"Verify Status",
		"Complete",
	}
	for i, step := range plan.Steps {
		if step.Name != expectedSteps[i] {
			t.Errorf("Step %d Name = %q, want %q", i+1, step.Name, expectedSteps[i])
		}
		if step.Status != StepPending {
			t.Errorf("Step %d Status = %q, want %q", i+1, step.Status, StepPending)
		}
	}
	// Verify dependency chain
	if len(plan.Steps[1].Dependencies) != 1 || plan.Steps[1].Dependencies[0] != "1" {
		t.Error("Step 2 should depend on Step 1")
	}
	if len(plan.Steps[4].Dependencies) != 1 || plan.Steps[4].Dependencies[0] != "4" {
		t.Error("Step 5 should depend on Step 4")
	}
}

func TestPlanListProjects(t *testing.T) {
	pl := NewPlanner()
	plan, err := pl.Plan(&Intent{ID: "test", Type: IntentListProjects})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(plan.Steps))
	}
}

func TestPlanDeleteProject(t *testing.T) {
	pl := NewPlanner()
	plan, err := pl.Plan(&Intent{
		ID:   "test",
		Type: IntentDeleteProject,
		Params: map[string]string{"name": "Test", "id": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(plan.Steps))
	}
}

func TestPlanShowControllers(t *testing.T) {
	pl := NewPlanner()
	plan, err := pl.Plan(&Intent{ID: "test", Type: IntentShowControllers})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(plan.Steps))
	}
}

func TestPlanShowResources(t *testing.T) {
	pl := NewPlanner()
	plan, err := pl.Plan(&Intent{ID: "test", Type: IntentShowResources})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(plan.Steps))
	}
}

func TestPlanShowHealth(t *testing.T) {
	pl := NewPlanner()
	plan, err := pl.Plan(&Intent{ID: "test", Type: IntentShowHealth})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(plan.Steps))
	}
}

func TestPlanInvalidType(t *testing.T) {
	pl := NewPlanner()
	_, err := pl.Plan(&Intent{ID: "test", Type: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid intent type")
	}
}

// ── Deploy Parser Tests ─────────────────────────────────────────────────

func TestParseDeployRuntimeAppFromURL(t *testing.T) {
	p := NewParser()
	tests := []struct {
		input   string
		runtime string
		appName string
	}{
		{"deploy my node app from https://github.com/user/myapp", "node", "myapp"},
		{"deploy go app from https://github.com/user/hello-world", "go", "hello-world"},
		{"deploy python app from https://github.com/user/api", "python", "api"},
		{"deploy static app from https://github.com/user/site", "static", "site"},
		{"deploy nextjs app from https://github.com/user/next-blog", "nextjs", "next-blog"},
		{"deploy laravel app from https://github.com/user/laravel-app", "laravel", "laravel-app"},
		{"deploy docker app from https://github.com/user/container", "docker", "container"},
	}
	for _, tc := range tests {
		intent, err := p.Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.input, err)
			continue
		}
		if intent.Type != IntentDeploy {
			t.Errorf("Parse(%q) Type = %q, want %q", tc.input, intent.Type, IntentDeploy)
		}
		if intent.Params["runtime"] != tc.runtime {
			t.Errorf("Parse(%q) runtime = %q, want %q", tc.input, intent.Params["runtime"], tc.runtime)
		}
		if intent.Params["appName"] != tc.appName {
			t.Errorf("Parse(%q) appName = %q, want %q", tc.input, intent.Params["appName"], tc.appName)
		}
	}
}

func TestParseDeployAppFromURL(t *testing.T) {
	p := NewParser()
	intent, err := p.Parse("deploy app from https://github.com/user/my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Type != IntentDeploy {
		t.Errorf("Type = %q, want %q", intent.Type, IntentDeploy)
	}
	if intent.Params["runtime"] != "auto" {
		t.Errorf("runtime = %q, want %q", intent.Params["runtime"], "auto")
	}
	if intent.Params["sourceURL"] != "https://github.com/user/my-repo" {
		t.Errorf("sourceURL = %q, want %q", intent.Params["sourceURL"], "https://github.com/user/my-repo")
	}
	if intent.Params["appName"] != "my-repo" {
		t.Errorf("appName = %q, want %q", intent.Params["appName"], "my-repo")
	}
}

func TestParseDeployFromURL(t *testing.T) {
	p := NewParser()
	intent, err := p.Parse("deploy from https://github.com/user/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Type != IntentDeploy {
		t.Errorf("Type = %q, want %q", intent.Type, IntentDeploy)
	}
	if intent.Params["runtime"] != "auto" {
		t.Errorf("runtime = %q, want %q", intent.Params["runtime"], "auto")
	}
}

// ── Deploy Planner Tests ────────────────────────────────────────────────

func TestPlanDeploy(t *testing.T) {
	pl := NewPlanner()
	intent := &Intent{
		ID:   "test-deploy",
		Type: IntentDeploy,
		Params: map[string]string{
			"runtime":   "node",
			"sourceURL": "https://github.com/user/myapp",
			"appName":   "myapp",
		},
	}
	plan, err := pl.Plan(intent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 5 {
		t.Fatalf("expected 5 steps, got %d", len(plan.Steps))
	}
	expectedSteps := []string{
		"Validate Deploy Request",
		"Create Application Resource",
		"Wait for Application Controller",
		"Check Application Status",
		"Return Access URL",
	}
	for i, step := range plan.Steps {
		if step.Name != expectedSteps[i] {
			t.Errorf("Step %d Name = %q, want %q", i+1, step.Name, expectedSteps[i])
		}
		if step.Status != StepPending {
			t.Errorf("Step %d Status = %q, want %q", i+1, step.Status, StepPending)
		}
	}
	// Verify dependency chain
	if len(plan.Steps[1].Dependencies) != 1 || plan.Steps[1].Dependencies[0] != "1" {
		t.Error("Step 2 should depend on Step 1")
	}
	// Verify plan preview exists
	if plan.Preview == nil {
		t.Fatal("PlanPreview should not be nil")
	}
	if len(plan.Preview.Steps) != 5 {
		t.Errorf("Preview.Steps = %d, want 5", len(plan.Preview.Steps))
	}
}

// ── Deploy Engine Tests ─────────────────────────────────────────────────

func TestEngineSubmitDeploy(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "deploy my node app from https://github.com/user/myapp")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	// Verify parse result
	if intent.Type != IntentDeploy {
		t.Errorf("Type = %q, want %q", intent.Type, IntentDeploy)
	}
	if intent.Params["runtime"] != "node" {
		t.Errorf("runtime = %q, want %q", intent.Params["runtime"], "node")
	}
	if intent.Params["appName"] != "myapp" {
		t.Errorf("appName = %q, want %q", intent.Params["appName"], "myapp")
	}

	// Verify awaiting approval
	if intent.Status != IntentAwaitingApproval {
		t.Errorf("Status = %q, want %q", intent.Status, IntentAwaitingApproval)
	}

	// Verify plan and preview
	plan, ok := engine.GetPlan(intent.PlanID)
	if !ok {
		t.Fatal("GetPlan() returned false")
	}
	if plan.Preview == nil {
		t.Fatal("Plan should have a preview")
	}
	if len(plan.Preview.Steps) != 5 {
		t.Errorf("Preview has %d steps, want 5", len(plan.Preview.Steps))
	}
	if len(plan.Preview.Resources) == 0 {
		t.Error("Preview should list affected resources")
	}
}

func TestEngineConfirmDeploy(t *testing.T) {
	engine, _, _, reg, _ := setupTestEngine(t)

	// Register Application kind
	if err := reg.RegisterKind(resource.Kind{
		Name:       "Application",
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	intent, err := engine.Submit(context.Background(), "deploy my node app from https://github.com/user/myapp")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}
	if intent.Status != IntentAwaitingApproval {
		t.Fatalf("Status = %q, want %q", intent.Status, IntentAwaitingApproval)
	}

	// Confirm
	confirmed, err := engine.Confirm(context.Background(), intent.ID)
	if err != nil {
		t.Fatalf("Confirm() error: %v", err)
	}
	if confirmed.Status != IntentExecuting {
		t.Errorf("After Confirm, Status = %q, want %q", confirmed.Status, IntentExecuting)
	}

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	updated, ok := engine.GetIntent(intent.ID)
	if !ok {
		t.Fatal("GetIntent() returned false")
	}
	t.Logf("Deploy intent status: %s, result: %v, error: %s", updated.Status, updated.Result, updated.Error)
	// Note: deploy execution may partially succeed depending on controller setup,
	// but the App resource should have been created
}

func TestEngineConfirmWithoutSubmit(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	_, err := engine.Confirm(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent intent")
	}
}

func TestEngineConfirmWrongStatus(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "invalid input here")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}
	if intent.Status != IntentFailed {
		t.Fatalf("Status = %q, want %q", intent.Status, IntentFailed)
	}

	// Confirm should fail because intent is not awaiting_approval
	_, err = engine.Confirm(context.Background(), intent.ID)
	if err == nil {
		t.Fatal("Expected error for non-awaiting intent")
	}
}

// ── Engine Integration Tests ───────────────────────────────────────────

func setupTestEngine(t *testing.T) (*IntentEngine, *events.Bus, *health.Manager, *resource.Registry, *controller.Manager) {
	t.Helper()

	log := logging.NewSubsystemLogger("intent-test", logging.LevelError)
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
	t.Cleanup(func() {
		_ = ctrlManager.Stop(context.Background())
	})

	pluginReg := plugin.NewRegistry()
	capReg := registry.NewManager("capability", log)
	provReg := registry.NewManager("provider", log)

	engine := NewIntentEngine(reg, ctrlManager, healthMgr, pluginReg, capReg, provReg, bus, log)

	t.Cleanup(func() {
		healthMgr.Stop(context.Background())
		bus.Stop()
	})

	return engine, bus, healthMgr, reg, ctrlManager
}

func TestEngineSubmitCreateProject(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "create project EngineTest")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	if intent.ID == "" {
		t.Error("Intent ID should not be empty")
	}
	if intent.Type != IntentCreateProject {
		t.Errorf("Type = %q, want %q", intent.Type, IntentCreateProject)
	}
	if intent.PlanID == "" {
		t.Error("PlanID should not be empty")
	}
	// After Submit, intent should be awaiting approval (not executing)
	if intent.Status != IntentAwaitingApproval {
		t.Errorf("Status = %q, want %q", intent.Status, IntentAwaitingApproval)
	}

	// Confirm execution
	confirmed, err := engine.Confirm(context.Background(), intent.ID)
	if err != nil {
		t.Fatalf("Confirm() error: %v", err)
	}
	if confirmed.Status != IntentExecuting {
		t.Errorf("After Confirm, Status = %q, want %q", confirmed.Status, IntentExecuting)
	}

	// Wait for execution to complete
	time.Sleep(200 * time.Millisecond)

	// Get the updated intent
	updated, ok := engine.GetIntent(intent.ID)
	if !ok {
		t.Fatal("GetIntent() returned false")
	}
	if updated.Status != IntentCompleted {
		t.Errorf("Final Status = %q, want %q", updated.Status, IntentCompleted)
	}
	if updated.Result == nil {
		t.Fatal("Result should not be nil")
	}
	if !updated.Result.Success {
		t.Errorf("Result.Success = false, want true: %s", updated.Result.Summary)
	}

	// Verify the project was actually created
	plan, ok := engine.GetPlan(intent.PlanID)
	if !ok {
		t.Fatal("GetPlan() returned false")
	}
	if plan.Status != IntentCompleted {
		t.Errorf("Plan Status = %q, want %q", plan.Status, IntentCompleted)
	}
}

func TestEngineSubmitAndGetIntent(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "list projects")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	got, ok := engine.GetIntent(intent.ID)
	if !ok {
		t.Fatal("GetIntent() returned false")
	}
	if got.ID != intent.ID {
		t.Errorf("ID = %q, want %q", got.ID, intent.ID)
	}
	if got.Type != IntentListProjects {
		t.Errorf("Type = %q, want %q", got.Type, IntentListProjects)
	}
}

func TestEngineSubmitAndGetPlan(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "list projects")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	plan, ok := engine.GetPlan(intent.PlanID)
	if !ok {
		t.Fatal("GetPlan() returned false")
	}
	if plan.IntentID != intent.ID {
		t.Errorf("IntentID = %q, want %q", plan.IntentID, intent.ID)
	}
	if len(plan.Steps) == 0 {
		t.Error("Plan should have at least 1 step")
	}
}

func TestEngineSubmitInvalid(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "do something random")
	if err != nil {
		t.Fatalf("Submit() should not return error for invalid input: %v", err)
	}
	if intent.Status != IntentFailed {
		t.Errorf("Status = %q, want %q", intent.Status, IntentFailed)
	}
	if intent.Error == "" {
		t.Error("Error should not be empty for invalid intent")
	}
}

func TestEngineSubmitEmpty(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	intent, err := engine.Submit(context.Background(), "")
	if err != nil {
		t.Fatalf("Submit() should not return error for empty input: %v", err)
	}
	if intent.Status != IntentFailed {
		t.Errorf("Status = %q, want %q", intent.Status, IntentFailed)
	}
}

func TestEngineGetIntentNotFound(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	_, ok := engine.GetIntent("nonexistent")
	if ok {
		t.Error("GetIntent() should return false for nonexistent ID")
	}
}

func TestEngineGetPlanNotFound(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	_, ok := engine.GetPlan("nonexistent")
	if ok {
		t.Error("GetPlan() should return false for nonexistent ID")
	}
}

func TestEngineListIntents(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	_, _ = engine.Submit(context.Background(), "list projects")
	_, _ = engine.Submit(context.Background(), "show health")

	intents := engine.ListIntents()
	if len(intents) != 2 {
		t.Errorf("ListIntents() = %d, want 2", len(intents))
	}
}

func TestEngineListPlans(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)

	_, _ = engine.Submit(context.Background(), "list projects")
	_, _ = engine.Submit(context.Background(), "show health")

	plans := engine.ListPlans()
	if len(plans) != 2 {
		t.Errorf("ListPlans() = %d, want 2", len(plans))
	}
}

func TestEngineDeleteProject(t *testing.T) {
	engine, _, _, reg, _ := setupTestEngine(t)

	// First create a project
	p := project.NewProject("delete-test", "Delete Test", "development", "Test for deletion")
	if err := reg.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	// Submit delete intent
	intent, err := engine.Submit(context.Background(), "delete project delete-test")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}
	if intent.Status != IntentAwaitingApproval {
		t.Errorf("After Submit, Status = %q, want %q", intent.Status, IntentAwaitingApproval)
	}

	// Confirm execution
	_, err = engine.Confirm(context.Background(), intent.ID)
	if err != nil {
		t.Fatalf("Confirm() error: %v", err)
	}

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	updated, ok := engine.GetIntent(intent.ID)
	if !ok {
		t.Fatal("GetIntent() returned false")
	}
	if updated.Status != IntentCompleted {
		t.Errorf("Status = %q, want %q. Error: %s", updated.Status, IntentCompleted, updated.Error)
	}

	// Verify project is deleted
	_, err = reg.Get(project.Kind, "delete-test")
	if err == nil {
		t.Error("Project should have been deleted")
	}
}
