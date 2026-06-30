package application

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/workflow"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Application Resource Tests ─────────────────────────────────────────────

func TestNewApplication_Defaults(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{
			Type: SourceGit,
			URL:  "https://github.com/user/my-app",
		},
		Runtime: ApplicationRuntime{
			Type: RuntimeNode,
		},
	}
	app := NewApplication("my-app", "My App", spec)

	if app.GetKind() != Kind {
		t.Errorf("GetKind() = %q, want %q", app.GetKind(), Kind)
	}
	if app.GetMetadata().ID != "my-app" {
		t.Errorf("ID = %q, want %q", app.GetMetadata().ID, "my-app")
	}
	if app.GetMetadata().Namespace != resource.NamespaceDefault {
		t.Errorf("Namespace = %q, want %q", app.GetMetadata().Namespace, resource.NamespaceDefault)
	}
	if app.Metadata_.Name != "My App" {
		t.Errorf("Name = %q, want %q", app.Metadata_.Name, "My App")
	}
	if app.Spec_.Source.Type != SourceGit {
		t.Errorf("Source.Type = %q, want %q", app.Spec_.Source.Type, SourceGit)
	}
	if app.Spec_.Runtime.Type != RuntimeNode {
		t.Errorf("Runtime.Type = %q, want %q", app.Spec_.Runtime.Type, RuntimeNode)
	}
	if app.Status_.Phase != PhaseCreating {
		t.Errorf("Phase = %q, want %q", app.Status_.Phase, PhaseCreating)
	}
	if app.Status_.Health != HealthHealthy {
		t.Errorf("Health = %q, want %q", app.Status_.Health, HealthHealthy)
	}
}

func TestNewApplication_DefaultCommand(t *testing.T) {
	tests := []struct {
		runtime   string
		wantCmd   string
		wantPort  int
	}{
		{RuntimeNode, "npm start", 3000},
		{RuntimePython, "python app.py", 8000},
		{RuntimeGo, "./app", 8080},
		{RuntimeStatic, "", 80},
		{RuntimeNextJS, "npm run start", 3000},
		{RuntimeLaravel, "php artisan serve", 8000},
		{RuntimeDocker, "", 0},
	}

	for _, tt := range tests {
		spec := ApplicationSpec{
			Source: ApplicationSource{Type: SourceLocal},
			Runtime: ApplicationRuntime{Type: tt.runtime},
		}
		app := NewApplication("test-"+tt.runtime, "Test", spec)
		app.EnsureDefaults()

		if app.Spec_.Runtime.Command != tt.wantCmd {
			t.Errorf("Runtime %q: default command = %q, want %q", tt.runtime, app.Spec_.Runtime.Command, tt.wantCmd)
		}
		if app.Spec_.Runtime.Port != tt.wantPort && tt.runtime != RuntimeDocker {
			t.Errorf("Runtime %q: default port = %d, want %d", tt.runtime, app.Spec_.Runtime.Port, tt.wantPort)
		}
	}
}

func TestApplication_Validate_Valid(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{
			Type: SourceGit,
			URL:  "https://github.com/user/repo",
		},
		Runtime: ApplicationRuntime{
			Type: RuntimeNode,
			Port: 3000,
		},
	}
	app := NewApplication("valid-app", "Valid App", spec)
	if err := app.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid application: %v", err)
	}
}

func TestApplication_Validate_EmptyID(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("", "No ID", spec)
	if err := app.Validate(); err == nil {
		t.Error("Validate() should return error for empty ID")
	}
}

func TestApplication_Validate_InvalidID(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	invalidIDs := []string{"UPPERCASE", "has spaces", "has_underscore", "", "-starts-with-hyphen", "ends-with-hyphen-"}
	for _, id := range invalidIDs {
		app := NewApplication(id, "Test", spec)
		if err := app.Validate(); err == nil {
			t.Errorf("Validate() should return error for invalid ID %q", id)
		}
	}
}

func TestApplication_Validate_InvalidSourceType(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: "invalid-source"},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)
	if err := app.Validate(); err == nil {
		t.Error("Validate() should return error for invalid source type")
	}
}

func TestApplication_Validate_GitSourceRequiresURL(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceGit},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)
	if err := app.Validate(); err == nil {
		t.Error("Validate() should return error for git source without URL")
	}
}

func TestApplication_Validate_InvalidRuntime(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: "invalid-runtime"},
	}
	app := NewApplication("test", "Test", spec)
	if err := app.Validate(); err == nil {
		t.Error("Validate() should return error for invalid runtime type")
	}
}

func TestApplication_Validate_WrongKind(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)
	app.Metadata_.Kind = "WrongKind"
	if err := app.Validate(); err == nil {
		t.Error("Validate() should return error for wrong kind")
	}
}

func TestApplication_Validate_PortRange(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode, Port: 70000},
	}
	app := NewApplication("test", "Test", spec)
	if err := app.Validate(); err == nil {
		t.Error("Validate() should return error for port out of range")
	}
}

func TestApplication_EnsureDefaults(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceGit, URL: "https://github.com/user/repo"},
		Runtime: ApplicationRuntime{Type: RuntimePython},
	}
	app := NewApplication("test", "Test", spec)
	app.Metadata_.Labels = nil
	app.Metadata_.Annotations = nil
	app.Spec_.Settings = nil
	app.Spec_.Source.Branch = ""
	app.Spec_.Runtime.Command = ""
	app.Spec_.Runtime.Port = 0
	app.Status_.Phase = ""
	app.Status_.Health = ""

	app.EnsureDefaults()

	if app.Metadata_.Labels == nil {
		t.Error("Labels should not be nil after EnsureDefaults")
	}
	if app.Metadata_.Annotations == nil {
		t.Error("Annotations should not be nil after EnsureDefaults")
	}
	if app.Spec_.Settings == nil {
		t.Error("Settings should not be nil after EnsureDefaults")
	}
	if app.Spec_.Source.Branch != "main" {
		t.Errorf("Branch after EnsureDefaults = %q, want %q", app.Spec_.Source.Branch, "main")
	}
	if app.Spec_.Runtime.Command != "python app.py" {
		t.Errorf("Command after EnsureDefaults = %q, want %q", app.Spec_.Runtime.Command, "python app.py")
	}
	if app.Spec_.Runtime.Port != 8000 {
		t.Errorf("Port after EnsureDefaults = %d, want %d", app.Spec_.Runtime.Port, 8000)
	}
	if app.Spec_.Deployment.Replicas != 1 {
		t.Errorf("Replicas after EnsureDefaults = %d, want 1", app.Spec_.Deployment.Replicas)
	}
	if app.Status_.Phase != PhaseCreating {
		t.Errorf("Phase after EnsureDefaults = %q, want %q", app.Status_.Phase, PhaseCreating)
	}
	if app.Status_.Health != HealthHealthy {
		t.Errorf("Health after EnsureDefaults = %q, want %q", app.Status_.Health, HealthHealthy)
	}
	if app.Metadata_.Labels["runtime"] != RuntimePython {
		t.Errorf("runtime label = %q, want %q", app.Metadata_.Labels["runtime"], RuntimePython)
	}
	if app.Metadata_.Labels["source"] != SourceGit {
		t.Errorf("source label = %q, want %q", app.Metadata_.Labels["source"], SourceGit)
	}
}

func TestApplication_ResourceInterface(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)

	// Compile-time check: Application implements resource.Resource.
	var _ resource.Resource = app

	// GetSpec/GetStatus should return the correct types.
	s := app.GetSpec()
	if _, ok := s.(ApplicationSpec); !ok {
		t.Errorf("GetSpec() returned %T, want ApplicationSpec", s)
	}
	status := app.GetStatus()
	if _, ok := status.(ApplicationStatus); !ok {
		t.Errorf("GetStatus() returned %T, want ApplicationStatus", status)
	}

	// SetStatus should work.
	newStatus := ApplicationStatus{Phase: PhaseRunning, Health: HealthHealthy}
	app.SetStatus(newStatus)
	if app.Status_.Phase != PhaseRunning {
		t.Errorf("After SetStatus, Phase = %q, want %q", app.Status_.Phase, PhaseRunning)
	}

	// SetStatus with wrong type should be ignored.
	app.SetStatus("wrong type")
	if app.Status_.Phase != PhaseRunning {
		t.Error("SetStatus with wrong type should be a no-op")
	}
}

func TestApplication_AddCondition(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)

	app.AddCondition("Created", "True", "AppCreated", "Application created")
	if len(app.Status_.Conditions) != 1 {
		t.Fatalf("Conditions length = %d, want 1", len(app.Status_.Conditions))
	}
	if app.Status_.Conditions[0].Type != "Created" {
		t.Errorf("Condition type = %q, want %q", app.Status_.Conditions[0].Type, "Created")
	}
	if app.Status_.Conditions[0].Status != "True" {
		t.Errorf("Condition status = %q, want %q", app.Status_.Conditions[0].Status, "True")
	}

	// Update existing condition.
	app.AddCondition("Created", "False", "AppFailed", "Something went wrong")
	if len(app.Status_.Conditions) != 1 {
		t.Fatalf("Conditions length after update = %d, want 1", len(app.Status_.Conditions))
	}
	if app.Status_.Conditions[0].Status != "False" {
		t.Errorf("Condition status after update = %q, want %q", app.Status_.Conditions[0].Status, "False")
	}
}

func TestApplication_GetCondition(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)
	app.AddCondition("Running", "True", "", "")

	c := app.GetCondition("Running")
	if c == nil {
		t.Fatal("GetCondition('Running') returned nil")
	}
	if c.Status != "True" {
		t.Errorf("Condition status = %q, want %q", c.Status, "True")
	}

	c = app.GetCondition("NonExistent")
	if c != nil {
		t.Error("GetCondition('NonExistent') should return nil")
	}
}

func TestApplication_Touch(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test", "Test", spec)
	before := app.Status_.UpdatedAt

	time.Sleep(time.Millisecond)
	app.Touch()

	if app.Status_.UpdatedAt.Equal(before) {
		t.Error("Touch() should update UpdatedAt")
	}
}

// ── Deployment Plan Tests ─────────────────────────────────────────────────

func TestBuildDeploymentPlan_GitSource(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{
			Type: SourceGit,
			URL:  "https://github.com/user/repo",
		},
		Runtime: ApplicationRuntime{
			Type: RuntimeNode,
		},
		Build: &ApplicationBuild{
			Command:    "npm run build",
			InstallCmd: "npm ci",
		},
	}
	app := NewApplication("my-app", "My App", spec)
	app.EnsureDefaults()

	plan := BuildDeploymentPlan(app)

	if len(plan) == 0 {
		t.Fatal("BuildDeploymentPlan returned empty plan")
	}

	// Verify plan structure.
	expectedSteps := []string{"validate", "clone", "install", "build", "deploy", "healthcheck", "complete"}
	if len(plan) != len(expectedSteps) {
		t.Fatalf("Plan length = %d, want %d", len(plan), len(expectedSteps))
	}
	for i, step := range plan {
		if step.ID != expectedSteps[i] {
			t.Errorf("Step %d: ID = %q, want %q", i, step.ID, expectedSteps[i])
		}
	}

	// Verify dependency chain.
	for i := 1; i < len(plan); i++ {
		if len(plan[i].DependsOn) == 0 {
			t.Errorf("Step %q has no dependencies", plan[i].ID)
		}
	}
}

func TestBuildDeploymentPlan_LocalSource(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimePython},
	}
	app := NewApplication("local-app", "Local App", spec)
	app.EnsureDefaults()

	plan := BuildDeploymentPlan(app)

	// Local source should NOT have a clone step.
	for _, step := range plan {
		if step.ID == "clone" {
			t.Error("Local source should not have a clone step")
		}
	}

	// Should still have validate, deploy, healthcheck, complete.
	hasValidate := false
	hasDeploy := false
	hasHealthCheck := false
	hasComplete := false
	for _, step := range plan {
		switch step.ID {
		case "validate":
			hasValidate = true
		case "deploy":
			hasDeploy = true
		case "healthcheck":
			hasHealthCheck = true
		case "complete":
			hasComplete = true
		}
	}
	if !hasValidate {
		t.Error("Plan missing validate step")
	}
	if !hasDeploy {
		t.Error("Plan missing deploy step")
	}
	if !hasHealthCheck {
		t.Error("Plan missing healthcheck step")
	}
	if !hasComplete {
		t.Error("Plan missing complete step")
	}
}

func TestBuildDeploymentPlan_NoBuild(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceGit, URL: "https://github.com/user/repo"},
		Runtime: ApplicationRuntime{Type: RuntimeGo},
		// No Build configured.
	}
	app := NewApplication("no-build", "No Build", spec)
	app.EnsureDefaults()

	plan := BuildDeploymentPlan(app)

	// Should NOT have install or build steps.
	for _, step := range plan {
		if step.ID == "install" || step.ID == "build" {
			t.Errorf("Plan should not have %q step when build is not configured", step.ID)
		}
	}
}

func TestBuildDeploymentPlan_TargetValues(t *testing.T) {
	spec := ApplicationSpec{
		Source: ApplicationSource{
			Type: SourceGit,
			URL:  "https://github.com/user/repo.git",
		},
		Runtime: ApplicationRuntime{
			Type:    RuntimeNode,
			Command: "npm run start",
		},
	}
	app := NewApplication("target-test", "Target Test", spec)
	app.EnsureDefaults()

	plan := BuildDeploymentPlan(app)

	for _, step := range plan {
		switch step.ID {
		case "validate":
			if step.Target != "Application:target-test" {
				t.Errorf("validate target = %q, want %q", step.Target, "Application:target-test")
			}
		case "clone":
			if step.Target != "https://github.com/user/repo.git" {
				t.Errorf("clone target = %q, want %q", step.Target, "https://github.com/user/repo.git")
			}
		case "deploy":
			if step.Target != "runtime:node" {
				t.Errorf("deploy target = %q, want %q", step.Target, "runtime:node")
			}
		case "complete":
			if step.Target != "target-test" {
				t.Errorf("complete target = %q, want %q", step.Target, "target-test")
			}
		}
	}
}

// ── Controller Tests ──────────────────────────────────────────────────────

// mockWorkflowService implements the WorkflowService interface for testing.
type mockWorkflowService struct {
	mu               sync.Mutex
	definitions      map[string]*workflow.WorkflowDefinition
	defCount         int
	submitCount      int
	submitError      error
	registerError    error
	lastSubmittedDef string
	// Optional: pre-set to control GetExecution behavior.
	executionResult *workflow.WorkflowExecution
	executionError  error
}

func newMockWorkflowService() *mockWorkflowService {
	return &mockWorkflowService{
		definitions: make(map[string]*workflow.WorkflowDefinition),
	}
}

func (m *mockWorkflowService) RegisterDefinition(def *workflow.WorkflowDefinition) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defCount++
	m.definitions[def.ID] = def
	return m.registerError
}

func (m *mockWorkflowService) GetExecution(runID string) (*workflow.WorkflowExecution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.executionError != nil {
		return nil, m.executionError
	}
	if m.executionResult != nil {
		return m.executionResult, nil
	}

	// Default: return a completed execution with a deploy URL.
	// Build a fake run with a deploy node result.
	deployNode := workflow.NewTaskNode("deploy", "Deploy Application", "provider.deploy", "")
	deployNode.Result = "Running at http://localhost:9000 (pid=12345, port=9000)"
	deployNode.SetStatus(workflow.NodeSucceeded)

	run := &workflow.WorkflowRun{
		ID:     runID,
		Status: workflow.WorkflowCompleted,
		Nodes: []workflow.Node{
			deployNode,
		},
		Result: &workflow.WorkflowResult{
			Success: true,
			Summary: "Completed successfully — 5 nodes",
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	return workflow.NewWorkflowExecution(run, workflow.WorkflowExecutionSpec{
		WorkflowID: "def-" + runID,
	}), nil
}

func (m *mockWorkflowService) Submit(def *workflow.WorkflowDefinition) (*workflow.WorkflowRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.submitError != nil {
		return nil, m.submitError
	}

	m.submitCount++
	m.lastSubmittedDef = def.ID
	return &workflow.WorkflowRun{
		ID:           "run-" + def.ID,
		DefinitionID: def.ID,
		Status:       workflow.WorkflowPending,
		Nodes:        def.Nodes,
	}, nil
}

func setupApplicationControllerTest(t *testing.T) (*ApplicationController, *resource.Registry, *events.Bus, *mockWorkflowService) {
	t.Helper()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	reg := resource.NewRegistry(bus, log)

	// Register the Application kind.
	if err := reg.RegisterKind(resource.Kind{
		Name:       Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	mockWS := newMockWorkflowService()
	ac := NewApplicationController(reg, bus, mockWS, log)
	return ac, reg, bus, mockWS
}

func TestApplicationController_Name(t *testing.T) {
	ac, _, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	if ac.Name() != "application" {
		t.Errorf("Name() = %q, want %q", ac.Name(), "application")
	}
	if ac.Kind() != Kind {
		t.Errorf("Kind() = %q, want %q", ac.Kind(), Kind)
	}
}

func TestApplicationController_Health(t *testing.T) {
	ac, _, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	h := ac.Health()
	if h.Name != "application" {
		t.Errorf("Health().Name = %q, want %q", h.Name, "application")
	}
	if h.State != "stopped" {
		t.Errorf("Health().State before start = %q, want %q", h.State, "stopped")
	}
}

func TestApplicationController_StartStop(t *testing.T) {
	ac, _, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := ac.Start(ctx); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}
	// Idempotent.
	if err := ac.Start(ctx); err != nil {
		t.Fatalf("second Start() returned error: %v", err)
	}

	if err := ac.Stop(ctx); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}
	// Idempotent.
	if err := ac.Stop(ctx); err != nil {
		t.Fatalf("second Stop() returned error: %v", err)
	}

	h := ac.Health()
	if h.State != "stopped" {
		t.Errorf("Health().State after stop = %q, want %q", h.State, "stopped")
	}
}

func TestApplicationController_ReconcileCreate(t *testing.T) {
	ac, reg, bus, mockWS := setupApplicationControllerTest(t)
	defer bus.Stop()

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceGit, URL: "https://github.com/user/repo"},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("test-app", "Test App", spec)
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	// First reconciliation: Creating → Deploying + workflow submission.
	// The controller transitions to Deploying and requeues (returns Requeue=true).
	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "test-app"})
	if result.Err != nil {
		t.Errorf("First Reconcile() returned error: %v", result.Err)
	}

	// Verify the application transitioned to Deploying phase.
	updated, err := reg.Get(Kind, "test-app")
	if err != nil {
		t.Fatal(err)
	}
	a := updated.(*Application)
	if a.Status_.Phase != PhaseDeploying {
		t.Errorf("Phase after first reconcile = %q, want %q", a.Status_.Phase, PhaseDeploying)
	}
	if a.GetCondition("Created") == nil {
		t.Error("Created condition was not set after first reconcile")
	}
	if a.GetCondition("Deploying") == nil {
		t.Error("Deploying condition was not set after first reconcile")
	}
	if a.Status_.CurrentDeploymentID == "" {
		t.Error("CurrentDeploymentID should be set after workflow submission")
	}

	// Verify the workflow service was called.
	if mockWS.submitCount == 0 {
		t.Error("Workflow service Submit was not called")
	}

	// Second reconciliation: Deploying → Running (simulates deployment completion).
	result = ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "test-app"})
	if result.Err != nil {
		t.Errorf("Second Reconcile() returned error: %v", result.Err)
	}

	// Verify the application is now Running.
	updated, err = reg.Get(Kind, "test-app")
	if err != nil {
		t.Fatal(err)
	}
	a = updated.(*Application)
	if a.Status_.Phase != PhaseRunning {
		t.Errorf("Phase after second reconcile = %q, want %q", a.Status_.Phase, PhaseRunning)
	}
	if a.GetCondition("Running") == nil {
		t.Error("Running condition was not set after second reconcile")
	}
	if a.Status_.DeploymentCount != 1 {
		t.Errorf("DeploymentCount = %d, want 1", a.Status_.DeploymentCount)
	}
}

func TestApplicationController_ReconcileRunning(t *testing.T) {
	ac, reg, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("running-app", "Running App", spec)
	app.Status_.Phase = PhaseRunning
	app.Status_.Health = HealthHealthy
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "running-app"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	updated, _ := reg.Get(Kind, "running-app")
	a := updated.(*Application)
	if a.Status_.Phase != PhaseRunning {
		t.Errorf("Phase = %q, want %q", a.Status_.Phase, PhaseRunning)
	}
	if a.Status_.Health != HealthHealthy {
		t.Errorf("Health = %q, want %q", a.Status_.Health, HealthHealthy)
	}
}

func TestApplicationController_ReconcileStopped(t *testing.T) {
	ac, reg, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("stopped-app", "Stopped App", spec)
	app.Status_.Phase = PhaseStopped
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "stopped-app"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	updated, _ := reg.Get(Kind, "stopped-app")
	a := updated.(*Application)
	if a.Status_.Phase != PhaseStopped {
		t.Errorf("Phase = %q, want %q", a.Status_.Phase, PhaseStopped)
	}
	c := a.GetCondition("Stopped")
	if c == nil || c.Status != "True" {
		t.Errorf("Stopped condition should be True for stopped applications, got %v", c)
	}
}

func TestApplicationController_ReconcileFailed(t *testing.T) {
	ac, reg, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("failed-app", "Failed App", spec)
	app.Status_.Phase = PhaseFailed
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "failed-app"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	updated, _ := reg.Get(Kind, "failed-app")
	a := updated.(*Application)
	if a.Status_.Phase != PhaseFailed {
		t.Errorf("Phase = %q, want %q", a.Status_.Phase, PhaseFailed)
	}
	if a.Status_.Health != HealthError {
		t.Errorf("Health = %q, want %q", a.Status_.Health, HealthError)
	}
}

func TestApplicationController_ReconcileDeleting(t *testing.T) {
	ac, reg, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceLocal},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("deleting-app", "Deleting App", spec)
	app.Status_.Phase = PhaseDeleting
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "deleting-app"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}
}

func TestApplicationController_ReconcileDeletedApplication(t *testing.T) {
	ac, _, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	// Reconcile an application that doesn't exist (was deleted).
	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "nonexistent"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error for deleted application: %v", result.Err)
	}
	if result.Requeue {
		t.Error("Reconcile() returned Requeue=true for deleted application")
	}
}

func TestApplicationController_ReconcileUnknownKind(t *testing.T) {
	ac, _, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	result := ac.Reconcile(controller.ReconcileRequest{Kind: "WrongKind", ID: "test"})
	if result.Err == nil {
		t.Error("Reconcile() should return error for unknown kind")
	}
}

func TestApplicationController_ErrorsInterface(t *testing.T) {
	ac, _, bus, _ := setupApplicationControllerTest(t)
	defer bus.Stop()

	// Compile-time check: ApplicationController implements controller.Controller.
	var _ controller.Controller = ac
}

// ── Workflow Service Failure Tests ─────────────────────────────────────────

func TestApplicationController_WorkflowSubmitError(t *testing.T) {
	ac, reg, bus, mockWS := setupApplicationControllerTest(t)
	defer bus.Stop()

	mockWS.submitError = fmt.Errorf("engine is down")

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceGit, URL: "https://github.com/user/repo"},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("submit-error", "Submit Error", spec)
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "submit-error"})
	if result.Err == nil {
		t.Error("Reconcile() should return error when workflow submit fails")
	}

	// Application should be in Failed phase.
	updated, _ := reg.Get(Kind, "submit-error")
	a := updated.(*Application)
	if a.Status_.Phase != PhaseFailed {
		t.Errorf("Phase = %q, want %q after submit error", a.Status_.Phase, PhaseFailed)
	}
}

func TestApplicationController_WorkflowRegisterError(t *testing.T) {
	ac, reg, bus, mockWS := setupApplicationControllerTest(t)
	defer bus.Stop()

	mockWS.registerError = fmt.Errorf("duplicate definition")

	spec := ApplicationSpec{
		Source: ApplicationSource{Type: SourceGit, URL: "https://github.com/user/repo"},
		Runtime: ApplicationRuntime{Type: RuntimeNode},
	}
	app := NewApplication("register-error", "Register Error", spec)
	if err := reg.Create(context.Background(), app); err != nil {
		t.Fatal(err)
	}

	result := ac.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "register-error"})
	if result.Err == nil {
		t.Error("Reconcile() should return error when workflow register fails")
	}

	updated, _ := reg.Get(Kind, "register-error")
	a := updated.(*Application)
	if a.Status_.Phase != PhaseFailed {
		t.Errorf("Phase = %q, want %q after register error", a.Status_.Phase, PhaseFailed)
	}
}

// ── Valid Runtimes Test ───────────────────────────────────────────────────

func TestValidRuntimes(t *testing.T) {
	expected := []string{RuntimeNode, RuntimePython, RuntimeGo, RuntimeStatic, RuntimeNextJS, RuntimeLaravel, RuntimeDocker}
	for _, r := range expected {
		if !ValidRuntimes[r] {
			t.Errorf("%q should be a valid runtime", r)
		}
	}
	if ValidRuntimes["invalid"] {
		t.Error("invalid should NOT be a valid runtime")
	}
}

func TestValidSourceTypes(t *testing.T) {
	expected := []string{SourceGit, SourceLocal, SourceDocker}
	for _, s := range expected {
		if !ValidSourceTypes[s] {
			t.Errorf("%q should be a valid source type", s)
		}
	}
	if ValidSourceTypes["invalid"] {
		t.Error("invalid should NOT be a valid source type")
	}
}

func TestValidPhases(t *testing.T) {
	expected := []string{PhaseCreating, PhaseRunning, PhaseStopped, PhaseFailed, PhaseDeleting, PhaseDeploying}
	for _, p := range expected {
		if !ValidPhases[p] {
			t.Errorf("%q should be a valid phase", p)
		}
	}
}

// ── Copy Map Helper ────────────────────────────────────────────────────────

func TestCopyMap(t *testing.T) {
	src := map[string]string{"a": "1", "b": "2"}
	dst := copyMap(src)

	if dst["a"] != "1" || dst["b"] != "2" {
		t.Error("copyMap did not copy correctly")
	}

	// Modifying dst should not affect src.
	dst["a"] = "changed"
	if src["a"] != "1" {
		t.Error("copyMap should create a deep copy")
	}

	// Nil source should return nil.
	if copyMap(nil) != nil {
		t.Error("copyMap(nil) should return nil")
	}
}
