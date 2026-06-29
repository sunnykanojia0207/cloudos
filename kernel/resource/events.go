package resource

// Resource lifecycle event types published through the kernel's event bus.
// Consumers (watchers, auditors, AI agents, dashboards) subscribe to these
// to react to resource state changes.
const (
	EventResourceCreated   = "resource.created"
	EventResourceUpdated   = "resource.updated"
	EventResourceDeleted   = "resource.deleted"
	EventResourceValidated = "resource.validated"
)
