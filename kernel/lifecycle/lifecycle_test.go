package lifecycle

import (
	"sync"
	"testing"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testComponent struct {
	name     string
	mu       sync.Mutex
	started  bool
	stopped  bool
	startErr error
	stopErr  error
}

func (c *testComponent) Name() string      { return c.name }
func (c *testComponent) Start() error       { c.mu.Lock(); defer c.mu.Unlock(); c.started = true; return c.startErr }
func (c *testComponent) Stop() error        { c.mu.Lock(); defer c.mu.Unlock(); c.stopped = true; return c.stopErr }
func (c *testComponent) Started() bool      { c.mu.Lock(); defer c.mu.Unlock(); return c.started }
func (c *testComponent) Stopped() bool      { c.mu.Lock(); defer c.mu.Unlock(); return c.stopped }

func testLogger() *logging.Logger {
	return logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestManagerRegister(t *testing.T) {
	m := NewManager(testLogger())
	c := &testComponent{name: "test"}
	err := m.Register(c)
	require.NoError(t, err)
}

func TestManagerRegisterDuplicate(t *testing.T) {
	m := NewManager(testLogger())
	c := &testComponent{name: "test"}
	require.NoError(t, m.Register(c))
	err := m.Register(c)
	assert.Error(t, err)
}

func TestManagerStart(t *testing.T) {
	m := NewManager(testLogger())
	c := &testComponent{name: "test"}
	require.NoError(t, m.Register(c))

	err := m.Start("test")
	require.NoError(t, err)

	state, ok := m.State("test")
	assert.True(t, ok)
	assert.Equal(t, StateRunning, state)
	assert.True(t, c.Started())
}

func TestManagerStartNotFound(t *testing.T) {
	m := NewManager(testLogger())
	err := m.Start("nonexistent")
	assert.Error(t, err)
}

func TestManagerStop(t *testing.T) {
	m := NewManager(testLogger())
	c := &testComponent{name: "test"}
	require.NoError(t, m.Register(c))
	require.NoError(t, m.Start("test"))

	err := m.Stop("test")
	require.NoError(t, err)

	state, ok := m.State("test")
	assert.True(t, ok)
	assert.Equal(t, StateStopped, state)
	assert.True(t, c.Stopped())
}

func TestManagerSnapshot(t *testing.T) {
	m := NewManager(testLogger())
	require.NoError(t, m.Register(&testComponent{name: "a"}))
	require.NoError(t, m.Register(&testComponent{name: "b"}))

	snap := m.Snapshot()
	assert.Len(t, snap, 2)
	assert.Equal(t, StatePending, snap["a"])
	assert.Equal(t, StatePending, snap["b"])
}

func TestManagerStartAll(t *testing.T) {
	m := NewManager(testLogger())
	c1 := &testComponent{name: "a"}
	c2 := &testComponent{name: "b"}
	require.NoError(t, m.Register(c1))
	require.NoError(t, m.Register(c2))

	err := m.StartAll()
	require.NoError(t, err)

	assert.True(t, c1.Started())
	assert.True(t, c2.Started())
}
