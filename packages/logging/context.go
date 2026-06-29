package logging

import "context"

// Private context key type to prevent collisions.
type ctxKey string

const (
	traceIDKey   ctxKey = "trace_id"
	requestIDKey ctxKey = "request_id"
	userIDKey    ctxKey = "user_id"
)

// WithTraceID attaches a trace ID to the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithRequestID attaches a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID attaches a user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// extractFields pulls known log fields from a context and returns them as a
// slog-friendly key-value slice.
func extractFields(ctx context.Context) []interface{} {
	var fields []interface{}
	if v, ok := ctx.Value(traceIDKey).(string); ok && v != "" {
		fields = append(fields, "trace_id", v)
	}
	if v, ok := ctx.Value(requestIDKey).(string); ok && v != "" {
		fields = append(fields, "request_id", v)
	}
	if v, ok := ctx.Value(userIDKey).(string); ok && v != "" {
		fields = append(fields, "user_id", v)
	}
	return fields
}
