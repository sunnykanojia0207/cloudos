package resource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Kind Registration Options ──────────────────────────────────────────────

// KindOption configures a registered resource kind.
type KindOption func(*KindRegistration)

// WithValidator registers a custom validation function for resources of this kind.
func WithValidator(fn func(Resource) error) KindOption {
	return func(kr *KindRegistration) { kr.Validator = fn }
}

// WithHooks registers lifecycle hooks for this kind.
func WithHooks(onCreate, onUpdate, onDelete func(Resource) error) KindOption {
	return func(kr *KindRegistration) {
		kr.OnCreate = onCreate
		kr.OnUpdate = onUpdate
		kr.OnDelete = onDelete
	}
}

// WithNamespace sets whether this kind is namespaced.
func WithNamespace(namespaced bool) KindOption {
	return func(kr *KindRegistration) { kr.Kind.Namespaced = namespaced }
}

// ── KindRegistration ────────────────────────────────────────────────────────

// KindRegistration holds the metadata and hooks for a registered resource kind.
type KindRegistration struct {
	Kind     Kind
	Validator func(Resource) error
	OnCreate  func(Resource) error
	OnUpdate  func(Resource) error
	OnDelete  func(Resource) error
}

// ── Store ──────────────────────────────────────────────────────────────────

// Store holds all resource instances of a single kind. It is concurrency-safe.
type Store struct {
	mu       sync.RWMutex
	kind     string
	items    map[string]Resource
	revision uint64
}

func newStore(kind string) *Store {
	return &Store{
		kind:  kind,
		items: make(map[string]Resource),
	}
}

func (s *Store) get(id string) (Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.items[id]
	return r, ok
}

func (s *Store) list() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Resource, 0, len(s.items))
	for _, r := range s.items {
		out = append(out, r)
	}
	return out
}

func (s *Store) create(r Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := r.GetMetadata().ID
	if _, exists := s.items[id]; exists {
		return &ErrResourceAlreadyExists{Kind: s.kind, ID: id}
	}
	s.revision++
	r.GetMetadata().ResourceVersion = s.revision
	r.GetMetadata().CreatedAt = time.Now()
	r.GetMetadata().UpdatedAt = r.GetMetadata().CreatedAt
	s.items[id] = r
	return nil
}

func (s *Store) update(r Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := r.GetMetadata().ID
	existing, ok := s.items[id]
	if !ok {
		return &ErrResourceNotFound{Kind: s.kind, ID: id}
	}
	// Preserve the original creation timestamp.
	r.GetMetadata().CreatedAt = existing.GetMetadata().CreatedAt
	s.revision++
	r.GetMetadata().ResourceVersion = s.revision
	r.GetMetadata().UpdatedAt = time.Now()
	s.items[id] = r
	return nil
}

func (s *Store) delete(id string) (Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.items[id]
	if !ok {
		return nil, &ErrResourceNotFound{Kind: s.kind, ID: id}
	}
	delete(s.items, id)
	return existing, nil
}

// ── Registry ───────────────────────────────────────────────────────────────

// Registry is the central resource registry for CloudOS. It manages resource
// kind registration, instance CRUD, lifecycle hooks, and event publishing.
//
// Usage:
//
//	reg := resource.NewRegistry(bus, log)
//	reg.RegisterKind(resource.Kind{Name: "Namespace"}, resource.WithNamespace(false))
//	ns := resource.NewGenericResource("Namespace", "default", "Default", spec, status)
//	reg.Create(ns)
type Registry struct {
	mu     sync.RWMutex
	kinds  map[string]*KindRegistration
	stores map[string]*Store
	bus    *events.Bus
	log    *logging.Logger
}

// NewRegistry creates an empty resource registry.
func NewRegistry(bus *events.Bus, log *logging.Logger) *Registry {
	return &Registry{
		kinds:  make(map[string]*KindRegistration),
		stores: make(map[string]*Store),
		bus:    bus,
		log:    log.WithContext(context.Background()),
	}
}

// ── Kind Management ────────────────────────────────────────────────────────

// RegisterKind registers a new resource kind. Returns an error if the kind
// is already registered.
func (r *Registry) RegisterKind(kind Kind, opts ...KindOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := kind.Name
	if _, exists := r.kinds[name]; exists {
		return fmt.Errorf("resource kind %q already registered", name)
	}

	kr := &KindRegistration{Kind: kind}
	for _, opt := range opts {
		opt(kr)
	}
	r.kinds[name] = kr
	r.stores[name] = newStore(name)

	r.log.Info("resource kind registered",
		"kind", name,
		"namespaced", kind.Namespaced,
	)
	return nil
}

// GetKind returns the registered Kind by name.
func (r *Registry) GetKind(name string) (Kind, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	kr, ok := r.kinds[name]
	if !ok {
		return Kind{}, false
	}
	return kr.Kind, true
}

// ListKinds returns all registered resource kinds.
func (r *Registry) ListKinds() []Kind {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Kind, 0, len(r.kinds))
	for _, kr := range r.kinds {
		out = append(out, kr.Kind)
	}
	return out
}

// ── CRUD Operations ────────────────────────────────────────────────────────

// Create stores a new resource and publishes a resource.created event.
// If the kind has an OnCreate hook, it is invoked before storage.
func (r *Registry) Create(ctx context.Context, res Resource) error {
	kind := res.GetKind()
	meta := res.GetMetadata()

	kr, store, err := r.getKindAndStore(kind)
	if err != nil {
		return err
	}

	// Set metadata defaults.
	meta.Kind = kind
	meta.APIVersion = APIVersion

	// Validate.
	if err := r.validate(kr, res); err != nil {
		return err
	}

	// Lifecycle hook.
	if kr.OnCreate != nil {
		if err := kr.OnCreate(res); err != nil {
			return fmt.Errorf("onCreate hook for %s/%s: %w", kind, meta.ID, err)
		}
	}

	// Store.
	if err := store.create(res); err != nil {
		return err
	}

	r.publish(ctx, EventResourceCreated, res)
	r.log.Info("resource created", "kind", kind, "id", meta.ID)
	return nil
}

// Get retrieves a resource by kind and ID.
func (r *Registry) Get(kind, id string) (Resource, error) {
	_, store, err := r.getKindAndStore(kind)
	if err != nil {
		return nil, err
	}
	res, ok := store.get(id)
	if !ok {
		return nil, &ErrResourceNotFound{Kind: kind, ID: id}
	}
	return res, nil
}

// List returns all resources of a given kind.
func (r *Registry) List(kind string) ([]Resource, error) {
	_, store, err := r.getKindAndStore(kind)
	if err != nil {
		return nil, err
	}
	return store.list(), nil
}

// Update replaces a resource and publishes a resource.updated event.
// If the kind has an OnUpdate hook, it is invoked before storage.
func (r *Registry) Update(ctx context.Context, res Resource) error {
	kind := res.GetKind()
	meta := res.GetMetadata()

	kr, store, err := r.getKindAndStore(kind)
	if err != nil {
		return err
	}

	if err := r.validate(kr, res); err != nil {
		return err
	}

	if kr.OnUpdate != nil {
		if err := kr.OnUpdate(res); err != nil {
			return fmt.Errorf("onUpdate hook for %s/%s: %w", kind, meta.ID, err)
		}
	}

	if err := store.update(res); err != nil {
		return err
	}

	r.publish(ctx, EventResourceUpdated, res)
	r.log.Info("resource updated", "kind", kind, "id", meta.ID)
	return nil
}

// Delete removes a resource and publishes a resource.deleted event.
// If the kind has an OnDelete hook, it is invoked before removal.
func (r *Registry) Delete(ctx context.Context, kind, id string) error {
	kr, store, err := r.getKindAndStore(kind)
	if err != nil {
		return err
	}

	// Get the resource first so we can pass it to the hook and event.
	var res Resource
	if existing, ok := store.get(id); !ok {
		return &ErrResourceNotFound{Kind: kind, ID: id}
	} else {
		res = existing
	}

	if kr.OnDelete != nil {
		if err := kr.OnDelete(res); err != nil {
			return fmt.Errorf("onDelete hook for %s/%s: %w", kind, id, err)
		}
	}

	if _, err := store.delete(id); err != nil {
		return err
	}

	r.publish(ctx, EventResourceDeleted, res)
	r.log.Info("resource deleted", "kind", kind, "id", id)
	return nil
}

// ── Internal ───────────────────────────────────────────────────────────────

func (r *Registry) getKindAndStore(kind string) (*KindRegistration, *Store, error) {
	r.mu.RLock()
	kr, ok := r.kinds[kind]
	store := r.stores[kind]
	r.mu.RUnlock()
	if !ok {
		return nil, nil, &ErrKindNotFound{Kind: kind}
	}
	return kr, store, nil
}

func (r *Registry) validate(kr *KindRegistration, res Resource) error {
	// Run the built-in Validate method.
	if err := res.Validate(); err != nil {
		return &ErrInvalidResource{
			Kind:    res.GetKind(),
			ID:      res.GetMetadata().ID,
			Details: err.Error(),
		}
	}
	// Run the kind-specific validator if one is registered.
	if kr.Validator != nil {
		if err := kr.Validator(res); err != nil {
			return &ErrInvalidResource{
				Kind:    res.GetKind(),
				ID:      res.GetMetadata().ID,
				Details: err.Error(),
			}
		}
	}
	return nil
}

func (r *Registry) publish(ctx context.Context, eventType string, res Resource) {
	if r.bus == nil {
		return
	}
	r.bus.Publish(ctx, events.Event{
		Type:   eventType,
		Source: "resource.registry",
		Payload: map[string]interface{}{
			"kind":            res.GetKind(),
			"id":              res.GetMetadata().ID,
			"name":            res.GetMetadata().Name,
			"namespace":       res.GetMetadata().Namespace,
			"resourceVersion": res.GetMetadata().ResourceVersion,
		},
	})
}
