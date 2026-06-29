package api

import (
	"net/http"
	"time"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/kernel/controller"
)

// ── Handler ────────────────────────────────────────────────────────────────

// ControllerHandler serves the controller runtime REST endpoints.
type ControllerHandler struct {
	k *kernel.Kernel
}

// NewControllerHandler creates a handler bound to the given kernel.
func NewControllerHandler(k *kernel.Kernel) *ControllerHandler {
	return &ControllerHandler{k: k}
}

// ── DTOs ───────────────────────────────────────────────────────────────────

// controllerDTO is the JSON representation of a registered controller.
type controllerDTO struct {
	Name    string            `json:"name"`
	Kind    string            `json:"kind"`
	State   string            `json:"state"`
	Message string            `json:"message,omitempty"`
	Health  controllerHealthDTO `json:"health,omitempty"`
}

// controllerHealthDTO is the health details for a controller.
type controllerHealthDTO struct {
	Name            string    `json:"name"`
	Kind            string    `json:"kind"`
	State           string    `json:"state"`
	Message         string    `json:"message,omitempty"`
	LastReconciled  time.Time `json:"lastReconciled,omitempty"`
	ReconcileCount  uint64    `json:"reconcileCount"`
	ErrorCount      uint64    `json:"errorCount"`
}

// controllerListResponse wraps a list of controllers.
type controllerListResponse struct {
	Controllers []controllerDTO `json:"controllers"`
	Total       int             `json:"total"`
}

// ── GET /api/v1/controllers ────────────────────────────────────────────────

// ListControllers returns every registered controller.
func (ch *ControllerHandler) ListControllers(w http.ResponseWriter, r *http.Request) {
	mgr := ch.k.ControllerManager()
	names := mgr.ControllerNames()

	dtos := make([]controllerDTO, 0, len(names))
	for _, name := range names {
		ctrl, ok := mgr.Get(name)
		if !ok {
			continue
		}
		h, ok := mgr.ControllerHealth(name)
		_ = ok

		dtos = append(dtos, controllerDTO{
			Name:    ctrl.Name(),
			Kind:    ctrl.Kind(),
			State:   h.State,
			Message: h.Message,
			Health: controllerHealthDTO{
				Name:           h.Name,
				Kind:           h.Kind,
				State:          h.State,
				Message:        h.Message,
				LastReconciled: h.LastReconciled,
				ReconcileCount: h.ReconcileCount,
				ErrorCount:     h.ErrorCount,
			},
		})
	}

	OK(w, controllerListResponse{
		Controllers: dtos,
		Total:       len(dtos),
	})
}

// ── GET /api/v1/controllers/{id} ──────────────────────────────────────────

// GetController returns a single controller by name.
func (ch *ControllerHandler) GetController(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Controller ID is required")
		return
	}

	mgr := ch.k.ControllerManager()
	ctrl, ok := mgr.Get(id)
	if !ok {
		NotFound(w, "CONTROLLER_NOT_FOUND", "Controller "+id+" not found")
		return
	}

	h, ok := mgr.ControllerHealth(id)
	_ = ok

	dto := controllerDTO{
		Name:    ctrl.Name(),
		Kind:    ctrl.Kind(),
		State:   h.State,
		Message: h.Message,
		Health: controllerHealthDTO{
			Name:           h.Name,
			Kind:           h.Kind,
			State:          h.State,
			Message:        h.Message,
			LastReconciled: h.LastReconciled,
			ReconcileCount: h.ReconcileCount,
			ErrorCount:     h.ErrorCount,
		},
	}

	OK(w, dto)
}

// ── GET /api/v1/controllers/{id}/health ───────────────────────────────────

// GetControllerHealth returns the health snapshot for a single controller.
func (ch *ControllerHandler) GetControllerHealth(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Controller ID is required")
		return
	}

	mgr := ch.k.ControllerManager()
	_, ok := mgr.Get(id)
	if !ok {
		NotFound(w, "CONTROLLER_NOT_FOUND", "Controller "+id+" not found")
		return
	}

	h, ok := mgr.ControllerHealth(id)
	if !ok {
		NotFound(w, "CONTROLLER_HEALTH_NOT_FOUND", "Health for controller "+id+" not found")
		return
	}

	dto := controllerHealthDTO{
		Name:           h.Name,
		Kind:           h.Kind,
		State:          h.State,
		Message:        h.Message,
		LastReconciled: h.LastReconciled,
		ReconcileCount: h.ReconcileCount,
		ErrorCount:     h.ErrorCount,
	}

	OK(w, dto)
}

// controllerToObject converts a controller + health to a ResourceObject
// envelope for use in GenericResource-style responses. (Not used in the
// current endpoints, but available for future ResourceObject consistency.)
func controllerToObject(c *kernel.Kernel, name string) (Object, bool) {
	mgr := c.ControllerManager()
	ctrl, ok := mgr.Get(name)
	if !ok {
		return Object{}, false
	}

	h, ok := mgr.ControllerHealth(name)
	if !ok {
		h = controller.ControllerHealth{
			Name:  ctrl.Name(),
			Kind:  ctrl.Kind(),
			State: "unknown",
		}
	}

	spec := map[string]interface{}{
		"name": ctrl.Name(),
		"kind": ctrl.Kind(),
	}

	status := map[string]interface{}{
		"state":           h.State,
		"message":         h.Message,
		"lastReconciled":  h.LastReconciled,
		"reconcileCount":  h.ReconcileCount,
		"errorCount":      h.ErrorCount,
	}

	return NewObject("Controller", ctrl.Name(), ctrl.Name(), spec, status), true
}
