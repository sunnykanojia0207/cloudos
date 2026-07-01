package safe_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/safe"
)

// ── Helper: Test Logger ─────────────────────────────────────────────────────

// testLogger captures panic messages for verification.
type testLogger struct {
	mu     sync.Mutex
	errors []string
}

func (l *testLogger) Error(msg string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, msg)
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestGo_NormalExecution(t *testing.T) {
	var ran atomic.Bool
	safe.Go(func() {
		ran.Store(true)
	})

	// Give goroutine time to execute.
	time.Sleep(50 * time.Millisecond)
	if !ran.Load() {
		t.Error("safe.Go did not execute the function")
	}
}

func TestGo_PanicRecovery(t *testing.T) {
	log := &testLogger{}
	safe.SetLogger(log)

	var afterPanic atomic.Bool
	safe.Go(func() {
		panic("test panic: something went wrong")
	})
	safe.Go(func() {
		afterPanic.Store(true)
	})

	// Give goroutines time to execute.
	time.Sleep(100 * time.Millisecond)

	log.mu.Lock()
	errorCount := len(log.errors)
	log.mu.Unlock()

	if errorCount == 0 {
		t.Error("expected panic to be logged, but no errors were recorded")
	}

	if !afterPanic.Load() {
		t.Error("subsequent goroutines should still run after a panic recovery")
	}
}

func TestGo_MultiplePanics(t *testing.T) {
	log := &testLogger{}
	safe.SetLogger(log)

	for i := 0; i < 10; i++ {
		i := i
		safe.Go(func() {
			if i%2 == 0 {
				panic("even iteration panic")
			}
		})
	}

	time.Sleep(100 * time.Millisecond)
	log.mu.Lock()
	count := len(log.errors)
	log.mu.Unlock()

	if count != 5 {
		t.Errorf("expected 5 panic logs for even iterations, got %d", count)
	}
}

func TestGo_ConcurrentSafety(t *testing.T) {
	// Spawn many goroutines rapidly to test for races.
	log := &testLogger{}
	safe.SetLogger(log)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		i := i
		safe.Go(func() {
			defer wg.Done()
			if i%10 == 0 {
				panic("periodic panic")
			}
		})
	}
	wg.Wait()

	log.mu.Lock()
	count := len(log.errors)
	log.mu.Unlock()

	// 100 iterations, i%10==0 means 10 panics.
	if count != 10 {
		t.Errorf("expected 10 panic logs, got %d", count)
	}
}

func TestGoWithCtx_NormalExecution(t *testing.T) {
	var ran atomic.Bool
	ctx := context.Background()

	safe.GoWithCtx(ctx, func(c context.Context) {
		if c == nil {
			t.Error("GoWithCtx context should not be nil")
		}
		ran.Store(true)
	})

	time.Sleep(50 * time.Millisecond)
	if !ran.Load() {
		t.Error("safe.GoWithCtx did not execute the function")
	}
}

func TestGoWithCtx_PanicRecovery(t *testing.T) {
	log := &testLogger{}
	safe.SetLogger(log)

	ctx := context.Background()
	safe.GoWithCtx(ctx, func(c context.Context) {
		panic("contextual panic")
	})

	time.Sleep(100 * time.Millisecond)

	log.mu.Lock()
	count := len(log.errors)
	log.mu.Unlock()

	if count == 0 {
		t.Error("expected panic to be logged for GoWithCtx")
	}
}

func TestGoWithCtx_CancelledContext(t *testing.T) {
	var ran atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	safe.GoWithCtx(ctx, func(c context.Context) {
		// The goroutine should receive the cancelled context.
		if c.Err() == nil {
			t.Error("expected cancelled context to have error")
		}
		ran.Store(true)
	})

	time.Sleep(50 * time.Millisecond)
	if !ran.Load() {
		t.Error("goroutine should execute even with cancelled context")
	}
}

func TestSetLogger_NilSafety(t *testing.T) {
	// Ensure that calling SetLogger with nil does not cause issues
	// (it will simply prevent future panic logs).
	safe.SetLogger(nil)

	var ran atomic.Bool
	safe.Go(func() {
		panic("panic with nil logger")
	})
	safe.Go(func() {
		ran.Store(true)
	})

	time.Sleep(100 * time.Millisecond)
	if !ran.Load() {
		t.Error("goroutines should still work after nil logger is set")
	}
}

func TestRecoverPanic_DoesNotRecoverNormalReturns(t *testing.T) {
	// recoverPanic should not interfere with normal function returns.
	var ran atomic.Bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		// Simulate what safe.Go does internally
		safe.Go(func() {
			ran.Store(true)
		})
	}()

	time.Sleep(50 * time.Millisecond)
	if !ran.Load() {
		t.Error("normal execution should not be affected by recovery")
	}
}

func TestGo_DoesNotBlockOnPanic(t *testing.T) {
	// A panicking goroutine should not block the caller.
	start := time.Now()

	safe.Go(func() {
		panic("slow panic simulation")
	})

	elapsed := time.Since(start)
	if elapsed > 10*time.Millisecond {
		t.Errorf("safe.Go should return immediately, took %v", elapsed)
	}
}
