// Package health provides a health-check manager that periodically probes
// registered subsystems and aggregates their status into a single report.
package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/types"
)

// Report contains the health status of a single component at a point in time.
type Report struct {
	State     types.ResourceState `json:"state"`
	Message   string              `json:"message,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
}

// Checkable is the interface components implement to expose health checks.
type Checkable interface {
	CheckHealth(ctx context.Context) Report
}

// Manager aggregates health checks from registered components.
type Manager struct {
	mu         sync.RWMutex
	checkables map[string]Checkable
	results    map[string]Report
	log        *logging.Logger
	running    bool
	cancel     context.CancelFunc
}

// NewManager creates a new health manager.
func NewManager(log *logging.Logger) *Manager {
	return &Manager{
		checkables: make(map[string]Checkable),
		results:    make(map[string]Report),
		log:        log,
	}
}

// Register adds a component to health monitoring.
func (m *Manager) Register(name string, c Checkable) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.checkables[name]; exists {
		return fmt.Errorf("health: %q already registered", name)
	}
	m.checkables[name] = c
	m.log.Debug("health check registered", "component", name)
	return nil
}

// Unregister removes a component from health monitoring.
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checkables, name)
	delete(m.results, name)
}

// Start begins periodic health checking. If the manager is already running
// this is a no-op.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.mu.Unlock()

	// Run an immediate check, then every 30 seconds.
	m.runChecks(ctx)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.runChecks(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	m.log.Info("health manager started")
	return nil
}

// Stop disables periodic health checking.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
	m.log.Info("health manager stopped")
	return nil
}

// runChecks iterates over all registered checkables and stores their reports.
func (m *Manager) runChecks(ctx context.Context) {
	m.mu.RLock()
	names := make([]string, 0, len(m.checkables))
	for n := range m.checkables {
		names = append(names, n)
	}
	m.mu.RUnlock()

	for _, name := range names {
		m.mu.RLock()
		checkable := m.checkables[name]
		m.mu.RUnlock()

		report := checkable.CheckHealth(ctx)

		m.mu.Lock()
		m.results[name] = report
		m.mu.Unlock()
	}
}

// Report returns the latest health report for a single component.
func (m *Manager) Report(name string) (Report, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.results[name]
	return r, ok
}

// All returns the latest health report for every registered component.
func (m *Manager) All() map[string]Report {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make(map[string]Report, len(m.results))
	for n, r := range m.results {
		out[n] = r
	}
	return out
}

// Overall returns a single aggregated health status. It returns healthy only
// when all components report healthy.
func (m *Manager) Overall() Report {
	all := m.All()

	overall := Report{
		State:     types.StateRunning,
		Message:   "all systems operational",
		Timestamp: time.Now(),
	}

	for name, r := range all {
		if r.State == types.StateFailed || r.State == types.StateDegraded {
			overall.State = types.StateDegraded
			overall.Message = fmt.Sprintf("%s is %s", name, r.State)
			break
		}
	}

	return overall
}

// Registered returns the names of all registered health checkables.
func (m *Manager) Registered() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.checkables))
	for n := range m.checkables {
		names = append(names, n)
	}
	return names
}
