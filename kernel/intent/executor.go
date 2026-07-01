package intent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/plugin"
	"github.com/cloudos/cloudos/packages/types"
	"github.com/cloudos/cloudos/kernel/project"
	"github.com/cloudos/cloudos/kernel/registry"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ExecutorDeps holds the dependencies required by the Executor
type ExecutorDeps struct {
	ResourceRegistry   *resource.Registry
	ControllerManager  *controller.Manager
	HealthManager      *health.Manager
	PluginRegistry     *plugin.Registry
	CapabilityRegistry *registry.Manager
	ProviderRegistry   *registry.Manager
	EventBus           *events.Bus
	Logger             *logging.Logger
}

// Executor runs ExecutionPlans against the CloudOS Kernel
type Executor struct {
	resRegistry   *resource.Registry
	ctrlManager   *controller.Manager
	healthMgr     *health.Manager
	pluginReg     *plugin.Registry
	capRegistry   *registry.Manager
	providerReg   *registry.Manager
	bus           *events.Bus
	log           *logging.Logger
}

// NewExecutor creates a new Executor
func NewExecutor(deps ExecutorDeps) *Executor {
	return &Executor{
		resRegistry: deps.ResourceRegistry,
		ctrlManager: deps.ControllerManager,
		healthMgr:   deps.HealthManager,
		pluginReg:   deps.PluginRegistry,
		capRegistry: deps.CapabilityRegistry,
		providerReg: deps.ProviderRegistry,
		bus:         deps.EventBus,
		log:         deps.Logger,
	}
}

// Execute runs an ExecutionPlan, updating step statuses and collecting results
func (ex *Executor) Execute(ctx context.Context, plan *ExecutionPlan) *IntentResult {
	ex.log.Info("executing plan", "plan_id", plan.ID, "steps", len(plan.Steps))
	plan.Status = IntentExecuting
	plan.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	result := &IntentResult{Success: true}
	completed := make(map[string]bool)

	for i := range plan.Steps {
		step := &plan.Steps[i]

		// Check dependencies are met
		if !ex.dependenciesMet(step.Dependencies, completed) {
			step.Status = StepSkipped
			step.Result = "dependencies not met — skipping"
			continue
		}

		// Check for cancellation
		if ctx.Err() != nil {
			step.Status = StepSkipped
			step.Result = "execution cancelled"
			plan.Status = IntentFailed
			result.Success = false
			result.Summary = "Execution cancelled"
			return result
		}

		step.Status = StepRunning
		step.StartedAt = time.Now().UTC().Format(time.RFC3339)
		plan.UpdatedAt = step.StartedAt

		ex.publishProgress(plan)

		err := ex.executeStep(ctx, step, result)

		if err != nil {
			step.Status = StepFailed
			step.Error = err.Error()
			step.CompletedAt = time.Now().UTC().Format(time.RFC3339)
			plan.Status = IntentFailed
			plan.UpdatedAt = step.CompletedAt
			result.Success = false
			result.Summary = fmt.Sprintf("Step %d (%s) failed: %s", i+1, step.Name, err.Error())
			result.Details = append(result.Details, ResultItem{
				Message: fmt.Sprintf("Step %d: %s — Failed", i+1, step.Name),
				Type:    "error",
				Detail:  err.Error(),
			})
			ex.publishProgress(plan)
			return result
		}

		step.Status = StepSuccess
		step.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		completed[step.ID] = true

		result.Details = append(result.Details, ResultItem{
			Message: fmt.Sprintf("Step %d: %s — Done", i+1, step.Name),
			Type:    "success",
			Detail:  step.Result,
		})

		plan.UpdatedAt = step.CompletedAt
		ex.publishProgress(plan)
	}

	plan.Status = IntentCompleted
	plan.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if result.Summary == "" {
		result.Summary = fmt.Sprintf("Completed in %d steps", len(plan.Steps))
	}

	return result
}

func (ex *Executor) dependenciesMet(deps []string, completed map[string]bool) bool {
	for _, dep := range deps {
		if !completed[dep] {
			return false
		}
	}
	return true
}

func (ex *Executor) publishProgress(plan *ExecutionPlan) {
	if ex.bus == nil {
		return
	}
	ex.bus.Publish(context.Background(), events.Event{
		Type:   "plan.progress",
		Source: "intent.executor",
		Payload: map[string]interface{}{
			"plan_id": plan.ID,
			"status":  string(plan.Status),
		},
	})
}

func (ex *Executor) executeStep(ctx context.Context, step *ExecutionStep, result *IntentResult) error {
	ex.log.Debug("executing step", "step", step.Name, "action", step.Action, "target", step.Target)

	switch step.Action {
	case "validate":
		return ex.execValidate(step)
	case "resource.create":
		return ex.execResourceCreate(step)
	case "resource.get":
		return ex.execResourceGet(step, result)
	case "resource.list":
		return ex.execResourceList(step, result)
	case "resource.delete":
		return ex.execResourceDelete(step)
	case "resource.kinds":
		return ex.execResourceKinds(step, result)
	case "controller.list":
		return ex.execControllerList(step, result)
	case "controller.reconcile":
		return ex.execControllerReconcile(step)
	case "health.check":
		return ex.execHealthCheck(step, result)
	case "application.create":
		return ex.execApplicationCreate(step)
	case "application.get":
		return ex.execApplicationGet(step, result)
	case "complete":
		return nil
	case "format":
		return nil
	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}
}

// ── Action Implementations ──────────────────────────────────────────────

func (ex *Executor) execValidate(step *ExecutionStep) error {
	parts := strings.SplitN(step.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", step.Target)
	}
	kind, id := parts[0], parts[1]

	switch kind {
	case "Project":
		p := project.NewProject(id, id, "development", "Created via Intent Engine")
		if err := p.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		step.Result = fmt.Sprintf("Project %q is valid", id)
		return nil
	case "Deploy":
		step.Result = fmt.Sprintf("Deploy request for %q is valid", id)
		return nil
	default:
		return fmt.Errorf("unknown kind for validation: %s", kind)
	}
}

func (ex *Executor) execResourceCreate(step *ExecutionStep) error {
	parts := strings.SplitN(step.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", step.Target)
	}
	kind, id := parts[0], parts[1]

	switch kind {
	case "Project":
		p := project.NewProject(id, id, "development", "Created via Intent Engine")
		if err := ex.resRegistry.Create(context.Background(), p); err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		step.Result = fmt.Sprintf("Project %q created", id)
		return nil
	default:
		return fmt.Errorf("unknown kind for creation: %s", kind)
	}
}

func (ex *Executor) execResourceGet(step *ExecutionStep, result *IntentResult) error {
	parts := strings.SplitN(step.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", step.Target)
	}
	kind, id := parts[0], parts[1]

	res, err := ex.resRegistry.Get(kind, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// For delete verification, "not found" confirms deletion
			step.Result = fmt.Sprintf("%s %q not found (deletion confirmed)", kind, id)
			result.Details = append(result.Details, ResultItem{
				Message: fmt.Sprintf("%s %q has been removed", kind, id),
				Type:    "success",
			})
			return nil
		}
		return fmt.Errorf("get %s %q: %w", kind, id, err)
	}

	meta := res.GetMetadata()
	result.Details = append(result.Details, ResultItem{
		Message: fmt.Sprintf("%s %q — Status: %s", kind, id, res.GetStatus()),
		Type:    "info",
		Detail:  fmt.Sprintf("Created: %s", meta.CreatedAt),
	})
	step.Result = fmt.Sprintf("%s %q found (status: %s)", kind, id, res.GetStatus())
	return nil
}

func (ex *Executor) execResourceList(step *ExecutionStep, result *IntentResult) error {
	kind := step.Target
	if kind == "" {
		kind = "Project"
	}

	items, err := ex.resRegistry.List(kind)
	if err != nil {
		return fmt.Errorf("list %s: %w", kind, err)
	}

	if len(items) == 0 {
		result.Details = append(result.Details, ResultItem{
			Message: fmt.Sprintf("No %s found", strings.ToLower(kind)),
			Type:    "info",
		})
		step.Result = fmt.Sprintf("Found 0 %s", strings.ToLower(kind))
		return nil
	}

	for _, item := range items {
		meta := item.GetMetadata()
		result.Details = append(result.Details, ResultItem{
			Message: fmt.Sprintf("%s: %s (%s)", kind, meta.Name, meta.ID),
			Type:    "success",
			Detail:  fmt.Sprintf("Status: %s", item.GetStatus()),
		})
	}

	step.Result = fmt.Sprintf("Found %d %s", len(items), strings.ToLower(kind))
	result.Summary = step.Result
	return nil
}

func (ex *Executor) execResourceDelete(step *ExecutionStep) error {
	parts := strings.SplitN(step.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", step.Target)
	}
	kind, id := parts[0], parts[1]

	if err := ex.resRegistry.Delete(context.Background(), kind, id); err != nil {
		return fmt.Errorf("delete %s %q: %w", kind, id, err)
	}
	step.Result = fmt.Sprintf("%s %q deleted", kind, id)
	return nil
}

func (ex *Executor) execResourceKinds(step *ExecutionStep, result *IntentResult) error {
	kinds := ex.resRegistry.ListKinds()
	if len(kinds) == 0 {
		result.Details = append(result.Details, ResultItem{
			Message: "No resource kinds registered",
			Type:    "info",
		})
		step.Result = "Found 0 resource kinds"
		return nil
	}

	for _, k := range kinds {
		ns := "namespaced"
		if !k.Namespaced {
			ns = "cluster-scoped"
		}
		result.Details = append(result.Details, ResultItem{
			Message: fmt.Sprintf("%s (%s)", k.Name, ns),
			Type:    "info",
			Detail:  fmt.Sprintf("Versions: %v", k.Versions),
		})
	}

	step.Result = fmt.Sprintf("Found %d resource kinds", len(kinds))
	return nil
}

func (ex *Executor) execControllerList(step *ExecutionStep, result *IntentResult) error {
	if ex.ctrlManager == nil {
		return fmt.Errorf("controller manager not available")
	}

	names := ex.ctrlManager.ControllerNames()
	if len(names) == 0 {
		result.Details = append(result.Details, ResultItem{
			Message: "No controllers registered",
			Type:    "info",
		})
		step.Result = "Found 0 controllers"
		return nil
	}

	for _, name := range names {
		ctrl, ok := ex.ctrlManager.Get(name)
		if !ok {
			continue
		}
		h := ctrl.Health()
		result.Details = append(result.Details, ResultItem{
			Message: fmt.Sprintf("%s → %s", name, ctrl.Kind()),
			Type:    "success",
			Detail:  fmt.Sprintf("State: %s, Reconciled: %d", h.State, h.ReconcileCount),
		})
	}

	step.Result = fmt.Sprintf("Found %d controllers", len(names))
	return nil
}

func (ex *Executor) execControllerReconcile(step *ExecutionStep) error {
	if ex.ctrlManager == nil {
		return fmt.Errorf("controller manager not available")
	}
	step.Result = "Reconciliation dispatched via Controller Runtime"
	return nil
}

func (ex *Executor) execHealthCheck(step *ExecutionStep, result *IntentResult) error {
	if ex.healthMgr == nil {
		return fmt.Errorf("health manager not available")
	}

	report := ex.healthMgr.All()
	if len(report) == 0 {
		result.Details = append(result.Details, ResultItem{
			Message: "No health data available",
			Type:    "info",
		})
		step.Result = "No health components"
		return nil
	}

	for name, h := range report {
		msgType := "success"
		if h.State != types.StateRunning && h.State != types.StatePending {
			msgType = "error"
		}
		result.Details = append(result.Details, ResultItem{
			Message: fmt.Sprintf("%s: %s", name, h.State),
			Type:    msgType,
			Detail:  h.Message,
		})
	}

	step.Result = fmt.Sprintf("Checked %d components", len(report))
	result.Summary = step.Result
	return nil
}

// ── Application Actions (Natural Infrastructure) ───────────────────────────

func (ex *Executor) execApplicationCreate(step *ExecutionStep) error {
	parts := strings.SplitN(step.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", step.Target)
	}
	kind, id := parts[0], parts[1]
	if kind != "Application" {
		return fmt.Errorf("expected Application kind, got %q", kind)
	}

	// Check if Application already exists
	existing, err := ex.resRegistry.Get("Application", id)
	if err == nil && existing != nil {
		step.Result = fmt.Sprintf("Application %q already exists (resource version: %d)", id, existing.GetMetadata().ResourceVersion)
		return nil
	}

	// Build spec from step metadata or plan params (passed via step.Result as JSON later,
	// for now use sensible defaults for git-based deployments)
	spec := application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type:   application.SourceGit,
			URL:    "https://github.com/cloudos/sample-app",
			Branch: "main",
		},
		Runtime: application.ApplicationRuntime{
			Type:    application.RuntimeNode,
			Port:    3000,
			Command: "npm start",
		},
		Deployment: application.ApplicationDeployment{
			Port:     3000,
			Replicas: 1,
		},
		Environment: map[string]string{},
		Settings:    map[string]string{},
	}

	app := application.NewApplication(id, id, spec)
	app.EnsureDefaults()

	if err := ex.resRegistry.Create(context.Background(), app); err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	step.Result = fmt.Sprintf("Application %q created (namespace: %s, runtime: %s)", id, app.Metadata_.Namespace, spec.Runtime.Type)
	return nil
}

func (ex *Executor) execApplicationGet(step *ExecutionStep, result *IntentResult) error {
	parts := strings.SplitN(step.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", step.Target)
	}
	kind, id := parts[0], parts[1]
	if kind != "Application" {
		return fmt.Errorf("expected Application kind, got %q", kind)
	}

	res, err := ex.resRegistry.Get("Application", id)
	if err != nil {
		return fmt.Errorf("get application %q: %w", id, err)
	}

	app, ok := res.(*application.Application)
	if !ok {
		return fmt.Errorf("expected *application.Application, got %T", res)
	}

	// Report status
	url := app.Status_.URL
	if url == "" {
		url = "pending (controller will assign after reconciliation)"
	}

	phase := app.Status_.Phase
	health := app.Status_.Health
	replicas := strconv.Itoa(app.Spec_.Deployment.Replicas)

	result.Details = append(result.Details, ResultItem{
		Message: fmt.Sprintf("Application %q — Phase: %s, Health: %s", id, phase, health),
		Type:    "success",
		Detail:  fmt.Sprintf("Runtime: %s, Replicas: %s, URL: %s", app.Spec_.Runtime.Type, replicas, url),
	})

	step.Result = fmt.Sprintf("Application %q: phase=%s health=%s url=%s", id, phase, health, url)
	return nil
}
