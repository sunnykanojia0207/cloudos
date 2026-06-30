package workflow

import (
	"context"

	"github.com/cloudos/cloudos/kernel/events"
)

// Event type constants for workflow lifecycle events.
const (
	EventWorkflowSubmitted  = "workflow.submitted"
	EventWorkflowStarted    = "workflow.started"
	EventWorkflowPaused     = "workflow.paused"
	EventWorkflowResumed    = "workflow.resumed"
	EventWorkflowCancelled  = "workflow.cancelled"
	EventWorkflowCompleted  = "workflow.completed"
	EventWorkflowFailed     = "workflow.failed"
	EventNodeStarted        = "workflow.node.started"
	EventNodeSucceeded      = "workflow.node.succeeded"
	EventNodeFailed         = "workflow.node.failed"
	EventNodeRetrying       = "workflow.node.retrying"
	EventNodeSkipped        = "workflow.node.skipped"
)

// EventPublisher publishes workflow events to the kernel event bus.
type EventPublisher struct {
	bus *events.Bus
}

// NewEventPublisher creates a new EventPublisher.
func NewEventPublisher(bus *events.Bus) *EventPublisher {
	return &EventPublisher{bus: bus}
}

// Publish sends a workflow event with the given type and payload.
func (ep *EventPublisher) Publish(eventType string, data map[string]interface{}) {
	if ep == nil || ep.bus == nil {
		return
	}
	ep.bus.Publish(context.Background(), events.Event{
		Type:    eventType,
		Source:  "workflow.engine",
		Payload: data,
	})
}

// PublishWorkflowEvent sends a workflow-scoped event.
func (ep *EventPublisher) PublishWorkflowEvent(eventType string, run *WorkflowRun) {
	ep.Publish(eventType, map[string]interface{}{
		"workflow_id":  run.ID,
		"definition_id": run.DefinitionID,
		"status":       string(run.Status),
		"progress":     run.Progress(),
		"completed":    run.CompletedCount(),
		"total":        len(run.Nodes),
	})
}

// PublishNodeEvent sends a node-scoped event.
func (ep *EventPublisher) PublishNodeEvent(eventType string, run *WorkflowRun, node Node) {
	ep.Publish(eventType, map[string]interface{}{
		"workflow_id":  run.ID,
		"definition_id": run.DefinitionID,
		"node_id":      node.ID(),
		"node_name":    node.Name(),
		"node_type":    string(node.Type()),
		"status":       string(node.Status()),
	})
}
