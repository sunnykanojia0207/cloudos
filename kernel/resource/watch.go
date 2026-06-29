package resource

import (
	"context"
	"sync"

	"github.com/cloudos/cloudos/kernel/events"
)

// ── Watch ──────────────────────────────────────────────────────────────────

// WatchEvent wraps an event bus event for resource watchers.
type WatchEvent struct {
	Type     string      `json:"type"`
	Resource Resource    `json:"resource"`
}

// Watch subscribes to resource lifecycle events for a given kind and returns
// a channel that receives WatchEvents. The caller must cancel the context to
// stop watching.
func (r *Registry) Watch(ctx context.Context, kind string) (<-chan WatchEvent, error) {
	// Verify the kind exists.
	if _, _, err := r.getKindAndStore(kind); err != nil {
		return nil, err
	}

	ch := make(chan WatchEvent, 64)
	var mu sync.Mutex
	active := true

	handler := func(ctx context.Context, evt events.Event) {
		mu.Lock()
		if !active {
			mu.Unlock()
			return
		}
		mu.Unlock()

		payload, ok := evt.Payload.(map[string]interface{})
		if !ok {
			return
		}
		evtKind, _ := payload["kind"].(string)
		if evtKind != kind {
			return
		}

		// Fetch the current state of the resource for the watch event.
		id, _ := payload["id"].(string)
		res, err := r.Get(kind, id)
		if err != nil {
			return
		}

		select {
		case ch <- WatchEvent{Type: evt.Type, Resource: res}:
		default:
			// Channel full — drop event to avoid blocking the event bus.
		}
	}

	// Subscribe to all resource event types.
	for _, et := range []string{EventResourceCreated, EventResourceUpdated, EventResourceDeleted} {
		if err := r.bus.Subscribe(et, handler); err != nil {
			return nil, err
		}
	}

	// Unsubscribe when context is cancelled.
	go func() {
		<-ctx.Done()
		mu.Lock()
		active = false
		mu.Unlock()
		for _, et := range []string{EventResourceCreated, EventResourceUpdated, EventResourceDeleted} {
			r.bus.Unsubscribe(et)
		}
		close(ch)
	}()

	return ch, nil
}
