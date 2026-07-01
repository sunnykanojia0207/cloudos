package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/safe"
	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/types"
)

// ── Manager ────────────────────────────────────────────────────────────────

// Manager is the central controller runtime. It manages the lifecycle of all
// registered controllers, watches resource events from the Resource Engine,
// dispatches ReconcileRequests to the appropriate controller, and tracks
// controller health.
//
// Lifecycle:
//
//	NewManager(resRegistry, eventBus, log)
//	Register(myController)
//	Start(ctx)
//	  → subscribes to resource events
//	  → starts each controller's reconcile loop
//	  → registers for health checking
//	Stop(ctx)
//	  → stops all controllers
//	  → unsubscribes from events
type Manager struct {
	mu          sync.RWMutex
	controllers map[string]Controller

	// Per-controller health snapshots.
	health map[string]ControllerHealth

	resRegistry *resource.Registry
	eventBus    *events.Bus
	healthMgr   *health.Manager
	log         *logging.Logger
	backoff     BackoffStrategy

	// Work queue for reconcile requests. Shared by all controllers.
	workQueue chan ReconcileRequest

	// Lifecycle.
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewManager creates a new controller manager bound to the given resource
// registry and event bus.
func NewManager(resRegistry *resource.Registry, eventBus *events.Bus, healthMgr *health.Manager, log *logging.Logger) *Manager {
	log.Debug("initialising controller runtime")
	return &Manager{
		controllers:  make(map[string]Controller),
		health:       make(map[string]ControllerHealth),
		resRegistry:  resRegistry,
		eventBus:     eventBus,
		healthMgr:    healthMgr,
		log:          logging.NewSubsystemLogger("controllers", logging.LevelInfo),
		backoff:      DefaultBackoff(),
		workQueue:    make(chan ReconcileRequest, 4096),
	}
}

// ── Controller Registration ────────────────────────────────────────────────

// Register adds a controller to the manager. Returns an error if a controller
// with the same Name is already registered or if the manager is already running.
func (m *Manager) Register(ctrl Controller) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("controller runtime is running, cannot register new controllers")
	}

	name := ctrl.Name()
	if _, exists := m.controllers[name]; exists {
		return &ErrControllerAlreadyRegistered{Name: name}
	}

	m.controllers[name] = ctrl
	m.health[name] = ControllerHealth{
		Name:  name,
		Kind:  ctrl.Kind(),
		State: "stopped",
	}

	m.log.Info("controller registered",
		"name", name,
		"kind", ctrl.Kind(),
	)
	return nil
}

// ── Lifecycle ──────────────────────────────────────────────────────────────

// Start begins all registered controllers and subscribes to resource events.
// This method is non-blocking — controllers run in goroutines.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.mu.Unlock()

	m.log.Info("controller runtime starting",
		"controllers", len(m.controllers),
	)

	// Subscribe to resource lifecycle events.
	m.eventBus.Subscribe(resource.EventResourceCreated, m.handleResourceEvent)
	m.eventBus.Subscribe(resource.EventResourceUpdated, m.handleResourceEvent)
	m.eventBus.Subscribe(resource.EventResourceDeleted, m.handleResourceEvent)

	// Register for health checking.
	_ = m.healthMgr.Register("controller.runtime", m)

	// Start each controller's reconcile loop.
	m.mu.RLock()
	names := make([]string, 0, len(m.controllers))
	for n := range m.controllers {
		names = append(names, n)
	}
	m.mu.RUnlock()

	for _, name := range names {
		ctrl := m.controllers[name]
		m.startController(ctx, ctrl)
	}

	m.log.Info("controller runtime started")
	return nil
}

// startController starts a single controller's reconcile loop in a goroutine.
func (m *Manager) startController(ctx context.Context, ctrl Controller) {
	m.mu.Lock()
	m.health[ctrl.Name()] = ControllerHealth{
		Name:  ctrl.Name(),
		Kind:  ctrl.Kind(),
		State: "starting",
	}
	m.mu.Unlock()

	// Call the controller's Start() — most implementations are lightweight
	// and start internal goroutines if needed.
	if err := ctrl.Start(ctx); err != nil {
		m.log.Error("controller start failed",
			"name", ctrl.Name(),
			"error", err,
		)
	}

	m.mu.Lock()
	h := m.health[ctrl.Name()]
	h.State = "running"
	m.health[ctrl.Name()] = h
	m.mu.Unlock()

	// Run the reconcile loop in a goroutine with panic recovery.
	m.wg.Add(1)
	safe.Go(func() {
		defer m.wg.Done()

		// Immediate reconciliation of all existing resources on startup.
		m.reconcileExistingResources(ctx, ctrl)

		// Then enter the event-driven loop.
		m.reconcileLoop(ctx, ctrl)
	})

	m.log.Info("controller started",
		"name", ctrl.Name(),
		"kind", ctrl.Kind(),
	)
}

// Stop gracefully shuts down all controllers and unsubscribes from events.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}

	// Copy controller names to call Stop outside the lock.
	names := make([]string, 0, len(m.controllers))
	for n := range m.controllers {
		names = append(names, n)
	}
	m.mu.Unlock()

	// Unsubscribe from resource events.
	m.eventBus.Unsubscribe(resource.EventResourceCreated)
	m.eventBus.Unsubscribe(resource.EventResourceUpdated)
	m.eventBus.Unsubscribe(resource.EventResourceDeleted)

	// Stop each controller.
	for _, name := range names {
		if ctrl, ok := m.Get(name); ok {
			if err := ctrl.Stop(ctx); err != nil {
				m.log.Error("controller stop error",
					"name", name,
					"error", err,
				)
			}
		}
	}

	// Unregister from health checking.
	m.healthMgr.Unregister("controller.runtime")

	// Wait for all controller goroutines to finish.
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		m.log.Warn("controller runtime shutdown timed out")
	}

	m.mu.Lock()
	for name := range m.controllers {
		m.health[name] = ControllerHealth{
			Name:  name,
			State: "stopped",
		}
	}
	m.mu.Unlock()

	m.log.Info("controller runtime stopped")
	return nil
}

// ── Event Handling ─────────────────────────────────────────────────────────

// handleResourceEvent is the callback for resource lifecycle events from the
// event bus. It enqueues a ReconcileRequest for any controller that owns
// the resource's kind.
func (m *Manager) handleResourceEvent(ctx context.Context, event events.Event) {
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		return
	}

	kind, _ := payload["kind"].(string)
	id, _ := payload["id"].(string)
	if kind == "" || id == "" {
		return
	}

	m.mu.RLock()
	hasController := false
	for _, ctrl := range m.controllers {
		if ctrl.Kind() == kind {
			hasController = true
			break
		}
	}
	m.mu.RUnlock()

	if !hasController {
		return
	}

	// Non-blocking enqueue.
	req := ReconcileRequest{Kind: kind, ID: id}
	select {
	case m.workQueue <- req:
	default:
		m.log.Warn("work queue full, dropping reconcile request",
			"kind", kind, "id", id,
		)
	}
}

// ── Accessors ──────────────────────────────────────────────────────────────

// Get returns a registered controller by name. Returns nil and false if
// not found.
func (m *Manager) Get(name string) (Controller, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ctrl, ok := m.controllers[name]
	return ctrl, ok
}

// List returns all registered controllers.
func (m *Manager) List() []Controller {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Controller, 0, len(m.controllers))
	for _, ctrl := range m.controllers {
		out = append(out, ctrl)
	}
	return out
}

// ControllerNames returns the names of all registered controllers.
func (m *Manager) ControllerNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.controllers))
	for n := range m.controllers {
		names = append(names, n)
	}
	return names
}

// ControllerHealth returns the health snapshot for a single controller.
func (m *Manager) ControllerHealth(name string) (ControllerHealth, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.health[name]
	return h, ok
}

// AllControllerHealth returns health snapshots for all controllers.
func (m *Manager) AllControllerHealth() []ControllerHealth {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]ControllerHealth, 0, len(m.health))
	for _, h := range m.health {
		out = append(out, h)
	}
	return out
}

// ── Internal ───────────────────────────────────────────────────────────────

// reconcileExistingResources enqueues reconcile requests for every existing
// resource of the controller's kind. This ensures all resources are reconciled
// on startup even if no events have been emitted.
func (m *Manager) reconcileExistingResources(ctx context.Context, ctrl Controller) {
	kind := ctrl.Kind()
	resources, err := m.resRegistry.List(kind)
	if err != nil {
		m.log.Warn("cannot list existing resources for controller",
			"controller", ctrl.Name(),
			"kind", kind,
			"error", err,
		)
		return
	}

	for _, res := range resources {
		req := ReconcileRequest{
			Kind: kind,
			ID:   res.GetMetadata().ID,
		}
		m.enqueue(req)
	}

	if len(resources) > 0 {
		m.log.Info("initial reconciliation enqueued",
			"controller", ctrl.Name(),
			"kind", kind,
			"count", len(resources),
		)
	}
}

// enqueue adds a request to the work queue (non-blocking).
func (m *Manager) enqueue(req ReconcileRequest) {
	select {
	case m.workQueue <- req:
	default:
		m.log.Warn("work queue full, dropping request",
			"kind", req.Kind, "id", req.ID,
		)
	}
}

// updateControllerHealth updates the health snapshot for a controller after
// a reconcile attempt.
func (m *Manager) updateControllerHealth(name string, fn func(h *ControllerHealth)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.health[name]
	if !ok {
		return
	}
	fn(&h)
	m.health[name] = h
}

// retryKey builds a unique key for retry tracking.
func retryKey(kind, id string) string {
	return kind + "/" + id
}

// CheckHealth implements health.Checkable for the controller runtime.
func (m *Manager) CheckHealth(ctx context.Context) health.Report {
	allHealth := m.AllControllerHealth()

	state := "running"
	message := "all controllers healthy"
	for _, h := range allHealth {
		if h.State == "failed" {
			state = "degraded"
			message = "controller " + h.Name + " is in failed state"
			break
		}
	}

	return health.Report{
		State:     healthStateToResource(state),
		Message:   message,
		Timestamp: time.Now(),
	}
}

// healthStateToResource maps controller state strings to resource states.
func healthStateToResource(s string) types.ResourceState {
	switch s {
	case "running":
		return types.StateRunning
	case "stopped":
		return types.StateStopped
	case "failed":
		return types.StateFailed
	default:
		return types.StateUnknown
	}
}
