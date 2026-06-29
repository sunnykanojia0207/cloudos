// Package resource provides the CloudOS Resource Engine — a generic,
// Kubernetes-style resource management layer. Every object managed by
// CloudOS (projects, deployments, databases, secrets, namespaces, etc.)
// is a Resource with a uniform lifecycle: Create, Read, Update, Delete,
// List, Watch, Validate.
//
// The Resource Engine is the foundation for all higher-level resource types.
// It is NOT an ORM — it is a runtime registry with lifecycle hooks,
// event publishing, watch support, and validation.
package resource

import (
	"fmt"
	"time"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// APIVersion is the current resource API version.
	APIVersion = "cloudos.io/v1"

	// NamespaceDefault is the name of the default namespace, created at boot.
	NamespaceDefault = "default"
)

// ── Kind ───────────────────────────────────────────────────────────────────

// Kind describes a registered resource type. Every resource kind must be
// registered before any resource instances of that kind can be created.
type Kind struct {
	// Name is the PascalCase resource type (e.g. "Namespace", "Project").
	Name string `json:"name"`

	// Namespaced indicates whether resources of this kind belong to a namespace.
	Namespaced bool `json:"namespaced"`

	// Versions lists the supported API versions for this kind.
	Versions []string `json:"versions,omitempty"`
}

// ── Metadata ───────────────────────────────────────────────────────────────

// Metadata is the standard identity and metadata block for every CloudOS
// resource. It follows the Kubernetes metadata convention.
type Metadata struct {
	// ID is the unique resource identifier within its kind/namespace scope.
	ID string `json:"id"`

	// Name is a short human-readable label.
	Name string `json:"name"`

	// Namespace is the namespace this resource belongs to.
	Namespace string `json:"namespace,omitempty"`

	// Kind is the PascalCase resource type.
	Kind string `json:"kind"`

	// APIVersion is the API group + version.
	APIVersion string `json:"apiVersion"`

	// Labels are arbitrary key-value pairs for filtering and grouping.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are non-identifying metadata.
	Annotations map[string]string `json:"annotations,omitempty"`

	// OwnerReferences links this resource to its parent resource(s).
	OwnerReferences []OwnerReference `json:"ownerReferences,omitempty"`

	// Finalizers block deletion until removed.
	Finalizers []string `json:"finalizers,omitempty"`

	// CreatedAt is when the resource was first created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is when the resource was last modified.
	UpdatedAt time.Time `json:"updatedAt"`

	// ResourceVersion is an opaque version counter that changes on every write.
	// Used for optimistic concurrency and watch cursors.
	ResourceVersion uint64 `json:"resourceVersion"`
}

// OwnerReference links a resource to its owner.
type OwnerReference struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// ── Resource Interface ─────────────────────────────────────────────────────

// Resource is the interface that every CloudOS managed object implements.
// It provides uniform access to identity, spec, status, and validation.
type Resource interface {
	// GetKind returns the PascalCase resource type (e.g. "Namespace").
	GetKind() string

	// GetMetadata returns the resource's metadata block.
	GetMetadata() *Metadata

	// GetSpec returns the resource's desired-state payload.
	GetSpec() interface{}

	// GetStatus returns the resource's current runtime state.
	GetStatus() interface{}

	// SetStatus replaces the resource's runtime state.
	SetStatus(interface{})

	// Validate checks the resource for semantic correctness.
	Validate() error
}

// ── GenericResource ─────────────────────────────────────────────────────────

// GenericResource is a concrete Resource implementation that stores spec
// and status as interface{}. It is used for ad-hoc resource kinds that do
// not define their own typed struct.
type GenericResource struct {
	Metadata_ *Metadata    `json:"metadata"`
	Spec_     interface{}  `json:"spec"`
	Status_   interface{}  `json:"status,omitempty"`
}

// NewGenericResource creates a fully populated GenericResource with sensible
// defaults for timestamps and the resource version.
func NewGenericResource(kind, id, name string, spec, status interface{}) *GenericResource {
	now := time.Now()
	return &GenericResource{
		Metadata_: &Metadata{
			ID:              id,
			Name:            name,
			Namespace:       NamespaceDefault,
			Kind:            kind,
			APIVersion:      APIVersion,
			CreatedAt:       now,
			UpdatedAt:       now,
			ResourceVersion: 1,
		},
		Spec_:   spec,
		Status_: status,
	}
}

func (r *GenericResource) GetKind() string           { return r.Metadata_.Kind }
func (r *GenericResource) GetMetadata() *Metadata     { return r.Metadata_ }
func (r *GenericResource) GetSpec() interface{}       { return r.Spec_ }
func (r *GenericResource) GetStatus() interface{}     { return r.Status_ }
func (r *GenericResource) SetStatus(s interface{})    { r.Status_ = s }
func (r *GenericResource) Validate() error            { return nil }

// ── Errors ─────────────────────────────────────────────────────────────────

// ErrKindNotFound is returned when an operation refers to an unregistered kind.
type ErrKindNotFound struct {
	Kind string
}

func (e *ErrKindNotFound) Error() string {
	return fmt.Sprintf("resource kind %q not found", e.Kind)
}

// ErrResourceNotFound is returned when a specific resource is not found.
type ErrResourceNotFound struct {
	Kind string
	ID   string
}

func (e *ErrResourceNotFound) Error() string {
	return fmt.Sprintf("resource %s/%s not found", e.Kind, e.ID)
}

// ErrResourceAlreadyExists is returned on duplicate create.
type ErrResourceAlreadyExists struct {
	Kind string
	ID   string
}

func (e *ErrResourceAlreadyExists) Error() string {
	return fmt.Sprintf("resource %s/%s already exists", e.Kind, e.ID)
}

// ErrInvalidResource is returned when validation fails.
type ErrInvalidResource struct {
	Kind    string
	ID      string
	Details string
}

func (e *ErrInvalidResource) Error() string {
	return fmt.Sprintf("invalid resource %s/%s: %s", e.Kind, e.ID, e.Details)
}
