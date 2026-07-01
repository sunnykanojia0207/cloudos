package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── ApplicationController ──────────────────────────────────────────────────

// ApplicationController manages the lifecycle of CloudOS Applications. It is
// registered with the Controller Runtime at boot and reconciles Application
// resources in response to create, update, and delete events.
//
// Responsibilities:
//   - Validate application configuration (delegates to Application.Validate())
//   - Ensure default settings and defaults are applied
//   - Generate deployment workflows and submit them via the Workflow Service
//   - Maintain application status (phase, health, conditions, timestamps)
//   - Publish lifecycle events for audit, AI, and dashboard consumption
//
// Key architectural principle: the controller does NOT deploy directly.
// Instead, it creates a WorkflowDefinition and submits it through the
// Workflow Service. Every deployment becomes a WorkflowExecution Resource.
//
//	Application Controller
//	    │
//	    ├─ validate app spec
//	    ├─ BuildDeploymentPlan → WorkflowDefinition
//	    ├─ workflowService.Submit(def)
//	    └─ update Application status
type ApplicationController struct {
	mu       sync.RWMutex
	name     string
	kind     string

	reg      *resource.Registry
	eventBus *events.Bus
	log      *logging.Logger
	running  bool
	stop     context.CancelFunc

	// Workflow Service reference. The controller creates deployment workflows
	// through this interface rather than calling the Engine directly.
	workflowService WorkflowService

	// Health tracking.
	health controller.ControllerHealth
}

// WorkflowService is the subset of the Workflow Service interface that the
// Application Controller needs. This keeps the controller testable without
// requiring a full Workflow Engine.
//
// The Application Controller builds deployment workflows using workflow.BuildDefinition()
// (converting DeploymentSteps to PlanNodes) and then submits them through this interface.
type WorkflowService interface {
	RegisterDefinition(def *workflow.WorkflowDefinition) error
	Submit(def *workflow.WorkflowDefinition) (*workflow.WorkflowRun, error)
	GetExecution(runID string) (*workflow.WorkflowExecution, error)
}

// NewApplicationController creates a new ApplicationController bound to the
// given resource registry, event bus, and workflow service.
func NewApplicationController(
	reg *resource.Registry,
	eventBus *events.Bus,
	workflowSvc WorkflowService,
	log *logging.Logger,
) *ApplicationController {
	return &ApplicationController{
		name:     "application",
		kind:     Kind,
		reg:      reg,
		eventBus: eventBus,
		workflowService: workflowSvc,
		log:      logging.NewSubsystemLogger("application-controller", logging.LevelInfo),
		health: controller.ControllerHealth{
			Name:  "application",
			Kind:  Kind,
			State: "stopped",
		},
	}
}

// ── Controller Interface ───────────────────────────────────────────────────

// Name returns "application".
func (ac *ApplicationController) Name() string { return ac.name }

// Kind returns "Application".
func (ac *ApplicationController) Kind() string { return ac.kind }

// Start begins the application controller's background work.
func (ac *ApplicationController) Start(ctx interface{ Done() <-chan struct{} }) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.running {
		return nil
	}
	ac.running = true
	c, cancel := context.WithCancel(context.Background())
	ac.stop = cancel

	// Start periodic reconciliation every 5 minutes.
	go ac.periodicReconcile(c)

	ac.log.Info("application controller started")
	return nil
}

// Stop shuts down the application controller.
func (ac *ApplicationController) Stop(ctx interface{ Done() <-chan struct{} }) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if !ac.running {
		return nil
	}
	ac.running = false
	if ac.stop != nil {
		ac.stop()
	}
	ac.health.State = "stopped"
	ac.health.Message = "controller stopped"

	ac.log.Info("application controller stopped")
	return nil
}

// Reconcile is called by the Controller Runtime when an Application resource
// event occurs (created, updated, deleted).
func (ac *ApplicationController) Reconcile(req controller.ReconcileRequest) controller.ReconcileResult {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.log.Debug("reconciling application", "id", req.ID)

	switch req.Kind {
	case Kind:
		return ac.reconcileApplication(req.ID)
	default:
		return controller.ReconcileResult{
			Requeue: false,
			Err:     fmt.Errorf("unknown kind %q for application controller", req.Kind),
		}
	}
}

// Health returns the controller's current health status.
func (ac *ApplicationController) Health() controller.ControllerHealth {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.health
}

// ── Reconciliation ─────────────────────────────────────────────────────────

// reconcileApplication handles the reconciliation of a single Application resource.
func (ac *ApplicationController) reconcileApplication(id string) controller.ReconcileResult {
	// Fetch the application resource.
	appResource, err := ac.reg.Get(Kind, id)
	if err != nil {
		// Resource was deleted — acknowledge and publish event.
		ac.publishEvent("application.reconciled.deleted", id, nil)
		ac.health.State = "running"
		ac.health.Message = fmt.Sprintf("application %q deleted, acknowledged", id)
		return controller.ReconcileResultSuccess
	}

	app, ok := appResource.(*Application)
	if !ok {
		return ac.reconcileGenericApplication(appResource)
	}

	return ac.reconcileTypedApplication(app)
}

// reconcileTypedApplication handles a typed Application resource.
func (ac *ApplicationController) reconcileTypedApplication(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID
	ac.log.Debug("reconciling typed application", "id", id, "phase", app.Status_.Phase, "runtime", app.Spec_.Runtime.Type)

	// Ensure defaults are applied.
	app.EnsureDefaults()

	switch app.Status_.Phase {
	case PhaseCreating:
		return ac.handleCreatingPhase(app)
	case PhaseDeploying:
		return ac.handleDeployingPhase(app)
	case PhaseRunning:
		return ac.handleRunningPhase(app)
	case PhaseStopped:
		return ac.handleStoppedPhase(app)
	case PhaseFailed:
		return ac.handleFailedPhase(app)
	case PhaseDeleting:
		return ac.handleDeletingPhase(app)
	default:
		// Unknown phase — set to Creating and requeue.
		app.Status_.Phase = PhaseCreating
		if err := ac.saveApplication(app); err != nil {
			return controller.RequeueWithError(err)
		}
		return controller.RequeueAfter(1 * time.Second)
	}
}

// reconcileGenericApplication handles an Application stored as GenericResource.
func (ac *ApplicationController) reconcileGenericApplication(res resource.Resource) controller.ReconcileResult {
	id := res.GetMetadata().ID
	ac.log.Debug("reconciling generic application", "id", id)

	// Create a typed Application from the generic resource.
	app := &Application{
		Metadata_: res.GetMetadata(),
		Spec_: ApplicationSpec{
			Source: ApplicationSource{
				Type: SourceLocal,
			},
			Runtime: ApplicationRuntime{
				Type: RuntimeNode,
			},
			Settings: copyMap(DefaultApplicationSettings),
		},
		Status_: ApplicationStatus{
			Phase:  PhaseCreating,
			Health: HealthHealthy,
		},
	}
	app.EnsureDefaults()

	// Update the registry with the typed version.
	if err := ac.saveApplication(app); err != nil {
		return controller.RequeueWithError(err)
	}

	return controller.ReconcileResultSuccess
}

// ── Phase Handlers ─────────────────────────────────────────────────────────

// handleCreatingPhase initializes a newly created application by starting
// the first deployment workflow.
func (ac *ApplicationController) handleCreatingPhase(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID
	ac.log.Info("initializing application", "id", id, "runtime", app.Spec_.Runtime.Type)

	// Apply defaults.
	app.EnsureDefaults()

	// Add initialized condition.
	app.AddCondition("Created", "True", "ApplicationCreated", "Application resource has been created")

	// Transition to Deploying phase.
	app.Status_.Phase = PhaseDeploying
	app.AddCondition("Deploying", "True", "DeploymentStarted", "First deployment workflow started")
	app.Touch()

	if err := ac.saveApplication(app); err != nil {
		return controller.RequeueWithError(err)
	}

	// Start the deployment workflow.
	result := ac.startDeploymentWorkflow(app)
	if result.Err != nil {
		return result
	}

	ac.publishEvent("application.created", id, app)

	ac.health.State = "running"
	ac.health.Message = fmt.Sprintf("application %q initialized, deployment started", id)

	// Requeue to check deployment status after a short delay.
	return controller.RequeueAfter(3 * time.Second)
}

// handleDeployingPhase checks the status of an ongoing deployment workflow.
// It queries the WorkflowExecution resource to determine if the deployment
// has completed, failed, or is still running.
func (ac *ApplicationController) handleDeployingPhase(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID
	deployID := app.Status_.CurrentDeploymentID
	if deployID == "" {
		ac.log.Warn("no deployment ID set, transitioning to failed", "id", id)
		app.Status_.Phase = PhaseFailed
		app.AddCondition("Failed", "True", "NoDeploymentID", "No deployment workflow ID was set")
		_ = ac.saveApplication(app)
		return controller.ReconcileResultSuccess
	}

	ac.log.Debug("checking deployment status", "id", id, "deployment_id", deployID)

	// Query the WorkflowExecution resource.
	exec, err := ac.workflowService.GetExecution(deployID)
	if err != nil {
		// Execution not found yet (might still be being created) — requeue.
		ac.log.Debug("deployment execution not yet available, requeueing", "id", id, "error", err)
		return controller.RequeueAfter(1 * time.Second)
	}

	execStatus := exec.GetStatus().(*workflow.WorkflowExecutionStatus)

	switch execStatus.Phase {
	case workflow.WorkflowPending, workflow.WorkflowRunning:
		// Still in progress — check back later.
		ac.log.Debug("deployment still in progress", "id", id, "phase", execStatus.Phase)
		return controller.RequeueAfter(2 * time.Second)

	case workflow.WorkflowCompleted:
		// Deployment succeeded — extract URL from the deploy node result.
		app.Status_.Phase = PhaseRunning
		app.Status_.Health = HealthHealthy
		app.Status_.DeploymentCount++
		app.Status_.LastDeploymentTime = time.Now()

		// Extract the deploy URL from the "deploy" node result.
		deployURL := ""
		buildpackName := ""
		for _, nr := range execStatus.NodeResults {
			if nr.Action == "provider.deploy" && nr.Result != "" {
				if url := extractDeployURL(nr.Result); url != "" {
					app.Status_.URL = url
					deployURL = url
					ac.log.Info("extracted deploy URL", "id", id, "url", url)
				}
				break
			}
			if nr.Action == "build.execute" && nr.Result != "" {
				// Buildpack name is embedded in the build result.
				buildpackName = extractBuildpackName(nr.Result)
			}
		}

		// Build and record the deployment report.
		report := &DeploymentReport{
			DeploymentNumber: app.Status_.DeploymentCount,
			StartedAt:        parseStartedAt(execStatus.StartedAt),
			CompletedAt:      time.Now(),
			Duration:         fmtDuration(time.Now().Sub(parseStartedAt(execStatus.StartedAt))),
			Repository:       app.Spec_.Source.URL,
			Branch:           app.Spec_.Source.Branch,
			DetectedRuntime:  app.Spec_.Runtime.Type,
			Buildpack:        buildpackName,
			BuildSuccess:     true,
			RuntimeName:      extractRuntimeName(app),
			RuntimeVersion:   "runtime.cloudos.io/v1",
			WorkflowID:       deployID,
			WorkflowSteps:    len(execStatus.NodeResults),
			HealthStatus:     string(HealthHealthy),
			Endpoint:         deployURL,
			Environment:      extractEnvironment(app),
		}
		ac.recordDeploymentReport(app, report)

		app.AddCondition("Running", "True", "ApplicationDeployed", "Application is deployed and running")
		app.Touch()

		if err := ac.saveApplication(app); err != nil {
			return controller.RequeueWithError(err)
		}

		ac.publishEvent("application.deployed", id, app)
		ac.health.State = "running"
		ac.health.Message = fmt.Sprintf("application %q deployed successfully", id)
		ac.log.Info("application deployment completed", "id", id, "url", app.Status_.URL)
		return controller.ReconcileResultSuccess

	case workflow.WorkflowFailed, workflow.WorkflowCancelled:
		// Deployment failed.
		errMsg := execStatus.Error
		if errMsg == "" {
			errMsg = "deployment workflow " + string(execStatus.Phase)
		}
		ac.log.Warn("deployment failed", "id", id, "error", errMsg)

		// Build and record a failure report.
		report := &DeploymentReport{
			DeploymentNumber: app.Status_.DeploymentCount + 1,
			StartedAt:        parseStartedAt(execStatus.StartedAt),
			CompletedAt:      time.Now(),
			Duration:         fmtDuration(time.Now().Sub(parseStartedAt(execStatus.StartedAt))),
			Repository:       app.Spec_.Source.URL,
			Branch:           app.Spec_.Source.Branch,
			BuildSuccess:     false,
			WorkflowID:       deployID,
			WorkflowSteps:    len(execStatus.NodeResults),
			HealthStatus:     string(HealthError),
			Endpoint:         "",
			Environment:      extractEnvironment(app),
			Errors:           []string{errMsg},
		}
		ac.recordDeploymentReport(app, report)

		app.Status_.Phase = PhaseFailed
		app.AddCondition("Failed", "True", "DeploymentFailed", errMsg)

		app.Touch()
		_ = ac.saveApplication(app)
		ac.publishEvent("application.failed", id, app)
		return controller.ReconcileResultSuccess

	default:
		// Unknown phase — requeue.
		ac.log.Debug("unknown deployment phase, requeueing", "id", id, "phase", execStatus.Phase)
		return controller.RequeueAfter(2 * time.Second)
	}
}

// parseStartedAt parses an RFC3339 timestamp string, returning the zero
// time if the string is empty or malformed.
func parseStartedAt(s string) time.Time {
	if s == "" {
		return time.Now().Add(-5 * time.Second)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now().Add(-5 * time.Second)
	}
	return t
}

// recordDeploymentReport stores a deployment report in the application's
// deployment history. It prepends the new report, caps the history at
// DeploymentHistoryMax, and updates LastReport to point to the latest entry.
func (ac *ApplicationController) recordDeploymentReport(app *Application, report *DeploymentReport) {
	// Prepend to history.
	app.Status_.DeploymentHistory = append([]DeploymentReport{*report}, app.Status_.DeploymentHistory...)

	// Cap at max entries.
	if len(app.Status_.DeploymentHistory) > DeploymentHistoryMax {
		app.Status_.DeploymentHistory = app.Status_.DeploymentHistory[:DeploymentHistoryMax]
	}

	// Update LastReport convenience pointer.
	reportCopy := app.Status_.DeploymentHistory[0]
	app.Status_.LastReport = &reportCopy
}

// extractEnvironment returns a human-readable environment name from the
// application's spec, labels, or runtime type.
func extractEnvironment(app *Application) string {
	if app == nil {
		return "unknown"
	}

	// Check labels for explicit environment annotation.
	if env, ok := app.Metadata_.Labels["environment"]; ok && env != "" {
		return env
	}

	// Infer from runtime type.
	rt := app.Spec_.Runtime.Type
	switch rt {
	case RuntimeDocker:
		return "docker"
	default:
		return "local"
	}
}

// extractBuildpackName attempts to extract the buildpack name from a
// build node result string. Returns empty string if not found.
func extractBuildpackName(result string) string {
	// For now, infer from common patterns.
	// In the future, the executor will pass structured metadata.
	if result == "" {
		return ""
	}
	return "inferred"
}

// extractRuntimeName returns a human-readable runtime name from the
// application's spec and labels.
func extractRuntimeName(app *Application) string {
	if app == nil {
		return ""
	}
	rt := app.Spec_.Runtime.Type
	switch rt {
	case RuntimeGo:
		return "Go Runtime"
	case RuntimeNode:
		return "Node.js Runtime"
	case RuntimePython:
		return "Python Runtime"
	case RuntimeStatic:
		return "Static Runtime"
	case RuntimeNextJS:
		return "Next.js Runtime"
	case RuntimeLaravel:
		return "Laravel Runtime"
	case RuntimeDocker:
		return "Docker Runtime"
	default:
		return rt + " Runtime"
	}
}

// fmtDuration formats a time.Duration into a human-readable string.
// Examples: "8.2s", "1m 23s", "2h 15m"
func fmtDuration(d time.Duration) string {
	d = d.Round(100 * time.Millisecond)
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := d.Seconds() - float64(hours*3600) - float64(minutes*60)

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, int(seconds))
	}
	return fmt.Sprintf("%.1fs", seconds)
}

// extractDeployURL parses the URL from a deploy node result string.
// The result is formatted as: "Running at <url> (pid=..., port=...)"
func extractDeployURL(result string) string {
	// Look for "Running at " prefix
	const prefix = "Running at "
	if !strings.HasPrefix(result, prefix) {
		return ""
	}
	rest := result[len(prefix):]
	// Find the space before "(pid="
	if idx := strings.Index(rest, " (pid="); idx > 0 {
		return rest[:idx]
	}
	// Fallback: take everything before the first space
	if idx := strings.Index(rest, " "); idx > 0 {
		return rest[:idx]
	}
	return rest
}

// handleRunningPhase maintains an active application's status.
func (ac *ApplicationController) handleRunningPhase(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID

	// Check health (placeholder — will integrate with actual health checks).
	app.AddCondition("Running", "True", "ApplicationHealthy", "Application is running")
	app.Status_.Health = HealthHealthy
	app.Touch()

	// Only save if there are changes.
	if err := ac.saveApplication(app); err != nil {
		return controller.RequeueWithError(err)
	}

	ac.health.State = "running"
	ac.health.Message = fmt.Sprintf("application %q running healthy", id)
	return controller.ReconcileResultSuccess
}

// handleStoppedPhase ensures a stopped application remains stable.
func (ac *ApplicationController) handleStoppedPhase(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID
	app.Status_.Health = HealthHealthy
	app.AddCondition("Stopped", "True", "ApplicationStopped", "Application is stopped")
	app.Touch()

	ac.publishEvent("application.stopped", id, app)

	ac.health.State = "running"
	ac.health.Message = fmt.Sprintf("application %q stopped", id)
	return controller.ReconcileResultSuccess
}

// handleFailedPhase handles an application in failed state.
func (ac *ApplicationController) handleFailedPhase(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID
	app.Status_.Health = HealthError
	app.AddCondition("Failed", "True", "DeploymentFailed", "Application deployment or health check failed")
	app.Touch()

	ac.health.State = "running"
	ac.health.Message = fmt.Sprintf("application %q in failed state", id)
	return controller.ReconcileResultSuccess
}

// handleDeletingPhase handles application deletion cleanup.
func (ac *ApplicationController) handleDeletingPhase(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID
	ac.log.Info("cleaning up application", "id", id)

	// In a full implementation, this would:
	//   1. Cancel any active deployment workflows
	//   2. Stop the running application instance
	//   3. Clean up associated resources

	ac.publishEvent("application.deleting", id, app)
	ac.health.State = "running"
	ac.health.Message = fmt.Sprintf("application %q deletion acknowledged", id)

	return controller.ReconcileResultSuccess
}

// ── Deployment Workflow ────────────────────────────────────────────────────

// startDeploymentWorkflow builds a WorkflowDefinition from the Application's
// deployment plan and submits it through the Workflow Service.
//
// This is the key architectural integration: the controller does NOT deploy
// directly. It generates a workflow and submits it. Every deployment becomes
// a WorkflowExecution Resource with full retry, replay, and history support.
//
// Build pipeline:
//
//	DeploymentStep[] → workflow.PlanNode[] → workflow.BuildDefinition()
//	 → *workflow.WorkflowDefinition → workflow.Service.RegisterDefinition()
//	 → workflow.Service.Submit() → *workflow.WorkflowRun
func (ac *ApplicationController) startDeploymentWorkflow(app *Application) controller.ReconcileResult {
	id := app.GetMetadata().ID

	// Build the deployment plan (application-specific steps).
	steps := BuildDeploymentPlan(app)

	// Convert deployment steps to workflow PlanNodes.
	planNodes := make([]workflow.PlanNode, 0, len(steps))
	for _, step := range steps {
		planNodes = append(planNodes, workflow.PlanNode{
			ID:        step.ID,
			Name:      step.Name,
			Action:    step.Action,
			Target:    step.Target,
			DependsOn: step.DependsOn,
		})
	}

	// Build the WorkflowDefinition from PlanNodes.
	defID := fmt.Sprintf("deploy-%s-%d", id, time.Now().UnixNano())
	defName := fmt.Sprintf("Deploy %s", app.Metadata_.Name)
	def, err := workflow.BuildDefinition(defID, defName, planNodes)
	if err != nil {
		ac.log.Error("failed to build deployment definition", "error", err)
		app.Status_.Phase = PhaseFailed
		app.AddCondition("Failed", "True", "DefinitionFailed", fmt.Sprintf("Failed to build deployment workflow: %v", err))
		_ = ac.saveApplication(app)
		return controller.RequeueWithError(fmt.Errorf("build deployment definition: %w", err))
	}

	// Register the definition with the Workflow Service.
	if err := ac.workflowService.RegisterDefinition(def); err != nil {
		ac.log.Error("failed to register deployment definition", "error", err)
		app.Status_.Phase = PhaseFailed
		app.AddCondition("Failed", "True", "RegistrationFailed", fmt.Sprintf("Failed to register deployment workflow: %v", err))
		_ = ac.saveApplication(app)
		return controller.RequeueWithError(fmt.Errorf("register deployment definition: %w", err))
	}

	// Submit the definition for execution.
	run, err := ac.workflowService.Submit(def)
	if err != nil {
		ac.log.Error("failed to submit deployment workflow", "error", err)
		app.Status_.Phase = PhaseFailed
		app.AddCondition("Failed", "True", "SubmitFailed", fmt.Sprintf("Failed to submit deployment workflow: %v", err))
		_ = ac.saveApplication(app)
		return controller.RequeueWithError(fmt.Errorf("submit deployment workflow: %w", err))
	}

	// Store the deployment reference in the application status.
	app.Status_.CurrentDeploymentID = run.ID
	app.AddCondition("Deploying", "True", "WorkflowSubmitted", "Deployment workflow has been submitted")
	_ = ac.saveApplication(app)

	ac.log.Info("deployment workflow submitted",
		"application", id,
		"workflow", defID,
		"run_id", app.Status_.CurrentDeploymentID,
	)

	return controller.ReconcileResultSuccess
}

// ── Internal ───────────────────────────────────────────────────────────────

// saveApplication persists an Application through the Resource Engine.
func (ac *ApplicationController) saveApplication(app *Application) error {
	ctx := context.Background()

	existing, err := ac.reg.Get(Kind, app.GetMetadata().ID)
	if err == nil && existing != nil {
		app.Metadata_.CreatedAt = existing.GetMetadata().CreatedAt
		return ac.reg.Update(ctx, app)
	}
	return ac.reg.Create(ctx, app)
}

// publishEvent publishes an application lifecycle event through the event bus.
func (ac *ApplicationController) publishEvent(eventType, id string, app *Application) {
	if ac.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"kind": Kind,
		"id":   id,
	}
	if app != nil {
		payload["name"] = app.Metadata_.Name
		payload["runtime"] = app.Spec_.Runtime.Type
		payload["phase"] = app.Status_.Phase
		payload["source"] = app.Spec_.Source.Type
	}

	ac.eventBus.Publish(context.Background(), events.Event{
		Type:    eventType,
		Source:  "application-controller",
		Payload: payload,
	})
}

// ── Periodic Reconciliation ────────────────────────────────────────────────

// periodicReconcile periodically checks all applications for health (every 5 min).
func (ac *ApplicationController) periodicReconcile(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ac.reconcileAllApplications()
		case <-ctx.Done():
			return
		}
	}
}

// reconcileAllApplications iterates over all registered applications and triggers
// reconciliation for each.
func (ac *ApplicationController) reconcileAllApplications() {
	apps, err := ac.reg.List(Kind)
	if err != nil {
		ac.log.Error("cannot list applications for periodic reconcile", "error", err)
		return
	}
	for _, res := range apps {
		req := controller.ReconcileRequest{Kind: Kind, ID: res.GetMetadata().ID}
		_ = ac.Reconcile(req)
	}
	ac.log.Debug("periodic application reconciliation complete", "count", len(apps))
}
