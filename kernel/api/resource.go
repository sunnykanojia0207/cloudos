package api

import "time"

// APIVersion is the current CloudOS API group + version string.
const APIVersion = "cloudos.io/v1"

// Object is the standard resource envelope for every CloudOS API resource.
// It follows a Kubernetes-style pattern: every managed object shares a common
// set of identity and metadata fields, while the type-specific payload lives
// in Spec and the runtime state lives in Status.
//
// This convention ensures that every CloudOS API response — whether for
// capabilities, providers, plugins, deployments, databases, or projects —
// has a uniform structure that CLIs, SDKs, the dashboard, and AI can
// consume generically.
type Object struct {
	// APIVersion is the API group + version (e.g. "cloudos.io/v1").
	APIVersion string `json:"apiVersion"`

	// Kind is the PascalCase resource type (e.g. "Capability", "Provider").
	Kind string `json:"kind"`

	// Metadata carries identity, labels, annotations, and timestamps.
	Metadata ObjectMeta `json:"metadata"`

	// Spec is the desired state / specification of the resource.
	Spec interface{} `json:"spec"`

	// Status is the current runtime state of the resource.
	Status interface{} `json:"status,omitempty"`
}

// ObjectMeta is the standard identity and metadata for every CloudOS resource.
type ObjectMeta struct {
	// ID is the unique resource identifier within its kind (e.g. "compute.local").
	ID string `json:"id"`

	// Name is a short human-readable label (e.g. "Compute Local").
	Name string `json:"name"`

	// Labels are arbitrary key-value pairs for filtering and grouping.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are non-identifying metadata.
	Annotations map[string]string `json:"annotations,omitempty"`

	// CreatedAt is when the resource was first created.
	CreatedAt time.Time `json:"createdAt,omitempty"`

	// UpdatedAt is when the resource was last modified.
	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	// ResourceVersion is an opaque version identifier that changes on every
	// write. Useful for optimistic concurrency and watch cursors.
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// ObjectList is the standard envelope for paginated lists of resources.
type ObjectList struct {
	// APIVersion is the API group + version.
	APIVersion string `json:"apiVersion"`

	// Kind is the PascalCase resource type with a "List" suffix.
	Kind string `json:"kind"`

	// Metadata contains pagination information.
	Metadata ListMeta `json:"metadata"`

	// Items is the list of resources.
	Items []Object `json:"items"`
}

// ListMeta carries pagination metadata.
type ListMeta struct {
	Total int `json:"total"`
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// NewObject creates a resource Object with the given identity fields.
// The apiVersion is set to "cloudos.io/v1" and CreatedAt to the current time.
func NewObject(kind, id, name string, spec, status interface{}) Object {
	return Object{
		APIVersion: APIVersion,
		Kind:       kind,
		Metadata: ObjectMeta{
			ID:        id,
			Name:      name,
			CreatedAt: time.Now(),
		},
		Spec:   spec,
		Status: status,
	}
}

// NewObjectList creates an ObjectList envelope with the given items.
func NewObjectList(kind string, items []Object) ObjectList {
	return ObjectList{
		APIVersion: APIVersion,
		Kind:       kind + "List",
		Items:      items,
		Metadata: ListMeta{
			Total: len(items),
		},
	}
}
