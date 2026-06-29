package registry

import (
	"testing"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testItem struct{ name string }

func (t testItem) Name() string { return t.name }

func testManager(t *testing.T) *Manager {
	t.Helper()
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	return NewManager("test", log)
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestRegister(t *testing.T) {
	m := testManager(t)
	err := m.Register(testItem{"alpha"})
	require.NoError(t, err)
}

func TestRegisterDuplicate(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register(testItem{"dup"}))
	err := m.Register(testItem{"dup"})
	assert.Error(t, err)
}

func TestGet(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register(testItem{"alpha"}))

	item, ok := m.Get("alpha")
	assert.True(t, ok)
	assert.Equal(t, "alpha", item.Name())
}

func TestGetNotFound(t *testing.T) {
	m := testManager(t)
	_, ok := m.Get("nonexistent")
	assert.False(t, ok)
}

func TestList(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register(testItem{"a"}))
	require.NoError(t, m.Register(testItem{"b"}))

	items := m.List()
	assert.Len(t, items, 2)
}

func TestNames(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register(testItem{"a"}))
	require.NoError(t, m.Register(testItem{"b"}))

	names := m.Names()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
}

func TestUnregister(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register(testItem{"a"}))
	m.Unregister("a")
	assert.Equal(t, 0, m.Len())
}

func TestMustGet(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register(testItem{"a"}))

	assert.NotPanics(t, func() { m.MustGet("a") })
	assert.Panics(t, func() { m.MustGet("nonexistent") })
}
