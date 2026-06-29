package project

import (
	"context"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Project Resource Tests ─────────────────────────────────────────────────

func TestNewProject_Defaults(t *testing.T) {
	p := NewProject("my-project", "My Project", EnvDevelopment, "A test project")

	if p.GetKind() != Kind {
		t.Errorf("GetKind() = %q, want %q", p.GetKind(), Kind)
	}
	if p.GetMetadata().ID != "my-project" {
		t.Errorf("ID = %q, want %q", p.GetMetadata().ID, "my-project")
	}
	if p.GetMetadata().Namespace != resource.NamespaceDefault {
		t.Errorf("Namespace = %q, want %q", p.GetMetadata().Namespace, resource.NamespaceDefault)
	}
	if p.Spec_.DisplayName != "My Project" {
		t.Errorf("DisplayName = %q, want %q", p.Spec_.DisplayName, "My Project")
	}
	if p.Spec_.Environment != EnvDevelopment {
		t.Errorf("Environment = %q, want %q", p.Spec_.Environment, EnvDevelopment)
	}
	if p.Status_.Phase != PhaseCreating {
		t.Errorf("Phase = %q, want %q", p.Status_.Phase, PhaseCreating)
	}
	if p.Status_.Health != HealthHealthy {
		t.Errorf("Health = %q, want %q", p.Status_.Health, HealthHealthy)
	}
	if p.Spec_.Settings == nil {
		t.Error("Settings should not be nil")
	}
}

func TestNewProject_DefaultEnvironment(t *testing.T) {
	p := NewProject("test", "Test", "invalid-env", "")
	if p.Spec_.Environment != EnvDevelopment {
		t.Errorf("invalid env defaulted to %q, want %q", p.Spec_.Environment, EnvDevelopment)
	}
}

func TestProject_Validate_Valid(t *testing.T) {
	p := NewProject("valid-project", "Valid Project", EnvProduction, "Production project")
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid project: %v", err)
	}
}

func TestProject_Validate_EmptyID(t *testing.T) {
	p := NewProject("", "No ID", EnvDevelopment, "")
	if err := p.Validate(); err == nil {
		t.Error("Validate() should return error for empty ID")
	}
}

func TestProject_Validate_InvalidID(t *testing.T) {
	invalidIDs := []string{"UPPERCASE", "has spaces", "has_underscore", "", "-starts-with-hyphen", "ends-with-hyphen-"}
	for _, id := range invalidIDs {
		p := NewProject(id, "Test", EnvDevelopment, "")
		if err := p.Validate(); err == nil {
			t.Errorf("Validate() should return error for invalid ID %q", id)
		}
	}
}

func TestProject_Validate_EmptyDisplayName(t *testing.T) {
	p := NewProject("test", "", EnvDevelopment, "")
	if err := p.Validate(); err == nil {
		t.Error("Validate() should return error for empty display name")
	}
}

func TestProject_Validate_InvalidEnvironment(t *testing.T) {
	p := NewProject("test", "Test", "invalid", "")
	// NewProject defaults invalid envs to development, so we must set it directly.
	p.Spec_.Environment = "invalid"
	if err := p.Validate(); err == nil {
		t.Error("Validate() should return error for invalid environment")
	}
}

func TestProject_Validate_WrongKind(t *testing.T) {
	p := NewProject("test", "Test", EnvDevelopment, "")
	p.Metadata_.Kind = "WrongKind"
	if err := p.Validate(); err == nil {
		t.Error("Validate() should return error for wrong kind")
	}
}

func TestProject_EnsureDefaults(t *testing.T) {
	p := NewProject("test", "Test", EnvStaging, "Staging project")
	p.Metadata_.Labels = nil
	p.Metadata_.Annotations = nil
	p.Spec_.Settings = nil
	p.Status_.Phase = ""
	p.Status_.Health = ""

	p.EnsureDefaults()

	if p.Metadata_.Labels == nil {
		t.Error("Labels should not be nil after EnsureDefaults")
	}
	if p.Metadata_.Annotations == nil {
		t.Error("Annotations should not be nil after EnsureDefaults")
	}
	if p.Spec_.Settings == nil {
		t.Error("Settings should not be nil after EnsureDefaults")
	}
	if p.Status_.Phase != PhaseCreating {
		t.Errorf("Phase = %q after EnsureDefaults, want %q", p.Status_.Phase, PhaseCreating)
	}
	if p.Status_.Health != HealthHealthy {
		t.Errorf("Health = %q after EnsureDefaults, want %q", p.Status_.Health, HealthHealthy)
	}
	if p.Metadata_.Labels["environment"] != EnvStaging {
		t.Errorf("environment label = %q, want %q", p.Metadata_.Labels["environment"], EnvStaging)
	}
}

func TestProject_ResourceInterface(t *testing.T) {
	p := NewProject("test", "Test", EnvDevelopment, "")
	// Compile-time check: Project implements resource.Resource.
	var _ resource.Resource = p

	// GetSpec/GetStatus should return the correct types.
	spec := p.GetSpec()
	if _, ok := spec.(ProjectSpec); !ok {
		t.Errorf("GetSpec() returned %T, want ProjectSpec", spec)
	}
	status := p.GetStatus()
	if _, ok := status.(ProjectStatus); !ok {
		t.Errorf("GetStatus() returned %T, want ProjectStatus", status)
	}

	// SetStatus should work.
	newStatus := ProjectStatus{Phase: PhaseActive, Health: HealthHealthy}
	p.SetStatus(newStatus)
	if p.Status_.Phase != PhaseActive {
		t.Errorf("After SetStatus, Phase = %q, want %q", p.Status_.Phase, PhaseActive)
	}

	// SetStatus with wrong type should be ignored.
	p.SetStatus("wrong type")
	if p.Status_.Phase != PhaseActive {
		t.Error("SetStatus with wrong type should be a no-op")
	}
}

func TestProject_AddCondition(t *testing.T) {
	p := NewProject("test", "Test", EnvDevelopment, "")

	p.AddCondition("Ready", "True", "ProjectReady", "Project is ready")
	if len(p.Status_.Conditions) != 1 {
		t.Fatalf("Conditions length = %d, want 1", len(p.Status_.Conditions))
	}
	if p.Status_.Conditions[0].Type != "Ready" {
		t.Errorf("Condition type = %q, want %q", p.Status_.Conditions[0].Type, "Ready")
	}
	if p.Status_.Conditions[0].Status != "True" {
		t.Errorf("Condition status = %q, want %q", p.Status_.Conditions[0].Status, "True")
	}

	// Update existing condition.
	p.AddCondition("Ready", "False", "ProjectNotReady", "Something went wrong")
	if len(p.Status_.Conditions) != 1 {
		t.Fatalf("Conditions length = %d after update, want 1", len(p.Status_.Conditions))
	}
	if p.Status_.Conditions[0].Status != "False" {
		t.Errorf("Condition status after update = %q, want %q", p.Status_.Conditions[0].Status, "False")
	}
	if p.Status_.Conditions[0].Reason != "ProjectNotReady" {
		t.Errorf("Condition reason after update = %q, want %q", p.Status_.Conditions[0].Reason, "ProjectNotReady")
	}
}

func TestProject_GetCondition(t *testing.T) {
	p := NewProject("test", "Test", EnvDevelopment, "")
	p.AddCondition("Ready", "True", "", "")

	c := p.GetCondition("Ready")
	if c == nil {
		t.Fatal("GetCondition('Ready') returned nil")
	}
	if c.Status != "True" {
		t.Errorf("Condition status = %q, want %q", c.Status, "True")
	}

	c = p.GetCondition("NonExistent")
	if c != nil {
		t.Error("GetCondition('NonExistent') should return nil")
	}
}

func TestProject_Touch(t *testing.T) {
	p := NewProject("test", "Test", EnvDevelopment, "")
	before := p.Status_.LastActivity

	time.Sleep(time.Millisecond)
	p.Touch()

	if p.Status_.LastActivity.Equal(before) {
		t.Error("Touch() should update LastActivity")
	}
}

// ── Controller Tests ────────────────────────────────────────────────────────

func setupProjectControllerTest(t *testing.T) (*ProjectController, *resource.Registry, *events.Bus) {
	t.Helper()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	reg := resource.NewRegistry(bus, log)

	// Register the Project kind.
	if err := reg.RegisterKind(resource.Kind{
		Name:       Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	pc := NewProjectController(reg, bus, log)
	return pc, reg, bus
}

func TestProjectController_Name(t *testing.T) {
	pc, _, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	if pc.Name() != "project" {
		t.Errorf("Name() = %q, want %q", pc.Name(), "project")
	}
	if pc.Kind() != Kind {
		t.Errorf("Kind() = %q, want %q", pc.Kind(), Kind)
	}
}

func TestProjectController_Health(t *testing.T) {
	pc, _, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	h := pc.Health()
	if h.Name != "project" {
		t.Errorf("Health().Name = %q, want %q", h.Name, "project")
	}
	if h.State != "stopped" {
		t.Errorf("Health().State before start = %q, want %q", h.State, "stopped")
	}
}

func TestProjectController_StartStop(t *testing.T) {
	pc, _, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := pc.Start(ctx); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}
	// Idempotent.
	if err := pc.Start(ctx); err != nil {
		t.Fatalf("second Start() returned error: %v", err)
	}

	if err := pc.Stop(ctx); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}
	// Idempotent.
	if err := pc.Stop(ctx); err != nil {
		t.Fatalf("second Stop() returned error: %v", err)
	}

	h := pc.Health()
	if h.State != "stopped" {
		t.Errorf("Health().State after stop = %q, want %q", h.State, "stopped")
	}
}

func TestProjectController_ReconcileCreate(t *testing.T) {
	pc, reg, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	// Create a project through the Resource Engine.
	p := NewProject("test-project", "Test Project", EnvDevelopment, "A project for testing")
	if err := reg.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	// Reconcile the created project.
	result := pc.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "test-project"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}
	if result.Requeue {
		t.Error("Reconcile() returned Requeue=true for successful reconciliation")
	}

	// Verify the project was updated.
	updated, err := reg.Get(Kind, "test-project")
	if err != nil {
		t.Fatal(err)
	}
	proj := updated.(*Project)
	if proj.Status_.Phase != PhaseActive {
		t.Errorf("Phase after reconcile = %q, want %q", proj.Status_.Phase, PhaseActive)
	}
	if proj.GetCondition("Initialized") == nil {
		t.Error("Initialized condition was not set")
	}
	if proj.GetCondition("Ready") == nil {
		t.Error("Ready condition was not set")
	}
}

func TestProjectController_ReconcileActive(t *testing.T) {
	pc, reg, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	// Create a project already in Active phase.
	p := NewProject("active-project", "Active Project", EnvProduction, "")
	p.Status_.Phase = PhaseActive
	if err := reg.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	result := pc.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "active-project"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	updated, _ := reg.Get(Kind, "active-project")
	proj := updated.(*Project)
	if proj.Status_.Phase != PhaseActive {
		t.Errorf("Phase = %q, want %q", proj.Status_.Phase, PhaseActive)
	}
	if proj.Status_.Health != HealthHealthy {
		t.Errorf("Health = %q, want %q", proj.Status_.Health, HealthHealthy)
	}
}

func TestProjectController_ReconcileArchived(t *testing.T) {
	pc, reg, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	p := NewProject("archived-project", "Archived", EnvDevelopment, "")
	p.Status_.Phase = PhaseArchived
	if err := reg.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	result := pc.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "archived-project"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	updated, _ := reg.Get(Kind, "archived-project")
	proj := updated.(*Project)
	if proj.Status_.Phase != PhaseArchived {
		t.Errorf("Phase = %q, want %q", proj.Status_.Phase, PhaseArchived)
	}
	c := proj.GetCondition("Ready")
	if c == nil || c.Status != "False" {
		t.Errorf("Ready condition should be False for archived projects, got %v", c)
	}
}

func TestProjectController_ReconcileDeleting(t *testing.T) {
	pc, reg, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	p := NewProject("deleting-project", "Deleting", EnvDevelopment, "")
	p.Status_.Phase = PhaseDeleting
	if err := reg.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	result := pc.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "deleting-project"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}
}

func TestProjectController_ReconcileDeletedProject(t *testing.T) {
	pc, _, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	// Reconcile a project that doesn't exist (was deleted).
	result := pc.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "nonexistent"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error for deleted project: %v", result.Err)
	}
	if result.Requeue {
		t.Error("Reconcile() returned Requeue=true for deleted project")
	}
}

func TestProjectController_ReconcileUnknownKind(t *testing.T) {
	pc, _, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	result := pc.Reconcile(controller.ReconcileRequest{Kind: "WrongKind", ID: "test"})
	if result.Err == nil {
		t.Error("Reconcile() should return error for unknown kind")
	}
}

func TestProjectController_ErrorsInterface(t *testing.T) {
	pc, _, bus := setupProjectControllerTest(t)
	defer bus.Stop()

	// Compile-time check.
	var _ controller.Controller = pc
	_ = pc
}

// ── Integration Tests ──────────────────────────────────────────────────────

func TestProjectIntegration_CreateAndReconcile(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()

	reg := resource.NewRegistry(bus, log)
	if err := reg.RegisterKind(resource.Kind{
		Name:       Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	pc := NewProjectController(reg, bus, log)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = pc.Start(ctx)
	defer func() { _ = pc.Stop(ctx) }()

	// Create a project via the Resource Engine.
	proj := NewProject("integration-test", "Integration Test", EnvProduction, "Integration test project")
	if err := reg.Create(context.Background(), proj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger reconciliation (simulates what Controller Runtime does).
	result := pc.Reconcile(controller.ReconcileRequest{Kind: Kind, ID: "integration-test"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	// Verify the project is now Active.
	updated, err := reg.Get(Kind, "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	p := updated.(*Project)
	if p.Status_.Phase != PhaseActive {
		t.Errorf("Phase = %q, want %q", p.Status_.Phase, PhaseActive)
	}
	if p.Spec_.Environment != EnvProduction {
		t.Errorf("Environment = %q, want %q", p.Spec_.Environment, EnvProduction)
	}
}

// ── Spec Defaults Tests ────────────────────────────────────────────────────

func TestDefaultProjectSettings(t *testing.T) {
	if DefaultProjectSettings["autoDeploy"] != "true" {
		t.Error("Default autoDeploy should be true")
	}
	if DefaultProjectSettings["monitoring"] != "basic" {
		t.Error("Default monitoring should be basic")
	}
}

func TestValidEnvironments(t *testing.T) {
	if !ValidEnvironments[EnvDevelopment] {
		t.Error("development should be a valid environment")
	}
	if !ValidEnvironments[EnvStaging] {
		t.Error("staging should be a valid environment")
	}
	if !ValidEnvironments[EnvProduction] {
		t.Error("production should be a valid environment")
	}
	if ValidEnvironments["invalid"] {
		t.Error("invalid should NOT be a valid environment")
	}
}

// ── Metadata / Labels Tests ────────────────────────────────────────────────

func TestProject_EnvironmentLabel(t *testing.T) {
	p := NewProject("test", "Test", EnvStaging, "Staging")
	p.EnsureDefaults()

	if p.Metadata_.Labels["environment"] != EnvStaging {
		t.Errorf("environment label = %q, want %q", p.Metadata_.Labels["environment"], EnvStaging)
	}
}

func TestProject_ResourceVersionIncremented(t *testing.T) {
	p := NewProject("test", "Test", EnvDevelopment, "")
	if p.Metadata_.ResourceVersion != 1 {
		t.Errorf("Initial ResourceVersion = %d, want 1", p.Metadata_.ResourceVersion)
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
