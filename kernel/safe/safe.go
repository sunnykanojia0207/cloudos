// Package safe provides panic-safe goroutine execution for the CloudOS kernel.
//
// A panic in any goroutine of a long-running daemon process causes the entire
// process to crash. Go's philosophy is "crash on panic" for servers, but in
// a kernel orchestrating many subsystems, a single buggy controller or handler
// should not take down the entire system. This package provides recovery-wrapped
// goroutine execution so that failures are isolated, logged, and survivable.
//
// Usage:
//
//	safe.Go(func() {
//	    riskyOperation()
//	})
//
//	safe.GoWithCtx(ctx, func(ctx context.Context) {
//	    controllerLoop(ctx)
//	})
package safe

import (
	"context"
	"fmt"
	"runtime/debug"
)

// defaultLogger is the package-level fallback when no logger is configured.
// It is replaced by SetLogger when the kernel initialises its logging system.
var defaultLogger interface {
	Error(msg string, args ...interface{})
}

// SetLogger configures a structured logger for panic reports.
// Call this during kernel boot so panic output uses the same logging
// pipeline as the rest of the system.
func SetLogger(l interface{ Error(msg string, args ...interface{}) }) {
	defaultLogger = l
}

// Go runs fn in a goroutine with panic recovery. If fn panics, the panic
// value and stack trace are logged and the goroutine exits cleanly.
//
// This prevents a single panic from taking down the entire kernel process.
// Without this, a nil pointer dereference in any goroutine — controller
// reconcile loop, health check ticker, event handler, workflow retry —
// would crash the entire daemon.
func Go(fn func()) {
	go func() {
		defer recoverPanic("safe.Go")
		fn()
	}()
}

// GoWithCtx is like Go but passes a context to fn. Use this when the
// goroutine needs a context for cancellation or deadline propagation.
func GoWithCtx(ctx context.Context, fn func(context.Context)) {
	go func() {
		defer recoverPanic("safe.GoWithCtx")
		fn(ctx)
	}()
}

// recoverPanic recovers from a panic and logs the error. It is designed to
// be called as `defer recoverPanic(label)` at the top of every goroutine
// that should not crash the process on failure.
func recoverPanic(label string) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		msg := fmt.Sprintf("PANIC in %s: %v", label, r)

		if defaultLogger != nil {
			defaultLogger.Error(msg,
				"panic", fmt.Sprintf("%v", r),
				"stack", stack,
			)
		}
	}
}
