package workflow

// Scheduler determines which nodes are eligible to run next.
// It operates on a WorkflowRun and uses the Graph to check dependencies.
//
// The Scheduler is stateless — it computes readiness purely from the
// current state of nodes in the run. This makes it safe to call from
// multiple goroutines without synchronization (the run's node statuses
// should be protected by the Engine).
type Scheduler struct{}

// NewScheduler creates a new Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// Ready returns the IDs of nodes that are eligible to run:
//   - Status is NodePending (not yet started, not terminal)
//   - All dependencies have status NodeSucceeded
//
// Nodes are returned in dependency order (topologically sorted).
func (s *Scheduler) Ready(run *WorkflowRun) []Node {
	g := NewGraph(run.Nodes)
	ready := g.ReadyNodes(run)

	// Sort ready nodes topologically for deterministic ordering
	sorted, err := g.TopologicalSort()
	if err != nil {
		// If there's a cycle, return ready nodes in original order
		return ready
	}

	// Filter sorted to only include ready nodes, preserving topological order
	var ordered []Node
	readySet := make(map[string]bool)
	for _, n := range ready {
		readySet[n.ID()] = true
	}
	for _, n := range sorted {
		if readySet[n.ID()] {
			ordered = append(ordered, n)
		}
	}
	return ordered
}

// IsComplete returns true if all nodes in the run are in a terminal state.
func (s *Scheduler) IsComplete(run *WorkflowRun) bool {
	for _, n := range run.Nodes {
		if !n.Status().IsTerminal() {
			return false
		}
	}
	return true
}

// HasFailures returns true if any node has failed.
func (s *Scheduler) HasFailures(run *WorkflowRun) bool {
	for _, n := range run.Nodes {
		if n.Status() == NodeFailed || n.Status() == NodeCancelled {
			return true
		}
	}
	return false
}
