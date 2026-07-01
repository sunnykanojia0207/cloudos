package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/packages/logging"
)

// WorkflowHandler serves WorkflowExecution-related endpoints, including the
// Deployment Timeline SSE stream.
type WorkflowHandler struct {
	reg *resource.Registry
	log *logging.Logger
}

// NewWorkflowHandler creates a new WorkflowHandler bound to the resource registry.
func NewWorkflowHandler(reg *resource.Registry) *WorkflowHandler {
	return &WorkflowHandler{
		reg: reg,
		log: logging.NewSubsystemLogger("api.workflow", logging.LevelInfo),
	}
}

// ── GET /api/v1/workflow-executions/{id} ──────────────────────────────────

// GetExecution returns a single WorkflowExecution by ID.
func (wh *WorkflowHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "WorkflowExecution ID is required")
		return
	}

	exec, err := wh.reg.Get(workflow.WorkflowExecutionKind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	obj := resourceToObject(exec)
	OK(w, obj)
}

// ── SSE Stream ─────────────────────────────────────────────────────────────

// timelineEvent is a single event in the deployment timeline stream.
type timelineEvent struct {
	Type    string      `json:"type"`    // "init", "progress", "node_started", "node_completed", "node_failed", "completed", "failed"
	Payload interface{} `json:"payload"` // the current status or node info
}

// sseTimelineData wraps the full WorkflowExecution status for a timeline update.
type sseTimelineData struct {
	Phase         string              `json:"phase"`
	Progress      float64             `json:"progress"`
	CurrentNode   string              `json:"currentNode,omitempty"`
	CompletedNodes []string           `json:"completedNodes,omitempty"`
	FailedNodes   []string            `json:"failedNodes,omitempty"`
	TotalNodes    int                 `json:"totalNodes"`
	Duration      string              `json:"duration,omitempty"`
	URL           string              `json:"url,omitempty"`
	Result        string              `json:"result,omitempty"`
	Error         string              `json:"error,omitempty"`
	NodeResults   []workflow.NodeResult `json:"nodeResults,omitempty"`
}

// StreamExecutionEvents streams WorkflowExecution status updates as SSE events.
//
//	GET /api/v1/workflow-executions/{id}/events
//
// The client receives a stream of JSON events:
//
//	event: execution
//	data: {"type":"init","payload":{...}}
//
//	event: execution
//	data: {"type":"progress","payload":{...}}
//
// The stream ends when the execution reaches a terminal state (completed/failed).
func (wh *WorkflowHandler) StreamExecutionEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		BadRequest(w, "MISSING_ID", "WorkflowExecution ID is required")
		return
	}

	// Check the execution exists
	exec, err := wh.reg.Get(workflow.WorkflowExecutionKind, id)
	if err != nil {
		resourceErrorToHTTP(w, err)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		InternalError(w, "STREAMING_UNSUPPORTED", "Streaming not supported")
		return
	}

	// Send initial state
	wh.writeSSEExecution(w, flusher, exec)

	// If already in terminal state, we're done
	status := extractStatus(exec)
	if status == nil {
		return
	}
	if isTerminal(status.Phase) {
		return
	}

	// Poll for changes
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	lastVersion := exec.GetMetadata().ResourceVersion

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			current, err := wh.reg.Get(workflow.WorkflowExecutionKind, id)
			if err != nil {
				wh.writeSSEEvent(w, flusher, "error", map[string]string{
					"error": fmt.Sprintf("execution not found: %v", err),
				})
				return
			}

			rv := current.GetMetadata().ResourceVersion
			if rv == lastVersion {
				continue // no change
			}
			lastVersion = rv

			wh.writeSSEExecution(w, flusher, current)

			// Check if terminal
			s := extractStatus(current)
			if s != nil && isTerminal(s.Phase) {
				return
			}
		}
	}
}

// ── SSE Helpers ─────────────────────────────────────────────────────────────

// writeSSEExecution sends a full execution status snapshot as an SSE event.
func (wh *WorkflowHandler) writeSSEExecution(w http.ResponseWriter, flusher http.Flusher, exec resource.Resource) {
	status := extractStatus(exec)
	if status == nil {
		return
	}

	eventType := "progress"
	if isTerminal(status.Phase) {
		eventType = string(status.Phase)
	}

	data := sseTimelineData{
		Phase:          string(status.Phase),
		Progress:       status.Progress,
		CurrentNode:    status.CurrentNode,
		CompletedNodes: status.CompletedNodes,
		FailedNodes:    status.FailedNodes,
		TotalNodes:     status.TotalNodes,
		Duration:       status.Duration,
		Result:         status.Result,
		Error:          status.Error,
		NodeResults:    status.NodeResults,
	}

	wh.writeSSEEvent(w, flusher, eventType, data)
}

// writeSSEEvent writes a single SSE event with JSON payload.
func (wh *WorkflowHandler) writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		wh.log.Error("sse marshal error", "error", err)
		return
	}

	_, err = fmt.Fprintf(w, "event: execution\ndata: %s\n\n", data)
	if err != nil {
		// Client disconnected — ignore write errors
		return
	}
	flusher.Flush()
}

// extractStatus extracts the WorkflowExecutionStatus from a resource.
func extractStatus(res resource.Resource) *workflow.WorkflowExecutionStatus {
	s := res.GetStatus()
	if s == nil {
		return nil
	}
	status, ok := s.(*workflow.WorkflowExecutionStatus)
	if !ok {
		// Try value type (status might be stored by value in some paths)
		if sv, ok := s.(workflow.WorkflowExecutionStatus); ok {
			return &sv
		}
		return nil
	}
	return status
}

// isTerminal returns true if the workflow phase is a terminal state.
func isTerminal(phase workflow.WorkflowStatus) bool {
	return phase == workflow.WorkflowCompleted ||
		phase == workflow.WorkflowFailed ||
		phase == workflow.WorkflowCancelled
}
