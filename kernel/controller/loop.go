package controller

import (
	"context"
	"sync"
	"time"

	"github.com/cloudos/cloudos/kernel/safe"
)

// ── Reconcile Loop ─────────────────────────────────────────────────────────

// reconcileLoop is the main event loop for a single controller. It:
//  1. Listens for ReconcileRequests from the shared work queue (that match
//     this controller's Kind).
//  2. Calls the controller's Reconcile method.
//  3. On success: logs, updates health, clears retry state.
//  4. On failure + requeue: re-enqueues with exponential backoff.
//  5. On failure + no requeue: logs error, updates health, clears retry state.
//
// The loop exits when ctx is cancelled (graceful shutdown).
func (m *Manager) reconcileLoop(ctx context.Context, ctrl Controller) {
	name := ctrl.Name()
	kind := ctrl.Kind()

	// Per-resource retry tracking for this controller.
	var retryMu sync.Mutex
	retryCounts := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-m.workQueue:
			// Only process requests for this controller's kind.
			if req.Kind != kind {
				// Re-enqueue for the correct controller — but to avoid
				// infinite loops, push it back to the shared queue.
				select {
				case m.workQueue <- req:
				default:
				}
				continue
			}

			key := retryKey(kind, req.ID)

			// Determine retry count.
			retryMu.Lock()
			retry := retryCounts[key]
			// After 10 retries, reset (avoids unbounded growth).
			if retry >= 10 {
				delete(retryCounts, key)
				retry = 0
			}
			retryCounts[key] = retry + 1
			retryCount := retry
			retryMu.Unlock()

			// Apply backoff delay before reconciling (skip for immediate
			// first attempt — retry 0 means first attempt after a success).
			if retryCount > 0 {
				delay := m.backoff.Delay(retryCount - 1)
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return
				}
			}

			// Execute reconciliation.
			result := ctrl.Reconcile(req)
			now := time.Now()

			// Update health.
			m.updateControllerHealth(name, func(h *ControllerHealth) {
				h.LastReconciled = now
				h.ReconcileCount++

				if result.Err != nil {
					h.ErrorCount++
				}

				if result.Err != nil && !result.Requeue {
					h.State = "failed"
					h.Message = "reconciliation failed: " + result.Err.Error()
				} else {
					h.State = "running"
					h.Message = "reconciliation completed"
				}
			})

			// Handle result.
			if result.Requeue {
				requeueAfter := result.RequeueAfter

				// Determine requeue delay.
				if result.Err != nil {
					// Error case: use explicit delay or exponential backoff.
					if requeueAfter <= 0 {
						requeueAfter = m.backoff.Delay(retryCount)
					}
				} else {
					// Success case: use explicit delay or default periodic check.
					// Reset retry count since reconciliation succeeded.
					retryMu.Lock()
					delete(retryCounts, key)
					retryMu.Unlock()
					if requeueAfter <= 0 {
						requeueAfter = 30 * time.Second // default periodic check
					}
				}

				// Re-enqueue after delay (with panic recovery).
				safe.Go(func() {
					select {
					case <-time.After(requeueAfter):
						m.enqueue(req)
					case <-ctx.Done():
					}
				})
			} else if result.Err == nil {
				// Success, no requeue: clear retry state.
				retryMu.Lock()
				delete(retryCounts, key)
				retryMu.Unlock()
			}
			// If !Requeue && Err != nil: do nothing (controller opted out
			// of retry even on error). Retry state is NOT cleared so it
			// will be retried on the next event.
		}
	}
}
