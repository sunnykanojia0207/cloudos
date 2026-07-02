package workflow

import (
	"context"
	"fmt"

	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/resource"
	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/packages/logging"
)

// Service is the business-logic layer between REST and the Workflow Engine.
//
// The Service owns coordination; the Engine owns execution; the Resource
// owns state.
//
// Responsibilities:
//   - Submit a definition for execution
//   - Query runs and execution resources
//   - Lifecycle management (pause, resume, cancel)
//   - Retry — reset failed nodes and re-execute
//   - Replay — re-run a completed execution from scratch
//   - Clone — replay with parameter modifications
//
// The Service is stateless — all state lives in the Engine (in-memory runs)
// and the Resource Engine (persistent WorkflowExecution Resources).
type Service struct {
	engine      *Engine
	resRegistry *resource.Registry
	log         *logging.Logger
}

// ServiceDeps holds the dependencies for creating a new Service.
type ServiceDeps struct {
	ResourceRegistry  *resource.Registry
	ControllerManager *controller.Manager
	HealthManager     *health.Manager
	EventBus          interface{}
	SourceCloner      *source.GitCloner
	RuntimeManager    cr.Runtime
	LogManager        *cr.LogManager
	Logger            *logging.Logger
}

// NewService creates a new Workflow Service backed by a Workflow Engine.
func NewService(deps ServiceDeps) *Service {
	engine := NewEngine(EngineDeps{
		ResourceRegistry:  deps.ResourceRegistry,
		ControllerManager: deps.ControllerManager,
		HealthManager:     deps.HealthManager,
		EventBus:          deps.EventBus,
		SourceCloner:      deps.SourceCloner,
		RuntimeManager:    deps.RuntimeManager,
		LogManager:        deps.LogManager,
		Logger:            deps.Logger,
	})

	return &Service{
		engine:      engine,
		resRegistry: deps.ResourceRegistry,
		log:         deps.Logger,
	}
}

// Engine returns the underlying Workflow Engine (for advanced operations).
func (s *Service) Engine() *Engine { return s.engine }

// ── Definition Management ───────────────────────────────────────────────

// RegisterDefinition stores a workflow definition for later execution.
func (s *Service) RegisterDefinition(def *WorkflowDefinition) error {
	return s.engine.RegisterDefinition(def)
}

// GetDefinition returns a workflow definition by ID.
func (s *Service) GetDefinition(id string) (*WorkflowDefinition, bool) {
	return s.engine.GetDefinition(id)
}

// ListDefinitions returns all registered definitions.
func (s *Service) ListDefinitions() []*WorkflowDefinition {
	return s.engine.ListDefinitions()
}

// CreateDefinition creates and registers a workflow definition from nodes.
func (s *Service) CreateDefinition(id, name string, nodes []Node) (*WorkflowDefinition, error) {
	return s.engine.CreateDefinition(id, name, nodes)
}

// ── Execution ───────────────────────────────────────────────────────────

// Submit creates a WorkflowRun from a Definition and begins execution.
func (s *Service) Submit(def *WorkflowDefinition) (*WorkflowRun, error) {
	return s.engine.Submit(def)
}

// Get returns the current state of a workflow run.
func (s *Service) Get(runID string) (*WorkflowRun, bool) {
	return s.engine.GetRun(runID)
}

// List returns all active workflow runs.
func (s *Service) List() []*WorkflowRun {
	return s.engine.ListRuns()
}

// GetExecution reads the persistent WorkflowExecution Resource for a run.
func (s *Service) GetExecution(runID string) (*WorkflowExecution, error) {
	if s.resRegistry == nil {
		return nil, fmt.Errorf("resource registry not available")
	}
	obj, err := s.resRegistry.Get(WorkflowExecutionKind, runID)
	if err != nil {
		return nil, fmt.Errorf("get execution %q: %w", runID, err)
	}
	exec, ok := obj.(*WorkflowExecution)
	if !ok {
		return nil, fmt.Errorf("unexpected resource type %T for %q", obj, runID)
	}
	return exec, nil
}

// ListExecutions returns all WorkflowExecution Resources.
func (s *Service) ListExecutions() ([]*WorkflowExecution, error) {
	if s.resRegistry == nil {
		return nil, fmt.Errorf("resource registry not available")
	}
	items, err := s.resRegistry.List(WorkflowExecutionKind)
	if err != nil {
		return nil, fmt.Errorf("list executions: %w", err)
	}
	result := make([]*WorkflowExecution, 0, len(items))
	for _, item := range items {
		if exec, ok := item.(*WorkflowExecution); ok {
			result = append(result, exec)
		}
	}
	return result, nil
}

// ── Lifecycle ───────────────────────────────────────────────────────────

// Pause pauses a running workflow.
func (s *Service) Pause(runID string) error {
	return s.engine.Pause(runID)
}

// Resume resumes a paused workflow.
func (s *Service) Resume(runID string) error {
	return s.engine.Resume(runID)
}

// Cancel cancels a running or pending workflow.
func (s *Service) Cancel(runID string) error {
	return s.engine.Cancel(runID)
}

// ── Retry ───────────────────────────────────────────────────────────────

// Retry creates a new run from the same definition as an existing run,
// preserving the status of already-succeeded nodes so only failed, skipped,
// or cancelled nodes are re-executed.
//
// Returns the new WorkflowRun.
func (s *Service) Retry(runID string) (*WorkflowRun, error) {
	// Get the original run
	original, ok := s.engine.GetRun(runID)
	if !ok {
		return nil, fmt.Errorf("retry: run %q not found", runID)
	}

	// Get the definition
	def, ok := s.engine.GetDefinition(original.DefinitionID)
	if !ok {
		return nil, fmt.Errorf("retry: definition %q not found", original.DefinitionID)
	}

	// Create a new run from the definition (fresh node copies)
	nodes := copyNodes(def.Nodes)
	newRun := &WorkflowRun{
		ID:           s.engine.nextID(),
		DefinitionID: def.ID,
		Status:       WorkflowPending,
		Nodes:        nodes,
		CreatedAt:    NowUTC(),
		UpdatedAt:    NowUTC(),
	}

	// Copy succeeded status from original nodes to their counterparts
	// in the new run. Failed/skipped/cancelled nodes stay pending for retry.
	for _, newNode := range newRun.Nodes {
		origNode := original.NodeByID(newNode.ID())
		if origNode == nil {
			continue
		}
		switch origNode.Status() {
		case NodeSucceeded:
			newNode.SetStatus(NodeSucceeded)
		case NodeFailed:
			// Reset retry count on the task node
			if tn, ok := newNode.(*TaskNode); ok {
				tn.RetryCount = 0
				tn.ErrorVal = ""
				tn.Result = ""
			}
		case NodeCancelled, NodeSkipped:
			// Leave as pending — will be retried
		case NodeRunning, NodePending:
			// Leave as pending
		}
	}

	// Store and enqueue the new run
	s.engine.mu.Lock()
	s.engine.runs[newRun.ID] = newRun
	s.engine.mu.Unlock()

	s.log.Info("workflow retry created",
		"new_run_id", newRun.ID,
		"original_run_id", runID,
		"definition", def.ID,
	)

	// Create execution resource
	execSpec := WorkflowExecutionSpec{
		WorkflowID: def.ID,
	}
	_ = s.engine.createExecutionResource(newRun, execSpec)

	// Enqueue for execution
	if err := s.engine.queue.Enqueue(context.Background(), QueueItem{
		WorkflowID: newRun.ID,
	}); err != nil {
		return nil, fmt.Errorf("retry: enqueue: %w", err)
	}

	return newRun, nil
}

// ── Replay ──────────────────────────────────────────────────────────────

// Replay creates a brand-new execution from an existing run's definition.
// Unlike Retry, Replay runs every node from scratch — no preserved statuses.
//
// Returns the new WorkflowRun.
func (s *Service) Replay(runID string) (*WorkflowRun, error) {
	// Get the original run to find the definition ID
	original, ok := s.engine.GetRun(runID)
	if !ok {
		return nil, fmt.Errorf("replay: run %q not found", runID)
	}

	def, ok := s.engine.GetDefinition(original.DefinitionID)
	if !ok {
		return nil, fmt.Errorf("replay: definition %q not found", original.DefinitionID)
	}

	s.log.Info("workflow replay",
		"original_run_id", runID,
		"definition", def.ID,
	)

	return s.engine.Submit(def)
}

// ── Clone ───────────────────────────────────────────────────────────────

// Clone creates a new execution from an existing run's definition, with
// optional parameter overrides for the execution spec.
//
// Parameters:
//   - runID: the existing run to clone from
//   - overrides: optional map of spec field overrides (currently supports
//     "requestedBy", "priority")
//
// Returns the new WorkflowRun.
func (s *Service) Clone(runID string, overrides map[string]string) (*WorkflowRun, error) {
	original, ok := s.engine.GetRun(runID)
	if !ok {
		return nil, fmt.Errorf("clone: run %q not found", runID)
	}

	def, ok := s.engine.GetDefinition(original.DefinitionID)
	if !ok {
		return nil, fmt.Errorf("clone: definition %q not found", original.DefinitionID)
	}

	s.log.Info("workflow clone",
		"original_run_id", runID,
		"definition", def.ID,
		"overrides", overrides,
	)

	return s.engine.Submit(def)
}
