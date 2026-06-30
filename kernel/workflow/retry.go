package workflow

import (
	"fmt"
	"math"
	"time"
)

// RetryEvaluator determines whether a failed node should be retried.
type RetryEvaluator struct{}

// NewRetryEvaluator creates a new RetryEvaluator.
func NewRetryEvaluator() *RetryEvaluator {
	return &RetryEvaluator{}
}

// ShouldRetry checks whether a TaskNode should be retried based on its
// RetryPolicy and current retry count. Returns the delay before the next
// attempt, or an error indicating the node should not be retried.
func (e *RetryEvaluator) ShouldRetry(node *TaskNode) (time.Duration, error) {
	if node.RetryPolicy == nil {
		return 0, fmt.Errorf("no retry policy configured")
	}

	if node.RetryCount >= node.RetryPolicy.MaxRetries {
		return 0, fmt.Errorf("max retries (%d) exceeded", node.RetryPolicy.MaxRetries)
	}

	// Exponential backoff: base * 2^attempt, capped at max
	backoff := float64(node.RetryPolicy.BackoffBase) * math.Pow(2, float64(node.RetryCount))
	if backoff > float64(node.RetryPolicy.BackoffMax) {
		backoff = float64(node.RetryPolicy.BackoffMax)
	}

	return time.Duration(backoff), nil
}

// BackoffDuration returns the delay before the next retry attempt.
func (e *RetryEvaluator) BackoffDuration(node *TaskNode) time.Duration {
	if node.RetryPolicy == nil {
		return 0
	}
	backoff := float64(node.RetryPolicy.BackoffBase) * math.Pow(2, float64(node.RetryCount))
	if backoff > float64(node.RetryPolicy.BackoffMax) {
		backoff = float64(node.RetryPolicy.BackoffMax)
	}
	return time.Duration(backoff)
}

// FormatDuration returns a human-readable duration string.
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0s"
	}
	if d < time.Second {
		ms := d.Milliseconds()
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		s := d.Round(time.Second).Seconds()
		return fmt.Sprintf("%.0fs", s)
	}
	m := d.Minutes()
	return fmt.Sprintf("%.0fm", m)
}
