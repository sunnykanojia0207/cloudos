package health

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCheckable struct {
	state   types.ResourceState
	message string
	mu      sync.Mutex
}

func (m *mockCheckable) CheckHealth(ctx context.Context) Report {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Report{State: m.state, Message: m.message, Timestamp: time.Now()}
}

func (m *mockCheckable) SetState(s types.ResourceState, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = s
	m.message = msg
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func testManager(t *testing.T) *Manager {
	t.Helper()
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	m := NewManager(log)
	require.NoError(t, m.Start(context.Background()))
	return m
}

func TestRegister(t *testing.T) {
	m := testManager(t)
	err := m.Register("test", &mockCheckable{state: types.StateRunning})
	require.NoError(t, err)
	assert.Contains(t, m.Registered(), "test")
}

func TestRegisterDuplicate(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register("dup", &mockCheckable{}))
	err := m.Register("dup", &mockCheckable{})
	assert.Error(t, err)
}

func TestReport(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	c := &mockCheckable{state: types.StateRunning, message: "all good"}
	require.NoError(t, m.Register("svc", c))
	m.runChecks(ctx)

	r, ok := m.Report("svc")
	assert.True(t, ok)
	assert.Equal(t, types.StateRunning, r.State)
	assert.Equal(t, "all good", r.Message)
}

func TestOverallHealthy(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	require.NoError(t, m.Register("a", &mockCheckable{state: types.StateRunning}))
	require.NoError(t, m.Register("b", &mockCheckable{state: types.StateRunning}))
	m.runChecks(ctx)

	overall := m.Overall()
	assert.Equal(t, types.StateRunning, overall.State)
}

func TestOverallDegraded(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	require.NoError(t, m.Register("a", &mockCheckable{state: types.StateRunning}))
	require.NoError(t, m.Register("b", &mockCheckable{state: types.StateFailed}))
	m.runChecks(ctx)

	overall := m.Overall()
	assert.Equal(t, types.StateDegraded, overall.State)
}

func TestUnregister(t *testing.T) {
	m := testManager(t)
	require.NoError(t, m.Register("x", &mockCheckable{}))
	m.Unregister("x")
	assert.NotContains(t, m.Registered(), "x")
}

func TestAll(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	require.NoError(t, m.Register("a", &mockCheckable{state: types.StateRunning}))
	m.runChecks(ctx)

	all := m.All()
	assert.Len(t, all, 1)
	assert.Equal(t, types.StateRunning, all["a"].State)
}
