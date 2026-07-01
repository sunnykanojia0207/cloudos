package api_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Timeline Tests ───────────────────────────────────────────────────────────

func TestTimeline_EmptyAppID(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Empty app ID in path is treated as a non-matching route by Go's mux.
	code, env := httpGet(t, ts.URL+"/api/v1/applications//deployments/1/timeline")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
}

func TestTimeline_EmptyNumber(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Empty number in path is treated as a non-matching route by Go's mux.
	code, env := httpGet(t, ts.URL+"/api/v1/applications/test-app/deployments//timeline")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
}

func TestTimeline_InvalidNumber(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/test-app/deployments/abc/timeline")
	assert.Equal(t, 400, code)
	assert.False(t, env.Success)
}

func TestTimeline_AppNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/nonexistent/deployments/1/timeline")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	if env.Error != nil {
		assert.Equal(t, "APPLICATION_NOT_FOUND", env.Error.Code)
	}
}

func TestTimeline_DeploymentNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	createTestApplication(t, reg, "timeline-app-1", nil)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/timeline-app-1/deployments/99/timeline")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	assert.Equal(t, "DEPLOYMENT_NOT_FOUND", env.Error.Code)
}

func TestTimeline_NoWorkflowExecution(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	// Create an application with a deployment report but no workflow execution.
	app := createTestApplication(t, reg, "timeline-app-2", nil)
	addDeploymentReport(t, app, 1, "wf-missing", time.Now(), time.Now())
	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/timeline-app-2/deployments/1/timeline")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "timeline-app-2", data["application"])
	assert.Equal(t, float64(1), data["deploymentNumber"])
	assert.Equal(t, "wf-missing", data["workflowId"])
	assert.Equal(t, "unknown", data["overallStatus"])

	steps, ok := data["steps"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, steps)
}

func TestTimeline_WithWorkflowExecution(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	// Create a workflow execution with node results.
	wfID := "wf-timeline-full"
	createWorkflowExecutionWithNodes(t, reg, wfID, workflow.WorkflowCompleted, []workflow.NodeResult{
		{ID: "validate", Name: "Validate Application", Action: "validate", Status: "succeeded", Result: "Configuration valid"},
		{ID: "clone", Name: "Clone Source Repository", Action: "source.clone", Status: "succeeded", Result: "Cloned 42 commits"},
		{ID: "build", Name: "Build Artifact", Action: "build.execute", Status: "succeeded", Result: "Build completed, artifact=app"},
		{ID: "deploy", Name: "Deploy Application", Action: "provider.deploy", Status: "succeeded", Result: "Deployed to runtime"},
		{ID: "healthcheck", Name: "Health Check", Action: "health.check", Status: "succeeded", Result: "HTTP 200 OK"},
		{ID: "complete", Name: "Complete Deployment", Action: "complete", Status: "succeeded", Result: "Deployment #1 complete"},
	})

	// Create an application referencing this workflow execution.
	app := createTestApplication(t, reg, "timeline-app-3", nil)
	addDeploymentReport(t, app, 1, wfID, time.Now().Add(-10*time.Second), time.Now())
	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/timeline-app-3/deployments/1/timeline")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "timeline-app-3", data["application"])
	assert.Equal(t, float64(1), data["deploymentNumber"])
	assert.Equal(t, wfID, data["workflowId"])
	assert.Contains(t, data, "overallStatus")

	steps, ok := data["steps"].([]interface{})
	require.True(t, ok)
	assert.Len(t, steps, 6)

	// Verify specific steps.
	firstStep := steps[0].(map[string]interface{})
	assert.Equal(t, "validate", firstStep["id"])
	assert.Equal(t, "succeeded", firstStep["status"])
	assert.Equal(t, "Configuration valid", firstStep["result"])

	lastStep := steps[5].(map[string]interface{})
	assert.Equal(t, "complete", lastStep["id"])
	assert.Equal(t, "Deployment #1 complete", lastStep["result"])
}

func TestTimeline_DeploymentWithErrors(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	wfID := "wf-timeline-failed"
	createWorkflowExecutionWithNodes(t, reg, wfID, workflow.WorkflowFailed, []workflow.NodeResult{
		{ID: "validate", Name: "Validate Application", Action: "validate", Status: "succeeded", Result: "Configuration valid"},
		{ID: "clone", Name: "Clone Source Repository", Action: "source.clone", Status: "succeeded", Result: "Cloned 42 commits"},
		{ID: "build", Name: "Build Artifact", Action: "build.execute", Status: "failed", Result: "", Error: "Build failed: missing package"},
	})

	app := createTestApplication(t, reg, "timeline-app-4", nil)
	addDeploymentReport(t, app, 1, wfID, time.Now().Add(-5*time.Second), time.Now())
	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/timeline-app-4/deployments/1/timeline")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	data := env.Data.(map[string]interface{})
	steps, ok := data["steps"].([]interface{})
	require.True(t, ok)
	assert.Len(t, steps, 3)

	failedStep := steps[2].(map[string]interface{})
	assert.Equal(t, "failed", failedStep["status"])
	assert.Equal(t, "Build failed: missing package", failedStep["error"])
}

// ── Compare Tests ────────────────────────────────────────────────────────────

func TestCompare_EmptyAppID(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Empty app ID in path is treated as a non-matching route by Go's mux.
	code, env := httpGet(t, ts.URL+"/api/v1/applications//deployments/compare?from=1&to=2")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
}

func TestCompare_MissingParams(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/test-app/deployments/compare")
	assert.Equal(t, 400, code)
	assert.False(t, env.Success)
	assert.Equal(t, "MISSING_PARAMS", env.Error.Code)
}

func TestCompare_AppNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/nonexistent/deployments/compare?from=1&to=2")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
}

func TestCompare_FromDeploymentNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	createTestApplication(t, reg, "cmp-app-1", nil)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/cmp-app-1/deployments/compare?from=1&to=2")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	assert.Equal(t, "FROM_DEPLOYMENT_NOT_FOUND", env.Error.Code)
}

func TestCompare_ToDeploymentNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	app := createTestApplication(t, reg, "cmp-app-2", nil)
	addDeploymentReport(t, app, 1, "wf-1", time.Now().Add(-20*time.Second), time.Now().Add(-10*time.Second))
	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/cmp-app-2/deployments/compare?from=1&to=2")
	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	assert.Equal(t, "TO_DEPLOYMENT_NOT_FOUND", env.Error.Code)
}

func TestCompare_TwoDeployments(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	// Create workflow executions for both deployments.
	createWorkflowExecutionWithNodes(t, reg, "wf-cmp-1", workflow.WorkflowCompleted, []workflow.NodeResult{
		{ID: "validate", Name: "Validate Application", Action: "validate", Status: "succeeded", Result: "Configuration valid"},
		{ID: "build", Name: "Build Artifact", Action: "build.execute", Status: "succeeded", Result: "Build completed"},
		{ID: "deploy", Name: "Deploy Application", Action: "provider.deploy", Status: "succeeded", Result: "Deployed"},
	})

	createWorkflowExecutionWithNodes(t, reg, "wf-cmp-2", workflow.WorkflowCompleted, []workflow.NodeResult{
		{ID: "validate", Name: "Validate Application", Action: "validate", Status: "succeeded", Result: "Configuration valid"},
		{ID: "build", Name: "Build Artifact", Action: "build.execute", Status: "succeeded", Result: "Build completed (cached)"},
		{ID: "deploy", Name: "Deploy Application", Action: "provider.deploy", Status: "succeeded", Result: "Deployed"},
	})

	// Create an application with two deployments.
	app := createTestApplication(t, reg, "cmp-app-3", nil)

	addDeploymentReport(t, app, 1, "wf-cmp-1", time.Now().Add(-20*time.Second), time.Now().Add(-10*time.Second))
	addDeploymentReport(t, app, 2, "wf-cmp-2", time.Now().Add(-10*time.Second), time.Now())

	// Modify reports to simulate different metadata.
	app.Status_.DeploymentHistory[1].CommitSHA = "abc123"
	app.Status_.DeploymentHistory[1].Branch = "main"
	app.Status_.DeploymentHistory[1].DetectedRuntime = "Go 1.24"
	app.Status_.DeploymentHistory[1].BuildSuccess = true
	app.Status_.DeploymentHistory[1].HealthStatus = "Healthy"
	app.Status_.DeploymentHistory[1].Duration = "5.2s"
	app.Status_.DeploymentHistory[1].WorkflowSteps = 3

	app.Status_.DeploymentHistory[0].CommitSHA = "def456"
	app.Status_.DeploymentHistory[0].Branch = "main"
	app.Status_.DeploymentHistory[0].DetectedRuntime = "Go 1.24"
	app.Status_.DeploymentHistory[0].BuildSuccess = true
	app.Status_.DeploymentHistory[0].HealthStatus = "Healthy"
	app.Status_.DeploymentHistory[0].Duration = "6.1s"
	app.Status_.DeploymentHistory[0].WorkflowSteps = 3

	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/cmp-app-3/deployments/compare?from=1&to=2")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})

	// Verify from/to summaries.
	fromData, ok := data["from"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), fromData["deploymentNumber"])
	assert.Equal(t, "abc123", fromData["commitSha"])

	toData, ok := data["to"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), toData["deploymentNumber"])
	assert.Equal(t, "def456", toData["commitSha"])

	// Verify node comparison.
	nodes, ok := data["nodeComparison"].([]interface{})
	require.True(t, ok)
	assert.Len(t, nodes, 3)

	// The build step should be changed (different result).
	buildNode := nodes[1].(map[string]interface{})
	assert.Equal(t, "build", buildNode["id"])
	assert.True(t, buildNode["changed"].(bool))
	assert.Equal(t, "Build completed", buildNode["fromResult"])
	assert.Equal(t, "Build completed (cached)", buildNode["toResult"])

	// Verify summary.
	summary, ok := data["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, summary["commitChanged"].(bool))
	assert.True(t, summary["durationChanged"].(bool))
	assert.True(t, summary["totalStepsMatch"].(bool))
	assert.Equal(t, float64(1), summary["changedNodeCount"])
}

func TestCompare_IdenticalDeployments(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	// Same workflow execution used for both deployments.
	createWorkflowExecutionWithNodes(t, reg, "wf-identical", workflow.WorkflowCompleted, []workflow.NodeResult{
		{ID: "validate", Name: "Validate Application", Action: "validate", Status: "succeeded", Result: "Configuration valid"},
		{ID: "build", Name: "Build Artifact", Action: "build.execute", Status: "succeeded", Result: "Build completed"},
	})

	app := createTestApplication(t, reg, "cmp-app-4", nil)
	addDeploymentReport(t, app, 1, "wf-identical", time.Now().Add(-20*time.Second), time.Now().Add(-10*time.Second))
	addDeploymentReport(t, app, 2, "wf-identical", time.Now().Add(-10*time.Second), time.Now())

	// Same metadata for both deployments.
	app.Status_.DeploymentHistory[1].CommitSHA = "abc123"
	app.Status_.DeploymentHistory[1].Duration = "5.0s"
	app.Status_.DeploymentHistory[1].HealthStatus = "Healthy"
	app.Status_.DeploymentHistory[1].BuildSuccess = true
	app.Status_.DeploymentHistory[1].WorkflowSteps = 2

	app.Status_.DeploymentHistory[0].CommitSHA = "abc123"
	app.Status_.DeploymentHistory[0].Duration = "5.0s"
	app.Status_.DeploymentHistory[0].HealthStatus = "Healthy"
	app.Status_.DeploymentHistory[0].BuildSuccess = true
	app.Status_.DeploymentHistory[0].WorkflowSteps = 2

	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/cmp-app-4/deployments/compare?from=1&to=2")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	data := env.Data.(map[string]interface{})
	summary, ok := data["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.False(t, summary["commitChanged"].(bool))
	assert.False(t, summary["durationChanged"].(bool))
	assert.True(t, summary["totalStepsMatch"].(bool))
	assert.Equal(t, float64(0), summary["changedNodeCount"])
}

func TestCompare_DeploymentWithErrors(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)
	reg := k.ResourceRegistry()

	wfID := "wf-cmp-failed"
	createWorkflowExecutionWithNodes(t, reg, wfID, workflow.WorkflowFailed, []workflow.NodeResult{
		{ID: "validate", Name: "Validate Application", Action: "validate", Status: "succeeded", Result: "Configuration valid"},
		{ID: "build", Name: "Build Artifact", Action: "build.execute", Status: "failed", Error: "Build failed"},
	})

	app := createTestApplication(t, reg, "cmp-app-5", nil)
	addDeploymentReport(t, app, 1, "wf-ok", time.Now().Add(-20*time.Second), time.Now().Add(-10*time.Second))
	addDeploymentReport(t, app, 2, wfID, time.Now().Add(-10*time.Second), time.Now())

	// First deployment (success)
	app.Status_.DeploymentHistory[1].HealthStatus = "Healthy"
	app.Status_.DeploymentHistory[1].BuildSuccess = true
	app.Status_.DeploymentHistory[1].Branch = "main"
	app.Status_.DeploymentHistory[1].Duration = "4.0s"
	app.Status_.DeploymentHistory[1].WorkflowSteps = 2

	// Second deployment (failed)
	app.Status_.DeploymentHistory[0].HealthStatus = "Error"
	app.Status_.DeploymentHistory[0].BuildSuccess = false
	app.Status_.DeploymentHistory[0].Branch = "main"
	app.Status_.DeploymentHistory[0].Duration = "3.2s"
	app.Status_.DeploymentHistory[0].WorkflowSteps = 2
	app.Status_.DeploymentHistory[0].Errors = []string{"Build failed: missing package"}

	updateApplication(t, reg, app)

	code, env := httpGet(t, ts.URL+"/api/v1/applications/cmp-app-5/deployments/compare?from=1&to=2")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	data := env.Data.(map[string]interface{})

	toData := data["to"].(map[string]interface{})
	assert.Equal(t, "Error", toData["healthStatus"])
	assert.Equal(t, false, toData["buildSuccess"])
	errors := toData["errors"].([]interface{})
	assert.Len(t, errors, 1)

	summary := data["summary"].(map[string]interface{})
	assert.True(t, summary["healthChanged"].(bool))
	assert.True(t, summary["buildChanged"].(bool))
	assert.True(t, summary["durationChanged"].(bool))
}

// ── Test Helpers ─────────────────────────────────────────────────────────────

// createTestApplication creates an Application resource for testing.
func createTestApplication(t *testing.T, reg *resource.Registry, id string, envVars map[string]string) *application.Application {
	t.Helper()

	spec := application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  "https://github.com/test/repo",
		},
		Runtime: application.ApplicationRuntime{
			Type: application.RuntimeGo,
		},
		Environment: envVars,
	}
	app := application.NewApplication(id, id, spec)
	app.Status_.Phase = application.PhaseRunning
	app.Status_.Health = application.HealthHealthy
	app.Status_.URL = "http://localhost:8080"

	err := reg.Create(context.Background(), app)
	require.NoError(t, err)
	return app
}

// addDeploymentReport adds a deployment report to the application's history.
func addDeploymentReport(t *testing.T, app *application.Application, number int, workflowID string, startedAt, completedAt time.Time) {
	t.Helper()

	report := application.DeploymentReport{
		DeploymentNumber: number,
		StartedAt:        startedAt,
		CompletedAt:      completedAt,
		Duration:         fmt.Sprintf("%.1fs", completedAt.Sub(startedAt).Seconds()),
		Repository:       "https://github.com/test/repo",
		Branch:           "main",
		CommitSHA:        "abc123",
		DetectedRuntime:  "Go 1.24",
		Buildpack:        "go",
		BuildSuccess:     true,
		RuntimeName:      "local",
		Environment:      "development",
		HealthStatus:     application.HealthHealthy,
		Endpoint:         "http://localhost:8080",
		WorkflowID:       workflowID,
		WorkflowSteps:    0, // will be set by caller if needed
	}

	// Prepend to history (newest first).
	app.Status_.DeploymentHistory = append([]application.DeploymentReport{report}, app.Status_.DeploymentHistory...)
	app.Status_.DeploymentCount = number
	app.Status_.LastReport = &app.Status_.DeploymentHistory[0]
}

// updateApplication saves the updated application back to the registry.
func updateApplication(t *testing.T, reg *resource.Registry, app *application.Application) {
	t.Helper()
	err := reg.Update(context.Background(), app)
	require.NoError(t, err)
}

// createWorkflowExecutionWithNodes creates a WorkflowExecution with the given
// phase and node results.
func createWorkflowExecutionWithNodes(t *testing.T, reg *resource.Registry, id string, phase workflow.WorkflowStatus, nodes []workflow.NodeResult) {
	t.Helper()

	spec := workflow.WorkflowExecutionSpec{
		WorkflowID: "test-workflow",
	}
	status := workflow.WorkflowExecutionStatus{
		Phase:       phase,
		Progress:    1.0,
		TotalNodes:  len(nodes),
		NodeResults: nodes,
	}

	if phase == workflow.WorkflowCompleted {
		status.Result = "Deployment completed successfully"
	} else if phase == workflow.WorkflowFailed {
		status.Result = "Deployment failed"
	}

	run := &workflow.WorkflowRun{ID: "test-run", Status: phase}
	exec := workflow.NewWorkflowExecution(run, spec)
	exec.GetMetadata().ID = id
	exec.SetStatus(&status)

	err := reg.Create(context.Background(), exec)
	require.NoError(t, err)
}


