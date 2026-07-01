// Package events provides an in-memory publish / subscribe event bus for
// communication between kernel subsystems. The bus is typed, synchronous
// (handlers are invoked on the publisher's goroutine), and safe for concurrent
// use.
package events

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
)

// Event is a typed message that travels through the bus.
type Event struct {
	Type      string      `json:"type"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Handler processes a single event.
type Handler func(ctx context.Context, event Event)

// Bus is a concurrent, in-memory pub/sub event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	log      *logging.Logger
	running  bool
}

// NewBus creates a new event bus.
func NewBus(log *logging.Logger) *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
		log:      log.WithContext(context.Background()),
	}
}

// Start enables event publishing and subscription.
func (b *Bus) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = true
}

// Stop disables event publishing and clears all handlers.
func (b *Bus) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = false
	b.handlers = make(map[string][]Handler)
}

// Subscribe registers a handler for the given event type. The handler will be
// invoked synchronously on the publisher's goroutine.
func (b *Bus) Subscribe(eventType string, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("events: nil handler")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes all handlers for the given event type.
func (b *Bus) Unsubscribe(eventType string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, eventType)
}

// Publish sends an event to all registered handlers. Handlers are invoked
// synchronously in registration order.
func (b *Bus) Publish(ctx context.Context, event Event) {
	event.Timestamp = time.Now()

	b.mu.RLock()
	handlers, ok := b.handlers[event.Type]
	running := b.running
	b.mu.RUnlock()

	if !running || !ok {
		return
	}

	for _, h := range handlers {
		// Recover from handler panics so a single misbehaving handler
		// does not crash the publisher goroutine or prevent other
		// handlers from receiving the event.
		func(handler Handler) {
			defer func() {
				if r := recover(); r != nil {
					// Use the bus's logger if available, otherwise fall back
					// to a package-level recovery that won't crash.
					if b.log != nil {
						b.log.Error("event handler panic",
							"event_type", event.Type,
							"source", event.Source,
							"panic", fmt.Sprintf("%v", r),
							"stack", string(debug.Stack()),
						)
					}
				}
			}()
			handler(ctx, event)
		}(h)
	}
}

// MustSubscribe is a convenience wrapper that panics on error.
func (b *Bus) MustSubscribe(eventType string, handler Handler) {
	if err := b.Subscribe(eventType, handler); err != nil {
		panic(fmt.Sprintf("events: subscribe %s: %v", eventType, err))
	}
}
