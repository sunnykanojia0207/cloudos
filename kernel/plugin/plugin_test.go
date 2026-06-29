package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPlugin struct {
	manifest Manifest
}

func (p *testPlugin) Manifest() Manifest        { return p.manifest }
func (p *testPlugin) Load(ctx context.Context) error    { return nil }
func (p *testPlugin) Activate(ctx context.Context) error { return nil }
func (p *testPlugin) Unload(ctx context.Context) error   { return nil }

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	p := &testPlugin{
		manifest: Manifest{
			Info: Info{Name: "test-plugin", Version: "1.0.0"},
		},
	}
	err := r.Register(p, StateDiscovered)
	require.NoError(t, err)

	got, ok := r.Get("test-plugin")
	assert.True(t, ok)
	assert.NotNil(t, got)
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	p := &testPlugin{manifest: Manifest{Info: Info{Name: "dup"}}}
	require.NoError(t, r.Register(p, StateDiscovered))
	err := r.Register(p, StateDiscovered)
	assert.Error(t, err)
}

func TestRegistryState(t *testing.T) {
	r := NewRegistry()
	p := &testPlugin{manifest: Manifest{Info: Info{Name: "p1"}}}
	require.NoError(t, r.Register(p, StateDiscovered))

	state, ok := r.State("p1")
	assert.True(t, ok)
	assert.Equal(t, StateDiscovered, state)
}

func TestRegistrySetState(t *testing.T) {
	r := NewRegistry()
	p := &testPlugin{manifest: Manifest{Info: Info{Name: "p1"}}}
	require.NoError(t, r.Register(p, StateDiscovered))

	err := r.SetState("p1", StateLoaded)
	require.NoError(t, err)

	state, ok := r.State("p1")
	assert.True(t, ok)
	assert.Equal(t, StateLoaded, state)
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(&testPlugin{manifest: Manifest{Info: Info{Name: "a"}}}, StateDiscovered))
	require.NoError(t, r.Register(&testPlugin{manifest: Manifest{Info: Info{Name: "b"}}}, StateDiscovered))

	manifests := r.List()
	assert.Len(t, manifests, 2)
}

func TestRegistryRemove(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(&testPlugin{manifest: Manifest{Info: Info{Name: "a"}}}, StateDiscovered))
	r.Remove("a")
	_, ok := r.Get("a")
	assert.False(t, ok)
}
