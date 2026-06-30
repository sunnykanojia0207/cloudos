// Package kernel implements the CloudOS operating system kernel. The kernel is
// the central orchestrator: it manages the lifecycle of every subsystem,
// coordinates the event bus, capability registry, provider registry, and
// exposes health and diagnostics.
//
// Architectural rule: the kernel must never import provider packages directly.
// Providers are discovered and loaded through the provider registry.
package kernel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/kernel/di"
	"github.com/cloudos/cloudos/providers"
	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/lifecycle"
	"github.com/cloudos/cloudos/kernel/project"
	"github.com/cloudos/cloudos/kernel/runtime/local"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/kernel/registry"
	"github.com/cloudos/cloudos/kernel/resource"
	"github.com/cloudos/cloudos/kernel/scheduler"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/kernel/security"
	"github.com/cloudos/cloudos/packages/config"
	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/types"
)

// Kernel is the central orchestrator. It owns every subsystem and coordinates
// their startup, runtime communication, and shutdown.
type Kernel struct {
	mu       sync.RWMutex
	cfg      config.Config
	log      *logging.Logger
	state    types.ResourceState

	// Subsystems.
	lifecycle           *lifecycle.Manager
	events              *events.Bus
	scheduler           *scheduler.Scheduler
	health              *health.Manager
	security            *security.Manager
	capRegistry         *registry.Manager
	provRegistry        *registry.Manager
	capDescRegistry     *capabilities.Registry
	provDescRegistry    *providers.Registry
	resRegistry         *resource.Registry
	ctrlManager         *controller.Manager
	container           *di.Container

	// Lifecycle.
	startedAt time.Time
	cancel    context.CancelFunc
}

// Config carries the subset of system configuration the kernel needs.
type Config struct {
	LogLevel  string
	DataDir   string
	PluginDir string
}

// New creates a new Kernel with the given configuration. It initialises all
// subsystems but does not start them. Call Kernel.Boot() to start.
func New(cfg config.Config) (*Kernel, error) {
	logLevel := logging.ParseLevel(cfg.Kernel.LogLevel)
	log := logging.NewSubsystemLogger("kernel", logLevel)

	_, cancel := context.WithCancel(context.Background())

	k := &Kernel{
		cfg:          cfg,
		log:          log,
		state:        types.StatePending,
		startedAt:    time.Time{},
		cancel:       cancel,
		lifecycle:       lifecycle.NewManager(log),
		events:          events.NewBus(log),
		scheduler:       scheduler.New(log),
		health:          health.NewManager(log),
		security:        security.NewManager(),
		capRegistry:      registry.NewManager("capability", log),
		provRegistry:     registry.NewManager("provider", log),
		capDescRegistry:  capabilities.NewRegistry(),
		provDescRegistry: providers.NewRegistry(),
		container:        di.NewContainer(log),
	}

	// Resource registry needs the event bus which is set above.
	k.resRegistry = resource.NewRegistry(k.events, log)

	// Controller runtime needs the resource registry and event bus.
	k.ctrlManager = controller.NewManager(k.resRegistry, k.events, k.health, log)

	// Register the kernel itself for health checking.
	k.health.Register("kernel", k)

	return k, nil
}

// Boot starts all kernel subsystems in dependency order and transitions the
// kernel to the running state.
func (k *Kernel) Boot(ctx context.Context) error {
	k.mu.Lock()
	if k.state == types.StateRunning {
		k.mu.Unlock()
		return fmt.Errorf("kernel is already running")
	}
	k.state = types.StateRunning
	k.startedAt = time.Now()
	k.mu.Unlock()

	k.log.Info("booting kernel",
		"dataDir", k.cfg.Kernel.DataDir,
		"pluginDir", k.cfg.Kernel.PluginDir,
	)

	// Boot sequence: subsystems are started in dependency order.
	bootOrder := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"lifecycle.manager", func(ctx context.Context) error {
			return nil // lifecycle starts empty; components register themselves
		}},
		{"events.bus", func(ctx context.Context) error {
			k.events.Start()
			// Publish a boot-started event for any listener.
			k.events.Publish(ctx, events.Event{
				Type:   "kernel.boot.started",
				Source: "kernel",
			})
			return nil
		}},
		{"scheduler", func(ctx context.Context) error {
			k.scheduler.Start(ctx)
			return nil
		}},
		{"health.manager", func(ctx context.Context) error {
			return k.health.Start(ctx)
		}},
		{"capability.descriptors", func(ctx context.Context) error {
			for _, d := range capabilities.DefaultDescriptors() {
				if err := k.capDescRegistry.Register(d); err != nil {
					return fmt.Errorf("register capability descriptor %s: %w", d.ID, err)
				}
				k.log.Info("capability descriptor registered",
					"id", d.ID,
					"version", d.Version.String(),
					"category", string(d.Category),
				)
			}
			return nil
		}},
		{"provider.descriptors", func(ctx context.Context) error {
			for _, d := range providers.DefaultDescriptors() {
				if err := k.provDescRegistry.Register(d); err != nil {
					return fmt.Errorf("register provider descriptor %s: %w", d.ID, err)
				}
				k.log.Info("provider descriptor registered",
					"id", d.ID,
					"version", d.Version,
				)
			}
			return nil
		}},
		{"resource.engine", func(ctx context.Context) error {
			// Register the built-in "Namespace" resource kind.
			if err := k.resRegistry.RegisterKind(resource.Kind{
				Name:       "Namespace",
				Namespaced: false,
				Versions:   []string{"v1"},
			}); err != nil {
				return fmt.Errorf("register resource kind namespace: %w", err)
			}

			// Create the default namespace.
			ns := resource.DefaultNamespace()
			if err := k.resRegistry.Create(ctx, ns); err != nil {
				return fmt.Errorf("create default namespace: %w", err)
			}

			// Register the built-in "Project" resource kind.
			if err := k.resRegistry.RegisterKind(resource.Kind{
				Name:       project.Kind,
				Namespaced: true,
				Versions:   []string{"v1"},
			}); err != nil {
				return fmt.Errorf("register resource kind %s: %w", project.Kind, err)
			}
			k.log.Info("project resource kind registered")

			// Register the "WorkflowExecution" resource kind.
			if err := k.resRegistry.RegisterKind(resource.Kind{
				Name:       workflow.WorkflowExecutionKind,
				Namespaced: true,
				Versions:   []string{"v1"},
			}); err != nil {
				return fmt.Errorf("register resource kind %s: %w", workflow.WorkflowExecutionKind, err)
			}
			k.log.Info("workflow execution resource kind registered")

			// Register the "Application" resource kind.
			if err := k.resRegistry.RegisterKind(resource.Kind{
				Name:       application.Kind,
				Namespaced: true,
				Versions:   []string{"v1"},
			}); err != nil {
				return fmt.Errorf("register resource kind %s: %w", application.Kind, err)
			}
			k.log.Info("application resource kind registered")

			k.log.Info("resource engine initialised",
				"namespace", ns.GetMetadata().ID,
			)
			return nil
		}},
		{"controller.runtime", func(ctx context.Context) error {
			// Register the built-in NamespaceController.
			nsCtrl := controller.NewNamespaceController(k.resRegistry, k.events, k.log)
			if err := k.ctrlManager.Register(nsCtrl); err != nil {
				return fmt.Errorf("register namespace controller: %w", err)
			}
			k.log.Info("namespace controller registered",
				"controller", nsCtrl.Name(),
				"kind", nsCtrl.Kind(),
			)

			// Register the ProjectController.
			projCtrl := project.NewProjectController(k.resRegistry, k.events, k.log)
			if err := k.ctrlManager.Register(projCtrl); err != nil {
				return fmt.Errorf("register project controller: %w", err)
			}
			k.log.Info("project controller registered",
				"controller", projCtrl.Name(),
				"kind", projCtrl.Kind(),
			)

			// Create the Source GitCloner and Local Runtime Manager
			// for the deployment workflow pipeline.
			sourceCloner := source.NewGitCloner(k.cfg.Kernel.DataDir)
			runtimeMgr := local.NewManager(k.cfg.Kernel.DataDir, k.log)
			k.log.Info("source cloner and runtime manager created",
				"workDir", k.cfg.Kernel.DataDir,
			)

			// Create the Workflow Service (needed by ApplicationController).
			workflowSvc := workflow.NewService(workflow.ServiceDeps{
				ResourceRegistry:  k.resRegistry,
				ControllerManager: k.ctrlManager,
				HealthManager:     k.health,
				EventBus:          k.events,
				SourceCloner:      sourceCloner,
				RuntimeManager:    runtimeMgr,
				Logger:            k.log,
			})
			k.log.Info("workflow service created")

			// Register the ApplicationController.
			appCtrl := application.NewApplicationController(k.resRegistry, k.events, workflowSvc, k.log)
			if err := k.ctrlManager.Register(appCtrl); err != nil {
				return fmt.Errorf("register application controller: %w", err)
			}
			k.log.Info("application controller registered",
				"controller", appCtrl.Name(),
				"kind", appCtrl.Kind(),
			)

			// Start the controller runtime (starts all controllers).
			if err := k.ctrlManager.Start(ctx); err != nil {
				return fmt.Errorf("start controller runtime: %w", err)
			}

			k.log.Info("controller runtime started")
			return nil
		}},
	}

	for _, step := range bootOrder {
		k.log.Debug("booting subsystem", "subsystem", step.name)
		if err := step.fn(ctx); err != nil {
			k.log.Error("subsystem boot failed", "subsystem", step.name, "error", err)
			k.state = types.StateFailed
			return fmt.Errorf("boot %s: %w", step.name, err)
		}
		k.log.Info("subsystem booted", "subsystem", step.name)
	}

	// Publish boot-complete event.
	k.events.Publish(ctx, events.Event{
		Type:   "kernel.boot.complete",
		Source: "kernel",
	})

	k.log.Info("kernel booted",
		"uptime", time.Since(k.startedAt).String(),
	)
	return nil
}

// Shutdown gracefully shuts down all subsystems in reverse dependency order.
func (k *Kernel) Shutdown(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.state != types.StateRunning {
		return nil
	}

	k.log.Info("shutting down kernel")
	k.cancel()

	k.events.Publish(ctx, events.Event{
		Type:   "kernel.shutdown.started",
		Source: "kernel",
	})

	// Shutdown in reverse order.
	if err := k.ctrlManager.Stop(ctx); err != nil {
		k.log.Error("controller runtime shutdown error", "error", err)
	}
	if err := k.health.Stop(ctx); err != nil {
		k.log.Error("health manager shutdown error", "error", err)
	}
	k.scheduler.Stop()
	k.events.Stop()
	k.lifecycle.StopAll()

	k.state = types.StateStopped
	k.log.Info("kernel shut down")
	return nil
}

// State returns the current kernel state.
func (k *Kernel) State() types.ResourceState {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.state
}

// Uptime returns the duration the kernel has been running.
func (k *Kernel) Uptime() time.Duration {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if k.state != types.StateRunning {
		return 0
	}
	return time.Since(k.startedAt)
}

// --- Accessors ---------------------------------------------------------------

// Events returns the event bus for publishing and subscribing to system events.
func (k *Kernel) Events() *events.Bus { return k.events }

// Scheduler returns the task scheduler.
func (k *Kernel) Scheduler() *scheduler.Scheduler { return k.scheduler }

// Health returns the health manager.
func (k *Kernel) Health() *health.Manager { return k.health }

// Security returns the security manager.
func (k *Kernel) Security() *security.Manager { return k.security }

// CapabilityRegistry returns the capability interface registry.
func (k *Kernel) CapabilityRegistry() *registry.Manager { return k.capRegistry }

// CapabilityDescriptorRegistry returns the capability metadata descriptor
// registry. This is the source of truth for the capability discovery API.
func (k *Kernel) CapabilityDescriptorRegistry() *capabilities.Registry { return k.capDescRegistry }

// ProviderRegistry returns the provider interface registry.
func (k *Kernel) ProviderRegistry() *registry.Manager { return k.provRegistry }

// ProviderDescriptorRegistry returns the provider metadata descriptor
// registry. This is the source of truth for the provider discovery API.
func (k *Kernel) ProviderDescriptorRegistry() *providers.Registry { return k.provDescRegistry }

// ResourceRegistry returns the resource engine registry.
func (k *Kernel) ResourceRegistry() *resource.Registry { return k.resRegistry }

// ControllerManager returns the controller runtime manager.
func (k *Kernel) ControllerManager() *controller.Manager { return k.ctrlManager }

// Container returns the dependency injection container.
func (k *Kernel) Container() *di.Container { return k.container }

// Logger returns the kernel's logger.
func (k *Kernel) Logger() *logging.Logger { return k.log }

// CheckHealth implements health.Checkable so the kernel can report its own health.
func (k *Kernel) CheckHealth(ctx context.Context) health.Report {
	return health.Report{
		State:     k.State(),
		Message:   "kernel operational",
		Timestamp: time.Now(),
	}
}
