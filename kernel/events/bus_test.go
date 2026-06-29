package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBus(t *testing.T) *Bus {
	t.Helper()
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &syncBuffer{})
	bus := NewBus(log)
	bus.Start()
	return bus
}

type syncBuffer struct {
	mu sync.Mutex
}

func (b *syncBuffer) Write(p []byte) (int, error) { return len(p), nil }

func TestNewBus(t *testing.T) {
	b := testBus(t)
	require.NotNil(t, b)
}

func TestSubscribePublish(t *testing.T) {
	b := testBus(t)
	var mu sync.Mutex
	received := false

	err := b.Subscribe("test.event", func(_ context.Context, e Event) {
		mu.Lock()
		received = true
		mu.Unlock()
	})
	require.NoError(t, err)

	b.Publish(context.Background(), Event{Type: "test.event", Source: "test"})
	time.Sleep(5 * time.Millisecond)

	mu.Lock()
	assert.True(t, received)
	mu.Unlock()
}

func TestMultipleHandlers(t *testing.T) {
	b := testBus(t)
	var mu sync.Mutex
	count := 0

	for i := 0; i < 3; i++ {
		require.NoError(t, b.Subscribe("test.event", func(_ context.Context, e Event) {
			mu.Lock()
			count++
			mu.Unlock()
		}))
	}

	b.Publish(context.Background(), Event{Type: "test.event", Source: "test"})
	time.Sleep(5 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 3, count)
	mu.Unlock()
}

func TestUnsubscribe(t *testing.T) {
	b := testBus(t)
	var mu sync.Mutex
	received := false

	require.NoError(t, b.Subscribe("test.event", func(_ context.Context, e Event) {
		mu.Lock()
		received = true
		mu.Unlock()
	}))

	b.Unsubscribe("test.event")
	b.Publish(context.Background(), Event{Type: "test.event", Source: "test"})
	time.Sleep(5 * time.Millisecond)

	mu.Lock()
	assert.False(t, received)
	mu.Unlock()
}

func TestNilHandler(t *testing.T) {
	b := testBus(t)
	err := b.Subscribe("test", nil)
	assert.Error(t, err)
}

func TestEventPayload(t *testing.T) {
	b := testBus(t)
	var got Event

	require.NoError(t, b.Subscribe("deploy.created", func(_ context.Context, e Event) {
		got = e
	}))

	b.Publish(context.Background(), Event{
		Type:   "deploy.created",
		Source: "compute",
		Payload: map[string]string{"id": "abc"},
	})

	assert.Equal(t, "deploy.created", got.Type)
	assert.Equal(t, "compute", got.Source)
	assert.NotNil(t, got.Payload)
}
