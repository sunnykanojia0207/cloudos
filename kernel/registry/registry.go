// Package registry provides a generic, thread-safe registry for storing and
// retrieving named items. It is used by the kernel for both capabilities and
// providers.
package registry

import (
	"fmt"
	"sync"

	"github.com/cloudos/cloudos/packages/logging"
)

// Item is any value that can be stored in the registry.
type Item interface {
	// Name returns the unique name of this item within its registry.
	Name() string
}

// Manager is a typed, thread-safe registry.
type Manager struct {
	mu   sync.RWMutex
	kind string        // e.g. "capability", "provider"
	items map[string]Item
	log  *logging.Logger
}

// NewManager creates a new registry for the given kind of items.
func NewManager(kind string, log *logging.Logger) *Manager {
	return &Manager{
		kind:  kind,
		items: make(map[string]Item),
		log:   log,
	}
}

// Register adds an item to the registry. Returns an error if an item with the
// same name already exists.
func (m *Manager) Register(item Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := item.Name()
	if _, exists := m.items[name]; exists {
		return fmt.Errorf("registry: %s %q already registered", m.kind, name)
	}
	m.items[name] = item
	m.log.Debug("registered", "kind", m.kind, "name", name)
	return nil
}

// Unregister removes an item from the registry.
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, name)
}

// Get retrieves an item by name. Returns nil and false if not found.
func (m *Manager) Get(name string) (Item, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.items[name]
	return item, ok
}

// MustGet retrieves an item by name, panicking if not found.
func (m *Manager) MustGet(name string) Item {
	item, ok := m.Get(name)
	if !ok {
		panic(fmt.Sprintf("registry: %s %q not found", m.kind, name))
	}
	return item
}

// List returns all registered items.
func (m *Manager) List() []Item {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Item, 0, len(m.items))
	for _, item := range m.items {
		out = append(out, item)
	}
	return out
}

// Names returns the names of all registered items.
func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]string, 0, len(m.items))
	for n := range m.items {
		out = append(out, n)
	}
	return out
}

// Len returns the number of registered items.
func (m *Manager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}
