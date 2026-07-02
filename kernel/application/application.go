// Package application implements the CloudOS Application resource — the
// first-class representation of user software. An Application is NOT Docker,
// NOT Kubernetes. It is the user's code: a React app, a Laravel API, a
// Next.js site, a Python worker, a Go service, or a static website.
//
// Every Application belongs to a Project. Every deployment of an Application
// is a WorkflowExecution. The Application Controller validates the spec,
// generates a deployment workflow, and submits it through the Workflow Service.
//
//	Project
//	    │
//	    Application (desired state — the user's software)
//	    │
//	    Application Controller (validates + builds workflow)
//	    │
//	    Workflow Service (submits + tracks deployment)
//	    │
//	    WorkflowExecution Resource (each deployment)
//
// This package provides the Application resource type (spec + status),
// validation, defaults, and the ApplicationController that reconciles
// application state by creating deployment workflows.
package application

import (
	"fmt"
	"regexp"
	"time"

	"github.com/cloudos/cloudos/kernel/resource"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// Kind is the resource kind string for Application.
	Kind = "Application"

	// DeploymentHistoryMax is the maximum number of deployment reports
	// retained in ApplicationStatus.DeploymentHistory.
	DeploymentHistoryMax = 20

	// Source types.
	SourceGit   = "git"
	SourceLocal = "local"
	SourceDocker = "docker"

	// Runtime types.
	RuntimeNode    = "node"
	RuntimePython  = "python"
	RuntimeGo      = "go"
	RuntimeStatic  = "static"
	RuntimeNextJS  = "nextjs"
	RuntimeLaravel = "laravel"
	RuntimeDocker  = "docker"

	// Phase values.
	PhaseCreating  = "Creating"
	PhaseRunning   = "Running"
	PhaseStopped   = "Stopped"
	PhaseFailed    = "Failed"
	PhaseDeleting  = "Deleting"
	PhaseDeploying = "Deploying"

	// Health values.
	HealthHealthy  = "Healthy"
	HealthDegraded = "Degraded"
	HealthError    = "Error"
)

// ValidSourceTypes is the set of allowed source type values.
var ValidSourceTypes = map[string]bool{
	SourceGit:   true,
	SourceLocal: true,
	SourceDocker: true,
}

// ValidRuntimes is the set of allowed runtime type values.
var ValidRuntimes = map[string]bool{
	RuntimeNode:    true,
	RuntimePython:  true,
	RuntimeGo:      true,
	RuntimeStatic:  true,
	RuntimeNextJS:  true,
	RuntimeLaravel: true,
	RuntimeDocker:  true,
}

// ValidPhases is the set of allowed lifecycle phase values.
var ValidPhases = map[string]bool{
	PhaseCreating:  true,
	PhaseRunning:   true,
	PhaseStopped:   true,
	PhaseFailed:    true,
	PhaseDeleting:  true,
	PhaseDeploying: true,
}

// ValidHealthStatuses is the set of allowed health values.
var ValidHealthStatuses = map[string]bool{
	HealthHealthy:  true,
	HealthDegraded: true,
	HealthError:    true,
}

// DefaultApplicationSettings are the settings applied when an Application is created.
var DefaultApplicationSettings = map[string]string{
	"autoDeploy":        "true",
	"healthCheckEnabled": "true",
	"healthCheckPath":   "/health",
	"healthCheckInterval": "30s",
	"maxReplicas":       "1",
	"minReplicas":       "1",
}

// applicationIDPattern enforces DNS-label-compatible application IDs.
var applicationIDPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ── Condition ──────────────────────────────────────────────────────────────

// Condition represents a single status condition for a resource.
// Inspired by Kubernetes condition types.
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty"`
}

// ── Source Spec ────────────────────────────────────────────────────────────

// ApplicationSource describes where the application's code comes from.
type ApplicationSource struct {
	// Type is the source type: "git", "local", or "docker".
	Type string `json:"type"`

	// URL is the repository or image URL (e.g. "https://github.com/user/repo").
	URL string `json:"url,omitempty"`

	// Branch is the git branch to deploy from (default: "main").
	Branch string `json:"branch,omitempty"`

	// Path is the subdirectory within the repo containing the application.
	Path string `json:"path,omitempty"`
}

// ── Runtime Spec ───────────────────────────────────────────────────────────

// ApplicationRuntime describes how the application runs.
type ApplicationRuntime struct {
	// Type is the runtime type: "node", "python", "go", "static", "nextjs",
	// "laravel", or "docker".
	Type string `json:"type"`

	// Command is the start command (e.g. "npm start", "python app.py").
	// If empty, a sensible default is inferred from the runtime type.
	Command string `json:"command,omitempty"`

	// Port is the port the application listens on.
	Port int `json:"port,omitempty"`

	// Args are additional arguments passed to the command.
	Args []string `json:"args,omitempty"`
}

// ── Build Spec ─────────────────────────────────────────────────────────────

// ApplicationBuild describes how to build the application.
type ApplicationBuild struct {
	// Command is the build command (e.g. "npm run build", "make build").
	// If empty, no build step is performed.
	Command string `json:"command,omitempty"`

	// OutputDir is the directory containing build output (e.g. "build", "dist").
	OutputDir string `json:"outputDir,omitempty"`

	// InstallCmd is the dependency install command (e.g. "npm ci", "pip install").
	InstallCmd string `json:"installCmd,omitempty"`
}

// ── Deployment Spec ────────────────────────────────────────────────────────

// ApplicationDeployment describes how the application is deployed.
type ApplicationDeployment struct {
	// Port is the port the application listens on.
	Port int `json:"port,omitempty"`

	// Domain is the custom domain for the application (optional).
	Domain string `json:"domain,omitempty"`

	// Replicas is the desired number of instances.
	Replicas int `json:"replicas,omitempty"`
}

// ── ApplicationSpec ─────────────────────────────────────────────────────────

// ApplicationSpec is the desired state of a CloudOS Application.
type ApplicationSpec struct {
	// Source describes where the code comes from.
	Source ApplicationSource `json:"source"`

	// Runtime describes how the application runs.
	Runtime ApplicationRuntime `json:"runtime"`

	// Build describes how to build the application (optional).
	Build *ApplicationBuild `json:"build,omitempty"`

	// Deployment configures deployment behavior.
	Deployment ApplicationDeployment `json:"deployment,omitempty"`

	// Environment is a map of environment variable names to values.
	Environment map[string]string `json:"environment,omitempty"`

	// Settings are arbitrary key-value configuration pairs.
	Settings map[string]string `json:"settings,omitempty"`
}

// ── ApplicationStatus ───────────────────────────────────────────────────────

// ApplicationStatus is the current observed state of a CloudOS Application.
type ApplicationStatus struct {
	// Phase is the application lifecycle phase.
	// One of: "Creating", "Running", "Stopped", "Failed", "Deleting", "Deploying".
	Phase string `json:"phase"`

	// Health is the overall operational health.
	// One of: "Healthy", "Degraded", "Error".
	Health string `json:"health"`

	// URL is the access URL for the running application.
	URL string `json:"url,omitempty"`

	// Conditions provide detailed status signals.
	Conditions []Condition `json:"conditions,omitempty"`

	// CurrentDeploymentID references the active WorkflowExecution.
	CurrentDeploymentID string `json:"currentDeploymentId,omitempty"`

	// LastDeploymentTime is when the last deployment completed.
	LastDeploymentTime time.Time `json:"lastDeploymentTime,omitempty"`

	// DeploymentCount is the total number of deployments.
	DeploymentCount int `json:"deploymentCount"`

	// DeploymentHistory is the ordered list of deployment reports, most
	// recent first (index 0 = latest). Each report captures the full
	// deployment story: repository, detected runtime, buildpack, build
	// time, runtime, endpoint, health, logs, and diagnostics.
	//
	// Capped at DeploymentHistoryMax entries. Oldest entries are pruned.
	//
	// LastReport is a convenience pointer to the most recent report
	// (equivalent to DeploymentHistory[0] when non-empty).
	DeploymentHistory []DeploymentReport `json:"deploymentHistory,omitempty"`

	// LastReport is the structured deployment report from the most recent
	// deployment. Populated by the controller when a deployment completes.
	// It is always DeploymentHistory[0] when history is non-empty.
	LastReport *DeploymentReport `json:"lastReport,omitempty"`

	// CreatedAt is when the application was created.
	CreatedAt time.Time `json:"createdAt,omitempty"`

	// UpdatedAt is when the application was last modified.
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// ── DeploymentReport ────────────────────────────────────────────────────────

// DeploymentReport is the structured story of a single deployment. It captures
// every stage from repository to running application, providing full visibility
// into what happened, how long it took, and where the app is running.
//
// This is the primary output users see after "cloudos deploy". It answers the
// question: "What happened?" with a single structured response.
//
// Populated by the Application Controller when a deployment workflow completes.
// Extended by the Executor when runtimes pass back structured metadata.
type DeploymentReport struct {
	// DeploymentNumber is the sequential deployment number for this application.
	// Starts at 1 for the first deployment.
	DeploymentNumber int `json:"deploymentNumber"`

	// StartedAt is when the deployment workflow was submitted.
	StartedAt time.Time `json:"startedAt"`

	// CompletedAt is when the deployment workflow reached a terminal phase.
	CompletedAt time.Time `json:"completedAt"`

	// Duration is a human-readable duration string (e.g. "8.2s").
	Duration string `json:"duration"`

	// Repository is the source repository URL.
	Repository string `json:"repository"`

	// Branch is the git branch deployed.
	Branch string `json:"branch"`

	// CommitSHA is the short commit hash deployed (e.g. "7f41ab2").
	// Populated from the source clone result when available.
	CommitSHA string `json:"commitSha,omitempty"`

	// DetectedRuntime is what the buildpack detected (e.g. "Go 1.24", "Node 20").
	DetectedRuntime string `json:"detectedRuntime,omitempty"`

	// Buildpack is the name of the buildpack that handled this deployment.
	Buildpack string `json:"buildpack,omitempty"`

	// BuildSuccess indicates whether the build phase completed successfully.
	BuildSuccess bool `json:"buildSuccess"`

	// RuntimeName is the name of the runtime that executed the application.
	RuntimeName string `json:"runtimeName,omitempty"`

	// RuntimeVersion is the API version of the runtime contract used.
	RuntimeVersion string `json:"runtimeVersion,omitempty"`

	// Environment labels the deployment environment (e.g. "local", "docker",
	// "production", "staging"). Inferred from the runtime type and labels.
	Environment string `json:"environment,omitempty"`

	// ArtifactType describes the build output type (e.g. "Go Binary",
	// "React Bundle", "Static HTML").
	ArtifactType string `json:"artifactType,omitempty"`

	// ArtifactSize is the size of the build artifact in bytes.
	ArtifactSize int64 `json:"artifactSize,omitempty"`

	// WorkflowID is the ID of the workflow execution for this deployment.
	WorkflowID string `json:"workflowId"`

	// WorkflowSteps is the total number of steps in the deployment workflow.
	WorkflowSteps int `json:"workflowSteps"`

	// HealthStatus is the health check result ("Healthy", "Degraded", "Error").
	HealthStatus string `json:"healthStatus"`

	// Endpoint is the URL where the application is accessible.
	Endpoint string `json:"endpoint"`

	// LogLineCount is the number of log lines available from this deployment.
	LogLineCount int `json:"logLineCount,omitempty"`

	// Warnings from the deployment process (non-fatal issues).
	Warnings []string `json:"warnings,omitempty"`

	// Errors from the deployment process (fatal issues, if deployment failed).
	Errors []string `json:"errors,omitempty"`
}

// ── Application Resource ───────────────────────────────────────────────────

// Application is the concrete CloudOS Application resource. It implements the
// resource.Resource interface and can be used with the Resource Engine.
type Application struct {
	Metadata_ *resource.Metadata `json:"metadata"`
	Spec_     ApplicationSpec    `json:"spec"`
	Status_   ApplicationStatus  `json:"status"`
}

// NewApplication creates a new Application resource with sensible defaults.
// The Application is placed in the same namespace as its parent Project and
// initialized with Creating phase and Pending health.
func NewApplication(id, name string, spec ApplicationSpec) *Application {
	now := time.Now()
	return &Application{
		Metadata_: &resource.Metadata{
			ID:              id,
			Name:            name,
			Namespace:       resource.NamespaceDefault,
			Kind:            Kind,
			APIVersion:      resource.APIVersion,
			Labels:          make(map[string]string),
			Annotations:     make(map[string]string),
			CreatedAt:       now,
			UpdatedAt:       now,
			ResourceVersion: 1,
		},
		Spec_: spec,
		Status_: ApplicationStatus{
			Phase:           PhaseCreating,
			Health:          HealthHealthy,
			Conditions:      []Condition{},
			DeploymentCount: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
}

// ── Resource Interface ─────────────────────────────────────────────────────

func (a *Application) GetKind() string              { return Kind }
func (a *Application) GetMetadata() *resource.Metadata { return a.Metadata_ }
func (a *Application) GetSpec() interface{}         { return a.Spec_ }
func (a *Application) GetStatus() interface{}        { return a.Status_ }

func (a *Application) SetStatus(s interface{}) {
	if st, ok := s.(ApplicationStatus); ok {
		a.Status_ = st
	}
}

// Validate checks the Application for semantic correctness.
func (a *Application) Validate() error {
	if a.Metadata_.ID == "" {
		return fmt.Errorf("application id is required")
	}
	if !applicationIDPattern.MatchString(a.Metadata_.ID) {
		return fmt.Errorf("application id %q must match %s (lowercase letters, digits, and hyphens; must start and end with alphanumeric)",
			a.Metadata_.ID, applicationIDPattern.String())
	}

	// Validate source type.
	if !ValidSourceTypes[a.Spec_.Source.Type] {
		return fmt.Errorf("source type %q is invalid; must be one of: git, local, docker",
			a.Spec_.Source.Type)
	}
	if a.Spec_.Source.Type == SourceGit && a.Spec_.Source.URL == "" {
		return fmt.Errorf("git source requires a URL")
	}

	// Validate runtime type.
	if !ValidRuntimes[a.Spec_.Runtime.Type] {
		return fmt.Errorf("runtime type %q is invalid; must be one of: node, python, go, static, nextjs, laravel, docker",
			a.Spec_.Runtime.Type)
	}

	// Validate port ranges.
	if a.Spec_.Runtime.Port < 0 || a.Spec_.Runtime.Port > 65535 {
		return fmt.Errorf("port %d is out of range (0–65535)", a.Spec_.Runtime.Port)
	}
	if a.Spec_.Deployment.Port < 0 || a.Spec_.Deployment.Port > 65535 {
		return fmt.Errorf("deployment port %d is out of range (0–65535)", a.Spec_.Deployment.Port)
	}

	if a.Metadata_.Kind != Kind {
		return fmt.Errorf("kind must be %q, got %q", Kind, a.Metadata_.Kind)
	}

	return nil
}

// ── Defaults ───────────────────────────────────────────────────────────────

// EnsureDefaults populates any missing default values in the Application's spec
// and metadata.
func (a *Application) EnsureDefaults() {
	if a.Metadata_.Labels == nil {
		a.Metadata_.Labels = make(map[string]string)
	}
	if a.Metadata_.Annotations == nil {
		a.Metadata_.Annotations = make(map[string]string)
	}

	// Set labels based on spec.
	a.Metadata_.Labels["runtime"] = a.Spec_.Runtime.Type
	a.Metadata_.Labels["source"] = a.Spec_.Source.Type

	if a.Spec_.Source.Branch == "" {
		a.Spec_.Source.Branch = "main"
	}
	if a.Spec_.Settings == nil {
		a.Spec_.Settings = copyMap(DefaultApplicationSettings)
	}
	if a.Spec_.Deployment.Replicas == 0 {
		a.Spec_.Deployment.Replicas = 1
	}

	// Infer default command from runtime type if not provided.
	if a.Spec_.Runtime.Command == "" {
		a.Spec_.Runtime.Command = defaultCommand(a.Spec_.Runtime.Type)
	}

	// Infer default port from runtime type if not provided.
	if a.Spec_.Runtime.Port == 0 {
		a.Spec_.Runtime.Port = defaultPort(a.Spec_.Runtime.Type)
	}
	if a.Spec_.Deployment.Port == 0 {
		a.Spec_.Deployment.Port = a.Spec_.Runtime.Port
	}

	if a.Status_.Phase == "" {
		a.Status_.Phase = PhaseCreating
	}
	if a.Status_.Health == "" {
		a.Status_.Health = HealthHealthy
	}
	if a.Status_.Conditions == nil {
		a.Status_.Conditions = []Condition{}
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

// AddCondition adds or updates a status condition.
func (a *Application) AddCondition(condType, status, reason, message string) {
	now := time.Now()
	for i, c := range a.Status_.Conditions {
		if c.Type == condType {
			a.Status_.Conditions[i].Status = status
			a.Status_.Conditions[i].Reason = reason
			a.Status_.Conditions[i].Message = message
			a.Status_.Conditions[i].LastTransitionTime = now
			return
		}
	}
	a.Status_.Conditions = append(a.Status_.Conditions, Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	})
}

// GetCondition returns a condition by type.
func (a *Application) GetCondition(condType string) *Condition {
	for _, c := range a.Status_.Conditions {
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// Touch updates the UpdatedAt timestamp.
func (a *Application) Touch() {
	a.Status_.UpdatedAt = time.Now()
}

// ── Default Commands per Runtime ───────────────────────────────────────────

// defaultCommand returns a sensible default start command for the given runtime.
func defaultCommand(runtime string) string {
	switch runtime {
	case RuntimeNode:
		return "npm start"
	case RuntimePython:
		return "python app.py"
	case RuntimeGo:
		return "./app"
	case RuntimeNextJS:
		return "npm run start"
	case RuntimeLaravel:
		return "php artisan serve"
	case RuntimeStatic:
		return "" // Static sites don't have a runtime command
	case RuntimeDocker:
		return "" // Docker images have their own CMD
	default:
		return ""
	}
}

// defaultPort returns a sensible default port for the given runtime.
func defaultPort(runtime string) int {
	switch runtime {
	case RuntimeNode, RuntimeNextJS:
		return 3000
	case RuntimePython, RuntimeLaravel:
		return 8000
	case RuntimeGo:
		return 8080
	case RuntimeStatic:
		return 80
	case RuntimeDocker:
		return 0 // Port is defined by the Docker image
	default:
		return 8080
	}
}

// ── Deployment Workflow Plan ───────────────────────────────────────────────

// DeploymentStep is a single step in a deployment workflow.
type DeploymentStep struct {
	ID       string
	Name     string
	Action   string
	Target   string
	DependsOn []string
}

// BuildDeploymentPlan creates a deployment plan for the Application.
// The plan is a list of steps that the Application Controller will convert
// into a WorkflowDefinition and submit via the Workflow Service.
//
// Each step maps to a workflow Node. The plan follows a standard deployment
// pipeline. Note: install and build are NOT separate nodes — they are handled
// inside the deploy step because the buildpack engine (running on the cloned
// source) is the authoritative source of install/build/start commands, not
// the Application spec which may only contain speculative values from the UI.
//
//	Validate → Clone → Deploy (detect+install+build+start) → HealthCheck → Complete
func BuildDeploymentPlan(app *Application) []DeploymentStep {
	runtime := app.Spec_.Runtime.Type
	sourceType := app.Spec_.Source.Type

	steps := []DeploymentStep{}

	// Step 1: Validate the application configuration.
	steps = append(steps, DeploymentStep{
		ID:     "validate",
		Name:   "Validate Application",
		Action: "validate",
		Target: fmt.Sprintf("Application:%s", app.Metadata_.ID),
	})

	lastID := "validate"

	// Step 2: Clone source (only for git sources).
	if sourceType == SourceGit {
		steps = append(steps, DeploymentStep{
			ID:        "clone",
			Name:      "Clone Source Repository",
			Action:    "source.clone",
			Target:    app.Spec_.Source.URL,
			DependsOn: []string{lastID},
		})
		lastID = "clone"
	}

	// Step 3: Deploy via provider.
	// The provider.deploy action is the authoritative deployment orchestrator.
	// It uses the buildpack engine against the cloned source to:
	//   1. Detect the runtime (Go, Node, Python, etc.)
	//   2. Plan install/build/start commands
	//   3. Run install dependencies
	//   4. Build the project (if applicable)
	//   5. Produce an Artifact
	//   6. Prepare the runtime (allocate port, env, directory)
	//   7. Start the application
	//   8. Return the running instance URL
	//
	// Target format: "runtime:{type}:{appID}" — the app ID is embedded so the
	// executor can route logs and metadata correctly.
	steps = append(steps, DeploymentStep{
		ID:        "deploy",
		Name:      "Deploy Application",
		Action:    "provider.deploy",
		Target:    fmt.Sprintf("runtime:%s:%s", runtime, app.Metadata_.ID),
		DependsOn: []string{lastID},
	})
	lastID = "deploy"

	// Step 4: Health check.
	steps = append(steps, DeploymentStep{
		ID:        "healthcheck",
		Name:      "Health Check",
		Action:    "health.check",
		Target:    app.Spec_.Runtime.Command,
		DependsOn: []string{lastID},
	})
	lastID = "healthcheck"

	// Step 5: Complete — mark deployment as successful.
	steps = append(steps, DeploymentStep{
		ID:        "complete",
		Name:      "Complete Deployment",
		Action:    "complete",
		Target:    app.Metadata_.ID,
		DependsOn: []string{lastID},
	})

	return steps
}

// copyMap creates a shallow copy of a string map.
func copyMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
