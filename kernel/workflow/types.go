// Package workflow provides a DAG-based execution engine for CloudOS.
//
// A workflow separates definition (the immutable blueprint) from execution
// (the runtime instance). The definition is a DAG of nodes; each run tracks
// the state of every node through its lifecycle.
//
// Node hierarchy (extensible):
//
//	Node (interface)
//	├── TaskNode       — executes an action against the kernel
//	├── ConditionNode  — branches based on previous node output (future)
//	├── ParallelNode   — fan-out / fan-in (future)
//	├── DelayNode      — waits for a duration or time (future)
//	├── ApprovalNode   — blocks until human approval (future)
//	├── EventNode      — waits for an external event (future)
//	└── EndNode        — terminal marker (future)
package workflow

import (
	"fmt"
	"time"
)

// ── API Version ─────────────────────────────────────────────────────────────
//
// WorkflowAPIVersion is the frozen version of the Workflow API.
// Per ADR-0011, this contract is declared v1.0 and will only receive
// additive extensions via optional interfaces.
const WorkflowAPIVersion = "workflow.cloudos.io/v1"

// ── Status Enums ─────────────────────────────────────────────────────────

// WorkflowStatus represents the lifecycle state of a WorkflowRun.
type WorkflowStatus string

const (
	WorkflowPending   WorkflowStatus = "pending"
	WorkflowRunning   WorkflowStatus = "running"
	WorkflowPaused    WorkflowStatus = "paused"
	WorkflowCancelled WorkflowStatus = "cancelled"
	WorkflowCompleted WorkflowStatus = "completed"
	WorkflowFailed    WorkflowStatus = "failed"
)

func (s WorkflowStatus) Valid() bool {
	switch s {
	case WorkflowPending, WorkflowRunning, WorkflowPaused,
		WorkflowCancelled, WorkflowCompleted, WorkflowFailed:
		return true
	default:
		return false
	}
}

// NodeStatus represents the state of a single node within a run.
type NodeStatus string

const (
	NodePending   NodeStatus = "pending"
	NodeRunning   NodeStatus = "running"
	NodeSucceeded NodeStatus = "succeeded"
	NodeFailed    NodeStatus = "failed"
	NodeSkipped   NodeStatus = "skipped"
	NodeCancelled NodeStatus = "cancelled"
)

func (s NodeStatus) Valid() bool {
	switch s {
	case NodePending, NodeRunning, NodeSucceeded,
		NodeFailed, NodeSkipped, NodeCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the status is a terminal state.
func (s NodeStatus) IsTerminal() bool {
	return s == NodeSucceeded || s == NodeFailed || s == NodeSkipped || s == NodeCancelled
}

// ── Node Type Enum ───────────────────────────────────────────────────────

// NodeType identifies the kind of a workflow node.
type NodeType string

const (
	NodeTypeTask     NodeType = "task"
	NodeTypeEnd      NodeType = "end"
	// Future types — not yet implemented but reserved as part of the type system:
	// NodeTypeCondition NodeType = "condition"
	// NodeTypeParallel  NodeType = "parallel"
	// NodeTypeDelay     NodeType = "delay"
	// NodeTypeApproval  NodeType = "approval"
	// NodeTypeEvent     NodeType = "event"
)

// ── Node Interface ───────────────────────────────────────────────────────

// Node is the interface implemented by every workflow node type.
type Node interface {
	// ID returns the unique identifier of this node within the workflow.
	ID() string
	// Name returns a human-readable name.
	Name() string
	// Type returns the node type discriminator.
	Type() NodeType
	// Status returns the current execution status within a run.
	Status() NodeStatus
	// SetStatus updates the node's status.
	SetStatus(s NodeStatus)
	// Dependencies returns the IDs of nodes that must complete before this one.
	Dependencies() []string
	// Timeout returns the maximum duration this node is allowed to run.
	Timeout() time.Duration
}

// ── Concrete Node Types ──────────────────────────────────────────────────

// RetryPolicy defines how a node should be retried on failure.
type RetryPolicy struct {
	MaxRetries  int           `json:"max_retries"`
	BackoffBase time.Duration `json:"backoff_base"`
	BackoffMax  time.Duration `json:"backoff_max"`
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  3,
		BackoffBase: 100 * time.Millisecond,
		BackoffMax:  10 * time.Second,
	}
}

// TaskNode executes a single action against the CloudOS kernel.
// This is the primary workhorse node — all current intent types produce TaskNodes.
type TaskNode struct {
	id           string
	name         string
	status       NodeStatus
	deps         []string
	Action       string        `json:"action"`        // e.g. "resource.create", "controller.list"
	Target       string        `json:"target"`        // e.g. "Project:my-app"
	RetryPolicy  *RetryPolicy  `json:"retry_policy,omitempty"`
	TimeoutVal   time.Duration `json:"timeout,omitempty"`
	Result       string        `json:"result,omitempty"`       // populated after execution
	ErrorVal     string        `json:"error,omitempty"`        // populated on failure
	RetryCount   int           `json:"retry_count,omitempty"`  // current retry attempt
}

// NewTaskNode creates a TaskNode with the given id, name, action, and target.
func NewTaskNode(id, name, action, target string, deps ...string) *TaskNode {
	return &TaskNode{
		id:     id,
		name:   name,
		status: NodePending,
		deps:   deps,
		Action: action,
		Target: target,
	}
}

func (n *TaskNode) ID() string            { return n.id }
func (n *TaskNode) Name() string           { return n.name }
func (n *TaskNode) Type() NodeType         { return NodeTypeTask }
func (n *TaskNode) Status() NodeStatus     { return n.status }
func (n *TaskNode) SetStatus(s NodeStatus) { n.status = s }
func (n *TaskNode) Dependencies() []string { return n.deps }
func (n *TaskNode) Timeout() time.Duration { return n.TimeoutVal }

// EndNode is a terminal marker that has no action and succeeds immediately.
// Every workflow definition should have exactly one EndNode.
type EndNode struct {
	id     string
	name   string
	status NodeStatus
	deps   []string
}

// NewEndNode creates an EndNode with the given dependencies.
func NewEndNode(id string, deps ...string) *EndNode {
	return &EndNode{
		id:     id,
		name:   "Complete",
		status: NodePending,
		deps:   deps,
	}
}

func (n *EndNode) ID() string            { return n.id }
func (n *EndNode) Name() string           { return n.name }
func (n *EndNode) Type() NodeType         { return NodeTypeEnd }
func (n *EndNode) Status() NodeStatus     { return n.status }
func (n *EndNode) SetStatus(s NodeStatus) { n.status = s }
func (n *EndNode) Dependencies() []string { return n.deps }
func (n *EndNode) Timeout() time.Duration { return 0 }

// ── WorkflowDefinition (Immutable Blueprint) ─────────────────────────────

// WorkflowDefinition is the immutable blueprint for a workflow execution.
// It is produced by a Builder (from an ExecutionPlan) and referenced by runs.
type WorkflowDefinition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Nodes     []Node `json:"nodes"`
	CreatedAt string `json:"created_at"`
}

// NodeByID returns the node with the given ID, or nil if not found.
func (wd *WorkflowDefinition) NodeByID(id string) Node {
	for _, n := range wd.Nodes {
		if n.ID() == id {
			return n
		}
	}
	return nil
}

// ── WorkflowRun (Execution Instance) ─────────────────────────────────────

// WorkflowRun is a single execution instance of a WorkflowDefinition.
// It tracks the live state of all nodes through the execution lifecycle.
type WorkflowRun struct {
	ID           string          `json:"id"`
	DefinitionID string          `json:"definition_id"`
	Status       WorkflowStatus  `json:"status"`
	Nodes        []Node          `json:"nodes"`
	Result       *WorkflowResult `json:"result,omitempty"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
	CompletedAt  string          `json:"completed_at,omitempty"`
}

// WorkflowResult holds the final outcome of a workflow run.
type WorkflowResult struct {
	Success bool          `json:"success"`
	Summary string        `json:"summary"`
	Details []ResultItem  `json:"details,omitempty"`
}

// ResultItem is a single item in a workflow result.
type ResultItem struct {
	Message string `json:"message"`
	Type    string `json:"type"`    // "info", "success", "error", "warning"
	Detail  string `json:"detail,omitempty"`
}

// NodeByID returns the node with the given ID from this run, or nil.
func (wr *WorkflowRun) NodeByID(id string) Node {
	for _, n := range wr.Nodes {
		if n.ID() == id {
			return n
		}
	}
	return nil
}

// CompletedCount returns the number of nodes in a terminal state.
func (wr *WorkflowRun) CompletedCount() int {
	count := 0
	for _, n := range wr.Nodes {
		if n.Status().IsTerminal() {
			count++
		}
	}
	return count
}

// Progress returns the completion ratio as a float (0.0 – 1.0).
func (wr *WorkflowRun) Progress() float64 {
	if len(wr.Nodes) == 0 {
		return 1.0
	}
	return float64(wr.CompletedCount()) / float64(len(wr.Nodes))
}

// ── Utility ──────────────────────────────────────────────────────────────

// NowUTC returns the current UTC time formatted as RFC3339.
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Ensure Node interface is satisfied at compile time.
var _ Node = (*TaskNode)(nil)
var _ Node = (*EndNode)(nil)

// ValidateDefinition checks that a WorkflowDefinition is structurally valid.
func ValidateDefinition(wd *WorkflowDefinition) error {
	if wd.ID == "" {
		return fmt.Errorf("workflow definition: empty ID")
	}
	if wd.Name == "" {
		return fmt.Errorf("workflow definition: empty name")
	}
	if len(wd.Nodes) == 0 {
		return fmt.Errorf("workflow definition: no nodes")
	}

	ids := make(map[string]bool)
	for _, n := range wd.Nodes {
		if n.ID() == "" {
			return fmt.Errorf("workflow definition: node with empty ID")
		}
		if ids[n.ID()] {
			return fmt.Errorf("workflow definition: duplicate node ID %q", n.ID())
		}
		ids[n.ID()] = true

		if !n.Type().Valid() {
			return fmt.Errorf("workflow definition: node %q has unknown type %q", n.ID(), n.Type())
		}
	}
	return nil
}

// Valid returns true if the NodeType is a known type.
func (nt NodeType) Valid() bool {
	switch nt {
	case NodeTypeTask, NodeTypeEnd:
		return true
	default:
		return false
	}
}
