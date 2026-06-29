// Package providers defines the base Provider interface that every CloudOS
// provider must implement. Providers are the concrete implementations of
// capability interfaces; they are loaded, configured, and orchestrated by the
// kernel through the provider registry.
//
// Architectural rules:
//   - Providers implement capabilities (from capabilities/), they do not define them.
//   - Providers must not import the kernel package.
//   - Providers communicate with the kernel through capability interfaces and events.
package providers

import (
	"context"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/packages/types"
)

// State represents the lifecycle state of a provider.
type State = types.ResourceState

const (
	StateDiscovered State = "discovered"
	StateInit       State = "initialising"
	StateReady      State = "ready"
	StateFailed     State = "failed"
	StateStopped    State = "stopped"
)

// Info carries identifying metadata for a provider.
type Info struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Capability  capabilities.ID   `json:"capability"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Provider is the base interface every provider must implement.
type Provider interface {
	// Info returns the provider's identifying metadata.
	Info() Info

	// Init initialises the provider with its configuration. This is called
	// after the provider is loaded but before it is started.
	Init(ctx context.Context, config map[string]interface{}) error

	// Start transitions the provider to the ready state. After Start returns
	// without error the provider is expected to be operational.
	Start(ctx context.Context) error

	// Stop shuts down the provider and releases any held resources.
	Stop(ctx context.Context) error

	// State returns the provider's current lifecycle state.
	State() State

	// Capability returns the capability this provider implements.
	Capability() capabilities.Capability
}
