package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/kernel/project"
	"github.com/cloudos/cloudos/kernel/resource"
)

// ── Handler ────────────────────────────────────────────────────────────────

// ProjectHandler serves the Project resource REST endpoints.
// It delegates CRUD operations to the existing Resource Engine and adds
// project-specific validation and response formatting.
type ProjectHandler struct {
	k *kernel.Kernel
}

// NewProjectHandler creates a handler bound to the given kernel.
func NewProjectHandler(k *kernel.Kernel) *ProjectHandler {
	return &ProjectHandler{k: k}
}

// ── DTOs ───────────────────────────────────────────────────────────────────

// createProjectRequest is the JSON body for POST /api/v1/projects.
type createProjectRequest struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Environment string `json:"environment"`
}

// updateProjectRequest is the JSON body for PUT /api/v1/projects/{id}.
type updateProjectRequest struct {
	DisplayName     string            `json:"displayName,omitempty"`
	Description     string            `json:"description,omitempty"`
	Environment     string            `json:"environment,omitempty"`
	DefaultRegion   string            `json:"defaultRegion,omitempty"`
	DefaultProviders []string          `json:"defaultProviders,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Settings        map[string]string `json:"settings,omitempty"`
}

// ---------------------------------------------------------------------------
// GET /api/v1/projects
// ---------------------------------------------------------------------------

// ListProjects returns every Project resource.
func (ph *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	reg := ph.k.ResourceRegistry()
	resources, err := reg.List(project.Kind)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	items := make([]Object, 0, len(resources))
	for _, res := range resources {
		items = append(items, projectResourceToObject(res))
	}

	list := NewObjectList(project.Kind, items)
	OK(w, list)
}

// ---------------------------------------------------------------------------
// POST /api/v1/projects
// ---------------------------------------------------------------------------

// CreateProject creates a new Project resource.
func (ph *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "INVALID_JSON", "Invalid request body: "+err.Error())
		return
	}

	if req.ID == "" {
		BadRequest(w, "MISSING_ID", "Project ID is required")
		return
	}
	if req.DisplayName == "" {
		BadRequest(w, "MISSING_NAME", "Project display name is required")
		return
	}

	proj := project.NewProject(req.ID, req.DisplayName, req.Environment, req.Description)
	proj.EnsureDefaults()

	reg := ph.k.ResourceRegistry()
	if err := reg.Create(r.Context(), proj); err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	// The Controller Runtime will reconcile the project asynchronously,
	// setting its phase to Active. Return the created resource immediately.
	obj := projectResourceToObject(proj)
	Created(w, obj)
}

// ---------------------------------------------------------------------------
// GET /api/v1/projects/{id}
// ---------------------------------------------------------------------------

// GetProject returns a single Project by ID.
func (ph *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Project ID is required")
		return
	}

	reg := ph.k.ResourceRegistry()
	res, err := reg.Get(project.Kind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	obj := projectResourceToObject(res)
	OK(w, obj)
}

// ---------------------------------------------------------------------------
// PUT /api/v1/projects/{id}
// ---------------------------------------------------------------------------

// UpdateProject updates an existing Project resource.
func (ph *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Project ID is required")
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "INVALID_JSON", "Invalid request body: "+err.Error())
		return
	}

	reg := ph.k.ResourceRegistry()
	existing, err := reg.Get(project.Kind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	proj, ok := existing.(*project.Project)
	if !ok {
		// If it's a GenericResource, convert it.
		proj = genericToProject(existing)
	}

	// Apply updates.
	if req.DisplayName != "" {
		proj.Spec_.DisplayName = req.DisplayName
		proj.Metadata_.Name = req.DisplayName
	}
	if req.Description != "" {
		proj.Spec_.Description = req.Description
	}
	if req.Environment != "" {
		proj.Spec_.Environment = req.Environment
	}
	if req.DefaultRegion != "" {
		proj.Spec_.DefaultRegion = req.DefaultRegion
	}
	if req.DefaultProviders != nil {
		proj.Spec_.DefaultProviders = req.DefaultProviders
	}
	if req.Tags != nil {
		proj.Spec_.Tags = req.Tags
	}
	if req.Settings != nil {
		for k, v := range req.Settings {
			if proj.Spec_.Settings == nil {
				proj.Spec_.Settings = make(map[string]string)
			}
			proj.Spec_.Settings[k] = v
		}
	}

	proj.EnsureDefaults()
	proj.Touch()

	if err := reg.Update(r.Context(), proj); err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	obj := projectResourceToObject(proj)
	OK(w, obj)
}

// ---------------------------------------------------------------------------
// DELETE /api/v1/projects/{id}
// ---------------------------------------------------------------------------

// DeleteProject removes a Project resource.
func (ph *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Project ID is required")
		return
	}

	reg := ph.k.ResourceRegistry()

	// Before deleting, set the project phase to Deleting.
	existing, err := reg.Get(project.Kind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	proj, ok := existing.(*project.Project)
	if ok {
		proj.Status_.Phase = project.PhaseDeleting
		proj.Touch()
		_ = reg.Update(r.Context(), proj)
	}

	if err := reg.Delete(r.Context(), project.Kind, id); err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	NoContent(w)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// projectResourceToObject converts a resource.Resource to a ResourceObject
// with Project-typed spec and status for consistent API responses.
func projectResourceToObject(res resource.Resource) Object {
	meta := res.GetMetadata()

	spec, ok := res.GetSpec().(project.ProjectSpec)
	if !ok {
		spec = project.ProjectSpec{
			DisplayName: meta.Name,
			Environment: project.EnvDevelopment,
		}
	}

	status, ok := res.GetStatus().(project.ProjectStatus)
	if !ok {
		status = project.ProjectStatus{
			Phase:  project.PhaseActive,
			Health: project.HealthHealthy,
		}
	}

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
		Kind:       project.Kind,
		Metadata:   om,
		Spec:       spec,
		Status:     status,
	}
}

// genericToProject converts a GenericResource to a typed Project.
// This handles the case where a Project is stored as a GenericResource
// (e.g., from an earlier version).
func genericToProject(res resource.Resource) *project.Project {
	meta := res.GetMetadata()
	now := time.Now()

	p := &project.Project{
		Metadata_: meta,
		Spec_: project.ProjectSpec{
			DisplayName: meta.Name,
			Environment: project.EnvDevelopment,
			Settings:    make(map[string]string),
		},
		Status_: project.ProjectStatus{
			Phase:        project.PhaseActive,
			Health:       project.HealthHealthy,
			LastActivity: now,
		},
	}

	if spec, ok := res.GetSpec().(map[string]interface{}); ok {
		if env, ok := spec["environment"].(string); ok {
			p.Spec_.Environment = env
		}
		if desc, ok := spec["description"].(string); ok {
			p.Spec_.Description = desc
		}
	}

	p.EnsureDefaults()
	return p
}

// formatResourceVersion converts uint64 to string for the ObjectMeta.
func formatResourceVersion(v uint64) string {
	if v == 0 {
		return "1"
	}
	return itoa(v)
}

// itoa is a simple uint64-to-string for use when strconv is not imported.
func itoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
