package project

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── ProjectController ──────────────────────────────────────────────────────

// ProjectController manages the lifecycle of CloudOS Projects. It is
// registered with the Controller Runtime at boot and reconciles Project
// resources in response to create, update, and delete events.
//
// Responsibilities:
//   - Validate project configuration (delegates to Project.Validate())
//   - Ensure default settings and defaults are applied
//   - Maintain project status (phase, health, conditions, timestamps)
//   - Publish lifecycle events for audit, AI, and dashboard consumption
//   - Verify quota isn't exceeded (future: after quota system)
//   - Reconcile desired state with observed state
type ProjectController struct {
	mu       sync.RWMutex
	name     string
	kind     string

	reg      *resource.Registry
	eventBus *events.Bus
	log      *logging.Logger
	running  bool
	stop     context.CancelFunc

	// Health tracking.
	health controller.ControllerHealth
}

// NewProjectController creates a new ProjectController bound to the given
// resource registry and event bus.
func NewProjectController(reg *resource.Registry, eventBus *events.Bus, log *logging.Logger) *ProjectController {
	return &ProjectController{
		name:     "project",
		kind:     Kind,
		reg:      reg,
		eventBus: eventBus,
		log:      logging.NewSubsystemLogger("project-controller", logging.LevelInfo),
		health: controller.ControllerHealth{
			Name:  "project",
			Kind:  Kind,
			State: "stopped",
		},
	}
}

// ── Controller Interface ───────────────────────────────────────────────────

// Name returns "project".
func (pc *ProjectController) Name() string { return pc.name }

// Kind returns "Project".
func (pc *ProjectController) Kind() string { return pc.kind }

// Start begins the project controller's background work.
func (pc *ProjectController) Start(ctx interface{ Done() <-chan struct{} }) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.running {
		return nil
	}
	pc.running = true
	c, cancel := context.WithCancel(context.Background())
	pc.stop = cancel

	// Start periodic reconciliation every 10 minutes to verify project health.
	go pc.periodicReconcile(c)

	pc.log.Info("project controller started")
	return nil
}

// Stop shuts down the project controller.
func (pc *ProjectController) Stop(ctx interface{ Done() <-chan struct{} }) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if !pc.running {
		return nil
	}
	pc.running = false
	if pc.stop != nil {
		pc.stop()
	}
	pc.health.State = "stopped"
	pc.health.Message = "controller stopped"

	pc.log.Info("project controller stopped")
	return nil
}

// Reconcile is called by the Controller Runtime when a Project resource event
// occurs (created, updated, deleted).
func (pc *ProjectController) Reconcile(req controller.ReconcileRequest) controller.ReconcileResult {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.log.Debug("reconciling project", "id", req.ID)

	switch req.Kind {
	case Kind:
		return pc.reconcileProject(req.ID)
	default:
		return controller.ReconcileResult{
			Requeue: false,
			Err:     fmt.Errorf("unknown kind %q for project controller", req.Kind),
		}
	}
}

// Health returns the controller's current health status.
func (pc *ProjectController) Health() controller.ControllerHealth {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.health
}

// ── Reconciliation ─────────────────────────────────────────────────────────

// reconcileProject handles the reconciliation of a single Project resource.
func (pc *ProjectController) reconcileProject(id string) controller.ReconcileResult {
	// Fetch the project resource.
	projResource, err := pc.reg.Get(Kind, id)
	if err != nil {
		// Resource was deleted — nothing to do. The controller acknowledges
		// the deletion and publishes a cleanup event.
		pc.publishEvent("project.reconciled.deleted", id, nil)
		pc.health.State = "running"
		pc.health.Message = fmt.Sprintf("project %q deleted, acknowledged", id)
		return controller.ReconcileResultSuccess
	}

	proj, ok := projResource.(*Project)
	if !ok {
		// Fallback: if it's not a typed Project, try to handle generically.
		return pc.reconcileGenericProject(projResource)
	}

	return pc.reconcileTypedProject(proj)
}

// reconcileTypedProject handles a typed Project resource.
func (pc *ProjectController) reconcileTypedProject(proj *Project) controller.ReconcileResult {
	id := proj.GetMetadata().ID
	pc.log.Debug("reconciling typed project", "id", id, "phase", proj.Status_.Phase)

	// Ensure defaults are applied.
	proj.EnsureDefaults()

	switch proj.Status_.Phase {
	case PhaseCreating:
		return pc.handleCreatingPhase(proj)
	case PhaseActive:
		return pc.handleActivePhase(proj)
	case PhaseArchived:
		return pc.handleArchivedPhase(proj)
	case PhaseDeleting:
		return pc.handleDeletingPhase(proj)
	default:
		// Unknown phase — set to Creating and requeue.
		proj.Status_.Phase = PhaseCreating
		if err := pc.saveProject(proj); err != nil {
			return controller.RequeueWithError(err)
		}
		return controller.RequeueAfter(1 * time.Second)
	}
}

// reconcileGenericProject handles a Project stored as GenericResource.
func (pc *ProjectController) reconcileGenericProject(res resource.Resource) controller.ReconcileResult {
	id := res.GetMetadata().ID
	pc.log.Debug("reconciling generic project", "id", id)

	// Convert generic to typed.
	now := time.Now()
	proj := &Project{
		Metadata_: res.GetMetadata(),
		Spec_: ProjectSpec{
			DisplayName: res.GetMetadata().Name,
			Environment: EnvDevelopment,
			Settings:    copyMap(DefaultProjectSettings),
		},
		Status_: ProjectStatus{
			Phase:        PhaseActive,
			Health:       HealthHealthy,
			LastActivity: now,
		},
	}
	proj.EnsureDefaults()

	// Update the registry with the typed version.
	if err := pc.saveProject(proj); err != nil {
		return controller.RequeueWithError(err)
	}

	return controller.ReconcileResultSuccess
}

// ── Phase Handlers ─────────────────────────────────────────────────────────

// handleCreatingPhase initializes a newly created project.
func (pc *ProjectController) handleCreatingPhase(proj *Project) controller.ReconcileResult {
	id := proj.GetMetadata().ID
	pc.log.Info("initializing project", "id", id)

	// Apply defaults.
	proj.EnsureDefaults()

	// Mark as initialized.
	proj.AddCondition("Initialized", "True", "ProjectCreated", "Project has been initialized")
	proj.AddCondition("Ready", "True", "ProjectReady", "Project is ready for use")
	proj.Status_.Phase = PhaseActive
	proj.Status_.Health = HealthHealthy
	proj.Touch()

	if err := pc.saveProject(proj); err != nil {
		return controller.RequeueWithError(err)
	}

	pc.publishEvent("project.created", id, proj)

	pc.health.State = "running"
	pc.health.Message = fmt.Sprintf("project %q initialized", id)
	return controller.ReconcileResultSuccess
}

// handleActivePhase maintains an active project's status.
func (pc *ProjectController) handleActivePhase(proj *Project) controller.ReconcileResult {
	id := proj.GetMetadata().ID
	proj.Touch()

	// Update resource count — count all resources in this project's namespace.
	resourceCount := pc.countProjectResources(id)

	// Update health based on conditions.
	health := pc.computeHealth(proj)

	// Only save if something changed.
	if proj.Status_.ResourceCount != resourceCount || proj.Status_.Health != health {
		proj.Status_.ResourceCount = resourceCount
		proj.Status_.Health = health
		proj.AddCondition("Ready", "True", "ProjectHealthy", "Project is healthy")

		if err := pc.saveProject(proj); err != nil {
			return controller.RequeueWithError(err)
		}
		pc.publishEvent("project.updated", id, proj)
	}

	pc.health.State = "running"
	pc.health.Message = fmt.Sprintf("project %q reconciled (%d resources)", id, resourceCount)
	return controller.ReconcileResultSuccess
}

// handleArchivedPhase ensures an archived project remains stable.
func (pc *ProjectController) handleArchivedPhase(proj *Project) controller.ReconcileResult {
	id := proj.GetMetadata().ID
	proj.Status_.Health = HealthHealthy
	proj.AddCondition("Ready", "False", "ProjectArchived", "Project is archived and read-only")

	pc.publishEvent("project.archived", id, proj)

	pc.health.State = "running"
	pc.health.Message = fmt.Sprintf("project %q archived", id)
	return controller.ReconcileResultSuccess
}

// handleDeletingPhase handles project deletion cleanup.
func (pc *ProjectController) handleDeletingPhase(proj *Project) controller.ReconcileResult {
	id := proj.GetMetadata().ID
	pc.log.Info("cleaning up project", "id", id)

	// In a full implementation, this would cascade-delete all resources
	// belonging to the project. For now, we just acknowledge the deletion.
	pc.publishEvent("project.deleting", id, proj)
	pc.health.State = "running"
	pc.health.Message = fmt.Sprintf("project %q deletion acknowledged", id)

	// Return success — the Resource Engine will handle the actual deletion.
	return controller.ReconcileResultSuccess
}

// ── Internal ───────────────────────────────────────────────────────────────

// saveProject persists a Project through the Resource Engine.
func (pc *ProjectController) saveProject(proj *Project) error {
	ctx := context.Background()

	// Try update first; if not found, create.
	existing, err := pc.reg.Get(Kind, proj.GetMetadata().ID)
	if err == nil && existing != nil {
		// Preserve original creation timestamp.
		proj.Metadata_.CreatedAt = existing.GetMetadata().CreatedAt
		return pc.reg.Update(ctx, proj)
	}
	return pc.reg.Create(ctx, proj)
}

// countProjectResources counts resources associated with this project.
// Currently returns 0 as a placeholder; will be implemented when other
// resource types exist.
func (pc *ProjectController) countProjectResources(projectID string) int {
	// Future: iterate over all resources matching this project's namespace.
	// For now, return 0 since only Project resources exist.
	return 1 // The project itself counts as 1.
}

// computeHealth determines project health from its conditions.
func (pc *ProjectController) computeHealth(proj *Project) string {
	for _, c := range proj.Status_.Conditions {
		if c.Type == "Ready" && c.Status == "False" {
			if c.Reason == "ProjectArchived" {
				return HealthHealthy // Archived is healthy, just read-only.
			}
			return HealthDegraded
		}
		if c.Type == "Ready" && c.Status == "Unknown" {
			return HealthDegraded
		}
	}
	return HealthHealthy
}

// publishEvent publishes a project lifecycle event through the event bus.
func (pc *ProjectController) publishEvent(eventType, id string, proj *Project) {
	if pc.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"kind": Kind,
		"id":   id,
	}
	if proj != nil {
		payload["name"] = proj.Spec_.DisplayName
		payload["environment"] = proj.Spec_.Environment
		payload["phase"] = proj.Status_.Phase
	}

	pc.eventBus.Publish(context.Background(), events.Event{
		Type:    eventType,
		Source:  "project-controller",
		Payload: payload,
	})
}

// ── Periodic Reconciliation ────────────────────────────────────────────────

// periodicReconcile periodically checks all projects for health (every 10 min).
func (pc *ProjectController) periodicReconcile(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pc.reconcileAllProjects()
		case <-ctx.Done():
			return
		}
	}
}

// reconcileAllProjects iterates over all registered projects and triggers
// reconciliation for each. This is used by the periodic loop to ensure
// project status stays current even without resource events.
func (pc *ProjectController) reconcileAllProjects() {
	projects, err := pc.reg.List(Kind)
	if err != nil {
		pc.log.Error("cannot list projects for periodic reconcile", "error", err)
		return
	}
	for _, res := range projects {
		req := controller.ReconcileRequest{Kind: Kind, ID: res.GetMetadata().ID}
		_ = pc.Reconcile(req)
	}
	pc.log.Debug("periodic project reconciliation complete", "count", len(projects))
}
