package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── DeploymentHandler ────────────────────────────────────────────────────────

// DeploymentHandler serves deployment timeline and comparison endpoints.
// It reads existing DeploymentReport and WorkflowExecution data — it does NOT
// create, modify, or delete any resources.
type DeploymentHandler struct {
	reg *resource.Registry
	log *logging.Logger
}

// NewDeploymentHandler creates a handler bound to the resource registry.
func NewDeploymentHandler(reg *resource.Registry) *DeploymentHandler {
	return &DeploymentHandler{
		reg: reg,
		log: logging.NewSubsystemLogger("api.deployments", logging.LevelInfo),
	}
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

// TimelineStep is a single step in the deployment timeline.
// It maps directly from a WorkflowExecution NodeResult.
type TimelineStep struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Action string `json:"action"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// TimelineResponse is the full deployment timeline for a single deployment.
type TimelineResponse struct {
	Application      string         `json:"application"`
	DeploymentNumber int            `json:"deploymentNumber"`
	WorkflowID       string         `json:"workflowId"`
	OverallStatus    string         `json:"overallStatus"`
	StartedAt        string         `json:"startedAt,omitempty"`
	CompletedAt      string         `json:"completedAt,omitempty"`
	Duration         string         `json:"duration,omitempty"`
	Steps            []TimelineStep `json:"steps"`
}

// DeploymentSummary is a compact view of a single deployment for comparison.
type DeploymentSummary struct {
	DeploymentNumber int      `json:"deploymentNumber"`
	StartedAt        string   `json:"startedAt,omitempty"`
	CompletedAt      string   `json:"completedAt,omitempty"`
	Duration         string   `json:"duration,omitempty"`
	Repository       string   `json:"repository,omitempty"`
	Branch           string   `json:"branch,omitempty"`
	CommitSHA        string   `json:"commitSha,omitempty"`
	DetectedRuntime  string   `json:"detectedRuntime,omitempty"`
	Buildpack        string   `json:"buildpack,omitempty"`
	BuildSuccess     bool     `json:"buildSuccess"`
	RuntimeName      string   `json:"runtimeName,omitempty"`
	Environment      string   `json:"environment,omitempty"`
	ArtifactType     string   `json:"artifactType,omitempty"`
	HealthStatus     string   `json:"healthStatus"`
	Endpoint         string   `json:"endpoint,omitempty"`
	WorkflowSteps    int      `json:"workflowSteps"`
	Errors           []string `json:"errors,omitempty"`
}

// NodeComparison compares a single workflow step between two deployments.
type NodeComparison struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Action     string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	FromResult string `json:"fromResult,omitempty"`
	ToResult   string `json:"toResult,omitempty"`
	FromError  string `json:"fromError,omitempty"`
	ToError    string `json:"toError,omitempty"`
	Changed    bool   `json:"changed"`
}

// ComparisonSummary highlights what changed between two deployments.
type ComparisonSummary struct {
	StatusChanged    bool   `json:"statusChanged"`
	HealthChanged    bool   `json:"healthChanged"`
	DurationChanged  bool   `json:"durationChanged"`
	DurationDiff     string `json:"durationDiff,omitempty"`
	CommitChanged    bool   `json:"commitChanged"`
	BuildChanged     bool   `json:"buildChanged"`
	TotalStepsMatch  bool   `json:"totalStepsMatch"`
	ChangedNodeCount int    `json:"changedNodeCount"`
}

// ComparisonResponse is the full comparison between two deployments.
type ComparisonResponse struct {
	From           DeploymentSummary `json:"from"`
	To             DeploymentSummary `json:"to"`
	NodeComparison []NodeComparison  `json:"nodeComparison,omitempty"`
	Summary        ComparisonSummary `json:"summary"`
}

// ── GET /api/v1/applications/{id}/deployments/{number}/timeline ──────────────

// Timeline returns the step-by-step timeline for a specific deployment.
//
//	GET /api/v1/applications/{id}/deployments/{number}/timeline
//
// The response includes:
//   - Deployment metadata (status, timing)
//   - Ordered list of workflow steps with per-step status, result, and error
//
// Steps are ordered as they appear in the WorkflowExecution node results.
// This data is read entirely from existing resources — no new computation.
func (dh *DeploymentHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	numberStr := r.PathValue("number")

	if id == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}
	if numberStr == "" {
		BadRequest(w, "MISSING_NUMBER", "Deployment number is required")
		return
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		BadRequest(w, "INVALID_NUMBER", "Deployment number must be an integer")
		return
	}
	if number < 1 {
		BadRequest(w, "INVALID_NUMBER", "Deployment number must be positive")
		return
	}

	// Look up the application.
	app, err := dh.getApplication(id)
	if err != nil {
		NotFound(w, "APPLICATION_NOT_FOUND", fmt.Sprintf("Application %q not found", id))
		return
	}

	// Find the deployment report by number.
	report := dh.findReportByNumber(app, number)
	if report == nil {
		NotFound(w, "DEPLOYMENT_NOT_FOUND", fmt.Sprintf("Deployment #%d not found for application %q", number, id))
		return
	}

	// Get the workflow execution for this deployment.
	exec, err := dh.getWorkflowExecution(report.WorkflowID)
	if err != nil {
		// Workflow execution data might be gone (e.g., pruned). Return what we have.
		dh.log.Warn("workflow execution not found for timeline", "workflowId", report.WorkflowID, "error", err)
		OK(w, TimelineResponse{
			Application:      id,
			DeploymentNumber: number,
			WorkflowID:       report.WorkflowID,
			OverallStatus:    "unknown",
			StartedAt:        formatTime(report.StartedAt),
			CompletedAt:      formatTime(report.CompletedAt),
			Duration:         report.Duration,
			Steps:            []TimelineStep{},
		})
		return
	}

	execStatus := exec.GetStatus().(*workflow.WorkflowExecutionStatus) //nolint:errcheck
	steps := extractTimelineSteps(execStatus)

	overallStatus := string(execStatus.Phase)
	if overallStatus == "" {
		overallStatus = workflowStatusFromReport(report)
	}

	OK(w, TimelineResponse{
		Application:      id,
		DeploymentNumber: number,
		WorkflowID:       report.WorkflowID,
		OverallStatus:    overallStatus,
		StartedAt:        execStatus.StartedAt,
		CompletedAt:      execStatus.FinishedAt,
		Duration:         execStatus.Duration,
		Steps:            steps,
	})
}

// ── GET /api/v1/applications/{id}/deployments/compare ───────────────────────

// Compare returns a side-by-side comparison of two deployments.
//
//	GET /api/v1/applications/{id}/deployments/compare?from=41&to=42
//
// The response includes:
//   - Side-by-side deployment metadata (commit, runtime, duration, health)
//   - Per-step comparison (status, result, error)
//   - Summary of changes
func (dh *DeploymentHandler) Compare(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if id == "" {
		BadRequest(w, "MISSING_ID", "Application ID is required")
		return
	}
	if fromStr == "" || toStr == "" {
		BadRequest(w, "MISSING_PARAMS", "Both 'from' and 'to' query parameters are required")
		return
	}

	fromNum, err := strconv.Atoi(fromStr)
	if err != nil {
		BadRequest(w, "INVALID_FROM", "from must be an integer")
		return
	}
	toNum, err := strconv.Atoi(toStr)
	if err != nil {
		BadRequest(w, "INVALID_TO", "to must be an integer")
		return
	}
	if fromNum < 1 || toNum < 1 {
		BadRequest(w, "INVALID_NUMBER", "Deployment numbers must be positive")
		return
	}

	// Look up the application.
	app, err := dh.getApplication(id)
	if err != nil {
		NotFound(w, "APPLICATION_NOT_FOUND", fmt.Sprintf("Application %q not found", id))
		return
	}

	// Find both reports.
	fromReport := dh.findReportByNumber(app, fromNum)
	if fromReport == nil {
		NotFound(w, "FROM_DEPLOYMENT_NOT_FOUND", fmt.Sprintf("Deployment #%d not found", fromNum))
		return
	}
	toReport := dh.findReportByNumber(app, toNum)
	if toReport == nil {
		NotFound(w, "TO_DEPLOYMENT_NOT_FOUND", fmt.Sprintf("Deployment #%d not found", toNum))
		return
	}

	// Get workflow executions for per-step comparison.
	fromExec, fromErr := dh.getWorkflowExecution(fromReport.WorkflowID)
	toExec, toErr := dh.getWorkflowExecution(toReport.WorkflowID)

	// Build per-step comparison.
	nodeCmp := buildNodeComparison(fromExec, fromErr, toExec, toErr)

	// Build summary.
	summary := buildComparisonSummary(fromReport, toReport, nodeCmp)

	resp := ComparisonResponse{
		From:           reportToSummary(fromReport),
		To:             reportToSummary(toReport),
		NodeComparison: nodeCmp,
		Summary:        summary,
	}

	OK(w, resp)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// getApplication retrieves an Application resource by ID.
func (dh *DeploymentHandler) getApplication(id string) (*application.Application, error) {
	res, err := dh.reg.Get(application.Kind, id)
	if err != nil {
		return nil, err
	}
	app, ok := res.(*application.Application)
	if !ok {
		return nil, fmt.Errorf("resource %q is not an Application", id)
	}
	return app, nil
}

// getWorkflowExecution retrieves a WorkflowExecution resource by ID.
func (dh *DeploymentHandler) getWorkflowExecution(id string) (*workflow.WorkflowExecution, error) {
	if id == "" {
		return nil, fmt.Errorf("empty workflow ID")
	}
	res, err := dh.reg.Get(workflow.WorkflowExecutionKind, id)
	if err != nil {
		return nil, err
	}
	exec, ok := res.(*workflow.WorkflowExecution)
	if !ok {
		return nil, fmt.Errorf("resource %q is not a WorkflowExecution", id)
	}
	return exec, nil
}

// findReportByNumber finds a DeploymentReport by its deployment number.
func (dh *DeploymentHandler) findReportByNumber(app *application.Application, number int) *application.DeploymentReport {
	for i := range app.Status_.DeploymentHistory {
		if app.Status_.DeploymentHistory[i].DeploymentNumber == number {
			return &app.Status_.DeploymentHistory[i]
		}
	}
	return nil
}

// extractTimelineSteps converts WorkflowExecution node results to TimelineSteps.
func extractTimelineSteps(status *workflow.WorkflowExecutionStatus) []TimelineStep {
	if status == nil || len(status.NodeResults) == 0 {
		return []TimelineStep{}
	}
	steps := make([]TimelineStep, 0, len(status.NodeResults))
	for _, nr := range status.NodeResults {
		step := TimelineStep{
			ID:     nr.ID,
			Name:   nr.Name,
			Action: nr.Action,
			Status: nr.Status,
			Result: nr.Result,
		}
		if nr.Error != "" {
			step.Error = nr.Error
		}
		steps = append(steps, step)
	}
	return steps
}

// reportToSummary converts a DeploymentReport to a DeploymentSummary DTO.
func reportToSummary(r *application.DeploymentReport) DeploymentSummary {
	return DeploymentSummary{
		DeploymentNumber: r.DeploymentNumber,
		StartedAt:        formatTime(r.StartedAt),
		CompletedAt:      formatTime(r.CompletedAt),
		Duration:         r.Duration,
		Repository:       r.Repository,
		Branch:           r.Branch,
		CommitSHA:        r.CommitSHA,
		DetectedRuntime:  r.DetectedRuntime,
		Buildpack:        r.Buildpack,
		BuildSuccess:     r.BuildSuccess,
		RuntimeName:      r.RuntimeName,
		Environment:      r.Environment,
		ArtifactType:     r.ArtifactType,
		HealthStatus:     r.HealthStatus,
		Endpoint:         r.Endpoint,
		WorkflowSteps:    r.WorkflowSteps,
		Errors:           r.Errors,
	}
}

// buildNodeComparison compares node results from two workflow executions.
func buildNodeComparison(fromExec *workflow.WorkflowExecution, fromErr error, toExec *workflow.WorkflowExecution, toErr error) []NodeComparison {
	// If either execution is unavailable, we can't do per-step comparison.
	if fromErr != nil || toErr != nil {
		return nil
	}

	fromStatus := fromExec.GetStatus().(*workflow.WorkflowExecutionStatus) //nolint:errcheck
	toStatus := toExec.GetStatus().(*workflow.WorkflowExecutionStatus)     //nolint:errcheck

	fromNodes := fromStatus.NodeResults
	toNodes := toStatus.NodeResults

	// Match nodes by ID.
	fromMap := make(map[string]workflow.NodeResult, len(fromNodes))
	for _, n := range fromNodes {
		fromMap[n.ID] = n
	}

	// Collect all unique node IDs.
	seen := make(map[string]bool)
	var allIDs []string
	for _, n := range fromNodes {
		if !seen[n.ID] {
			allIDs = append(allIDs, n.ID)
			seen[n.ID] = true
		}
	}
	for _, n := range toNodes {
		if !seen[n.ID] {
			allIDs = append(allIDs, n.ID)
			seen[n.ID] = true
		}
	}

	var result []NodeComparison
	for _, id := range allIDs {
		fn, fOK := fromMap[id]
		var tn workflow.NodeResult
		tOK := false
		for _, n := range toNodes {
			if n.ID == id {
				tn = n
				tOK = true
				break
			}
		}

		fromStat := ""
		fromRes := ""
		fromErrStr := ""
		toStat := ""
		toRes := ""
		toErrStr := ""

		if fOK {
			fromStat = fn.Status
			fromRes = fn.Result
			fromErrStr = fn.Error
		}
		if tOK {
			toStat = tn.Status
			toRes = tn.Result
			toErrStr = tn.Error
		}

		changed := fromStat != toStat || fromRes != toRes || fromErrStr != toErrStr

		name := ""
		action := ""
		if fOK {
			name = fn.Name
			action = fn.Action
		} else if tOK {
			name = tn.Name
			action = tn.Action
		}

		result = append(result, NodeComparison{
			ID:         id,
			Name:       name,
			Action:     action,
			FromStatus: fromStat,
			ToStatus:   toStat,
			FromResult: fromRes,
			ToResult:   toRes,
			FromError:  fromErrStr,
			ToError:    toErrStr,
			Changed:    changed,
		})
	}

	return result
}

// buildComparisonSummary computes a summary of changes between two reports.
func buildComparisonSummary(from, to *application.DeploymentReport, nodes []NodeComparison) ComparisonSummary {
	summary := ComparisonSummary{
		StatusChanged:   from.HealthStatus != to.HealthStatus,
		HealthChanged:   from.HealthStatus != to.HealthStatus,
		DurationChanged: from.Duration != to.Duration,
		CommitChanged:   from.CommitSHA != to.CommitSHA,
		BuildChanged:    from.BuildSuccess != to.BuildSuccess || from.DetectedRuntime != to.DetectedRuntime,
		TotalStepsMatch: from.WorkflowSteps == to.WorkflowSteps,
	}

	if summary.DurationChanged {
		summary.DurationDiff = computeDurationDiff(from.Duration, to.Duration)
	}

	changedCount := 0
	for _, n := range nodes {
		if n.Changed {
			changedCount++
		}
	}
	summary.ChangedNodeCount = changedCount

	return summary
}

// computeDurationDiff computes a human-readable duration difference.
func computeDurationDiff(from, to string) string {
	// Simple heuristic: if they differ, show which is faster.
	// A full parse would require time.Duration parsing, but we
	// store durations as human-readable strings like "8.2s".
	// For now, return a directional indicator.
	return fmt.Sprintf("%s → %s", from, to)
}

// formatTime formats a time.Time as an RFC3339 string, or empty if zero.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// workflowStatusFromReport derives an overall status string from a DeploymentReport
// when the WorkflowExecution is unavailable.
func workflowStatusFromReport(r *application.DeploymentReport) string {
	if len(r.Errors) > 0 {
		return "Failed"
	}
	if r.HealthStatus == application.HealthHealthy {
		return "Completed"
	}
	return "Unknown"
}

// Ensure we never silently import.
var _ = strings.TrimSpace
