// Package plugin defines the interfaces and lifecycle for loading external
// plugins into the CloudOS kernel. Plugins provide additional providers and
// capabilities at runtime.
package plugin

import (
	"context"
	"fmt"
)

// State represents the current state of a loaded plugin.
type State string

const (
	StateDiscovered State = "discovered"
	StateLoaded     State = "loaded"
	StateActivated  State = "activated"
	StateFailed     State = "failed"
	StateUnloaded   State = "unloaded"
)

// Info contains metadata about a plugin.
type Info struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

// Manifest describes a plugin's identity, capabilities, and dependencies.
type Manifest struct {
	Info         Info     `json:"info"`
	Capabilities []string `json:"capabilities"`
	Dependencies []string `json:"dependencies"`
}

// Plugin is the interface every plugin must implement.
type Plugin interface {
	// Manifest returns the plugin's metadata and capabilities.
	Manifest() Manifest

	// Load initialises the plugin. The plugin should register its providers
	// and capabilities during Load.
	Load(ctx context.Context) error

	// Activate transitions the plugin from loaded to active. This is a
	// separate step from Load so that dependency resolution can happen
	// between the two phases.
	Activate(ctx context.Context) error

	// Unload shuts down the plugin and releases resources.
	Unload(ctx context.Context) error
}

// Loader is the interface for plugin loading strategies.
// Different implementations can load plugins from the filesystem (WASM),
// shared libraries (.so/.dll), or remote registries.
type Loader interface {
	// Discover returns manifests for all available plugins.
	Discover(ctx context.Context) ([]Manifest, error)

	// Load loads a specific plugin by name.
	Load(ctx context.Context, name string) (Plugin, error)

	// Lookup returns the manifest for a specific plugin.
	Lookup(ctx context.Context, name string) (Manifest, error)
}

// Registry tracks loaded plugins and their states.
type Registry struct {
	plugins map[string]*entry
}

type entry struct {
	Plugin
	state State
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*entry),
	}
}

// Register adds a plugin to the registry in its current state.
func (r *Registry) Register(p Plugin, state State) error {
	name := p.Manifest().Info.Name
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}
	r.plugins[name] = &entry{Plugin: p, state: state}
	return nil
}

// Get returns a registered plugin by name.
func (r *Registry) Get(name string) (Plugin, bool) {
	e, ok := r.plugins[name]
	if !ok {
		return nil, false
	}
	return e.Plugin, true
}

// State returns the current state of a plugin.
func (r *Registry) State(name string) (State, bool) {
	e, ok := r.plugins[name]
	if !ok {
		return StateUnloaded, false
	}
	return e.state, true
}

// SetState updates the state of a registered plugin.
func (r *Registry) SetState(name string, state State) error {
	e, ok := r.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %q not found", name)
	}
	e.state = state
	return nil
}

// List returns all registered plugin manifests.
func (r *Registry) List() []Manifest {
	manifests := make([]Manifest, 0, len(r.plugins))
	for _, e := range r.plugins {
		manifests = append(manifests, e.Manifest())
	}
	return manifests
}

// Remove removes a plugin from the registry.
func (r *Registry) Remove(name string) {
	delete(r.plugins, name)
}
