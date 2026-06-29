package capabilities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	d := &Descriptor{
		ID:          "test-cap",
		Name:        "Test Capability",
		Description: "A test capability",
		Version:     Version{Major: 1, Minor: 0, Patch: 0},
		Status:      StatusExperimental,
		Category:    CategoryCore,
	}

	err := r.Register(d)
	require.NoError(t, err)

	got, ok := r.Get("test-cap")
	require.True(t, ok)
	assert.Equal(t, "Test Capability", got.Name)
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(&Descriptor{ID: "dup", Name: "first"}))
	err := r.Register(&Descriptor{ID: "dup", Name: "second"})
	assert.Error(t, err)
}

func TestRegistry_RegisterOrReplace(t *testing.T) {
	r := NewRegistry()
	r.RegisterOrReplace(&Descriptor{ID: "x", Name: "original"})
	r.RegisterOrReplace(&Descriptor{ID: "x", Name: "replaced"})

	d, ok := r.Get("x")
	require.True(t, ok)
	assert.Equal(t, "replaced", d.Name)
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.RegisterOrReplace(&Descriptor{ID: "a", Name: "A"})
	r.RegisterOrReplace(&Descriptor{ID: "b", Name: "B"})

	list := r.List()
	assert.Len(t, list, 2)
}

func TestRegistry_Len(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, 0, r.Len())
	r.RegisterOrReplace(&Descriptor{ID: "a"})
	assert.Equal(t, 1, r.Len())
}

func TestRegistry_IDs(t *testing.T) {
	r := NewRegistry()
	r.RegisterOrReplace(&Descriptor{ID: "foo"})
	r.RegisterOrReplace(&Descriptor{ID: "bar"})

	ids := r.IDs()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, ID("foo"))
	assert.Contains(t, ids, ID("bar"))
}

func TestDefaultDescriptors_Count(t *testing.T) {
	descs := DefaultDescriptors()
	assert.Len(t, descs, 5)
}

func TestDefaultDescriptors_AllHaveRequiredFields(t *testing.T) {
	for _, d := range DefaultDescriptors() {
		assert.NotEmpty(t, d.ID, "descriptor %q should have an ID", d.ID)
		assert.NotEmpty(t, d.Name, "descriptor %q should have a Name", d.ID)
		assert.NotEmpty(t, d.Description, "descriptor %q should have a Description", d.ID)
		assert.NotEmpty(t, d.Category, "descriptor %q should have a Category", d.ID)
		assert.NotEmpty(t, d.Status, "descriptor %q should have a Status", d.ID)
		assert.NotEmpty(t, d.Operations, "descriptor %q should have Operations", d.ID)
	}
}

func TestDefaultDescriptors_OperationsHaveNames(t *testing.T) {
	for _, d := range DefaultDescriptors() {
		for _, op := range d.Operations {
			assert.NotEmpty(t, op.Name, "operation in %q should have a name", d.ID)
		}
	}
}


