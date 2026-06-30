package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// ── Health Checker ─────────────────────────────────────────────────────────

// HealthChecker performs health checks against an application instance
// according to a configurable HealthPolicy.
//
// It follows Kubernetes-style probe semantics:
//  1. Wait InitialDelay before the first check
//  2. Check every Interval
//  3. Each check has a Timeout
//  4. Consider healthy after SuccessThreshold consecutive successes
//  5. Consider unhealthy after FailureThreshold consecutive failures
type HealthChecker struct {
	policy *HealthPolicy
	client *http.Client
}

// NewHealthChecker creates a new HealthChecker with the given policy.
// If policy is nil, DefaultHealthPolicy is used.
func NewHealthChecker(policy *HealthPolicy) *HealthChecker {
	if policy == nil {
		policy = DefaultHealthPolicy()
	}
	return &HealthChecker{
		policy: policy,
		client: &http.Client{
			Timeout: policy.Timeout,
			// Don't follow redirects — we just care about the first response.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Policy returns the health checker's policy.
func (hc *HealthChecker) Policy() *HealthPolicy {
	return hc.policy
}

// Check performs a single health check against the given URL.
// It returns a HealthReport with the result.
func (hc *HealthChecker) Check(ctx context.Context, url string) *HealthReport {
	start := time.Now()

	// Create a context with the policy timeout.
	checkCtx, cancel := context.WithTimeout(ctx, hc.policy.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, url, nil)
	if err != nil {
		return &HealthReport{
			Status:     StatusFailed,
			Message:    fmt.Sprintf("request creation failed: %v", err),
			LastChecked: time.Now(),
		}
	}

	resp, err := hc.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		msg := err.Error()
		// Categorize common errors.
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			msg = "connection timed out"
		}
		return &HealthReport{
			Status:       StatusFailed,
			Message:      msg,
			LastChecked:  time.Now(),
			ResponseTime: elapsed,
		}
	}
	defer resp.Body.Close()

	status := StatusRunning
	if resp.StatusCode >= 500 {
		status = StatusFailed
	} else if resp.StatusCode >= 400 {
		status = StatusRunning // Client errors (4xx) mean the app is up
	}

	return &HealthReport{
		Status:       status,
		Message:      fmt.Sprintf("HTTP %d", resp.StatusCode),
		LastChecked:  time.Now(),
		ResponseTime: elapsed,
		StatusCode:   resp.StatusCode,
	}
}

// PollingChecker performs repeated health checks according to the policy.
// It reports results through a channel and stops when the context is done.
type PollingChecker struct {
	checker      *HealthChecker
	policy       *HealthPolicy
	url          string
	ready        chan struct{} // closed when initial check succeeds
	status       RuntimeStatus
	successCount int
	failureCount int
	mu           sync.Mutex
}

// NewPollingChecker creates a polling health checker.
func NewPollingChecker(url string, policy *HealthPolicy) *PollingChecker {
	if policy == nil {
		policy = DefaultHealthPolicy()
	}
	return &PollingChecker{
		checker: NewHealthChecker(policy),
		policy:  policy,
		url:     url,
		ready:   make(chan struct{}),
		status:  StatusStarting,
	}
}

// Start begins the polling loop. It returns a channel that receives
// HealthReport values for each check. The loop stops when ctx is cancelled
// or the application is considered unhealthy beyond the failure threshold.
func (pc *PollingChecker) Start(ctx context.Context) <-chan HealthReport {
	reports := make(chan HealthReport, 8)

	go func() {
		defer close(reports)

		// Initial delay before first check.
		select {
		case <-time.After(pc.policy.InitialDelay):
		case <-ctx.Done():
			return
		}

		ticker := time.NewTicker(pc.policy.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				report := pc.checker.Check(ctx, pc.url)
				pc.evaluate(report)
				reports <- *report
			case <-ctx.Done():
				return
			}
		}
	}()

	return reports
}

// Ready returns a channel that is closed when the application is
// considered healthy (success threshold met).
func (pc *PollingChecker) Ready() <-chan struct{} {
	return pc.ready
}

// Status returns the current health status.
func (pc *PollingChecker) Status() RuntimeStatus {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.status
}

func (pc *PollingChecker) evaluate(report *HealthReport) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if report.Status == StatusRunning {
		pc.successCount++
		pc.failureCount = 0
		if pc.successCount >= pc.policy.SuccessThreshold {
			if pc.status != StatusRunning {
				pc.status = StatusRunning
				close(pc.ready)
			}
		}
	} else {
		pc.failureCount++
		pc.successCount = 0
		if pc.failureCount >= pc.policy.FailureThreshold {
			pc.status = StatusFailed
		}
	}
}
