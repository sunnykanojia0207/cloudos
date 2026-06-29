package providers

import (
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/capabilities"
)

// --- Descriptor --------------------------------------------------------------

// Descriptor is the canonical metadata for a CloudOS provider. It describes
// everything about a provider — its identity, capabilities, health,
// configuration, and supported platforms — independently of how it is loaded.
//
// Every provider discovered by the kernel has a corresponding Descriptor
// registered in the provider descriptor registry.
type Descriptor struct {
	// ID is the unique, fully-qualified provider identifier (e.g. "compute.local").
	ID string `json:"id" yaml:"id"`

	// Name is a short human-readable label (e.g. "Compute Local").
	Name string `json:"name" yaml:"name"`

	// DisplayName is a longer human-readable label (e.g. "Local Compute Engine").
	DisplayName string `json:"displayName,omitempty" yaml:"display_name,omitempty"`

	// Description explains what this provider does.
	Description string `json:"description" yaml:"description"`

	// Version is the semantic version of the provider itself (not the capability).
	Version string `json:"version" yaml:"version"`

	// ProviderType describes the deployment model (e.g. "built-in", "plugin", "external").
	ProviderType string `json:"providerType" yaml:"provider_type"`

	// Status indicates the operational state of the provider.
	Status string `json:"status" yaml:"status"`

	// Author is the person or organisation that created the provider.
	Author string `json:"author,omitempty" yaml:"author,omitempty"`

	// License is the SPDX license identifier (e.g. "MIT", "Apache-2.0").
	License string `json:"license,omitempty" yaml:"license,omitempty"`

	// Homepage is a URL to the provider's project website.
	Homepage string `json:"homepage,omitempty" yaml:"homepage,omitempty"`

	// DocumentationURL is a URL to the provider's documentation.
	DocumentationURL string `json:"documentationUrl,omitempty" yaml:"documentation_url,omitempty"`

	// SourceRepository is a URL to the source code repository.
	SourceRepository string `json:"sourceRepository,omitempty" yaml:"source_repository,omitempty"`

	// Tags are arbitrary labels for search and filtering.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Experimental marks this provider as not yet production-ready.
	Experimental bool `json:"experimental" yaml:"experimental"`

	// Enterprise marks this provider as requiring a commercial license.
	Enterprise bool `json:"enterprise" yaml:"enterprise"`

	// SupportedPlatforms lists OS/arch tuples the provider runs on.
	SupportedPlatforms []Platform `json:"supportedPlatforms,omitempty" yaml:"supported_platforms,omitempty"`

	// Dependencies lists capability IDs that this provider depends on.
	Dependencies []capabilities.ID `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// ConfigSchema is the JSON Schema for the provider's configuration.
	ConfigSchema map[string]interface{} `json:"configSchema,omitempty" yaml:"config_schema,omitempty"`

	// Capabilities lists every capability this provider implements, along
	// with negotiation metadata (operations, features, limits, extensions).
	Capabilities []CapabilityClaim `json:"capabilities" yaml:"capabilities"`
}

// ---------------------------------------------------------------------------
// Capability negotiation
// ---------------------------------------------------------------------------

// CapabilityClaim describes how a provider implements a specific capability.
// This is the "capability negotiation" mechanism: it lets the AI, dashboard,
// and automation engine choose the best provider based on what operations
// are supported, what features are available, and what limits apply.
type CapabilityClaim struct {
	// ID is the capability identifier (e.g. "compute", "storage").
	ID capabilities.ID `json:"id" yaml:"id"`

	// Version is the capability interface version implemented.
	Version capabilities.Version `json:"version" yaml:"version"`

	// Operations lists every operation this provider supports for this capability.
	Operations []string `json:"operations" yaml:"operations"`

	// Features are optional feature flags (e.g. "versioning", "encryption", "snapshots").
	Features []string `json:"features,omitempty" yaml:"features,omitempty"`

	// Limits describe resource constraints (e.g. max object size, max connections).
	Limits map[string]string `json:"limits,omitempty" yaml:"limits,omitempty"`

	// Extensions are capability-specific extensions the provider supports.
	Extensions []string `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}

// ---------------------------------------------------------------------------
// Supporting types
// ---------------------------------------------------------------------------

// Platform describes a supported operating system and architecture.
type Platform struct {
	OS   string `json:"os" yaml:"os"`
	Arch string `json:"arch" yaml:"arch"`
}

// HealthInfo carries the runtime health of a single provider.
type HealthInfo struct {
	// Status is the overall health status (e.g. "healthy", "degraded", "unhealthy").
	Status string `json:"status" yaml:"status"`

	// Version is the provider version as reported at runtime.
	Version string `json:"version" yaml:"version"`

	// State is the provider's lifecycle state.
	State string `json:"state" yaml:"state"`

	// Available indicates whether the provider is ready to serve requests.
	Available bool `json:"available" yaml:"available"`

	// Dependencies describes the health of this provider's dependencies.
	Dependencies map[string]string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// LastCheck is when the health was last verified.
	LastCheck time.Time `json:"lastCheck" yaml:"last_check"`

	// Message is a human-readable status description.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

// Registry stores provider descriptors by ID. It is the source of truth for
// the provider discovery API. The kernel pre-populates it at boot time with
// all built-in providers, and plugins register themselves on load.
type Registry struct {
	mu    sync.RWMutex
	items map[string]*Descriptor
}

// NewRegistry creates an empty provider descriptor registry.
func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]*Descriptor),
	}
}

// Register adds a descriptor. Returns an error if the ID is already registered.
func (r *Registry) Register(d *Descriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[d.ID]; exists {
		return fmt.Errorf("provider descriptor %q already registered", d.ID)
	}
	r.items[d.ID] = d
	return nil
}

// RegisterOrReplace adds or replaces a descriptor.
func (r *Registry) RegisterOrReplace(d *Descriptor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[d.ID] = d
}

// Get retrieves a descriptor by ID. Returns nil and false if not found.
func (r *Registry) Get(id string) (*Descriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.items[id]
	return d, ok
}

// List returns all registered descriptors.
func (r *Registry) List() []*Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Descriptor, 0, len(r.items))
	for _, d := range r.items {
		out = append(out, d)
	}
	return out
}

// IDs returns the IDs of all registered descriptors.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.items))
	for id := range r.items {
		out = append(out, id)
	}
	return out
}

// Len returns the number of registered descriptors.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// ---------------------------------------------------------------------------
// Built-in provider descriptors
// ---------------------------------------------------------------------------

// DefaultDescriptors returns the built-in provider descriptors that ship
// with every CloudOS kernel.
func DefaultDescriptors() []*Descriptor {
	return []*Descriptor{
		{
			ID:                "compute.local",
			Name:              "Compute Local",
			DisplayName:       "Local Compute Engine",
			Description:       "Built-in provider that runs deployments as local operating system processes. Suitable for development and single-node deployments.",
			Version:           "0.1.0",
			ProviderType:      "built-in",
			Status:            "ready",
			Author:            "CloudOS Authors",
			License:           "MIT",
			Homepage:          "https://cloudos.io/providers/compute-local",
			DocumentationURL:  "https://cloudos.io/docs/providers/compute-local",
			SourceRepository:  "https://github.com/cloudos/cloudos",
			Tags:              []string{"compute", "local", "development"},
			Experimental:      false,
			Enterprise:        false,
			SupportedPlatforms: []Platform{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
				{OS: "windows", Arch: "amd64"},
			},
			Capabilities: []CapabilityClaim{
				{
					ID:         "compute",
					Version:    capabilities.Version{Major: 1, Minor: 0, Patch: 0},
					Operations: []string{"deploy", "get_deployment", "list_deployments", "remove_deployment", "exec", "logs"},
					Features:   []string{"env_vars", "port_mapping"},
					Limits:     map[string]string{"max_replicas": "1", "max_memory": "512MB"},
					Extensions: []string{"local_only"},
				},
			},
		},
		{
			ID:                "storage.local",
			Name:              "Storage Local",
			DisplayName:       "Local Object Storage",
			Description:       "Built-in provider that stores objects on the local filesystem. Supports buckets, content-type detection, and directory-based organisation.",
			Version:           "0.1.0",
			ProviderType:      "built-in",
			Status:            "ready",
			Author:            "CloudOS Authors",
			License:           "MIT",
			Homepage:          "https://cloudos.io/providers/storage-local",
			DocumentationURL:  "https://cloudos.io/docs/providers/storage-local",
			SourceRepository:  "https://github.com/cloudos/cloudos",
			Tags:              []string{"storage", "local", "filesystem", "development"},
			Experimental:      false,
			Enterprise:        false,
			SupportedPlatforms: []Platform{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
				{OS: "windows", Arch: "amd64"},
			},
			Capabilities: []CapabilityClaim{
				{
					ID:         "storage",
					Version:    capabilities.Version{Major: 1, Minor: 0, Patch: 0},
					Operations: []string{"put", "get", "delete", "list", "create_bucket", "delete_bucket"},
					Features:   []string{"content_type_detection", "directory_buckets"},
					Limits:     map[string]string{"max_object_size": "100MB", "max_buckets": "100"},
					Extensions: []string{"local_only"},
				},
			},
		},
		{
			ID:                "database.sqlite",
			Name:              "Database SQLite",
			DisplayName:       "SQLite Database",
			Description:       "Built-in provider that embeds SQLite for local database operations. Supports WAL mode, raw SQL execution, and schema migrations.",
			Version:           "0.1.0",
			ProviderType:      "built-in",
			Status:            "ready",
			Author:            "CloudOS Authors",
			License:           "MIT",
			Homepage:          "https://cloudos.io/providers/database-sqlite",
			DocumentationURL:  "https://cloudos.io/docs/providers/database-sqlite",
			SourceRepository:  "https://github.com/cloudos/cloudos",
			Tags:              []string{"database", "sqlite", "sql", "embedded", "development"},
			Experimental:      false,
			Enterprise:        false,
			SupportedPlatforms: []Platform{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
				{OS: "windows", Arch: "amd64"},
			},
			Capabilities: []CapabilityClaim{
				{
					ID:         "database",
					Version:    capabilities.Version{Major: 1, Minor: 0, Patch: 0},
					Operations: []string{"exec", "query", "migrate", "ping"},
					Features:   []string{"wal_mode", "foreign_keys", "transactions"},
					Limits:     map[string]string{"max_connections": "1", "max_db_size": "1GB"},
					Extensions: []string{"local_only", "single_user"},
				},
			},
		},
	}
}
