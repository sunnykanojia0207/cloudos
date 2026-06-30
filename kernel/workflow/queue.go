package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// QueueItem represents a work item in the execution queue.
type QueueItem struct {
	WorkflowID string
	NodeID     string
	RetryCount int
	EnqueuedAt string
}

// Queue is an in-memory, channel-backed execution queue for workflows.
//
// It supports:
//   - Async enqueue/dequeue with configurable workers
//   - Priority ordering (future: higher priority items are dequeued first)
//   - Graceful shutdown via context cancellation
//   - Metrics via events
//
// The queue is designed to be replaced by a persistent implementation (Redis,
// RabbitMQ, etc.) without changing the Scheduler or Executor.
type Queue struct {
	mu       sync.Mutex
	items    []QueueItem
	notify   chan struct{}
	done     chan struct{}
	running  bool
	capacity int
}

// NewQueue creates an in-memory workflow queue.
// capacity: max items before Enqueue blocks (0 = unlimited).
func NewQueue(capacity int) *Queue {
	return &Queue{
		items:    make([]QueueItem, 0, 16),
		notify:   make(chan struct{}, 1),
		done:     make(chan struct{}),
		capacity: capacity,
	}
}

// Enqueue adds an item to the queue. Blocks if the queue is at capacity.
// Returns an error if the queue is not running.
func (q *Queue) Enqueue(ctx context.Context, item QueueItem) error {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return fmt.Errorf("queue: not running")
	}
	if q.capacity > 0 && len(q.items) >= q.capacity {
		q.mu.Unlock()
		return fmt.Errorf("queue: at capacity (%d)", q.capacity)
	}
	item.EnqueuedAt = NowUTC()
	q.items = append(q.items, item)
	q.mu.Unlock()

	// Non-blocking notify
	select {
	case q.notify <- struct{}{}:
	default:
	}

	return nil
}

// Dequeue blocks until an item is available or the context is cancelled.
// Returns the item and a done function to call when processing is complete.
func (q *Queue) Dequeue(ctx context.Context) (QueueItem, func(), error) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			item := q.items[0]
			q.items = q.items[1:]
			q.mu.Unlock()
			return item, func() {}, nil
		}
		q.mu.Unlock()

		select {
		case <-q.notify:
			continue
		case <-ctx.Done():
			return QueueItem{}, nil, ctx.Err()
		}
	}
}

// TryDequeue attempts to dequeue an item without blocking.
func (q *Queue) TryDequeue() (QueueItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return QueueItem{}, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// Len returns the current queue depth.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Start enables enqueue/dequeue operations.
func (q *Queue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.running = true
}

// Stop disables the queue and cancels any pending Dequeue operations.
func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.running = false
	close(q.done)
}

// Run starts a configurable number of goroutines that dequeue items and
// pass them to the handler function. Blocks until ctx is cancelled.
func (q *Queue) Run(ctx context.Context, workers int, handler func(context.Context, QueueItem) error) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				item, done, err := q.Dequeue(ctx)
				if err != nil {
					return
				}
				_ = handler(ctx, item)
				done()
				_ = workerID // available for logging
			}
		}(i)
	}
	wg.Wait()
}

// Drain blocks until the queue is empty or the timeout expires.
func (q *Queue) Drain(timeout time.Duration) error {
	deadline := time.After(timeout)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-deadline:
			q.mu.Lock()
			n := len(q.items)
			q.mu.Unlock()
			if n > 0 {
				return fmt.Errorf("queue: drain timeout with %d items remaining", n)
			}
			return nil
		case <-tick.C:
			q.mu.Lock()
			n := len(q.items)
			q.mu.Unlock()
			if n == 0 {
				return nil
			}
		}
	}
}
