// Package controller provides the CloudOS Controller Runtime — a generic,
// Kubernetes-inspired reconciliation engine. Controllers are active components
// that continuously reconcile desired state (spec) with observed state (status).
//
// The Controller Runtime sits above the Resource Engine and brings it to life:
//
//	User/API
//	   |
//	Resource Engine (passive — stores desired state)
//	   |
//	Controller Runtime (active — reconciles state)
//	   |
//	Providers (actual execution)
//
// Every resource kind can have a controller. The controller receives
// ReconcileRequest events whenever a resource is created, updated, or deleted,
// and is responsible for making the real world match the resource's spec.
package controller

import (
	"time"
)

// ── Controller Interface ───────────────────────────────────────────────────

// Controller is the interface every CloudOS controller must implement.
// A controller owns the reconciliation loop for a single resource kind.
type Controller interface {
	// Name returns a unique controller identifier (e.g. "namespace").
	Name() string

	// Kind returns the resource kind this controller reconciles
	// (e.g. "Namespace"). Must match a registered resource kind.
	Kind() string

	// Start begins the controller's reconcile loop. Implementations should
	// block until ctx is cancelled. The Manager calls this in a goroutine.
	Start(ctx interface{ Done() <-chan struct{} }) error

	// Stop signals the controller to shut down gracefully.
	Stop(ctx interface{ Done() <-chan struct{} }) error

	// Reconcile is called when a resource needs to be reconciled. It receives
	// a request with the resource kind and ID, and must return a result
	// indicating success, failure, or if re-queuing is needed.
	Reconcile(req ReconcileRequest) ReconcileResult

	// Health returns the controller's current health status.
	Health() ControllerHealth
}

// ── Reconcile Types ────────────────────────────────────────────────────────

// ReconcileRequest identifies the resource to reconcile.
type ReconcileRequest struct {
	// Kind is the resource kind (e.g. "Namespace").
	Kind string `json:"kind"`

	// ID is the resource identifier within its kind.
	ID string `json:"id"`
}

// ReconcileResult tells the runtime what to do after a reconcile attempt.
type ReconcileResult struct {
	// Requeue indicates the resource should be re-queued for reconciliation,
	// typically after a delay (RequeueAfter).
	Requeue bool `json:"requeue,omitempty"`

	// RequeueAfter is the delay before re-queuing. If zero, the default
	// backoff applies.
	RequeueAfter time.Duration `json:"requeueAfter,omitempty"`

	// Err is set when reconciliation failed. The runtime uses this to
	// determine retry behaviour and update controller health.
	Err error `json:"-"`
}

// ── Controller Health ──────────────────────────────────────────────────────

// ControllerHealth is the runtime health snapshot for a single controller.
type ControllerHealth struct {
	// Name is the controller's unique identifier.
	Name string `json:"name"`

	// Kind is the resource kind this controller reconciles.
	Kind string `json:"kind"`

	// State is the controller's current operational state.
	State string `json:"state"` // "running", "stopped", "failed"

	// Message describes the current health in human-readable form.
	Message string `json:"message,omitempty"`

	// LastReconciled is when the controller last completed a reconcile loop
	// (successful or failed).
	LastReconciled time.Time `json:"lastReconciled,omitempty"`

	// ReconcileCount is the total number of reconcile attempts.
	ReconcileCount uint64 `json:"reconcileCount"`

	// ErrorCount is the number of failed reconcile attempts.
	ErrorCount uint64 `json:"errorCount"`
}

// ── Backoff Strategy ───────────────────────────────────────────────────────

// BackoffStrategy configures exponential backoff for retrying failed
// reconciliation.
type BackoffStrategy struct {
	// BaseDelay is the initial delay before the first retry.
	BaseDelay time.Duration `json:"baseDelay"`

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration `json:"maxDelay"`

	// Factor is the multiplier applied to the delay after each retry.
	Factor float64 `json:"factor"`
}

// DefaultBackoff returns a sensible default backoff strategy:
// base 100ms, max 60s, factor 2.0.
func DefaultBackoff() BackoffStrategy {
	return BackoffStrategy{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  60 * time.Second,
		Factor:    2.0,
	}
}

// Delay computes the delay for the nth retry (0-indexed).
func (b BackoffStrategy) Delay(retry int) time.Duration {
	if retry < 0 {
		retry = 0
	}
	d := float64(b.BaseDelay)
	for i := 0; i < retry; i++ {
		d *= b.Factor
		if d > float64(b.MaxDelay) {
			return b.MaxDelay
		}
	}
	return time.Duration(d)
}

// ── Errors ─────────────────────────────────────────────────────────────────

// ErrControllerNotFound is returned when a controller is not registered.
type ErrControllerNotFound struct {
	Name string
}

func (e *ErrControllerNotFound) Error() string {
	return "controller " + e.Name + " not found"
}

// ErrControllerAlreadyRegistered is returned on duplicate registration.
type ErrControllerAlreadyRegistered struct {
	Name string
}

func (e *ErrControllerAlreadyRegistered) Error() string {
	return "controller " + e.Name + " already registered"
}

// ReconcileResultSuccess is a convenience for a successful reconciliation
// that does not need re-queuing.
var ReconcileResultSuccess = ReconcileResult{}

// ReconcileResultRequeue is a convenience for a reconciliation that should
// be re-queued immediately (e.g. for periodic reconciliation).
var ReconcileResultRequeue = ReconcileResult{Requeue: true}

// RequeueAfter returns a result that re-queues after the given delay.
func RequeueAfter(d time.Duration) ReconcileResult {
	return ReconcileResult{Requeue: true, RequeueAfter: d}
}

// RequeueWithError returns a result that re-queues with an error
// (triggers retry backoff).
func RequeueWithError(err error) ReconcileResult {
	return ReconcileResult{Requeue: true, Err: err}
}
