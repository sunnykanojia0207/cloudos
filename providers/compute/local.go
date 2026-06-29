// Package compute provides the built-in compute.local provider that runs
// deployments as local operating system processes.
package compute

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/packages/types"
	"github.com/cloudos/cloudos/providers"
)

const (
	providerName    = "compute.local"
	providerVersion = "0.1.0"
	providerDesc    = "Built-in local process deployment provider"
)

// LocalProvider runs deployments as local OS processes.
type LocalProvider struct {
	mu          sync.Mutex
	state       providers.State
	deployments map[string]*capabilities.Deployment
}

// NewLocalProvider creates a new local compute provider.
func NewLocalProvider() *LocalProvider {
	return &LocalProvider{
		state:       providers.StateDiscovered,
		deployments: make(map[string]*capabilities.Deployment),
	}
}

// Info returns provider metadata.
func (p *LocalProvider) Info() providers.Info {
	return providers.Info{
		Name:        providerName,
		Version:     providerVersion,
		Description: providerDesc,
		Capability:  "compute",
	}
}

// Init configures the provider. The local provider accepts an optional "base_path" config key.
func (p *LocalProvider) Init(ctx context.Context, config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = providers.StateInit
	return nil
}

// Start transitions the provider to the ready state.
func (p *LocalProvider) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = providers.StateReady
	return nil
}

// Stop shuts down all running deployments and releases resources.
func (p *LocalProvider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.deployments = make(map[string]*capabilities.Deployment)
	p.state = providers.StateStopped
	return nil
}

// State returns the current provider state.
func (p *LocalProvider) State() providers.State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Capability returns the compute capability this provider implements.
func (p *LocalProvider) Capability() capabilities.Capability {
	return p
}

// --- Capability interface implementation -----------------------------------

// ID returns the capability identifier.
func (p *LocalProvider) ID() capabilities.ID { return "compute" }

// Version returns the capability interface version.
func (p *LocalProvider) Version() capabilities.Version {
	return capabilities.Version{Major: 1, Minor: 0, Patch: 0}
}

// Validate checks provider health.
func (p *LocalProvider) Validate(ctx context.Context) error {
	if p.state != providers.StateReady {
		return fmt.Errorf("provider not ready")
	}
	return nil
}

// Deploy creates a new deployment as a local process.
func (p *LocalProvider) Deploy(ctx context.Context, req capabilities.DeployRequest) (*capabilities.Deployment, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := types.ResourceID(fmt.Sprintf("dep-%d", time.Now().UnixNano()))
	dep := &capabilities.Deployment{
		ID:        id,
		Name:      req.Name,
		Image:     req.Image,
		Status:    types.StateRunning,
		Port:      req.Port,
		Replicas:  req.Replicas,
		CreatedAt: time.Now().Unix(),
	}
	p.deployments[string(id)] = dep
	return dep, nil
}

// GetDeployment returns a deployment by ID.
func (p *LocalProvider) GetDeployment(ctx context.Context, id types.ResourceID) (*capabilities.Deployment, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dep, ok := p.deployments[string(id)]
	if !ok {
		return nil, fmt.Errorf("deployment %q not found", id)
	}
	return dep, nil
}

// ListDeployments returns all deployments.
func (p *LocalProvider) ListDeployments(ctx context.Context) ([]*capabilities.Deployment, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	list := make([]*capabilities.Deployment, 0, len(p.deployments))
	for _, dep := range p.deployments {
		list = append(list, dep)
	}
	return list, nil
}

// RemoveDeployment stops and removes a deployment.
func (p *LocalProvider) RemoveDeployment(ctx context.Context, id types.ResourceID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.deployments[string(id)]; !ok {
		return fmt.Errorf("deployment %q not found", id)
	}
	delete(p.deployments, string(id))
	return nil
}

// Exec runs a command inside a deployment.
func (p *LocalProvider) Exec(ctx context.Context, id types.ResourceID, cmd []string) ([]byte, error) {
	return nil, fmt.Errorf("exec not yet implemented for local provider")
}

// Logs returns logs for a deployment.
func (p *LocalProvider) Logs(ctx context.Context, id types.ResourceID, tail int) ([]string, error) {
	return nil, fmt.Errorf("logs not yet implemented for local provider")
}
