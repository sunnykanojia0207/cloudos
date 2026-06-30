package workflow

import (
	"fmt"
	"time"

	"github.com/cloudos/cloudos/kernel/resource"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// WorkflowExecutionKind is the resource kind string for WorkflowExecution.
	WorkflowExecutionKind = "WorkflowExecution"
)

// ── ExecutionCondition ─────────────────────────────────────────────────────

// ExecutionCondition represents a single status condition for a workflow execution.
// Modeled after Kubernetes condition types for dashboard and automation compatibility.
type ExecutionCondition struct {
	// Type is the condition type (e.g. "Scheduled", "Running", "Completed", "Failed").
	Type string `json:"type"`

	// Status is one of "True", "False", "Unknown".
	Status string `json:"status"`

	// Reason is a machine-readable reason code.
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable explanation.
	Message string `json:"message,omitempty"`

	// LastTransitionTime is when the condition last changed.
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

// Condition status constants.
const (
	ConditionTrue    = "True"
	ConditionFalse   = "False"
	ConditionUnknown = "Unknown"
)

// Condition type constants for WorkflowExecution lifecycle.
const (
	ConditionScheduled = "Scheduled"
	ConditionRunning   = "Running"
	ConditionCompleted = "Completed"
	ConditionFailed    = "Failed"
	ConditionCancelled = "Cancelled"
	ConditionPaused    = "Paused"
)

// ── Spec ───────────────────────────────────────────────────────────────────

// WorkflowExecutionSpec defines the desired state of a workflow execution.
type WorkflowExecutionSpec struct {
	// WorkflowID references the WorkflowDefinition this execution is based on.
	WorkflowID string `json:"workflowId"`

	// IntentID is the intent that triggered this execution (if any).
	IntentID string `json:"intentId,omitempty"`

	// RequestedBy identifies who or what requested this execution.
	RequestedBy string `json:"requestedBy,omitempty"`

	// Priority controls execution ordering (higher = more urgent).
	Priority int `json:"priority,omitempty"`

	// Parameters are key-value inputs passed from the intent.
	Parameters map[string]string `json:"parameters,omitempty"`

	// Timeout is the maximum duration for the entire execution.
	Timeout string `json:"timeout,omitempty"`
}

// ── Status ─────────────────────────────────────────────────────────────────

// NodeResult captures the output of a single workflow node.
type NodeResult struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Action string `json:"action"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// WorkflowExecutionStatus defines the observed state of a workflow execution.
type WorkflowExecutionStatus struct {
	// Phase is the high-level lifecycle phase (pending, running, paused, etc.).
	Phase WorkflowStatus `json:"phase"`

	// Progress is a 0.0–1.0 completion ratio.
	Progress float64 `json:"progress"`

	// CurrentNode is the ID of the currently executing node (if any).
	CurrentNode string `json:"currentNode,omitempty"`

	// CompletedNodes lists IDs of nodes that have succeeded.
	CompletedNodes []string `json:"completedNodes,omitempty"`

	// FailedNodes lists IDs of nodes that have failed.
	FailedNodes []string `json:"failedNodes,omitempty"`

	// TotalNodes is the total number of nodes in the workflow.
	TotalNodes int `json:"totalNodes"`

	// StartedAt is when execution began.
	StartedAt string `json:"startedAt,omitempty"`

	// FinishedAt is when execution completed.
	FinishedAt string `json:"finishedAt,omitempty"`

	// Duration is the human-readable execution duration.
	Duration string `json:"duration,omitempty"`

	// Result is the success/failure summary.
	Result string `json:"result,omitempty"`

	// Error is the failure detail (if any).
	Error string `json:"error,omitempty"`

	// NodeResults captures the output of each workflow node.
	NodeResults []NodeResult `json:"nodeResults,omitempty"`

	// Conditions provide structured lifecycle tracking.
	Conditions []ExecutionCondition `json:"conditions,omitempty"`
}

// ── Resource Type ──────────────────────────────────────────────────────────

// WorkflowExecution is a CloudOS Resource that represents a single workflow run.
// It implements resource.Resource for automatic CRUD, REST API, watch, and events.
type WorkflowExecution struct {
	metadata resource.Metadata
	spec     WorkflowExecutionSpec
	status   WorkflowExecutionStatus
}

// NewWorkflowExecution creates a new WorkflowExecution resource from a WorkflowRun.
func NewWorkflowExecution(run *WorkflowRun, spec WorkflowExecutionSpec) *WorkflowExecution {
	now := time.Now()

	status := WorkflowExecutionStatus{
		Phase:      run.Status,
		Progress:   run.Progress(),
		TotalNodes: len(run.Nodes),
		Conditions: []ExecutionCondition{
			{
				Type:               ConditionScheduled,
				Status:             ConditionTrue,
				LastTransitionTime: now,
				Reason:             "ExecutionScheduled",
				Message:            "Workflow execution has been submitted",
			},
		},
	}

	if run.Status == WorkflowRunning || run.Status == WorkflowCompleted {
		status.StartedAt = run.CreatedAt
		status.Conditions = append(status.Conditions, ExecutionCondition{
			Type:               ConditionRunning,
			Status:             ConditionTrue,
			LastTransitionTime: now,
			Reason:             "ExecutionStarted",
			Message:            "Workflow execution is running",
		})
	}

	if run.CompletedAt != "" {
		status.FinishedAt = run.CompletedAt
	}
	if run.Result != nil {
		status.Result = run.Result.Summary
		if !run.Result.Success {
			status.Error = run.Result.Summary
		}
	}

	// Populate node-level results from the run's nodes.
	for _, n := range run.Nodes {
		if tn, ok := n.(*TaskNode); ok {
			nr := NodeResult{
				ID:     tn.ID(),
				Name:   tn.Name(),
				Action: tn.Action,
				Status: string(n.Status()),
				Result: tn.Result,
			}
			if tn.ErrorVal != "" {
				nr.Error = tn.ErrorVal
			}
			status.NodeResults = append(status.NodeResults, nr)
		}
	}

	return &WorkflowExecution{
		metadata: resource.Metadata{
			ID:              run.ID,
			Name:            fmt.Sprintf("execution-%s", run.ID),
			Namespace:       resource.NamespaceDefault,
			Kind:            WorkflowExecutionKind,
			APIVersion:      resource.APIVersion,
			CreatedAt:       now,
			UpdatedAt:       now,
			ResourceVersion: 1,
			Labels: map[string]string{
				"workflow.cloudos.io/status": string(run.Status),
			},
		},
		spec:   spec,
		status: status,
	}
}

// ── Resource Interface Implementation ──────────────────────────────────────

func (e *WorkflowExecution) GetKind() string           { return WorkflowExecutionKind }
func (e *WorkflowExecution) GetMetadata() *resource.Metadata { return &e.metadata }
func (e *WorkflowExecution) GetSpec() interface{}       { return &e.spec }
func (e *WorkflowExecution) GetStatus() interface{}     { return &e.status }
func (e *WorkflowExecution) SetStatus(s interface{}) {
	if s, ok := s.(*WorkflowExecutionStatus); ok {
		e.status = *s
	}
}
func (e *WorkflowExecution) Validate() error {
	if e.metadata.ID == "" {
		return fmt.Errorf("workflow execution: empty ID")
	}
	if e.spec.WorkflowID == "" {
		return fmt.Errorf("workflow execution: empty WorkflowID")
	}
	return nil
}

// ── Conditions Helpers ────────────────────────────────────────────────────

// SetCondition adds or updates a condition on the execution status.
func (s *WorkflowExecutionStatus) SetCondition(condType, status, reason, message string) {
	now := time.Now()
	for i, c := range s.Conditions {
		if c.Type == condType {
			if c.Status != status {
				s.Conditions[i].Status = status
				s.Conditions[i].LastTransitionTime = now
			}
			s.Conditions[i].Reason = reason
			s.Conditions[i].Message = message
			return
		}
	}
	// Condition not found — append new one
	s.Conditions = append(s.Conditions, ExecutionCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	})
}

// GetCondition returns the condition with the given type, or nil if not found.
func (s *WorkflowExecutionStatus) GetCondition(condType string) *ExecutionCondition {
	for i, c := range s.Conditions {
		if c.Type == condType {
			return &s.Conditions[i]
		}
	}
	return nil
}

// IsConditionTrue returns true if the given condition type has status True.
func (s *WorkflowExecutionStatus) IsConditionTrue(condType string) bool {
	c := s.GetCondition(condType)
	return c != nil && c.Status == ConditionTrue
}

// ── Status Sync ───────────────────────────────────────────────────────────

// SyncFromRun updates the execution status from a WorkflowRun.
// This is called by the Engine after each scheduling cycle.
func (s *WorkflowExecutionStatus) SyncFromRun(run *WorkflowRun) {
	s.Phase = run.Status
	s.Progress = run.Progress()
	s.StartedAt = run.CreatedAt

	// Collect node states
	var completed, failed []string
	for _, n := range run.Nodes {
		switch n.Status() {
		case NodeSucceeded:
			completed = append(completed, n.ID())
		case NodeFailed, NodeCancelled:
			failed = append(failed, n.ID())
		}
		if n.Status() == NodeRunning {
			s.CurrentNode = n.ID()
		}
	}
	s.CompletedNodes = completed
	s.FailedNodes = failed
	s.TotalNodes = len(run.Nodes)

	if run.CompletedAt != "" {
		s.FinishedAt = run.CompletedAt
	}
	if run.Result != nil {
		s.Result = run.Result.Summary
		if !run.Result.Success {
			s.Error = run.Result.Summary
		}
	}

	// Populate node-level results from the run's nodes.
	s.NodeResults = nil
	for _, n := range run.Nodes {
		if tn, ok := n.(*TaskNode); ok {
			nr := NodeResult{
				ID:     tn.ID(),
				Name:   tn.Name(),
				Action: tn.Action,
				Status: string(n.Status()),
				Result: tn.Result,
			}
			if tn.ErrorVal != "" {
				nr.Error = tn.ErrorVal
			}
			s.NodeResults = append(s.NodeResults, nr)
		}
	}

	// Sync conditions based on workflow status
	switch run.Status {
	case WorkflowPending:
		s.SetCondition(ConditionScheduled, ConditionTrue, "ExecutionScheduled", "Workflow execution has been submitted")
	case WorkflowRunning:
		s.SetCondition(ConditionRunning, ConditionTrue, "ExecutionStarted", "Workflow execution is running")
		s.SetCondition(ConditionScheduled, ConditionFalse, "ExecutionStarted", "")
	case WorkflowPaused:
		s.SetCondition(ConditionPaused, ConditionTrue, "ExecutionPaused", "Workflow execution has been paused")
		s.SetCondition(ConditionRunning, ConditionFalse, "ExecutionPaused", "")
	case WorkflowCompleted:
		s.SetCondition(ConditionCompleted, ConditionTrue, "ExecutionCompleted", "Workflow execution completed successfully")
		s.SetCondition(ConditionRunning, ConditionFalse, "ExecutionCompleted", "")
	case WorkflowFailed:
		s.SetCondition(ConditionFailed, ConditionTrue, "ExecutionFailed", run.Result.Summary)
		s.SetCondition(ConditionRunning, ConditionFalse, "ExecutionFailed", "")
	case WorkflowCancelled:
		s.SetCondition(ConditionCancelled, ConditionTrue, "ExecutionCancelled", "Workflow execution was cancelled")
		s.SetCondition(ConditionRunning, ConditionFalse, "ExecutionCancelled", "")
	}
}

// Ensure WorkflowExecution implements resource.Resource at compile time.
var _ resource.Resource = (*WorkflowExecution)(nil)
