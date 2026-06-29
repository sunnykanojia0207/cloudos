// Package lifecycle provides a manager that tracks and controls the lifecycle
// state of kernel subsystems. Each subsystem transitions through a well-defined
// state machine: Pending → Starting → Running → Stopping → Stopped.
package lifecycle

import (
	"fmt"
	"sync"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/types"
)

// State represents a component's lifecycle state.
type State = types.ResourceState

const (
	StatePending State = "pending"
	StateStarting State = "starting"
	StateRunning State = "running"
	StateStopping State = "stopping"
	StateStopped State = "stopped"
	StateFailed State = "failed"
)

// Component is any subsystem whose lifecycle is managed by the Manager.
type Component interface {
	// Name returns a stable identifier for the component.
	Name() string

	// Start initialises the component. It must block until the component is
	// fully running or return an error.
	Start() error

	// Stop shuts down the component. It must block until shutdown is complete.
	Stop() error
}

// Manager tracks the lifecycle state of registered components.
type Manager struct {
	mu         sync.RWMutex
	components map[string]*entry
	log        *logging.Logger
}

type entry struct {
	Component
	state State
}

// NewManager creates a new lifecycle manager.
func NewManager(log *logging.Logger) *Manager {
	return &Manager{
		components: make(map[string]*entry),
		log:        log,
	}
}

// Register adds a component to lifecycle management.
func (m *Manager) Register(c Component) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := c.Name()
	if _, exists := m.components[name]; exists {
		return fmt.Errorf("lifecycle: component %q already registered", name)
	}
	m.components[name] = &entry{Component: c, state: StatePending}
	m.log.Debug("component registered", "component", name)
	return nil
}

// Unregister removes a component from lifecycle management.
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.components, name)
}

// Start transitions the named component to Running. It returns an error if
// the component is already running or fails to start.
func (m *Manager) Start(name string) error {
	m.mu.Lock()
	e, ok := m.components[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("lifecycle: component %q not found", name)
	}
	if e.state == StateRunning {
		m.mu.Unlock()
		return fmt.Errorf("lifecycle: component %q already running", name)
	}
	e.state = StateStarting
	m.mu.Unlock()

	m.log.Info("starting component", "component", name)
	if err := e.Component.Start(); err != nil {
		m.mu.Lock()
		e.state = StateFailed
		m.mu.Unlock()
		return fmt.Errorf("lifecycle: start %q: %w", name, err)
	}

	m.mu.Lock()
	e.state = StateRunning
	m.mu.Unlock()
	m.log.Info("component started", "component", name)
	return nil
}

// Stop transitions the named component to Stopped.
func (m *Manager) Stop(name string) error {
	m.mu.Lock()
	e, ok := m.components[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("lifecycle: component %q not found", name)
	}
	if e.state == StateStopped {
		m.mu.Unlock()
		return nil
	}
	e.state = StateStopping
	m.mu.Unlock()

	m.log.Info("stopping component", "component", name)
	if err := e.Component.Stop(); err != nil {
		m.mu.Lock()
		e.state = StateFailed
		m.mu.Unlock()
		return fmt.Errorf("lifecycle: stop %q: %w", name, err)
	}

	m.mu.Lock()
	e.state = StateStopped
	m.mu.Unlock()
	m.log.Info("component stopped", "component", name)
	return nil
}

// State returns the current lifecycle state of a component.
func (m *Manager) State(name string) (State, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.components[name]
	if !ok {
		return StatePending, false
	}
	return e.state, true
}

// StartAll starts all registered components. Components are started in
// registration order.
func (m *Manager) StartAll() error {
	m.mu.RLock()
	names := make([]string, 0, len(m.components))
	for n := range m.components {
		names = append(names, n)
	}
	m.mu.RUnlock()

	for _, n := range names {
		if err := m.Start(n); err != nil {
			return err
		}
	}
	return nil
}

// StopAll stops all registered components in reverse registration order.
func (m *Manager) StopAll() {
	m.mu.RLock()
	names := make([]string, 0, len(m.components))
	for n := range m.components {
		names = append(names, n)
	}
	m.mu.RUnlock()

	for i := len(names) - 1; i >= 0; i-- {
		m.Stop(names[i]) //nolint:errcheck
	}
}

// Snapshot returns the state of every registered component.
func (m *Manager) Snapshot() map[string]State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap := make(map[string]State, len(m.components))
	for name, e := range m.components {
		snap[name] = e.state
	}
	return snap
}
