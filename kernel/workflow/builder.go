package workflow

import (
	"fmt"
)

// ── ExecutionPlan to WorkflowDefinition ──────────────────────────────────

// PlanNode is a simplified node description used to build a workflow.
// It mirrors the structure of intent.ExecutionStep but is workflow-agnostic.
type PlanNode struct {
	ID       string
	Name     string
	Action   string
	Target   string
	DependsOn []string // IDs of nodes that must complete first
}

// BuildDefinition creates a WorkflowDefinition from an ordered list of PlanNodes.
//
// The builder automatically:
//   - Assigns sequential IDs (step-1, step-2, …)
//   - Sets up dependency edges for serial execution
//   - Adds an EndNode as the terminal node
//   - Validates the resulting DAG
func BuildDefinition(id, name string, plan []PlanNode) (*WorkflowDefinition, error) {
	if len(plan) == 0 {
		return nil, fmt.Errorf("build definition: plan is empty")
	}

	var nodes []Node
	lastID := ""

	for i, pn := range plan {
		nodeID := pn.ID
		if nodeID == "" {
			nodeID = fmt.Sprintf("step-%d", i+1)
		}

		deps := make([]string, 0, len(pn.DependsOn))

		// If no explicit dependencies and not the first node, depend on previous
		if len(pn.DependsOn) == 0 && lastID != "" {
			deps = append(deps, lastID)
		} else {
			deps = append(deps, pn.DependsOn...)
		}

		node := NewTaskNode(nodeID, pn.Name, pn.Action, pn.Target, deps...)

		// Apply retry policy to certain actions
		switch pn.Action {
		case "resource.create", "resource.delete", "validate":
			node.RetryPolicy = DefaultRetryPolicy()
		}

		nodes = append(nodes, node)
		lastID = nodeID
	}

	// Add EndNode depending on the last task node
	lastTaskID := lastID
	if lastTaskID == "" {
		lastTaskID = "step-1"
	}
	nodes = append(nodes, NewEndNode("end", lastTaskID))

	def := &WorkflowDefinition{
		ID:        id,
		Name:      name,
		Nodes:     nodes,
		CreatedAt: NowUTC(),
	}

	if err := ValidateDefinition(def); err != nil {
		return nil, fmt.Errorf("build definition: %w", err)
	}

	return def, nil
}

// ── Common Plan Templates ───────────────────────────────────────────────

// CreateProjectPlan returns a PlanNode list for creating a project.
func CreateProjectPlan(projectID string) []PlanNode {
	return []PlanNode{
		{ID: "1", Name: "Validate Request", Action: "validate", Target: "Project:" + projectID},
		{ID: "2", Name: "Create Project Resource", Action: "resource.create", Target: "Project:" + projectID, DependsOn: []string{"1"}},
		{ID: "3", Name: "Wait for Reconciliation", Action: "controller.reconcile", Target: "Project:" + projectID, DependsOn: []string{"2"}},
		{ID: "4", Name: "Verify Status", Action: "resource.get", Target: "Project:" + projectID, DependsOn: []string{"3"}},
		{ID: "5", Name: "Complete", Action: "complete", Target: projectID, DependsOn: []string{"4"}},
	}
}

// ListProjectsPlan returns a PlanNode list for listing projects.
func ListProjectsPlan() []PlanNode {
	return []PlanNode{
		{ID: "1", Name: "List Projects", Action: "resource.list", Target: "Project"},
		{ID: "2", Name: "Format Results", Action: "format", Target: "list", DependsOn: []string{"1"}},
	}
}

// DeleteProjectPlan returns a PlanNode list for deleting a project.
func DeleteProjectPlan(projectID string) []PlanNode {
	return []PlanNode{
		{ID: "1", Name: "Delete Project", Action: "resource.delete", Target: "Project:" + projectID},
		{ID: "2", Name: "Verify Deletion", Action: "resource.get", Target: "Project:" + projectID, DependsOn: []string{"1"}},
		{ID: "3", Name: "Complete", Action: "complete", Target: projectID, DependsOn: []string{"2"}},
	}
}

// ShowControllersPlan returns a PlanNode list for listing controllers.
func ShowControllersPlan() []PlanNode {
	return []PlanNode{
		{ID: "1", Name: "List Controllers", Action: "controller.list", Target: ""},
		{ID: "2", Name: "Format Results", Action: "format", Target: "list", DependsOn: []string{"1"}},
	}
}

// ShowResourcesPlan returns a PlanNode list for listing resource kinds.
func ShowResourcesPlan() []PlanNode {
	return []PlanNode{
		{ID: "1", Name: "List Resource Kinds", Action: "resource.kinds", Target: ""},
		{ID: "2", Name: "Format Results", Action: "format", Target: "list", DependsOn: []string{"1"}},
	}
}

// ShowHealthPlan returns a PlanNode list for checking system health.
func ShowHealthPlan() []PlanNode {
	return []PlanNode{
		{ID: "1", Name: "Check Health", Action: "health.check", Target: ""},
		{ID: "2", Name: "Format Results", Action: "format", Target: "health", DependsOn: []string{"1"}},
	}
}
