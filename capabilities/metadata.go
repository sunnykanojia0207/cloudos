package capabilities

import (
	"fmt"
	"sync"
)

// --- Category constants ------------------------------------------------------

// Category groups capabilities by domain.
type Category string

const (
	CategoryCore       Category = "core"
	CategoryCompute    Category = "compute"
	CategoryStorage    Category = "storage"
	CategoryDatabase   Category = "database"
	CategoryAI         Category = "ai"
	CategoryNetworking Category = "networking"
	CategorySecurity   Category = "security"
	CategoryIntegration Category = "integration"
)

// --- Status constants --------------------------------------------------------

// Status represents the maturity of a capability interface.
type Status string

const (
	StatusStable       Status = "stable"
	StatusExperimental Status = "experimental"
	StatusDeprecated   Status = "deprecated"
)

// --- Descriptor --------------------------------------------------------------

// Descriptor is the canonical metadata for a CloudOS capability. It is
// independent of any provider implementation — it describes what a capability
// *is*, not how it is *backed*.
//
// Every capability in the system MUST have a Descriptor registered in the
// capability metadata registry before it can be discovered through the API.
type Descriptor struct {
	// ID is the well-known, unique identifier (e.g. "compute", "storage").
	ID ID `json:"id" yaml:"id"`

	// Name is a short human-readable label (e.g. "Compute").
	Name string `json:"name" yaml:"name"`

	// DisplayName is a longer human-readable label (e.g. "Compute Engine").
	DisplayName string `json:"displayName" yaml:"display_name"`

	// Description explains what this capability provides.
	Description string `json:"description" yaml:"description"`

	// Version is the semantic version of the capability interface contract.
	Version Version `json:"version" yaml:"version"`

	// Status indicates the maturity of the capability interface.
	Status Status `json:"status" yaml:"status"`

	// Category groups this capability into a domain.
	Category Category `json:"category" yaml:"category"`

	// Tags are arbitrary labels for search, filtering, and UI grouping.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Operations lists every operation this capability exposes.
	Operations []Operation `json:"operations" yaml:"operations"`

	// Dependencies lists capability IDs that this capability requires.
	Dependencies []ID `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// Operation describes a single RPC-style operation exposed by a capability.
type Operation struct {
	// Name is the operation identifier (e.g. "deploy", "exec", "logs").
	Name string `json:"name" yaml:"name"`

	// Description explains what the operation does.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// HTTPMethod is the suggested HTTP verb (GET, POST, PUT, DELETE).
	HTTPMethod string `json:"httpMethod,omitempty" yaml:"http_method,omitempty"`

	// Path is the suggested API path template (e.g. "/deployments/{id}").
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// --- Registry ----------------------------------------------------------------

// Registry stores capability descriptors by ID. It is the source of truth for
// the capability discovery API. The kernel pre-populates it at boot time with
// all known capability interfaces.
type Registry struct {
	mu   sync.RWMutex
	items map[ID]*Descriptor
}

// NewRegistry creates an empty descriptor registry.
func NewRegistry() *Registry {
	return &Registry{
		items: make(map[ID]*Descriptor),
	}
}

// Register adds a descriptor. Returns an error if the ID is already registered.
func (r *Registry) Register(d *Descriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[d.ID]; exists {
		return fmt.Errorf("capability descriptor %q already registered", d.ID)
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
func (r *Registry) Get(id ID) (*Descriptor, bool) {
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
func (r *Registry) IDs() []ID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]ID, 0, len(r.items))
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

// --- Built-in capability descriptors -----------------------------------------

// DefaultDescriptors returns the built-in capability descriptors that ship with
// every CloudOS kernel. These are registered at boot time.
func DefaultDescriptors() []*Descriptor {
	return []*Descriptor{
		{
			ID:          "compute",
			Name:        "Compute",
			DisplayName: "Compute Engine",
			Description: "Deploy, scale, and manage applications and processes. Supports container images, environment variables, scaling, logs, and remote execution.",
			Version:     Version{Major: 1, Minor: 0, Patch: 0},
			Status:      StatusStable,
			Category:    CategoryCompute,
			Tags:        []string{"compute", "deployment", "containers"},
			Operations: []Operation{
				{Name: "deploy", Description: "Create a new deployment", HTTPMethod: "POST", Path: "/deployments"},
				{Name: "get_deployment", Description: "Get a deployment by ID", HTTPMethod: "GET", Path: "/deployments/{id}"},
				{Name: "list_deployments", Description: "List all deployments", HTTPMethod: "GET", Path: "/deployments"},
				{Name: "remove_deployment", Description: "Remove a deployment", HTTPMethod: "DELETE", Path: "/deployments/{id}"},
				{Name: "exec", Description: "Execute a command in a deployment", HTTPMethod: "POST", Path: "/deployments/{id}/exec"},
				{Name: "logs", Description: "Stream logs from a deployment", HTTPMethod: "GET", Path: "/deployments/{id}/logs"},
			},
		},
		{
			ID:          "storage",
			Name:        "Storage",
			DisplayName: "Object Storage",
			Description: "Store, retrieve, and manage objects in buckets. Supports content types, listing, and bucket lifecycle management.",
			Version:     Version{Major: 1, Minor: 0, Patch: 0},
			Status:      StatusStable,
			Category:    CategoryStorage,
			Tags:        []string{"storage", "objects", "buckets"},
			Operations: []Operation{
				{Name: "put", Description: "Store an object", HTTPMethod: "PUT", Path: "/buckets/{bucket}/{key}"},
				{Name: "get", Description: "Retrieve an object", HTTPMethod: "GET", Path: "/buckets/{bucket}/{key}"},
				{Name: "delete", Description: "Delete an object", HTTPMethod: "DELETE", Path: "/buckets/{bucket}/{key}"},
				{Name: "list", Description: "List objects in a bucket", HTTPMethod: "GET", Path: "/buckets/{bucket}"},
				{Name: "create_bucket", Description: "Create a new bucket", HTTPMethod: "POST", Path: "/buckets"},
				{Name: "delete_bucket", Description: "Delete a bucket", HTTPMethod: "DELETE", Path: "/buckets/{bucket}"},
			},
		},
		{
			ID:          "database",
			Name:        "Database",
			DisplayName: "SQL Database",
			Description: "Execute SQL queries, run migrations, and manage database connectivity. Supports raw SQL execution and schema migration workflows.",
			Version:     Version{Major: 1, Minor: 0, Patch: 0},
			Status:      StatusStable,
			Category:    CategoryDatabase,
			Tags:        []string{"database", "sql", "migrations"},
			Operations: []Operation{
				{Name: "exec", Description: "Execute a write query", HTTPMethod: "POST", Path: "/exec"},
				{Name: "query", Description: "Execute a read query", HTTPMethod: "POST", Path: "/query"},
				{Name: "migrate", Description: "Run database migrations", HTTPMethod: "POST", Path: "/migrate"},
				{Name: "ping", Description: "Check database connectivity", HTTPMethod: "GET", Path: "/ping"},
			},
		},
		{
			ID:          "ai",
			Name:        "AI",
			DisplayName: "AI Engine",
			Description: "Interact with large language models and embedding services. Supports chat completions, streaming, embeddings, and model discovery.",
			Version:     Version{Major: 1, Minor: 0, Patch: 0},
			Status:      StatusExperimental,
			Category:    CategoryAI,
			Tags:        []string{"ai", "llm", "embeddings", "chat"},
			Operations: []Operation{
				{Name: "chat", Description: "Send a chat completion request", HTTPMethod: "POST", Path: "/chat"},
				{Name: "embed", Description: "Generate embeddings for text", HTTPMethod: "POST", Path: "/embeddings"},
				{Name: "list_models", Description: "List available AI models", HTTPMethod: "GET", Path: "/models"},
				{Name: "stream", Description: "Stream a chat completion", HTTPMethod: "POST", Path: "/chat/stream"},
			},
		},
		{
			ID:          "network",
			Name:        "Network",
			DisplayName: "Networking",
			Description: "Manage IP addresses, virtual networks, and DNS resolution. Supports network creation, IP allocation, and DNS lookups.",
			Version:     Version{Major: 1, Minor: 0, Patch: 0},
			Status:      StatusExperimental,
			Category:    CategoryNetworking,
			Tags:        []string{"network", "dns", "ip", "vpc"},
			Operations: []Operation{
				{Name: "allocate_ip", Description: "Allocate a new IP address", HTTPMethod: "POST", Path: "/ips"},
				{Name: "release_ip", Description: "Release an IP address", HTTPMethod: "DELETE", Path: "/ips/{ip}"},
				{Name: "create_network", Description: "Create a virtual network", HTTPMethod: "POST", Path: "/networks"},
				{Name: "delete_network", Description: "Delete a virtual network", HTTPMethod: "DELETE", Path: "/networks/{id}"},
				{Name: "resolve_dns", Description: "Resolve a DNS name", HTTPMethod: "GET", Path: "/dns/{name}"},
			},
		},
	}
}
