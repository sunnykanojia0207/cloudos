package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── NamespaceController ─────────────────────────────────────────────────────

// NamespaceController manages the lifecycle of CloudOS namespaces.
//
// Responsibilities:
//   - Ensures the default namespace always exists.
//   - Validates namespace lifecycle (no deletion of "default").
//   - Updates namespace status with resource counts.
//   - Publishes reconciliation events.
//
// The NamespaceController is a built-in controller that is registered at
// kernel boot time. It demonstrates the Controller pattern and serves as
// the foundation for all namespace-scoped resources.
type NamespaceController struct {
	mu    sync.RWMutex
	name  string
	kind  string

	reg        *resource.Registry
	eventBus   *events.Bus
	log        *logging.Logger
	running    bool
	cancel     context.CancelFunc

	// Health tracking.
	health ControllerHealth
}

// NewNamespaceController creates a new NamespaceController bound to the
// given resource registry and event bus.
func NewNamespaceController(reg *resource.Registry, eventBus *events.Bus, log *logging.Logger) *NamespaceController {
	return &NamespaceController{
		name:     "namespace",
		kind:     "Namespace",
		reg:      reg,
		eventBus: eventBus,
		log:      logging.NewSubsystemLogger("namespace-controller", logging.LevelInfo),
		health: ControllerHealth{
			Name:  "namespace",
			Kind:  "Namespace",
			State: "stopped",
		},
	}
}

// ── Controller Interface ───────────────────────────────────────────────────

// Name returns "namespace".
func (nc *NamespaceController) Name() string { return nc.name }

// Kind returns "Namespace".
func (nc *NamespaceController) Kind() string { return nc.kind }

// Start begins the namespace controller's background work. This is a
// lightweight controller that does not start its own reconcile loop —
// the Manager's event-driven loop handles Reconcile calls.
func (nc *NamespaceController) Start(ctx interface{ Done() <-chan struct{} }) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if nc.running {
		return nil
	}
	nc.running = true
	c, cancel := context.WithCancel(context.Background())
	nc.cancel = cancel

	// Start a periodic reconciliation goroutine (every 5 minutes) to
	// update namespace resource counts and ensure the default namespace
	// still exists.
	go nc.periodicReconcile(c)

	nc.log.Info("namespace controller started")
	return nil
}

// Stop shuts down the namespace controller's background work.
func (nc *NamespaceController) Stop(ctx interface{ Done() <-chan struct{} }) error {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if !nc.running {
		return nil
	}
	nc.running = false
	if nc.cancel != nil {
		nc.cancel()
	}

	nc.health.State = "stopped"
	nc.health.Message = "controller stopped"

	nc.log.Info("namespace controller stopped")
	return nil
}

// Reconcile is called by the manager when a Namespace resource event occurs.
func (nc *NamespaceController) Reconcile(req ReconcileRequest) ReconcileResult {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	nc.log.Debug("reconciling namespace",
		"id", req.ID,
	)

	// Get the namespace resource.
	nsResource, err := nc.reg.Get(nc.kind, req.ID)
	if err != nil {
		// If the namespace was deleted and it's the default, recreate it.
		if req.ID == resource.NamespaceDefault {
			nc.log.Warn("default namespace was deleted, recreating")
			ns := resource.DefaultNamespace()
			if createErr := nc.reg.Create(context.Background(), ns); createErr != nil {
				return ReconcileResult{
					Requeue: true,
					Err:     fmt.Errorf("recreate default namespace: %w", createErr),
				}
			}
			nc.publishEvent("namespace.recreated", ns)

			nc.health.State = "running"
			nc.health.Message = "default namespace recreated"
			return ReconcileResultSuccess
		}
		// Non-default namespace deletion is allowed.
		nc.publishEvent("namespace.deleted", nil)
		return ReconcileResultSuccess
	}

	ns, ok := nsResource.(*resource.Namespace)
	if !ok {
		// If it's not a typed Namespace, use the generic interface.
		return nc.reconcileGenericNamespace(nsResource)
	}

	return nc.reconcileTypedNamespace(ns)
}

// reconcileTypedNamespace handles a typed Namespace resource.
func (nc *NamespaceController) reconcileTypedNamespace(ns *resource.Namespace) ReconcileResult {
	id := ns.GetMetadata().ID

	// Validate lifecycle: "default" namespace cannot be deleted.
	if ns.GetStatus() != nil {
		if status, ok := ns.GetStatus().(resource.NamespaceStatus); ok {
			if status.Phase == "Terminating" && id == resource.NamespaceDefault {
				// Prevent deletion of default namespace by setting it back to Active.
				ns.SetStatus(resource.NamespaceStatus{
					Phase: "Active",
				})
				if err := nc.reg.Update(context.Background(), ns); err != nil {
					return ReconcileResult{
						Requeue: true,
						Err:     fmt.Errorf("restore default namespace: %w", err),
					}
				}
				nc.publishEvent("namespace.protected", ns)
				nc.health.Message = "default namespace protected from deletion"
				return ReconcileResultSuccess
			}
		}
	}

	// Count resources in this namespace.
	count := nc.countResourcesInNamespace(id)

	// Update namespace status with resource count if it changed.
	if status, ok := ns.GetStatus().(resource.NamespaceStatus); ok {
		if status.ResourceCount != count {
			ns.SetStatus(resource.NamespaceStatus{
				Phase:         "Active",
				ResourceCount: count,
			})
			if err := nc.reg.Update(context.Background(), ns); err != nil {
				return ReconcileResult{
					Requeue: true,
					Err:     fmt.Errorf("update namespace status: %w", err),
				}
			}
			nc.publishEvent("namespace.status.updated", ns)
		}
	}

	nc.health.State = "running"
	nc.health.Message = fmt.Sprintf("namespace %q reconciled (%d resources)", id, count)
	return ReconcileResultSuccess
}

// reconcileGenericNamespace handles a Namespace stored as GenericResource.
func (nc *NamespaceController) reconcileGenericNamespace(res resource.Resource) ReconcileResult {
	_ = res.GetMetadata().ID
	// For generic namespaces, just ensure they exist.
	return ReconcileResultSuccess
}

// ── Periodic Reconciliation ────────────────────────────────────────────────

// periodicReconcile periodically checks controller health and ensures the
// default namespace exists (runs every 5 minutes).
func (nc *NamespaceController) periodicReconcile(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nc.ensureDefaultNamespace()
		case <-ctx.Done():
			return
		}
	}
}

// ensureDefaultNamespace checks that the default namespace exists and
// recreates it if missing.
func (nc *NamespaceController) ensureDefaultNamespace() {
	_, err := nc.reg.Get(nc.kind, resource.NamespaceDefault)
	if err != nil {
		nc.log.Warn("default namespace missing, recreating")
		ns := resource.DefaultNamespace()
		if createErr := nc.reg.Create(context.Background(), ns); createErr != nil {
			nc.log.Error("cannot recreate default namespace", "error", createErr)
			nc.health.State = "failed"
			nc.health.Message = "cannot recreate default namespace"
			return
		}
		nc.publishEvent("namespace.recreated", ns)
	}
}

// ── Health ──────────────────────────────────────────────────────────────────

// Health returns the controller's current health status.
func (nc *NamespaceController) Health() ControllerHealth {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.health
}

// ── Internal ───────────────────────────────────────────────────────────────

// publishEvent publishes a namespace-related event through the event bus.
func (nc *NamespaceController) publishEvent(eventType string, ns resource.Resource) {
	if nc.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"kind": nc.kind,
	}
	if ns != nil {
		payload["id"] = ns.GetMetadata().ID
		payload["name"] = ns.GetMetadata().Name
	} else {
		payload["id"] = "unknown"
	}

	nc.eventBus.Publish(context.Background(), events.Event{
		Type:    eventType,
		Source:  "namespace-controller",
		Payload: payload,
	})
}

// countResourcesInNamespace counts all resources in a given namespace.
// For now, it only counts the namespace itself.
func (nc *NamespaceController) countResourcesInNamespace(nsID string) int {
	count := 1 // The namespace itself counts as 1
	// In a future iteration, this will query the resource registry for
	// all resources with matching namespace metadata.
	return count
}
