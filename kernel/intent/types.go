// Package intent transforms user intent (natural language) into executable
// plans. It sits between the user and the Kernel — it parses intent, creates
// execution plans, and runs them against the Resource Engine, Controller
// Runtime, and Project system.
package intent

// IntentType represents the type of user intent.
type IntentType string

const (
	IntentCreateProject   IntentType = "create_project"
	IntentListProjects    IntentType = "list_projects"
	IntentDeleteProject   IntentType = "delete_project"
	IntentShowControllers IntentType = "show_controllers"
	IntentShowResources   IntentType = "show_resources"
	IntentShowHealth      IntentType = "show_health"
	IntentDeploy          IntentType = "deploy"
)

// IntentStatus represents the lifecycle status of an intent.
type IntentStatus string

const (
	IntentPending           IntentStatus = "pending"
	IntentParsing           IntentStatus = "parsing"
	IntentValidating        IntentStatus = "validating"
	IntentPlanning          IntentStatus = "planning"
	IntentAwaitingApproval  IntentStatus = "awaiting_approval"
	IntentExecuting         IntentStatus = "executing"
	IntentCompleted         IntentStatus = "completed"
	IntentFailed            IntentStatus = "failed"
)

// StepStatus represents the status of an execution step.
type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepRunning StepStatus = "running"
	StepSuccess StepStatus = "success"
	StepFailed  StepStatus = "failed"
	StepSkipped StepStatus = "skipped"
)

// Intent represents a parsed user intent.
type Intent struct {
	ID          string            `json:"id"`
	Type        IntentType        `json:"type"`
	Raw         string            `json:"raw"`
	Status      IntentStatus      `json:"status"`
	Params      map[string]string `json:"params,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	CompletedAt string            `json:"completedAt,omitempty"`
	Error       string            `json:"error,omitempty"`
	PlanID      string            `json:"planId,omitempty"`
	Result      *IntentResult     `json:"result,omitempty"`
}

// IntentResult holds the structured result of intent execution.
type IntentResult struct {
	Summary string       `json:"summary"`
	Details []ResultItem `json:"details,omitempty"`
	Success bool         `json:"success"`
}

// ResultItem is a single item in an intent execution result.
type ResultItem struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "info", "success", "error", "warning"
	Detail  string `json:"detail,omitempty"`
}

// PlanPreview is a human-readable preview shown to the user before execution.
// It enables the Explain Mode / Trust layer — users see what will happen before it happens.
type PlanPreview struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Steps       []string `json:"steps"`       // bullet-point list of actions
	Resources   []string `json:"resources"`   // resources that will be created/modified
	EstimatedAt string   `json:"estimatedAt"` // when this preview was generated
}

// ExecutionPlan is an ordered sequence of steps that fulfills an intent.
type ExecutionPlan struct {
	ID         string          `json:"id"`
	IntentID   string          `json:"intentId"`
	IntentType IntentType      `json:"intentType"`
	Status     IntentStatus    `json:"status"`
	Preview    *PlanPreview    `json:"preview,omitempty"` // shown to user before confirmation
	Steps      []ExecutionStep `json:"steps"`
	CreatedAt  string          `json:"createdAt"`
	UpdatedAt  string          `json:"updatedAt"`
}

// ExecutionStep is a single unit of work within an execution plan.
type ExecutionStep struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Status       StepStatus `json:"status"`
	Action       string   `json:"action"`                 // e.g., "resource.create", "resource.get", "controller.reconcile"
	Target       string   `json:"target"`                 // e.g., "Project:my-project", "Namespace:default"
	Dependencies []string `json:"dependencies,omitempty"` // step IDs that must complete first
	Result       string   `json:"result,omitempty"`
	Error        string   `json:"error,omitempty"`
	StartedAt    string   `json:"startedAt,omitempty"`
	CompletedAt  string   `json:"completedAt,omitempty"`
}
