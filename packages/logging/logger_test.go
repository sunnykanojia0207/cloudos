package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewLoggerWithWriter(LevelInfo, &buf)
	require.NotNil(t, l)
	assert.Equal(t, "", l.subsystem)
}

func TestSubsystemLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewSubsystemLoggerWithWriter("kernel", LevelDebug, &buf)
	l.Info("booted")

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "kernel", m["subsystem"])
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		fn    func(*Logger, string, ...interface{})
		want  string
	}{
		{"debug", LevelDebug, (*Logger).Debug, "DEBUG"},
		{"info", LevelInfo, (*Logger).Info, "INFO"},
		{"warn", LevelWarn, (*Logger).Warn, "WARN"},
		{"error", LevelError, (*Logger).Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := NewLoggerWithWriter(tt.level, &buf)
			tt.fn(l, "msg", "key", "val")

			var m map[string]interface{}
			require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
			assert.Equal(t, tt.want, m["level"])
			assert.Equal(t, "msg", m["msg"])
			assert.Equal(t, "val", m["key"])
		})
	}
}

func TestStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	l := NewLoggerWithWriter(LevelInfo, &buf)
	l.Info("deploy", "id", "abc", "replicas", float64(3))

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "abc", m["id"])
	assert.Equal(t, float64(3), m["replicas"])
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-1")
	ctx = WithRequestID(ctx, "req-1")
	ctx = WithUserID(ctx, "user-1")

	var buf bytes.Buffer
	l := NewLoggerWithWriter(LevelInfo, &buf)
	l.WithContext(ctx).Info("request")

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "trace-1", m["trace_id"])
	assert.Equal(t, "req-1", m["request_id"])
	assert.Equal(t, "user-1", m["user_id"])
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in  string
		out Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo},
		{"", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.out, ParseLevel(tt.in))
		})
	}
}

func TestTimestampFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewLoggerWithWriter(LevelInfo, &buf)
	l.Info("hi")

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))

	ts, ok := m["time"].(string)
	require.True(t, ok, "time field should be a string")
	assert.Contains(t, ts, "T", "time should be ISO8601: %s", ts)
	assert.Contains(t, ts, "Z", "time should be UTC: %s", ts)
}
