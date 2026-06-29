// Package logging provides structured, level-based logging for CloudOS.
// It wraps Go 1.24's log/slog with a simplified API, subsystem scoping,
// context propagation, and optional file output with rotation.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents a log severity level.
type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the upper-case string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger wraps slog.Logger with CloudOS-specific features.
type Logger struct {
	logger    *slog.Logger
	subsystem string
	mu        sync.Mutex
}

// NewLogger creates a Logger that writes JSON-formatted entries to stdout.
func NewLogger(level Level) *Logger {
	return NewLoggerWithWriter(level, os.Stdout)
}

// NewLoggerWithWriter creates a Logger that writes JSON-formatted entries to
// the given io.Writer. Useful in tests to capture output.
func NewLoggerWithWriter(level Level, w io.Writer) *Logger {
	lvl := slog.LevelInfo
	switch level {
	case LevelDebug:
		lvl = slog.LevelDebug
	case LevelInfo:
		lvl = slog.LevelInfo
	case LevelWarn:
		lvl = slog.LevelWarn
	case LevelError:
		lvl = slog.LevelError
	}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, time.Now().UTC().Format(time.RFC3339))
			}
			return a
		},
	})

	return &Logger{
		logger: slog.New(handler),
	}
}

// NewSubsystemLogger creates a Logger scoped to a subsystem. Every emitted
// entry automatically includes the "subsystem" field.
func NewSubsystemLogger(subsystem string, level Level) *Logger {
	return &Logger{
		logger:    NewLoggerWithWriter(level, os.Stdout).logger,
		subsystem: subsystem,
	}
}

// NewSubsystemLoggerWithWriter creates a subsystem-scoped Logger that writes
// to the given writer.
func NewSubsystemLoggerWithWriter(subsystem string, level Level, w io.Writer) *Logger {
	return &Logger{
		logger:    NewLoggerWithWriter(level, w).logger,
		subsystem: subsystem,
	}
}

// Debug logs a debug-level message with optional structured key-value pairs.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, l.attachSubsystem(args...)...)
}

// Info logs an info-level message with optional structured key-value pairs.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, l.attachSubsystem(args...)...)
}

// Warn logs a warning-level message with optional structured key-value pairs.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, l.attachSubsystem(args...)...)
}

// Error logs an error-level message with optional structured key-value pairs.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, l.attachSubsystem(args...)...)
}

// WithContext returns a new Logger that automatically includes fields from the
// context (trace_id, request_id, user_id) in every emitted entry.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := extractFields(ctx)
	if len(fields) == 0 {
		return l
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	cl := &Logger{
		logger:    l.logger.With(fields...),
		subsystem: l.subsystem,
	}
	return cl
}

// attachSubsystem prepends the subsystem field when the logger is scoped.
func (l *Logger) attachSubsystem(args ...interface{}) []interface{} {
	if l.subsystem == "" {
		return args
	}
	out := make([]interface{}, 0, len(args)+2)
	out = append(out, "subsystem", l.subsystem)
	out = append(out, args...)
	return out
}

// ParseLevel converts a case-insensitive string to a Level.
func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
