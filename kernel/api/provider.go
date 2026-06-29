package api

import (
	"net/http"
	"time"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/providers"
)

// ProviderHandler serves provider discovery endpoints. It reads provider
// metadata from the kernel's descriptor registry and wraps every response
// in the standard ResourceObject envelope.
type ProviderHandler struct {
	k *kernel.Kernel
}

// NewProviderHandler creates a handler bound to the given kernel.
func NewProviderHandler(k *kernel.Kernel) *ProviderHandler {
	return &ProviderHandler{k: k}
}

// ---------------------------------------------------------------------------
// Spec and Status types
// ---------------------------------------------------------------------------

// providerSpec is the desired-state portion of a provider resource.
type providerSpec struct {
	DisplayName       string                      `json:"displayName,omitempty"`
	Description       string                      `json:"description"`
	Version           string                      `json:"version"`
	ProviderType      string                      `json:"providerType"`
	Author            string                      `json:"author,omitempty"`
	License           string                      `json:"license,omitempty"`
	Homepage          string                      `json:"homepage,omitempty"`
	DocumentationURL  string                      `json:"documentationUrl,omitempty"`
	SourceRepository  string                      `json:"sourceRepository,omitempty"`
	Tags              []string                    `json:"tags,omitempty"`
	Experimental      bool                        `json:"experimental"`
	Enterprise        bool                        `json:"enterprise"`
	SupportedPlatforms []providers.Platform       `json:"supportedPlatforms,omitempty"`
	Dependencies      []string                    `json:"dependencies,omitempty"`
	Capabilities      []providers.CapabilityClaim `json:"capabilities"`
}

// providerStatus is the runtime-state portion of a provider resource.
type providerStatus struct {
	Status   string `json:"status"`
	Healthy  bool   `json:"healthy"`
	Ready    bool   `json:"ready"`
	Message  string `json:"message,omitempty"`
}

// providerHealthResponse is the payload for the per-provider health endpoint.
type providerHealthResponse struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	State        string            `json:"state"`
	Available    bool              `json:"available"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	LastCheck    time.Time         `json:"lastCheck"`
	Message      string            `json:"message,omitempty"`
}

// providerCapabilitiesResponse wraps a provider's capability claims.
type providerCapabilitiesResponse struct {
	ProviderID   string                      `json:"providerId"`
	ProviderName string                      `json:"providerName"`
	Capabilities []providers.CapabilityClaim `json:"capabilities"`
}

// ---------------------------------------------------------------------------
// GET /api/v1/providers
// ---------------------------------------------------------------------------

// ListProviders returns every registered provider descriptor wrapped in
// the ResourceObject envelope.
func (ph *ProviderHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	reg := ph.k.ProviderDescriptorRegistry()
	descriptors := reg.List()

	items := make([]Object, 0, len(descriptors))
	for _, d := range descriptors {
		items = append(items, providerDescriptorToObject(d))
	}

	list := NewObjectList("Provider", items)
	OK(w, list)
}

// ---------------------------------------------------------------------------
// GET /api/v1/providers/{id}
// ---------------------------------------------------------------------------

// GetProvider returns a single provider descriptor by ID.
func (ph *ProviderHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Provider ID is required")
		return
	}

	reg := ph.k.ProviderDescriptorRegistry()
	d, ok := reg.Get(id)
	if !ok {
		NotFound(w, "PROVIDER_NOT_FOUND", "Provider "+id+" not found")
		return
	}

	obj := providerDescriptorToObject(d)
	OK(w, obj)
}

// ---------------------------------------------------------------------------
// GET /api/v1/providers/{id}/health
// ---------------------------------------------------------------------------

// GetProviderHealth returns health information for a single provider.
func (ph *ProviderHandler) GetProviderHealth(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Provider ID is required")
		return
	}

	reg := ph.k.ProviderDescriptorRegistry()
	d, ok := reg.Get(id)
	if !ok {
		NotFound(w, "PROVIDER_NOT_FOUND", "Provider "+id+" not found")
		return
	}

	// Compute runtime health from the descriptor's status.
	available := d.Status == "ready" || d.Status == "running"
	healthy := d.Status != "failed" && d.Status != "stopped"

	resp := providerHealthResponse{
		Status:    d.Status,
		Version:   d.Version,
		State:     d.Status,
		Available: available,
		LastCheck: time.Now(),
		Message:   "provider is " + d.Status,
	}

	if !healthy {
		resp.Message = "provider is in state: " + d.Status
	}

	OK(w, resp)
}

// ---------------------------------------------------------------------------
// GET /api/v1/providers/{id}/capabilities
// ---------------------------------------------------------------------------

// GetProviderCapabilities returns the capability claims for a single provider.
func (ph *ProviderHandler) GetProviderCapabilities(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "Provider ID is required")
		return
	}

	reg := ph.k.ProviderDescriptorRegistry()
	d, ok := reg.Get(id)
	if !ok {
		NotFound(w, "PROVIDER_NOT_FOUND", "Provider "+id+" not found")
		return
	}

	resp := providerCapabilitiesResponse{
		ProviderID:   d.ID,
		ProviderName: d.Name,
		Capabilities: d.Capabilities,
	}
	OK(w, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// providerDescriptorToObject converts a provider Descriptor into a ResourceObject.
func providerDescriptorToObject(d *providers.Descriptor) Object {
	spec := providerSpec{
		DisplayName:       d.DisplayName,
		Description:       d.Description,
		Version:           d.Version,
		ProviderType:      d.ProviderType,
		Author:            d.Author,
		License:           d.License,
		Homepage:          d.Homepage,
		DocumentationURL:  d.DocumentationURL,
		SourceRepository:  d.SourceRepository,
		Tags:              d.Tags,
		Experimental:      d.Experimental,
		Enterprise:        d.Enterprise,
		SupportedPlatforms: d.SupportedPlatforms,
		Dependencies:      idSliceToStrings(d.Dependencies),
		Capabilities:      d.Capabilities,
	}

	healthy := d.Status != "failed" && d.Status != "stopped"
	ready := d.Status == "ready" || d.Status == "running"

	status := providerStatus{
		Status:  d.Status,
		Healthy: healthy,
		Ready:   ready,
		Message: "provider is " + d.Status,
	}

	return NewObject("Provider", d.ID, d.Name, spec, status)
}

// idSliceToStrings converts a slice of capabilities.ID to a slice of string.
func idSliceToStrings(ids []capabilities.ID) []string {
	if ids == nil {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	return out
}
