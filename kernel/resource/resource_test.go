package resource_test

import (
	"context"
	"testing"

	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

func testRegistry(t *testing.T) *resource.Registry {
	t.Helper()
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	bus := events.NewBus(log)
	bus.Start()
	reg := resource.NewRegistry(bus, log)
	return reg
}

// nopWriter discards log output.
type nopWriter struct{}

func (n *nopWriter) Write(p []byte) (int, error) { return len(p), nil }

// ── Kind Registration ──────────────────────────────────────────────────────

func TestRegistry_RegisterKind(t *testing.T) {
	reg := testRegistry(t)

	err := reg.RegisterKind(resource.Kind{Name: "TestItem", Namespaced: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Duplicate registration should fail.
	err = reg.RegisterKind(resource.Kind{Name: "TestItem"})
	if err == nil {
		t.Fatal("expected error for duplicate kind registration")
	}
}

func TestRegistry_GetKind(t *testing.T) {
	reg := testRegistry(t)

	reg.RegisterKind(resource.Kind{Name: "Widget", Namespaced: false})

	kind, ok := reg.GetKind("Widget")
	if !ok {
		t.Fatal("expected to find kind")
	}
	if kind.Name != "Widget" {
		t.Fatalf("expected Widget, got %s", kind.Name)
	}

	_, ok = reg.GetKind("Nonexistent")
	if ok {
		t.Fatal("expected not to find nonexistent kind")
	}
}

func TestRegistry_ListKinds(t *testing.T) {
	reg := testRegistry(t)

	kinds := reg.ListKinds()
	if len(kinds) != 0 {
		t.Fatalf("expected 0 kinds, got %d", len(kinds))
	}

	reg.RegisterKind(resource.Kind{Name: "A"})
	reg.RegisterKind(resource.Kind{Name: "B"})

	kinds = reg.ListKinds()
	if len(kinds) != 2 {
		t.Fatalf("expected 2 kinds, got %d", len(kinds))
	}
}

// ── CRUD ───────────────────────────────────────────────────────────────────

func TestRegistry_CreateAndGet(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	reg.RegisterKind(resource.Kind{Name: "Item"})

	spec := map[string]string{"hello": "world"}
	res := resource.NewGenericResource("Item", "item-1", "Item One", spec, nil)

	if err := reg.Create(ctx, res); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := reg.Get("Item", "item-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.GetMetadata().ID != "item-1" {
		t.Fatalf("expected item-1, got %s", got.GetMetadata().ID)
	}
	if got.GetMetadata().Name != "Item One" {
		t.Fatalf("expected Item One, got %s", got.GetMetadata().Name)
	}
	if got.GetMetadata().Kind != "Item" {
		t.Fatalf("expected Item, got %s", got.GetMetadata().Kind)
	}
	if got.GetMetadata().APIVersion != "cloudos.io/v1" {
		t.Fatalf("expected cloudos.io/v1, got %s", got.GetMetadata().APIVersion)
	}
	if got.GetMetadata().ResourceVersion != 1 {
		t.Fatalf("expected ResourceVersion 1, got %d", got.GetMetadata().ResourceVersion)
	}
}

func TestRegistry_CreateDuplicate(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	reg.RegisterKind(resource.Kind{Name: "Item"})

	res := resource.NewGenericResource("Item", "dup", "Duplicate", nil, nil)
	if err := reg.Create(ctx, res); err != nil {
		t.Fatal(err)
	}

	res2 := resource.NewGenericResource("Item", "dup", "Duplicate Again", nil, nil)
	if err := reg.Create(ctx, res2); err == nil {
		t.Fatal("expected error for duplicate create")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	reg := testRegistry(t)
	reg.RegisterKind(resource.Kind{Name: "Item"})

	_, err := reg.Get("Item", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent resource")
	}
}

func TestRegistry_List(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	reg.RegisterKind(resource.Kind{Name: "Item"})

	items, err := reg.List("Item")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}

	reg.Create(ctx, resource.NewGenericResource("Item", "a", "A", nil, nil))
	reg.Create(ctx, resource.NewGenericResource("Item", "b", "B", nil, nil))

	items, _ = reg.List("Item")
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestRegistry_Update(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	reg.RegisterKind(resource.Kind{Name: "Item"})

	original := resource.NewGenericResource("Item", "upd", "Original Name", map[string]string{"val": "1"}, nil)
	if err := reg.Create(ctx, original); err != nil {
		t.Fatal(err)
	}

	// Get the resource and verify the version.
	got, _ := reg.Get("Item", "upd")
	if got.GetMetadata().ResourceVersion != 1 {
		t.Fatalf("expected version 1, got %d", got.GetMetadata().ResourceVersion)
	}

	// Update with new spec.
	updated := resource.NewGenericResource("Item", "upd", "Updated Name", map[string]string{"val": "2"}, nil)
	if err := reg.Update(ctx, updated); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, _ = reg.Get("Item", "upd")
	if got.GetMetadata().Name != "Updated Name" {
		t.Fatalf("expected Updated Name, got %s", got.GetMetadata().Name)
	}
	if got.GetMetadata().ResourceVersion != 2 {
		t.Fatalf("expected version 2, got %d", got.GetMetadata().ResourceVersion)
	}
	// Verify creation time is preserved.
	if got.GetMetadata().CreatedAt.IsZero() {
		t.Fatal("CreatedAt should not be zero")
	}
}

func TestRegistry_Delete(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	reg.RegisterKind(resource.Kind{Name: "Item"})

	reg.Create(ctx, resource.NewGenericResource("Item", "del", "Delete Me", nil, nil))

	if err := reg.Delete(ctx, "Item", "del"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should be gone.
	_, err := reg.Get("Item", "del")
	if err == nil {
		t.Fatal("expected error after delete")
	}

	// Delete again should fail.
	err = reg.Delete(ctx, "Item", "del")
	if err == nil {
		t.Fatal("expected error for double delete")
	}
}

// ── Validation ─────────────────────────────────────────────────────────────

func TestRegistry_Validator(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	// Register a kind with a custom validator that rejects empty names.
	reg.RegisterKind(
		resource.Kind{Name: "ValidatedItem"},
		resource.WithValidator(func(r resource.Resource) error {
			if r.GetMetadata().Name == "" {
				return &resource.ErrInvalidResource{
					Kind:    r.GetKind(),
					ID:      r.GetMetadata().ID,
					Details: "name is required",
				}
			}
			return nil
		}),
	)

	// Valid resource.
	err := reg.Create(ctx, resource.NewGenericResource("ValidatedItem", "v1", "Valid", nil, nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Invalid resource (empty name).
	err = reg.Create(ctx, resource.NewGenericResource("ValidatedItem", "v2", "", nil, nil))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// ── Kind Not Found ─────────────────────────────────────────────────────────

func TestRegistry_KindNotFound(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	err := reg.Create(ctx, resource.NewGenericResource("Nonexistent", "x", "X", nil, nil))
	if err == nil {
		t.Fatal("expected kind not found error")
	}

	_, err = reg.List("Nonexistent")
	if err == nil {
		t.Fatal("expected kind not found error")
	}

	_, err = reg.Get("Nonexistent", "x")
	if err == nil {
		t.Fatal("expected kind not found error")
	}
}

// ── Namespace ──────────────────────────────────────────────────────────────

func TestNamespace_Create(t *testing.T) {
	reg := testRegistry(t)
	ctx := context.Background()

	reg.RegisterKind(resource.Kind{Name: "Namespace", Namespaced: false})

	ns := resource.NewNamespace("my-ns", "My Namespace", "Test namespace")
	if err := reg.Create(ctx, ns); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := reg.Get("Namespace", "my-ns")
	if err != nil {
		t.Fatal(err)
	}

	if got.GetMetadata().Name != "My Namespace" {
		t.Fatalf("expected My Namespace, got %s", got.GetMetadata().Name)
	}
}

func TestNamespace_Validate(t *testing.T) {
	// Valid namespace IDs.
	valid := []string{"default", "my-namespace", "test123", "a", "z"}
	for _, id := range valid {
		ns := resource.NewNamespace(id, "Test", "")
		if err := ns.Validate(); err != nil {
			t.Fatalf("expected valid namespace %q: %v", id, err)
		}
	}

	// Invalid namespace IDs.
	invalid := []string{"", "-start", "end-", "UPPERCASE", "has space", "a_b"}
	for _, id := range invalid {
		ns := resource.NewNamespace(id, "Test", "")
		if err := ns.Validate(); err == nil {
			t.Fatalf("expected invalid namespace %q to fail validation", id)
		}
	}
}

func TestDefaultNamespace(t *testing.T) {
	ns := resource.DefaultNamespace()
	if ns.GetMetadata().ID != "default" {
		t.Fatalf("expected default namespace id 'default', got %q", ns.GetMetadata().ID)
	}
	if ns.GetKind() != "Namespace" {
		t.Fatalf("expected kind Namespace, got %q", ns.GetKind())
	}
	if err := ns.Validate(); err != nil {
		t.Fatalf("default namespace should be valid: %v", err)
	}
}

// ── Events ─────────────────────────────────────────────────────────────────

func TestRegistry_CreatesEvent(t *testing.T) {
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	bus := events.NewBus(log)
	bus.Start()

	reg := resource.NewRegistry(bus, log)
	reg.RegisterKind(resource.Kind{Name: "Item"})

	received := make(chan string, 1)
	bus.Subscribe("resource.created", func(ctx context.Context, evt events.Event) {
		received <- evt.Type
	})

	ctx := context.Background()
	reg.Create(ctx, resource.NewGenericResource("Item", "e1", "Event Test", nil, nil))

	select {
	case etype := <-received:
		if etype != "resource.created" {
			t.Fatalf("expected resource.created, got %s", etype)
		}
	default:
		t.Fatal("expected event to be published")
	}
}

func TestRegistry_UpdateEvent(t *testing.T) {
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	bus := events.NewBus(log)
	bus.Start()

	reg := resource.NewRegistry(bus, log)
	reg.RegisterKind(resource.Kind{Name: "Item"})

	received := make(chan string, 1)
	bus.Subscribe("resource.updated", func(ctx context.Context, evt events.Event) {
		received <- evt.Type
	})

	ctx := context.Background()
	reg.Create(ctx, resource.NewGenericResource("Item", "eu", "Event Update", nil, nil))
	reg.Update(ctx, resource.NewGenericResource("Item", "eu", "Updated", nil, nil))

	select {
	case etype := <-received:
		if etype != "resource.updated" {
			t.Fatalf("expected resource.updated, got %s", etype)
		}
	default:
		t.Fatal("expected event to be published")
	}
}

func TestRegistry_DeleteEvent(t *testing.T) {
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	bus := events.NewBus(log)
	bus.Start()

	reg := resource.NewRegistry(bus, log)
	reg.RegisterKind(resource.Kind{Name: "Item"})

	received := make(chan string, 1)
	bus.Subscribe("resource.deleted", func(ctx context.Context, evt events.Event) {
		received <- evt.Type
	})

	ctx := context.Background()
	reg.Create(ctx, resource.NewGenericResource("Item", "ed", "Event Delete", nil, nil))
	reg.Delete(ctx, "Item", "ed")

	select {
	case etype := <-received:
		if etype != "resource.deleted" {
			t.Fatalf("expected resource.deleted, got %s", etype)
		}
	default:
		t.Fatal("expected event to be published")
	}
}
