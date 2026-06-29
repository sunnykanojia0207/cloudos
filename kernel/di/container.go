// Package di provides a lightweight dependency injection container for the
// CloudOS kernel. The container enables loose coupling between subsystems by
// allowing them to declare and receive their dependencies through a central
// registry rather than hard-coding construction.
//
// Architecture: the container uses a simple string-keyed registry. Dependencies
// are registered as singletons (shared instances) and retrieved by name. This
// avoids the complexity and runtime reflection of general-purpose DI frameworks
// while still providing the decoupling benefit.
package di

import (
	"fmt"
	"sync"

	"github.com/cloudos/cloudos/packages/logging"
)

// Container is a lightweight, thread-safe dependency injection container.
// Dependencies are registered as singletons and retrieved by name at
// initialisation time.
type Container struct {
	mu   sync.RWMutex
	deps map[string]interface{}
	log  *logging.Logger
}

// NewContainer creates a new DI container.
func NewContainer(log *logging.Logger) *Container {
	return &Container{
		deps: make(map[string]interface{}),
		log:  log,
	}
}

// Register adds a dependency to the container. Returns an error if a
// dependency with the same name already exists.
func (c *Container) Register(name string, dep interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.deps[name]; exists {
		return fmt.Errorf("di: dependency %q already registered", name)
	}
	c.deps[name] = dep
	c.log.Debug("registered dependency", "name", name)
	return nil
}

// RegisterOrReplace adds a dependency to the container, silently replacing any
// existing value with the same name.
func (c *Container) RegisterOrReplace(name string, dep interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deps[name] = dep
}

// Get retrieves a dependency by name. Returns nil and false if not found.
func (c *Container) Get(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	dep, ok := c.deps[name]
	return dep, ok
}

// MustGet retrieves a dependency by name, panicking if not found.
func (c *Container) MustGet(name string) interface{} {
	dep, ok := c.Get(name)
	if !ok {
		panic(fmt.Sprintf("di: required dependency %q not found", name))
	}
	return dep
}

// Names returns the names of all registered dependencies.
func (c *Container) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.deps))
	for n := range c.deps {
		names = append(names, n)
	}
	return names
}

// Unregister removes a dependency from the container.
func (c *Container) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.deps, name)
}

// Clear removes all dependencies from the container.
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deps = make(map[string]interface{})
}
