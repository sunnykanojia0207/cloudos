package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudos/cloudos/kernel/application"
	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Structured Log Event ─────────────────────────────────────────────────────

// LogEvent is a structured log entry from any source (workflow, runtime, health).
// It provides a unified format for all log consumers (CLI, Dashboard, API).
//
// The Source field identifies the origin: "workflow", "runtime", "health", "system".
// The Step field identifies the workflow or lifecycle step that produced the event.
type LogEvent struct {
	// Timestamp is when the event occurred (RFC3339).
	Timestamp string `json:"timestamp"`

	// Source identifies the log origin: "workflow", "runtime", "health", "system".
	Source string `json:"source"`

	// Level is the severity: "debug", "info", "warn", "error".
	Level string `json:"level"`

	// Step is the workflow or lifecycle step (e.g. "clone", "build", "deploy").
	Step string `json:"step,omitempty"`

	// Message is the human-readable log content.
	Message string `json:"message"`

	// Metadata carries optional structured data.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ── LogEvent constructors ────────────────────────────────────────────────────

// NewLogEvent creates a LogEvent with the current timestamp.
func NewLogEvent(source, level, step, message string) LogEvent {
	return LogEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Source:    source,
		Level:     level,
		Step:      step,
		Message:   message,
	}
}

// ── LogHandler ───────────────────────────────────────────────────────────────

// LogHandler serves application log endpoints — streaming, snapshot, and download.
type LogHandler struct {
	reg     *resource.Registry
	runtime cr.Runtime
	log     *logging.Logger
}

// NewLogHandler creates a LogHandler bound to the resource registry and runtime.
func NewLogHandler(reg *resource.Registry, runtime cr.Runtime) *LogHandler {
	return &LogHandler{
		reg:     reg,
		runtime: runtime,
		log:     logging.NewSubsystemLogger("api.logs", logging.LevelInfo),
	}
}

// ── Identity Helper ─────────────────────────────────────────────────────────

// resolveAppID resolves the application ID from the request path, supporting both
// direct app IDs and "apps/{id}" patterns.
func (lh *LogHandler) resolveAppID(r *http.Request) string {
	id := r.PathValue("id")
	return id
}

// ── GET /api/v1/applications/{id}/logs ───────────────────────────────────────

// SnapshotLogs returns the latest N lines of application logs.
//
//	GET /api/v1/applications/{id}/logs?tail=50
//
// Returns a JSON array of LogEvent objects sorted oldest-first.
func (lh *LogHandler) SnapshotLogs(w http.ResponseWriter, r *http.Request) {
	id := lh.resolveAppID(r)
	if id == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}

	tail := 50 // default
	if t := r.URL.Query().Get("tail"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 && v <= 1000 {
			tail = v
		}
	}

	events, err := lh.collectLogs(r.Context(), id, tail, false)
	if err != nil {
		InternalError(w, "LOG_ERROR", fmt.Sprintf("Failed to read logs: %v", err))
		return
	}

	OK(w, events)
}

// ── GET /api/v1/applications/{id}/logs/stream ────────────────────────────────

// StreamLogs streams application logs as SSE events.
//
//	GET /api/v1/applications/{id}/logs/stream?tail=10
//
// Each event is a JSON LogEvent:
//
//	event: log
//	data: {"timestamp":"...", "source":"runtime", "level":"info", "message":"..."}
//
// The stream stays open until the client disconnects or the application stops.
func (lh *LogHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	id := lh.resolveAppID(r)
	if id == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}

	// Set SSE headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		InternalError(w, "STREAMING_UNSUPPORTED", "Streaming not supported")
		return
	}

	// Send recent history first.
	tail := 10
	if t := r.URL.Query().Get("tail"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 && v <= 1000 {
			tail = v
		}
	}

	history, err := lh.collectLogs(r.Context(), id, tail, false)
	if err == nil && len(history) > 0 {
		for _, event := range history {
			lh.writeSSELogEvent(w, flusher, event)
		}
	}

	// Stream runtime logs.
	app, err := lh.getApplication(id)
	if err != nil {
		lh.writeSSELogEvent(w, flusher, NewLogEvent("system", "error", "", fmt.Sprintf("Application not found: %v", err)))
		return
	}

	// Find the running instance ID from the application status.
	instanceID := lh.findInstanceID(app)
	if instanceID == "" {
		// No running instance — just send a notice and return.
		lh.writeSSELogEvent(w, flusher, NewLogEvent("system", "info", "", "Application is not currently running"))
		return
	}

	// Stream logs from the runtime.
	stream, err := lh.runtime.Logs(r.Context(), instanceID, cr.LogOptions{
		Tail:   0, // we already sent history
		Follow: true,
		Source: "", // all sources
	})
	if err != nil {
		lh.writeSSELogEvent(w, flusher, NewLogEvent("system", "error", "", fmt.Sprintf("Cannot open log stream: %v", err)))
		return
	}
	defer stream.Close()

	for {
		select {
		case <-r.Context().Done():
			return
		case entry, ok := <-stream.Lines():
			if !ok {
				return
			}
			event := LogEvent{
				Timestamp: entry.Timestamp.Format(time.RFC3339),
				Source:    "runtime",
				Level:     logLevelFromEntry(entry),
				Step:      "",
				Message:   entry.Line,
			}
			lh.writeSSELogEvent(w, flusher, event)
		}
	}
}

// ── GET /api/v1/applications/{id}/logs/download ──────────────────────────────

// DownloadLogs returns all available logs as plain text.
//
//	GET /api/v1/applications/{id}/logs/download
//
// Returns a text/plain file with one log line per entry.
func (lh *LogHandler) DownloadLogs(w http.ResponseWriter, r *http.Request) {
	id := lh.resolveAppID(r)
	if id == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}

	events, err := lh.collectLogs(r.Context(), id, 0, false) // 0 = all
	if err != nil {
		InternalError(w, "LOG_ERROR", fmt.Sprintf("Failed to read logs: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.log"`, id))

	for _, event := range events {
		var step string
		if event.Step != "" {
			step = " [" + event.Step + "]"
		}
		line := fmt.Sprintf("[%s] [%s]%s %s\n", event.Timestamp, event.Source+step, event.Level, event.Message)
		if _, err := io.WriteString(w, line); err != nil {
			return
		}
	}
}

// ── Internal ─────────────────────────────────────────────────────────────────

// collectLogs gathers logs from the runtime for a given application.
// If live is true, it returns only currently available lines.
func (lh *LogHandler) collectLogs(ctx context.Context, appID string, tail int, live bool) ([]LogEvent, error) {
	app, err := lh.getApplication(appID)
	if err != nil {
		return nil, err
	}

	instanceID := lh.findInstanceID(app)
	if instanceID == "" {
		// Send workflow/reconciliation events instead.
		return lh.workflowEventsAsLogs(ctx, app), nil
	}

	stream, err := lh.runtime.Logs(ctx, instanceID, cr.LogOptions{
		Tail:   tail,
		Follow: false,
		Source: "",
	})
	if err != nil {
		// Fall back to workflow events.
		return lh.workflowEventsAsLogs(ctx, app), nil
	}
	defer stream.Close()

	var events []LogEvent
	for entry := range stream.Lines() {
		events = append(events, LogEvent{
			Timestamp: entry.Timestamp.Format(time.RFC3339),
			Source:    "runtime",
			Level:     logLevelFromEntry(entry),
			Message:   entry.Line,
		})
	}
	return events, nil
}

// getApplication retrieves an Application resource by ID.
func (lh *LogHandler) getApplication(id string) (*application.Application, error) {
	res, err := lh.reg.Get(application.Kind, id)
	if err != nil {
		return nil, err
	}
	app, ok := res.(*application.Application)
	if !ok {
		return nil, fmt.Errorf("resource %q is not an Application", id)
	}
	return app, nil
}

// findInstanceID extracts the runtime instance ID from an Application status.
// It checks the current deployment's runtime instance reference.
func (lh *LogHandler) findInstanceID(app *application.Application) string {
	if app == nil || app.Status_.CurrentDeploymentID == "" {
		return ""
	}
	// For now, the runtime instance ID is derived from the current deployment.
	// In a full implementation, the runtime would expose a mapping from app ID
	// to instance ID.
	return app.Status_.CurrentDeploymentID
}

// workflowEventsAsLogs generates structured log events from the application's
// workflow execution status. This provides visibility even when the runtime
// hasn't started yet or the app is in a workflow phase.
func (lh *LogHandler) workflowEventsAsLogs(ctx context.Context, app *application.Application) []LogEvent {
	if app == nil {
		return nil
	}

	var events []LogEvent

	events = append(events, LogEvent{
		Timestamp: app.Status_.CreatedAt.Format(time.RFC3339),
		Source:    "workflow",
		Level:     "info",
		Step:      "init",
		Message:   fmt.Sprintf("Application %q created (runtime: %s)", app.GetMetadata().Name, app.Spec_.Runtime.Type),
	})

	if app.Status_.Phase == application.PhaseRunning {
		events = append(events, LogEvent{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source:    "workflow",
			Level:     "info",
			Step:      "running",
			Message:   fmt.Sprintf("Application running at %s", app.Status_.URL),
		})
	}

	if app.Status_.LastReport != nil {
		report := app.Status_.LastReport
		events = append(events, LogEvent{
			Timestamp: report.CompletedAt.Format(time.RFC3339),
			Source:    "workflow",
			Level:     "info",
			Step:      "complete",
			Message:   fmt.Sprintf("Deployment #%d completed in %s", report.DeploymentNumber, report.Duration),
		})
		if len(report.Errors) > 0 {
			for _, err := range report.Errors {
				events = append(events, LogEvent{
					Timestamp: report.CompletedAt.Format(time.RFC3339),
					Source:    "workflow",
					Level:     "error",
					Step:      "failed",
					Message:   err,
				})
			}
		}
	}

	return events
}

// logLevelFromEntry maps runtime log entry types to level strings.
func logLevelFromEntry(entry cr.LogEntry) string {
	switch {
	case entry.Line == "":
		return "info"
	case strings.Contains(entry.Line, "error") || strings.Contains(entry.Line, "Error"):
		return "error"
	case strings.Contains(entry.Line, "warn") || strings.Contains(entry.Line, "Warn"):
		return "warn"
	default:
		return "info"
	}
}

// writeSSELogEvent writes a structured LogEvent as an SSE event.
func (lh *LogHandler) writeSSELogEvent(w http.ResponseWriter, flusher http.Flusher, event LogEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		lh.log.Error("sse marshal error", "error", err)
		return
	}
	_, err = fmt.Fprintf(w, "event: log\ndata: %s\n\n", data)
	if err != nil {
		return // client disconnected
	}
	flusher.Flush()
}

// ── Compatibility: LogReader from stdout/stderr ─────────────────────────────

// LogReaderFromScanner wraps a bufio.Scanner as a LogReader for reading
// process stdout/stderr. This bridges the gap between raw output and the
// structured LogStream interface.
type LogReaderFromScanner struct {
	scanner *bufio.Scanner
	closed  bool
}

// NewLogReaderFromScanner creates a LogReader wrapping a scanner.
func NewLogReaderFromScanner(scanner *bufio.Scanner) *LogReaderFromScanner {
	return &LogReaderFromScanner{scanner: scanner}
}

func (r *LogReaderFromScanner) Read(p []byte) (int, error) {
	if r.closed || !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}
	data := r.scanner.Bytes()
	data = append(data, '\n')
	n := copy(p, data)
	return n, nil
}

func (r *LogReaderFromScanner) ReadLines(n int) ([]string, error) {
	if r.closed {
		return nil, io.EOF
	}
	var lines []string
	for i := 0; i < n || n <= 0; i++ {
		if !r.scanner.Scan() {
			break
		}
		lines = append(lines, r.scanner.Text())
	}
	if len(lines) == 0 {
		return nil, io.EOF
	}
	return lines, r.scanner.Err()
}

func (r *LogReaderFromScanner) Follow(ctx context.Context) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for {
			if r.closed {
				return
			}
			if !r.scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}
			select {
			case ch <- r.scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func (r *LogReaderFromScanner) Close() error {
	r.closed = true
	return nil
}

// Ensure interface compliance.
var _ cr.LogReader = (*LogReaderFromScanner)(nil)
