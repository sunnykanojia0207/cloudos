package di

import (
	"testing"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func testContainer(t *testing.T) *Container {
	t.Helper()
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	return NewContainer(log)
}

func TestRegister(t *testing.T) {
	c := testContainer(t)
	err := c.Register("logger", "test-logger")
	require.NoError(t, err)
}

func TestRegisterDuplicate(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("key", "val"))
	err := c.Register("key", "val2")
	assert.Error(t, err)
}

func TestRegisterOrReplace(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("key", "original"))
	c.RegisterOrReplace("key", "replaced")

	val, ok := c.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "replaced", val)
}

func TestGet(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("db", "postgres"))

	val, ok := c.Get("db")
	assert.True(t, ok)
	assert.Equal(t, "postgres", val)
}

func TestGetNotFound(t *testing.T) {
	c := testContainer(t)
	_, ok := c.Get("nonexistent")
	assert.False(t, ok)
}

func TestMustGet(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("a", 42))

	assert.NotPanics(t, func() { c.MustGet("a") })
	assert.Panics(t, func() { c.MustGet("missing") })
}

func TestNames(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("x", 1))
	require.NoError(t, c.Register("y", 2))

	names := c.Names()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "x")
	assert.Contains(t, names, "y")
}

func TestUnregister(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("x", 1))
	c.Unregister("x")
	assert.Len(t, c.Names(), 0)
}

func TestClear(t *testing.T) {
	c := testContainer(t)
	require.NoError(t, c.Register("a", 1))
	require.NoError(t, c.Register("b", 2))
	c.Clear()
	assert.Len(t, c.Names(), 0)
}
