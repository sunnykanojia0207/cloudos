package workflow

import (
	"fmt"
)

// Graph provides DAG operations on a collection of Nodes.
// It is stateless and computed from the nodes' dependency declarations.
type Graph struct {
	nodes []Node
}

// NewGraph creates a Graph from a slice of Nodes.
func NewGraph(nodes []Node) *Graph {
	return &Graph{nodes: nodes}
}

// Nodes returns the underlying node slice.
func (g *Graph) Nodes() []Node { return g.nodes }

// ── Topological Sort ─────────────────────────────────────────────────────

// TopologicalSort returns the nodes in topological order (dependencies first).
// Uses Kahn's algorithm. Returns an error if a cycle is detected.
func (g *Graph) TopologicalSort() ([]Node, error) {
	inDegree := make(map[string]int)
	adj := make(map[string][]string) // adjacency list (forward edges)
	nodeMap := make(map[string]Node)

	for _, n := range g.nodes {
		id := n.ID()
		nodeMap[id] = n
		inDegree[id] = 0
		if adj[id] == nil {
			adj[id] = []string{}
		}
	}

	// Build adjacency list and compute in-degrees
	for _, n := range g.nodes {
		id := n.ID()
		for _, dep := range n.Dependencies() {
			adj[dep] = append(adj[dep], id)
			inDegree[id]++
		}
	}

	// Queue nodes with in-degree 0
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []Node
	visited := make(map[string]bool)

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if visited[id] {
			continue
		}
		visited[id] = true

		if n, ok := nodeMap[id]; ok {
			sorted = append(sorted, n)
		}

		for _, neighbor := range adj[id] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(g.nodes) {
		return nil, fmt.Errorf("graph: cycle detected — %d of %d nodes sorted", len(sorted), len(g.nodes))
	}

	return sorted, nil
}

// ── Cycle Detection ──────────────────────────────────────────────────────

// HasCycle returns true if the graph contains at least one cycle.
func (g *Graph) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}

// ── Dependency Resolution ────────────────────────────────────────────────

// DependenciesOf returns the IDs of all transitive dependencies of the given node.
// Includes direct dependencies.
func (g *Graph) DependenciesOf(nodeID string) []string {
	visited := make(map[string]bool)
	var collect func(id string)
	collect = func(id string) {
		n := g.nodeByID(id)
		if n == nil {
			return
		}
		for _, dep := range n.Dependencies() {
			if !visited[dep] {
				visited[dep] = true
				collect(dep)
			}
		}
	}
	collect(nodeID)

	result := make([]string, 0, len(visited))
	for id := range visited {
		result = append(result, id)
	}
	return result
}

// ── Ready Nodes ──────────────────────────────────────────────────────────

// ReadyNodes returns the nodes whose dependencies are all satisfied
// (i.e., all deps have status NodeSucceeded).
func (g *Graph) ReadyNodes(run *WorkflowRun) []Node {
	var ready []Node
	nodeMap := makeNodeMap(run.Nodes)

	for _, n := range g.nodes {
		// Check the run's node status, not the definition node's status
		rn, ok := nodeMap[n.ID()]
		if !ok {
			continue
		}
		if rn.Status().IsTerminal() {
			continue
		}
		if rn.Status() == NodeRunning {
			continue
		}

		allMet := true
		for _, dep := range n.Dependencies() {
			if dn, ok := nodeMap[dep]; ok {
				if dn.Status() != NodeSucceeded {
					allMet = false
					break
				}
			}
		}
		if allMet {
			ready = append(ready, n)
		}
	}
	return ready
}

// ── Helpers ──────────────────────────────────────────────────────────────

func (g *Graph) nodeByID(id string) Node {
	for _, n := range g.nodes {
		if n.ID() == id {
			return n
		}
	}
	return nil
}

func makeNodeMap(nodes []Node) map[string]Node {
	m := make(map[string]Node, len(nodes))
	for _, n := range nodes {
		m[n.ID()] = n
	}
	return m
}

// ── Builder Helpers ──────────────────────────────────────────────────────

// AddEdge adds a dependency from `from` to `to` in the node slice.
// Returns an error if either node is not found.
func AddEdge(nodes []Node, from, to string) error {
	var fromFound, toFound bool
	for _, n := range nodes {
		if n.ID() == from {
			fromFound = true
		}
		if n.ID() == to {
			toFound = true
		}
	}
	if !fromFound {
		return fmt.Errorf("edge: source node %q not found", from)
	}
	if !toFound {
		return fmt.Errorf("edge: target node %q not found", to)
	}

	// Add "to" as a dependency of "from" (from must complete before to starts)
	for _, n := range nodes {
		if n.ID() == to {
			// We need to modify dependencies — use type assertion for known types
			switch node := n.(type) {
			case *TaskNode:
				node.deps = append(node.deps, from)
			case *EndNode:
				node.deps = append(node.deps, from)
			}
		}
	}
	return nil
}
