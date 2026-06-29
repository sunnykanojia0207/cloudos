package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

func testScheduler(t *testing.T) *Scheduler {
	t.Helper()
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})
	s := New(log)
	ctx := context.Background()
	s.Start(ctx)
	return s
}

func TestScheduleAndRunOnce(t *testing.T) {
	s := testScheduler(t)
	var mu sync.Mutex
	executed := false

	err := s.Schedule(Task{
		Name:    "test-once",
		RunOnce: true,
		Func: func(ctx context.Context) error {
			mu.Lock()
			executed = true
			mu.Unlock()
			return nil
		},
	})
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.True(t, executed)
	mu.Unlock()
}

func TestSchedulePeriodic(t *testing.T) {
	s := testScheduler(t)
	var mu sync.Mutex
	count := 0

	err := s.Schedule(Task{
		Name:     "test-periodic",
		Interval: 10 * time.Millisecond,
		Func: func(ctx context.Context) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	})
	require.NoError(t, err)
	time.Sleep(35 * time.Millisecond)

	mu.Lock()
	assert.GreaterOrEqual(t, count, 2, "periodic task should fire at least twice")
	mu.Unlock()

	s.Stop()
}

func TestScheduleDuplicate(t *testing.T) {
	s := testScheduler(t)

	err := s.Schedule(Task{
		Name:    "dup",
		RunOnce: true,
		Func:    func(ctx context.Context) error { return nil },
	})
	require.NoError(t, err)

	err = s.Schedule(Task{
		Name:    "dup",
		RunOnce: true,
		Func:    func(ctx context.Context) error { return nil },
	})
	assert.Error(t, err)
}

func TestScheduleNilFunc(t *testing.T) {
	s := testScheduler(t)
	err := s.Schedule(Task{
		Name:    "nil",
		RunOnce: true,
		Func:    nil,
	})
	assert.Error(t, err)
}

func TestUnschedule(t *testing.T) {
	s := testScheduler(t)
	var mu sync.Mutex
	count := 0

	err := s.Schedule(Task{
		Name:     "kill",
		Interval: 10 * time.Millisecond,
		Func: func(ctx context.Context) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	})
	require.NoError(t, err)

	time.Sleep(25 * time.Millisecond)
	s.Unschedule("kill")

	before := func() int { mu.Lock(); defer mu.Unlock(); return count }()
	time.Sleep(30 * time.Millisecond)
	after := func() int { mu.Lock(); defer mu.Unlock(); return count }()

	assert.Equal(t, before, after, "count should not increase after unschedule")
}

func TestTasks(t *testing.T) {
	s := testScheduler(t)
	require.NoError(t, s.Schedule(Task{Name: "a", RunOnce: true, Func: func(ctx context.Context) error { return nil }}))
	require.NoError(t, s.Schedule(Task{Name: "b", RunOnce: true, Func: func(ctx context.Context) error { return nil }}))

	names := s.Tasks()
	assert.Len(t, names, 2)
}
