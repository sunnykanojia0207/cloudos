package resource

import (
	"fmt"
	"regexp"
)

// ── Namespace Types ─────────────────────────────────────────────────────────

// NamespaceSpec is the desired state of a CloudOS namespace.
type NamespaceSpec struct {
	// DisplayName is an optional human-readable label.
	DisplayName string `json:"displayName,omitempty"`

	// Description explains the purpose of this namespace.
	Description string `json:"description,omitempty"`
}

// NamespaceStatus is the current state of a namespace.
type NamespaceStatus struct {
	// Phase indicates the namespace lifecycle phase.
	Phase string `json:"phase"` // "Active" or "Terminating"

	// ResourceCount is the number of resources in this namespace.
	ResourceCount int `json:"resourceCount"`
}

// ── Namespace Resource ─────────────────────────────────────────────────────

// Namespace is a concrete Resource for namespace management.
type Namespace struct {
	Metadata_ *Metadata        `json:"metadata"`
	Spec_     NamespaceSpec    `json:"spec"`
	Status_   NamespaceStatus  `json:"status"`
}

// NewNamespace creates a new Namespace resource.
func NewNamespace(id, displayName, description string) *Namespace {
	return &Namespace{
		Metadata_: &Metadata{
			ID:         id,
			Name:       displayName,
			Kind:       "Namespace",
			APIVersion: APIVersion,
			Labels:     map[string]string{},
			Annotations: map[string]string{},
		},
		Spec_: NamespaceSpec{
			DisplayName: displayName,
			Description: description,
		},
		Status_: NamespaceStatus{
			Phase: "Active",
		},
	}
}

func (n *Namespace) GetKind() string        { return "Namespace" }
func (n *Namespace) GetMetadata() *Metadata  { return n.Metadata_ }
func (n *Namespace) GetSpec() interface{}    { return n.Spec_ }
func (n *Namespace) GetStatus() interface{}  { return n.Status_ }
func (n *Namespace) SetStatus(s interface{}) {
	if st, ok := s.(NamespaceStatus); ok {
		n.Status_ = st
	}
}

// ValidateNamespaceID enforces that namespace IDs are valid DNS labels.
var namespaceIDPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func (n *Namespace) Validate() error {
	if n.Metadata_.ID == "" {
		return fmt.Errorf("namespace id is required")
	}
	if !namespaceIDPattern.MatchString(n.Metadata_.ID) {
		return fmt.Errorf("namespace id %q must match %s", n.Metadata_.ID, namespaceIDPattern.String())
	}
	if n.Metadata_.Kind != "Namespace" {
		return fmt.Errorf("kind must be 'Namespace', got %q", n.Metadata_.Kind)
	}
	return nil
}

// ── Default Namespace ──────────────────────────────────────────────────────

// DefaultNamespace returns the default namespace that is created at kernel boot.
func DefaultNamespace() *Namespace {
	return NewNamespace(
		"default",
		"Default",
		"The default CloudOS namespace. All unnamespaced resources belong here.",
	)
}
