package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/resource"
)

// ── Handler ────────────────────────────────────────────────────────────────

// AppHandler serves the Application REST endpoints for creating applications
// and triggering deployments.
type AppHandler struct {
	k *kernel.Kernel
}

// NewAppHandler creates a handler bound to the given kernel.
func NewAppHandler(k *kernel.Kernel) *AppHandler {
	return &AppHandler{k: k}
}

// ── DTOs ───────────────────────────────────────────────────────────────────

// createApplicationRequest is the JSON body for POST /api/v1/applications.
type createApplicationRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// ProjectID associates this Application with a parent Project.
	// When set, the Application appears in the project dashboard and is
	// subject to project-level lifecycle management (e.g. project delete
	// is blocked until all Applications are removed).
	// Optional — standalone Applications are permitted for backward compatibility.
	ProjectID string `json:"projectId,omitempty"`

	Source struct {
		URL    string `json:"url"`
		Branch string `json:"branch,omitempty"`
		Path   string `json:"path,omitempty"`
	} `json:"source"`

	Runtime struct {
		Type    string `json:"type"`
		Command string `json:"command,omitempty"`
		Port    int    `json:"port,omitempty"`
	} `json:"runtime"`

	Build *struct {
		Command    string `json:"command,omitempty"`
		OutputDir  string `json:"outputDir,omitempty"`
		InstallCmd string `json:"installCmd,omitempty"`
	} `json:"build,omitempty"`

	Environment map[string]string `json:"environment,omitempty"`
}

// ── POST /api/v1/applications ──────────────────────────────────────────────

// CreateApplication creates a new Application resource and triggers the first
// deployment. The Application Controller will reconcile the resource and start
// a deployment workflow automatically.
func (ah *AppHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var req createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "INVALID_JSON", "Invalid request body: "+err.Error())
		return
	}

	if req.ID == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}
	if req.Name == "" {
		req.Name = req.ID
	}
	if req.Source.URL == "" {
		BadRequest(w, "MISSING_SOURCE_URL", "Source repository URL is required")
		return
	}
	if req.Runtime.Type == "" {
		BadRequest(w, "MISSING_RUNTIME", "Runtime type is required")
		return
	}
	if req.Source.Branch == "" {
		req.Source.Branch = "main"
	}

	// Build the ApplicationSpec from the request.
	spec := application.ApplicationSpec{
		ProjectID: req.ProjectID,
		Source: application.ApplicationSource{
			Type:   application.SourceGit,
			URL:    req.Source.URL,
			Branch: req.Source.Branch,
			Path:   req.Source.Path,
		},
		Runtime: application.ApplicationRuntime{
			Type:    req.Runtime.Type,
			Command: req.Runtime.Command,
			Port:    req.Runtime.Port,
		},
		Environment: req.Environment,
		Settings:    copyMap(application.DefaultApplicationSettings),
	}

	if req.Build != nil {
		spec.Build = &application.ApplicationBuild{
			Command:    req.Build.Command,
			OutputDir:  req.Build.OutputDir,
			InstallCmd: req.Build.InstallCmd,
		}
	}

	// Set deployment port from runtime port.
	if spec.Runtime.Port != 0 {
		spec.Deployment.Port = spec.Runtime.Port
	}

	// Create the Application resource.
	app := application.NewApplication(req.ID, req.Name, spec)
	app.EnsureDefaults()

	reg := ah.k.ResourceRegistry()
	if err := reg.Create(r.Context(), app); err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	obj := appResourceToObject(app)
	Created(w, obj)
}

// ── POST /api/v1/applications/{id}/deploy ──────────────────────────────────

// TriggerDeploy triggers a new deployment for an existing application by
// transitioning it back through the Creating phase. The Application Controller
// will pick up the change and start a new deployment workflow.
func (ah *AppHandler) TriggerDeploy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}

	reg := ah.k.ResourceRegistry()
	existing, err := reg.Get(application.Kind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	app, ok := existing.(*application.Application)
	if !ok {
		// Try generic resource conversion.
		if generic := genericToApplication(existing); generic != nil {
			app = generic
		} else {
			InternalError(w, "INVALID_RESOURCE", fmt.Sprintf("resource %q is not an Application", id))
			return
		}
	}

	// Reset to Creating phase so the controller starts a new deployment.
	app.Status_.Phase = application.PhaseCreating
	app.Status_.CurrentDeploymentID = ""
	app.AddCondition("Deploying", "True", "DeploymentTriggered", "New deployment triggered by user")

	app.EnsureDefaults()
	app.Touch()

	// Get the latest report's deployment number and increment.
	deploymentNumber := app.Status_.DeploymentCount + 1
	app.Status_.DeploymentCount = deploymentNumber

	if err := reg.Update(r.Context(), app); err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	OK(w, appResourceToObject(app))
}

// ── Helpers ────────────────────────────────────────────────────────────────

// appResourceToObject converts an Application resource to an Object envelope.
func appResourceToObject(app *application.Application) Object {
	meta := app.GetMetadata()

	om := ObjectMeta{
		ID:              meta.ID,
		Name:            meta.Name,
		Labels:          meta.Labels,
		Annotations:     meta.Annotations,
		CreatedAt:       meta.CreatedAt,
		UpdatedAt:       meta.UpdatedAt,
		ResourceVersion: formatResourceVersion(meta.ResourceVersion),
	}

	return Object{
		APIVersion: APIVersion,
		Kind:       application.Kind,
		Metadata:   om,
		Spec:       app.Spec_,
		Status:     app.Status_,
	}
}

// genericToApplication converts a GenericResource to a typed Application.
func genericToApplication(res resource.Resource) *application.Application {
	meta := res.GetMetadata()
	now := time.Now()

	app := &application.Application{
		Metadata_: meta,
		Spec_: application.ApplicationSpec{
			Source: application.ApplicationSource{
				Type: application.SourceGit,
			},
			Runtime: application.ApplicationRuntime{},
			Settings: copyMap(application.DefaultApplicationSettings),
		},
		Status_: application.ApplicationStatus{
			Phase:     application.PhaseCreating,
			Health:    application.HealthHealthy,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	if spec, ok := res.GetSpec().(map[string]interface{}); ok {
		if pid, ok := spec["projectId"].(string); ok {
			app.Spec_.ProjectID = pid
		}
		if source, ok := spec["source"].(map[string]interface{}); ok {
			if url, ok := source["url"].(string); ok {
				app.Spec_.Source.URL = url
			}
			if branch, ok := source["branch"].(string); ok {
				app.Spec_.Source.Branch = branch
			}
			if sType, ok := source["type"].(string); ok {
				app.Spec_.Source.Type = sType
			}
		}
		if runtime, ok := spec["runtime"].(map[string]interface{}); ok {
			if rType, ok := runtime["type"].(string); ok {
				app.Spec_.Runtime.Type = rType
			}
		}
	}

	app.EnsureDefaults()
	return app
}

// copyMap returns a shallow copy of a string-to-string map.
func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
