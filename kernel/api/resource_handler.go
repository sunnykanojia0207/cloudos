package api

import (
	"fmt"
	"net/http"

	"github.com/cloudos/cloudos/kernel/resource"
)

// ResourceHandler serves the Resource Engine REST endpoints. It maps between
// the internal resource.Registry and the standard ResourceObject envelope.
type ResourceHandler struct {
	reg *resource.Registry
}

// NewResourceHandler creates a handler bound to the given resource registry.
func NewResourceHandler(reg *resource.Registry) *ResourceHandler {
	return &ResourceHandler{reg: reg}
}

// ── DTOs ───────────────────────────────────────────────────────────────────

// resourceKindDTO is the JSON representation of a registered resource kind.
type resourceKindDTO struct {
	Name       string   `json:"name"`
	Namespaced bool     `json:"namespaced"`
	Versions   []string `json:"versions,omitempty"`
}

// resourceKindListResponse wraps a list of resource kinds.
type resourceKindListResponse struct {
	Kinds []resourceKindDTO `json:"kinds"`
	Total int               `json:"total"`
}

// ── GET /api/v1/resources ──────────────────────────────────────────────────

// ListResourceKinds returns every registered resource kind.
func (rh *ResourceHandler) ListResourceKinds(w http.ResponseWriter, r *http.Request) {
	kinds := rh.reg.ListKinds()
	dtos := make([]resourceKindDTO, 0, len(kinds))
	for _, k := range kinds {
		dtos = append(dtos, resourceKindDTO{
			Name:       k.Name,
			Namespaced: k.Namespaced,
			Versions:   k.Versions,
		})
	}
	OK(w, resourceKindListResponse{
		Kinds: dtos,
		Total: len(dtos),
	})
}

// ── GET /api/v1/resources/{kind} ──────────────────────────────────────────

// ListResources returns every resource of the given kind.
func (rh *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	if kind == "" {
		BadRequest(w, "MISSING_KIND", "Resource kind is required")
		return
	}

	resources, err := rh.reg.List(kind)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	items := make([]Object, 0, len(resources))
	for _, res := range resources {
		items = append(items, resourceToObject(res))
	}

	list := NewObjectList(kind, items)
	OK(w, list)
}

// ── GET /api/v1/resources/{kind}/{id} ─────────────────────────────────────

// GetResource returns a single resource by kind and ID.
func (rh *ResourceHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id := r.PathValue("id")

	if kind == "" {
		BadRequest(w, "MISSING_KIND", "Resource kind is required")
		return
	}
	if id == "" {
		BadRequest(w, "MISSING_ID", "Resource ID is required")
		return
	}

	res, err := rh.reg.Get(kind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	obj := resourceToObject(res)
	OK(w, obj)
}

// ── Helpers ────────────────────────────────────────────────────────────────

// resourceToObject converts an internal Resource to a ResourceObject envelope.
func resourceToObject(res resource.Resource) Object {
	meta := res.GetMetadata()

	om := ObjectMeta{
		ID:              meta.ID,
		Name:            meta.Name,
		Labels:          meta.Labels,
		Annotations:     meta.Annotations,
		CreatedAt:       meta.CreatedAt,
		UpdatedAt:       meta.UpdatedAt,
		ResourceVersion: fmt.Sprintf("%d", meta.ResourceVersion),
	}

	return Object{
		APIVersion: APIVersion,
		Kind:       meta.Kind,
		Metadata:   om,
		Spec:       res.GetSpec(),
		Status:     res.GetStatus(),
	}
}

// resourceErrorToHTTP maps resource package errors to HTTP responses.
func resourceErrorToHTTP(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *resource.ErrKindNotFound:
		NotFound(w, "KIND_NOT_FOUND", e.Error())
	case *resource.ErrResourceNotFound:
		NotFound(w, "RESOURCE_NOT_FOUND", e.Error())
	case *resource.ErrResourceAlreadyExists:
		// This shouldn't happen on reads, but future-proof.
		BadRequest(w, "RESOURCE_ALREADY_EXISTS", e.Error())
	case *resource.ErrInvalidResource:
		BadRequest(w, "INVALID_RESOURCE", e.Error())
	default:
		InternalError(w, "INTERNAL_ERROR", err.Error())
	}
}


