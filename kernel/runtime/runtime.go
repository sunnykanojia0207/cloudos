// Package runtime defines the Runtime abstraction for CloudOS.
//
// A Runtime knows HOW to execute user applications. It is the interface
// between the workflow engine and the execution environment — local
// processes, Docker containers, Kubernetes pods, SSH targets, or
// Firecracker microVMs.
//
// The Runtime interface is provider-agnostic. Any Runtime implementation
// can be used with any Application. The choice of Runtime is a deployment
// concern, not an application concern.
//
// Architecture:
//
//	Application Controller
//	      │
//	      ▼
//	  Workflow Engine
//	      │
//	      ▼
//	  Executor (provider.deploy)
//	      │
//	      ▼
//	  Runtime interface
//	      ├── LocalRuntime   (local processes)
//	      ├── DockerRuntime  (Docker containers)
//	      ├── K8sRuntime     (Kubernetes)
//	      └── ...
//
// Every Runtime implementation provides:
//   - Start / Stop / Restart lifecycle
//   - Health checking with configurable policy
//   - Log capture via LogManager
//   - Process listing and inspection
package runtime

import (
	"context"
	"io"
	"time"
)

// ── Runtime Status ─────────────────────────────────────────────────────────

// RuntimeStatus represents the current state of a managed application instance.
type RuntimeStatus string

const (
	StatusPending  RuntimeStatus = "pending"  // Not yet started
	StatusStarting RuntimeStatus = "starting" // Startup in progress
	StatusRunning  RuntimeStatus = "running"  // Healthy and accepting traffic
	StatusStopping RuntimeStatus = "stopping" // Graceful shutdown in progress
	StatusStopped  RuntimeStatus = "stopped"  // Stopped (clean exit)
	StatusFailed   RuntimeStatus = "failed"   // Unhealthy or crashed
	StatusDeleted  RuntimeStatus = "deleted"  // Removed / cleaned up
)

// ── Runtime Interface ──────────────────────────────────────────────────────

// Runtime is the interface for executing and managing application workloads.
//
// Implementations must be safe for concurrent use.
type Runtime interface {
	// Name returns the runtime name (e.g. "local", "docker", "k8s").
	Name() string

	// Start launches an application instance and returns immediately once
	// the process has been started (not necessarily when it's healthy).
	Start(ctx context.Context, config StartConfig) (*RunningInstance, error)

	// Stop terminates a running instance by ID. It should attempt a
	// graceful shutdown before force-killing. Returns an error if the
	// instance is not found.
	Stop(ctx context.Context, id string) error

	// Restart stops and re-starts an instance.
	Restart(ctx context.Context, id string) error

	// List returns all managed instances.
	List(ctx context.Context) ([]RunningInstance, error)

	// Get returns a single instance by ID, or nil if not found.
	Get(ctx context.Context, id string) (*RunningInstance, error)

	// Health returns the health status of a specific instance.
	// Returns a HealthReport with the current state, or an error if
	// the instance is not found.
	Health(ctx context.Context, id string) (*HealthReport, error)

	// Logs returns a LogReader for streaming logs from an instance.
	// The reader should follow the io.Reader pattern, returning new
	// log data as it becomes available.
	Logs(ctx context.Context, id string) (LogReader, error)
}

// ── StartConfig ────────────────────────────────────────────────────────────

// StartConfig contains all parameters needed to start an application instance.
type StartConfig struct {
	// AppID is the CloudOS Application identifier.
	AppID string

	// Name is a human-readable name for this instance.
	Name string

	// WorkDir is the working directory containing the built application.
	WorkDir string

	// Command is the command to execute (e.g. "npm start", "./app").
	Command string

	// Args are additional arguments passed to the command.
	Args []string

	// Port is the port the application should listen on.
	// If 0, the runtime allocates one.
	Port int

	// EnvVars are environment variables for the application process.
	EnvVars map[string]string

	// HealthPolicy defines how health checks should be performed.
	// If nil, a default policy is used.
	HealthPolicy *HealthPolicy
}

// ── RunningInstance ────────────────────────────────────────────────────────

// RunningInstance represents a running or managed application instance.
type RunningInstance struct {
	// ID is a unique identifier for this instance.
	ID string `json:"id"`

	// AppID is the CloudOS Application ID this instance belongs to.
	AppID string `json:"appId"`

	// Name is a human-readable name.
	Name string `json:"name"`

	// Status is the current runtime status.
	Status RuntimeStatus `json:"status"`

	// Port is the port the application is listening on (0 if unknown).
	Port int `json:"port,omitempty"`

	// URL is the access URL (e.g. "http://localhost:9001").
	URL string `json:"url,omitempty"`

	// PID is the OS process ID (0 for non-process runtimes).
	PID int `json:"pid,omitempty"`

	// ContainerID is the container ID (for container runtimes).
	ContainerID string `json:"containerId,omitempty"`

	// StartTime is when the instance was started.
	StartTime time.Time `json:"startTime"`

	// RestartCount is the number of times this instance has been restarted.
	RestartCount int `json:"restartCount"`

	// HealthStatus is the result of the last health check.
	HealthStatus string `json:"healthStatus,omitempty"`

	// Labels are arbitrary key-value metadata.
	Labels map[string]string `json:"labels,omitempty"`
}

// ── Health Report ──────────────────────────────────────────────────────────

// HealthReport describes the health of a running instance.
type HealthReport struct {
	// Status is the overall health status.
	Status RuntimeStatus `json:"status"`

	// Message provides a human-readable description.
	Message string `json:"message,omitempty"`

	// LastChecked is when the health was last verified.
	LastChecked time.Time `json:"lastChecked"`

	// ResponseTime is how long the last health check took.
	ResponseTime time.Duration `json:"responseTime,omitempty"`

	// StatusCode is the HTTP status code from the last check (0 if N/A).
	StatusCode int `json:"statusCode,omitempty"`
}

// ── Log Reader ─────────────────────────────────────────────────────────────

// LogReader provides access to application logs.
// It combines read, historical, and streaming access.
type LogReader interface {
	io.Reader

	// ReadLines reads up to n log lines, oldest first.
	// If n <= 0, returns all available lines.
	ReadLines(n int) ([]string, error)

	// Follow returns a channel that receives new log lines as they
	// are written. The channel is closed when the log source is
	// closed or the context is cancelled.
	Follow(ctx context.Context) <-chan string

	// Close releases any resources held by the reader.
	Close() error
}

// ── Health Policy ──────────────────────────────────────────────────────────

// HealthPolicy configures how health checks are performed for an application.
//
// The policy follows Kubernetes-style probe semantics:
//   - Wait InitialDelay before first check
//   - Check every Interval
//   - Each check has Timeout to complete
//   - Consider healthy after SuccessThreshold consecutive successes
//   - Consider unhealthy after FailureThreshold consecutive failures
type HealthPolicy struct {
	// InitialDelay is how long to wait before the first health check.
	InitialDelay time.Duration `json:"initialDelay"`

	// Interval is how often to perform health checks.
	Interval time.Duration `json:"interval"`

	// Timeout is the maximum time for a single health check request.
	Timeout time.Duration `json:"timeout"`

	// SuccessThreshold is the number of consecutive successes required
	// to consider the application healthy.
	SuccessThreshold int `json:"successThreshold"`

	// FailureThreshold is the number of consecutive failures required
	// to consider the application unhealthy.
	FailureThreshold int `json:"failureThreshold"`
}

// DefaultHealthPolicy returns a sensible default health check policy.
func DefaultHealthPolicy() *HealthPolicy {
	return &HealthPolicy{
		InitialDelay:     5 * time.Second,
		Interval:         5 * time.Second,
		Timeout:          3 * time.Second,
		SuccessThreshold: 2,
		FailureThreshold: 3,
	}
}
