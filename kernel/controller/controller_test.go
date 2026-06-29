package controller

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Test Helpers ───────────────────────────────────────────────────────────

// testController is a simple controller for testing.
type testController struct {
	mu             sync.Mutex
	name           string
	kind           string
	reconcileCount int
	reconciled     map[string]int // id → count
	failOnID       string         // if set, Reconcile returns error for this ID
	requeueOnID    string         // if set, Reconcile returns Requeue for this ID
	healthState    string
}

func newTestController(name, kind string) *testController {
	return &testController{
		name:        name,
		kind:        kind,
		reconciled:  make(map[string]int),
		healthState: "running",
	}
}

func (tc *testController) Name() string { return tc.name }
func (tc *testController) Kind() string { return tc.kind }
func (tc *testController) Start(ctx interface{ Done() <-chan struct{} }) error { return nil }
func (tc *testController) Stop(ctx interface{ Done() <-chan struct{} }) error  { return nil }

func (tc *testController) Reconcile(req ReconcileRequest) ReconcileResult {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.reconcileCount++
	tc.reconciled[req.ID]++

	if tc.failOnID != "" && req.ID == tc.failOnID {
		return ReconcileResult{
			Requeue: false,
			Err:     errTestFailed{id: req.ID},
		}
	}
	if tc.requeueOnID != "" && req.ID == tc.requeueOnID {
		return ReconcileResultRequeue
	}
	return ReconcileResultSuccess
}

type errTestFailed struct{ id string }

func (e errTestFailed) Error() string { return "test failure for " + e.id }

func (tc *testController) Health() ControllerHealth {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return ControllerHealth{
		Name:    tc.name,
		Kind:    tc.kind,
		State:   tc.healthState,
		Message: "ok",
	}
}

// ── Test Setup ─────────────────────────────────────────────────────────────

func setupTest(t *testing.T) (*Manager, *resource.Registry, *events.Bus, *health.Manager) {
	t.Helper()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	healthMgr := health.NewManager(log)
	_ = healthMgr.Start(context.Background())
	reg := resource.NewRegistry(bus, log)
	err := reg.RegisterKind(resource.Kind{
		Name:       "TestResource",
		Namespaced: false,
		Versions:   []string{"v1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(reg, bus, healthMgr, log)
	return mgr, reg, bus, healthMgr
}

func teardownTest(t *testing.T, bus *events.Bus, healthMgr *health.Manager) {
	t.Helper()
	bus.Stop()
	_ = healthMgr.Stop(context.Background())
}

// ── Manager Tests ──────────────────────────────────────────────────────────

func TestManager_Register(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("test-ctrl", "TestResource")
	err := mgr.Register(ctrl)
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	got, ok := mgr.Get("test-ctrl")
	if !ok {
		t.Fatal("Get() returned false for registered controller")
	}
	if got.Name() != "test-ctrl" {
		t.Errorf("Get().Name() = %q, want %q", got.Name(), "test-ctrl")
	}
	if got.Kind() != "TestResource" {
		t.Errorf("Get().Kind() = %q, want %q", got.Kind(), "TestResource")
	}
}

func TestManager_Register_Duplicate(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("dup", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatalf("first Register() returned error: %v", err)
	}

	err := mgr.Register(ctrl)
	if err == nil {
		t.Fatal("expected error for duplicate register")
	}
	if _, ok := err.(*ErrControllerAlreadyRegistered); !ok {
		t.Errorf("expected ErrControllerAlreadyRegistered, got %T: %v", err, err)
	}
}

func TestManager_List(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl1 := newTestController("ctrl-1", "TestResource")
	ctrl2 := newTestController("ctrl-2", "TestResource")

	if err := mgr.Register(ctrl1); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Register(ctrl2); err != nil {
		t.Fatal(err)
	}

	names := mgr.ControllerNames()
	if len(names) != 2 {
		t.Errorf("ControllerNames() = %v, want 2 controllers", names)
	}

	allHealth := mgr.AllControllerHealth()
	if len(allHealth) != 2 {
		t.Errorf("AllControllerHealth() = %v, want 2 health snapshots", allHealth)
	}
}

func TestManager_RegisterAfterStart(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	// Start the manager.
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	ctrl := newTestController("late", "TestResource")
	err := mgr.Register(ctrl)
	if err == nil {
		t.Fatal("expected error registering after start")
	}
}

// ── Lifecycle Tests ────────────────────────────────────────────────────────

func TestManager_StartStop(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("lc-test", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}

	// Should be idempotent.
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("second Start() returned error: %v", err)
	}

	if err := mgr.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}

	// Should be idempotent.
	if err := mgr.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop() returned error: %v", err)
	}

	// After stop, health should be "stopped".
	h, ok := mgr.ControllerHealth("lc-test")
	if !ok {
		t.Fatal("ControllerHealth() returned false after stop")
	}
	if h.State != "stopped" {
		t.Errorf("health.State = %q after stop, want %q", h.State, "stopped")
	}
}

// ── Event-Driven Reconciliation ────────────────────────────────────────────

func TestManager_EventTriggersReconcile(t *testing.T) {
	mgr, reg, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("event-test", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	// Create a resource of the controller's kind.
	res := resource.NewGenericResource("TestResource", "test-1", "Test 1", nil, nil)
	if err := reg.Create(context.Background(), res); err != nil {
		t.Fatal(err)
	}

	// Give the event loop time to process.
	time.Sleep(150 * time.Millisecond)

	ctrl.mu.Lock()
	count := ctrl.reconciled["test-1"]
	ctrl.mu.Unlock()

	if count == 0 {
		t.Error("controller was not called for resource created event")
	}
}

func TestManager_EventFilterByKind(t *testing.T) {
	mgr, reg, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	// Register a controller for "TestResource".
	ctrl := newTestController("kind-filter", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	// Register a different kind with no controller.
	err := reg.RegisterKind(resource.Kind{
		Name:       "Unwatched",
		Namespaced: false,
		Versions:   []string{"v1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a resource of the unwatched kind — should not trigger reconcile.
	unwatched := resource.NewGenericResource("Unwatched", "unwatched-1", "Unwatched", nil, nil)
	if err := reg.Create(context.Background(), unwatched); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	ctrl.mu.Lock()
	total := ctrl.reconcileCount
	ctrl.mu.Unlock()

	if total != 0 {
		t.Errorf("controller called %d times for unwatched kind, want 0", total)
	}
}

func TestManager_DeleteEventTriggersReconcile(t *testing.T) {
	mgr, reg, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("delete-test", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	// Create then delete.
	res := resource.NewGenericResource("TestResource", "del-1", "Del 1", nil, nil)
	if err := reg.Create(context.Background(), res); err != nil {
		t.Fatal(err)
	}
	if err := reg.Delete(context.Background(), "TestResource", "del-1"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)

	ctrl.mu.Lock()
	count, ok := ctrl.reconciled["del-1"]
	ctrl.mu.Unlock()

	if !ok {
		t.Error("controller was not called for delete event")
	}
	if count < 1 {
		t.Errorf("controller called %d times for del-1, want at least 1", count)
	}
}

// ── Reconcile Loop Tests ───────────────────────────────────────────────────

func TestReconcileLoop_ExistingResources(t *testing.T) {
	mgr, reg, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	// Create some resources before the controller is registered.
	for i := 0; i < 3; i++ {
		id := "pre-" + string(rune('a'+i))
		res := resource.NewGenericResource("TestResource", id, "Pre "+id, nil, nil)
		if err := reg.Create(context.Background(), res); err != nil {
			t.Fatal(err)
		}
	}

	ctrl := newTestController("existing-test", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	// Give the initial reconciliation time to process.
	time.Sleep(200 * time.Millisecond)

	ctrl.mu.Lock()
	count := ctrl.reconcileCount
	ctrl.mu.Unlock()

	if count < 3 {
		t.Errorf("controller called %d times, want at least 3 for 3 existing resources", count)
	}
}

// ── Backoff Strategy Tests ─────────────────────────────────────────────────

func TestBackoffStrategy_Delay(t *testing.T) {
	b := DefaultBackoff()
	tests := []struct {
		retry int
		want  time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1600 * time.Millisecond},
		{5, 3200 * time.Millisecond},
		{6, 6400 * time.Millisecond},
		{7, 12800 * time.Millisecond},
		{8, 25600 * time.Millisecond},
		{9, 51200 * time.Millisecond},
		{10, 60000 * time.Millisecond}, // capped at MaxDelay
		{100, 60000 * time.Millisecond}, // capped
	}
	for _, tt := range tests {
		got := b.Delay(tt.retry)
		if got != tt.want {
			t.Errorf("Delay(%d) = %v, want %v", tt.retry, got, tt.want)
		}
	}
}

func TestBackoffStrategy_NegativeRetry(t *testing.T) {
	b := DefaultBackoff()
	got := b.Delay(-1)
	want := 100 * time.Millisecond
	if got != want {
		t.Errorf("Delay(-1) = %v, want %v", got, want)
	}
}

// ── Namespace Controller Tests ─────────────────────────────────────────────

func TestNamespaceController_Name(t *testing.T) {
	nc := NewNamespaceController(nil, nil, logging.NewSubsystemLogger("test", logging.LevelDebug))
	if nc.Name() != "namespace" {
		t.Errorf("Name() = %q, want %q", nc.Name(), "namespace")
	}
	if nc.Kind() != "Namespace" {
		t.Errorf("Kind() = %q, want %q", nc.Kind(), "Namespace")
	}
}

func TestNamespaceController_Health(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()
	reg := resource.NewRegistry(bus, log)

	nc := NewNamespaceController(reg, bus, log)
	h := nc.Health()
	if h.Name != "namespace" {
		t.Errorf("Health().Name = %q, want %q", h.Name, "namespace")
	}
	if h.State != "stopped" {
		t.Errorf("Health().State = %q before start, want %q", h.State, "stopped")
	}
}

func TestNamespaceController_StartStop(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()
	reg := resource.NewRegistry(bus, log)

	nc := NewNamespaceController(reg, bus, log)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-canceled context for immediate stop

	if err := nc.Start(ctx); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}
	// Should be idempotent.
	if err := nc.Start(ctx); err != nil {
		t.Fatalf("second Start() returned error: %v", err)
	}

	if err := nc.Stop(ctx); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}
	// Should be idempotent.
	if err := nc.Stop(ctx); err != nil {
		t.Fatalf("second Stop() returned error: %v", err)
	}

	h := nc.Health()
	if h.State != "stopped" {
		t.Errorf("Health().State after stop = %q, want %q", h.State, "stopped")
	}
}

func TestNamespaceController_ReconcileDefaultNamespace(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()
	reg := resource.NewRegistry(bus, log)

	// Register the Namespace kind and create the default namespace.
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

	nc := NewNamespaceController(reg, bus, log)

	// Reconcile the default namespace.
	result := nc.Reconcile(ReconcileRequest{Kind: "Namespace", ID: "default"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}
	if result.Requeue {
		t.Error("Reconcile() returned Requeue=true for healthy namespace")
	}

	h := nc.Health()
	if h.State != "running" {
		t.Errorf("Health().State after reconcile = %q, want %q", h.State, "running")
	}
}

func TestNamespaceController_ReconcileMissingDefault(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()
	reg := resource.NewRegistry(bus, log)

	// Register the Namespace kind but do NOT create the default namespace.
	if err := reg.RegisterKind(resource.Kind{
		Name:       "Namespace",
		Namespaced: false,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	nc := NewNamespaceController(reg, bus, log)

	// Reconcile the default namespace — controller should recreate it.
	result := nc.Reconcile(ReconcileRequest{Kind: "Namespace", ID: "default"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}

	// Verify the default namespace was recreated.
	_, err := reg.Get("Namespace", "default")
	if err != nil {
		t.Errorf("default namespace was not recreated: %v", err)
	}
}

func TestNamespaceController_ReconcileNonDefaultDeleted(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()
	reg := resource.NewRegistry(bus, log)

	if err := reg.RegisterKind(resource.Kind{
		Name:       "Namespace",
		Namespaced: false,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	// Create a non-default namespace.
	ns := resource.NewNamespace("my-ns", "My Namespace", "Test namespace")
	if err := reg.Create(context.Background(), ns); err != nil {
		t.Fatal(err)
	}
	// Delete it.
	if err := reg.Delete(context.Background(), "Namespace", "my-ns"); err != nil {
		t.Fatal(err)
	}

	nc := NewNamespaceController(reg, bus, log)

	// Reconcile the deleted namespace — should succeed without recreating.
	result := nc.Reconcile(ReconcileRequest{Kind: "Namespace", ID: "my-ns"})
	if result.Err != nil {
		t.Errorf("Reconcile() returned error: %v", result.Err)
	}
	if result.Requeue {
		t.Error("Reconcile() returned Requeue=true for deleted non-default namespace")
	}
}

// ── Health Check Integration ───────────────────────────────────────────────

func TestManager_HealthCheckable(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("health-test", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	// The Manager should be registered in the health manager.
	registered := hm.Registered()
	found := false
	for _, name := range registered {
		if name == "controller.runtime" {
			found = true
			break
		}
	}
	if !found {
		t.Error("controller.runtime not registered in health manager")
	}
}

// ── Edge Cases ─────────────────────────────────────────────────────────────

func TestManager_StartStopEmpty(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	// Start with no controllers.
	if err := mgr.Start(context.Background()); err != nil {
		t.Fatalf("Start() with no controllers: %v", err)
	}
	if err := mgr.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() with no controllers: %v", err)
	}
}

func TestManager_GetNotFound(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	_, ok := mgr.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for nonexistent controller")
	}

	_, ok = mgr.ControllerHealth("nonexistent")
	if ok {
		t.Error("ControllerHealth() returned true for nonexistent controller")
	}
}

// ── Reconcile Results ──────────────────────────────────────────────────────

func TestReconcileHelpers(t *testing.T) {
	if ReconcileResultSuccess.Requeue {
		t.Error("ReconcileResultSuccess.Requeue should be false")
	}
	if !ReconcileResultRequeue.Requeue {
		t.Error("ReconcileResultRequeue.Requeue should be true")
	}
	after := RequeueAfter(5 * time.Second)
	if !after.Requeue {
		t.Error("RequeueAfter().Requeue should be true")
	}
	if after.RequeueAfter != 5*time.Second {
		t.Errorf("RequeueAfter().RequeueAfter = %v, want %v", after.RequeueAfter, 5*time.Second)
	}
	err := errTestFailed{id: "x"}
	withErr := RequeueWithError(err)
	if !withErr.Requeue {
		t.Error("RequeueWithError().Requeue should be true")
	}
	if withErr.Err == nil {
		t.Error("RequeueWithError().Err should not be nil")
	}
}

// ── Integration: Full Runtime with Resource Engine ─────────────────────────

func TestIntegration_ResourceEngineAndController(t *testing.T) {
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()

	healthMgr := health.NewManager(log)
	_ = healthMgr.Start(context.Background())
	defer func() {
		_ = healthMgr.Stop(context.Background())
	}()

	reg := resource.NewRegistry(bus, log)

	// Register a test resource kind.
	if err := reg.RegisterKind(resource.Kind{
		Name:       "IntegrationTest",
		Namespaced: false,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(reg, bus, healthMgr, log)
	ctrl := newTestController("integration-ctrl", "IntegrationTest")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	// Create a resource through the Resource Engine — controller should reconcile.
	res := resource.NewGenericResource("IntegrationTest", "it-1", "Integration 1", nil, nil)
	if err := reg.Create(context.Background(), res); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)

	ctrl.mu.Lock()
	count := ctrl.reconciled["it-1"]
	ctrl.mu.Unlock()

	if count == 0 {
		t.Error("controller was not called for integration test resource")
	}

	// Update the resource — controller should reconcile again.
	res2 := resource.NewGenericResource("IntegrationTest", "it-1", "Integration 1 Updated", nil, nil)
	if err := reg.Update(context.Background(), res2); err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)

	ctrl.mu.Lock()
	count = ctrl.reconciled["it-1"]
	ctrl.mu.Unlock()

	if count < 2 {
		t.Errorf("controller called %d times for it-1, want >= 2", count)
	}

	// Verify controller health.
	h, ok := mgr.ControllerHealth("integration-ctrl")
	if !ok {
		t.Fatal("ControllerHealth() returned false")
	}
	if h.State != "running" {
		t.Errorf("Health().State = %q, want %q", h.State, "running")
	}
}

// ── Verify types implement expected interfaces ─────────────────────────────

func TestTypes_ImplementInterfaces(t *testing.T) {
	// Compile-time interface checks — these will fail at compile time if
	// the types don't implement the interface.
	var _ Controller = (*testController)(nil)
	var _ Controller = (*NamespaceController)(nil)

	// Just verifying no panics.
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)
	_ = mgr
}

// ── Concurrency Test ───────────────────────────────────────────────────────

func TestManager_ConcurrentAccess(t *testing.T) {
	mgr, _, bus, hm := setupTest(t)
	defer teardownTest(t, bus, hm)

	ctrl := newTestController("concurrent", "TestResource")
	if err := mgr.Register(ctrl); err != nil {
		t.Fatal(err)
	}

	if err := mgr.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = mgr.Stop(context.Background())
	}()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mgr.ControllerNames()
			_, _ = mgr.Get("concurrent")
			_, _ = mgr.ControllerHealth("concurrent")
			_ = mgr.AllControllerHealth()
		}()
	}
	wg.Wait()
}

// ── NamespaceController: Ensure log is set ─────────────────────────────────

func TestNamespaceController_Logging(t *testing.T) {
	nc := NewNamespaceController(nil, nil, logging.NewSubsystemLogger("test", logging.LevelInfo))
	if nc.log == nil {
		t.Error("NewNamespaceController log is nil")
	}
}
