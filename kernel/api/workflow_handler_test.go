package api_test

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Workflow Execution Handlers Tests ──────────────────────────────────────

func TestGetExecution_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/workflow-executions/nonexistent")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	if env.Error != nil {
		assert.Contains(t, env.Error.Message, "not found")
	}
}

func TestGetExecution_Success(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	// Create a WorkflowExecution via the resource registry
	id := "exec-integration-1"
	createWorkflowExecution(t, reg, id, workflow.WorkflowPending)

	// Test GET
	code, env := httpGet(t, ts.URL+"/api/v1/workflow-executions/"+id)
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	// Verify response has execution data
	data, ok := env.Data.(map[string]interface{})
	require.True(t, ok, "response data should be an object")
	assert.Equal(t, "WorkflowExecution", data["kind"])
	metadata, ok := data["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, id, metadata["id"])
}

// ── SSE Deployment Timeline Tests ──────────────────────────────────────────

func TestStreamExecutionEvents_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	resp, err := http.Get(ts.URL + "/api/v1/workflow-executions/nonexistent/events")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 404 JSON (not SSE) when execution doesn't exist
	assert.Equal(t, 404, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestStreamExecutionEvents_InitialState(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	id := "exec-sse-init"
	createWorkflowExecution(t, reg, id, workflow.WorkflowPending)

	// Connect to SSE stream
	resp, err := http.Get(ts.URL + "/api/v1/workflow-executions/" + id + "/events")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Read the first SSE event
	scanner := bufio.NewScanner(resp.Body)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if strings.HasPrefix(line, "data: ") {
			// Got a data line — we have enough
			break
		}
		if len(lines) > 20 {
			break
		}
	}

	combined := strings.Join(lines, "\n")
	assert.Contains(t, combined, "event: execution")
	assert.Contains(t, combined, `"phase":"pending"`)
}

func TestStreamExecutionEvents_TerminalState(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	id := "exec-sse-terminal"
	createWorkflowExecution(t, reg, id, workflow.WorkflowCompleted)

	// For completed state, the SSE should send one event then close
	resp, err := http.Get(ts.URL + "/api/v1/workflow-executions/" + id + "/events")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Read the event
	scanner := bufio.NewScanner(resp.Body)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 10 {
			break
		}
	}

	combined := strings.Join(lines, "\n")
	assert.Contains(t, combined, `"phase":"completed"`)
	assert.Contains(t, combined, `event: execution`)
}

// ── Helpers ─────────────────────────────────────────────────────────────────

// createWorkflowExecution creates a WorkflowExecution resource with the given
// phase and registers it in the resource registry.
func createWorkflowExecution(t *testing.T, reg *resource.Registry, id string, phase workflow.WorkflowStatus) {
	t.Helper()

	spec := workflow.WorkflowExecutionSpec{
		WorkflowID: "test-workflow",
		IntentID:   "test-intent",
	}
	status := workflow.WorkflowExecutionStatus{
		Phase:      phase,
		Progress:   0.0,
		TotalNodes: 3,
		CompletedNodes: []string{},
		NodeResults: []workflow.NodeResult{},
	}

	// Create a minimal WorkflowRun (no nodes needed for our tests)
	run := &workflow.WorkflowRun{ID: "test-run", Status: phase}
	exec := workflow.NewWorkflowExecution(run, spec)
	exec.GetMetadata().ID = id
	exec.SetStatus(&status)

	err := reg.Create(context.Background(), exec)
	require.NoError(t, err)
}

