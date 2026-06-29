package api

import (
	"net/http"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/kernel"
)

// CapabilityHandler serves capability discovery endpoints. It reads
// capability metadata from the kernel's descriptor registry and wraps
// every response in the standard ResourceObject envelope.
type CapabilityHandler struct {
	k *kernel.Kernel
}

// NewCapabilityHandler creates a handler bound to the given kernel.
func NewCapabilityHandler(k *kernel.Kernel) *CapabilityHandler {
	return &CapabilityHandler{k: k}
}

// ---------------------------------------------------------------------------
// Spec and Status types
// ---------------------------------------------------------------------------

// capabilitySpec is the desired-state portion of a capability resource.
type capabilitySpec struct {
	DisplayName  string                   `json:"displayName"`
	Description  string                   `json:"description"`
	Version      capabilities.Version     `json:"version"`
	Category     capabilities.Category    `json:"category"`
	Tags         []string                 `json:"tags,omitempty"`
	Operations   []capabilities.Operation `json:"operations"`
	Dependencies []capabilities.ID        `json:"dependencies,omitempty"`
}

// capabilityStatus is the runtime-state portion of a capability resource.
type capabilityStatus struct {
	Status        capabilities.Status `json:"status"`
	Available     bool                `json:"available"`
	ProviderCount int                 `json:"providerCount"`
}

// ---------------------------------------------------------------------------
// GET /api/v1/capabilities
// ---------------------------------------------------------------------------

// ListCapabilities returns every registered capability descriptor wrapped in
// the ResourceObject envelope.
func (ch *CapabilityHandler) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	reg := ch.k.CapabilityDescriptorRegistry()
	descriptors := reg.List()

	items := make([]Object, 0, len(descriptors))
	for _, d := range descriptors {
		items = append(items, descriptorToObject(d, ch.k))
	}

	list := NewObjectList("Capability", items)
	OK(w, list)
}

// ---------------------------------------------------------------------------
// GET /api/v1/capabilities/{id}
// ---------------------------------------------------------------------------

// GetCapability returns a single capability descriptor by ID.
func (ch *CapabilityHandler) GetCapability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Capability ID is required")
		return
	}

	reg := ch.k.CapabilityDescriptorRegistry()
	d, ok := reg.Get(capabilities.ID(id))
	if !ok {
		NotFound(w, "CAPABILITY_NOT_FOUND", "Capability "+id+" not found")
		return
	}

	obj := descriptorToObject(d, ch.k)
	OK(w, obj)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// descriptorToObject converts a capability Descriptor into a ResourceObject,
// computing the runtime status from the kernel state.
func descriptorToObject(d *capabilities.Descriptor, k *kernel.Kernel) Object {
	// Count providers that implement this capability.
	providerCount := 0
	for _, p := range k.ProviderRegistry().List() {
		if p.Name() == string(d.ID) {
			providerCount++
		}
	}
	available := providerCount > 0

	spec := capabilitySpec{
		DisplayName:  d.DisplayName,
		Description:  d.Description,
		Version:      d.Version,
		Category:     d.Category,
		Tags:         d.Tags,
		Operations:   d.Operations,
		Dependencies: d.Dependencies,
	}

	status := capabilityStatus{
		Status:        d.Status,
		Available:     available,
		ProviderCount: providerCount,
	}

	obj := NewObject("Capability", string(d.ID), d.Name, spec, status)
	obj.Metadata.Labels = map[string]string{
		"category": string(d.Category),
		"status":   string(d.Status),
	}
	return obj
}
