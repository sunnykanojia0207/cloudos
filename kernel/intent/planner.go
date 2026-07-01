package intent

import (
	"fmt"
	"time"
)

// Planner converts parsed Intents into ordered, dependency-aware ExecutionPlans
type Planner struct{}

// NewPlanner creates a new Planner
func NewPlanner() *Planner {
	return &Planner{}
}

// buildPreview generates a PlanPreview for an intent and its steps.
func buildPreview(intent *Intent, steps []ExecutionStep) *PlanPreview {
	stepBullets := make([]string, 0, len(steps))
	for _, s := range steps {
		stepBullets = append(stepBullets, s.Description)
	}

	resources := []string{}
	switch intent.Type {
	case IntentDeploy:
		resources = append(resources, fmt.Sprintf("Application %q", intent.Params["appName"]))
	case IntentCreateProject:
		resources = append(resources, fmt.Sprintf("Project %q", intent.Params["name"]))
	}

	title := fmt.Sprintf("Plan: %s", intent.Raw)
	summary := fmt.Sprintf("This will execute %d steps to %s.", len(steps), intent.Raw)
	if len(resources) > 0 {
		summary += fmt.Sprintf(" Resources affected: %s", resources[0])
	}

	return &PlanPreview{
		Title:       title,
		Summary:     summary,
		Steps:       stepBullets,
		Resources:   resources,
		EstimatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// Plan converts an Intent into an ExecutionPlan with ordered steps and dependencies
func (pl *Planner) Plan(intent *Intent) (*ExecutionPlan, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	plan := &ExecutionPlan{
		ID:         fmt.Sprintf("plan_%s", intent.ID),
		IntentID:   intent.ID,
		IntentType: intent.Type,
		Status:     IntentPending,
		CreatedAt:  now,
	}

	switch intent.Type {
	case IntentCreateProject:
		name := intent.Params["name"]
		id := intent.Params["id"]
		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "Validate Request",
				Description:  fmt.Sprintf("Validating project creation request for %q", name),
				Status:       StepPending,
				Action:       "validate",
				Target:       fmt.Sprintf("Project:%s", id),
			},
			{
				ID: "2", Name: "Create Project Resource",
				Description:  fmt.Sprintf("Creating project %q with ID %q", name, id),
				Status:       StepPending,
				Action:       "resource.create",
				Target:       fmt.Sprintf("Project:%s", id),
				Dependencies: []string{"1"},
			},
			{
				ID: "3", Name: "Wait for Reconciliation",
				Description:  "Waiting for the Project Controller to reconcile the new project",
				Status:       StepPending,
				Action:       "controller.reconcile",
				Target:       "Project",
				Dependencies: []string{"2"},
			},
			{
				ID: "4", Name: "Verify Status",
				Description:  fmt.Sprintf("Verifying project %q status is Active", id),
				Status:       StepPending,
				Action:       "resource.get",
				Target:       fmt.Sprintf("Project:%s", id),
				Dependencies: []string{"3"},
			},
			{
				ID: "5", Name: "Complete",
				Description:  fmt.Sprintf("Project %q created successfully", name),
				Status:       StepPending,
				Action:       "complete",
				Target:       "",
				Dependencies: []string{"4"},
			},
		}

	case IntentListProjects:
		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "List Projects",
				Description:  "Fetching all projects from the Resource Engine",
				Status:       StepPending,
				Action:       "resource.list",
				Target:       "Project",
			},
			{
				ID: "2", Name: "Format Results",
				Description:  "Preparing project list",
				Status:       StepPending,
				Action:       "format",
				Target:       "",
				Dependencies: []string{"1"},
			},
		}

	case IntentDeleteProject:
		name := intent.Params["name"]
		id := intent.Params["id"]
		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "Validate Request",
				Description:  fmt.Sprintf("Validating deletion request for project %q", name),
				Status:       StepPending,
				Action:       "validate",
				Target:       fmt.Sprintf("Project:%s", id),
			},
			{
				ID: "2", Name: "Delete Project Resource",
				Description:  fmt.Sprintf("Deleting project %q", id),
				Status:       StepPending,
				Action:       "resource.delete",
				Target:       fmt.Sprintf("Project:%s", id),
				Dependencies: []string{"1"},
			},
			{
				ID: "3", Name: "Verify Deletion",
				Description:  "Confirming project has been removed",
				Status:       StepPending,
				Action:       "resource.get",
				Target:       fmt.Sprintf("Project:%s", id),
				Dependencies: []string{"2"},
			},
		}

	case IntentShowControllers:
		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "List Controllers",
				Description:  "Fetching all registered controllers from the Controller Runtime",
				Status:       StepPending,
				Action:       "controller.list",
				Target:       "",
			},
			{
				ID: "2", Name: "Format Results",
				Description:  "Preparing controller list",
				Status:       StepPending,
				Action:       "format",
				Target:       "",
				Dependencies: []string{"1"},
			},
		}

	case IntentShowResources:
		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "List Resource Kinds",
				Description:  "Fetching all registered resource kinds",
				Status:       StepPending,
				Action:       "resource.kinds",
				Target:       "",
			},
			{
				ID: "2", Name: "Format Results",
				Description:  "Preparing resource kind list",
				Status:       StepPending,
				Action:       "format",
				Target:       "",
				Dependencies: []string{"1"},
			},
		}

	case IntentShowHealth:
		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "Check System Health",
				Description:  "Checking all system components health status",
				Status:       StepPending,
				Action:       "health.check",
				Target:       "",
			},
			{
				ID: "2", Name: "Format Results",
				Description:  "Preparing health report",
				Status:       StepPending,
				Action:       "format",
				Target:       "",
				Dependencies: []string{"1"},
			},
		}

	// ── Deploy Intent ───────────────────────────────────────────────
	case IntentDeploy:
		appName := intent.Params["appName"]
		runtime := intent.Params["runtime"]
		sourceURL := intent.Params["sourceURL"]
		appID := toID(appName)

		plan.Steps = []ExecutionStep{
			{
				ID: "1", Name: "Validate Deploy Request",
				Description:  fmt.Sprintf("Validating deploy request: %s app from %s", runtime, sourceURL),
				Status:       StepPending,
				Action:       "validate",
				Target:       fmt.Sprintf("Deploy:%s", appID),
			},
			{
				ID: "2", Name: "Create Application Resource",
				Description:  fmt.Sprintf("Creating Application %q (runtime: %s) from %s", appName, runtime, sourceURL),
				Status:       StepPending,
				Action:       "application.create",
				Target:       fmt.Sprintf("Application:%s", appID),
				Dependencies: []string{"1"},
			},
			{
				ID: "3", Name: "Wait for Application Controller",
				Description:  "Waiting for the Application Controller to reconcile and generate deployment workflow",
				Status:       StepPending,
				Action:       "controller.reconcile",
				Target:       "Application",
				Dependencies: []string{"2"},
			},
			{
				ID: "4", Name: "Check Application Status",
				Description:  fmt.Sprintf("Verifying Application %q phase and health", appName),
				Status:       StepPending,
				Action:       "application.get",
				Target:       fmt.Sprintf("Application:%s", appID),
				Dependencies: []string{"3"},
			},
			{
				ID: "5", Name: "Return Access URL",
				Description:  fmt.Sprintf("Application %q is ready. Returning access URL.", appName),
				Status:       StepPending,
				Action:       "complete",
				Target:       "",
				Dependencies: []string{"4"},
			},
		}

	default:
		return nil, fmt.Errorf("unsupported intent type: %s", intent.Type)
	}

	// Build the PlanPreview for every plan
	plan.Preview = buildPreview(intent, plan.Steps)

	return plan, nil
}
