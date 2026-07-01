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
//   - Prepare / Start / Stop / Restart / Destroy lifecycle
//   - Health checking with configurable policy
//   - Log capture and streaming
//   - Metrics collection
//
// Compatibility Promise (see ADR-0008):
//   - The Runtime interface is the only supported execution contract
//   - Workflows never call Docker, SSH, Kubernetes, or processes directly
//   - Controllers never manage processes directly
//   - Buildpacks never start applications
//   - Runtimes never modify source code
package runtime

import (
	"context"
	"io"
	"time"
)

// ── API Version ─────────────────────────────────────────────────────────────
//
// RuntimeAPIVersion is the frozen version of the Runtime interface.
// Per ADR-0011, this contract is declared v1.0 and will only receive
// additive extensions via optional interfaces.
//
// Every Runtime implementation declares which API version it implements.
// The certification harness validates that declaration matches reality.
//
//	const APIVersion = runtime.RuntimeAPIVersion
const RuntimeAPIVersion = "runtime.cloudos.io/v1"

// ── Runtime Type ────────────────────────────────────────────────────────────

// RuntimeType categorizes Runtime implementations.
type RuntimeType string

const (
	RuntimeTypeLocal      RuntimeType = "local"
	RuntimeTypeDocker     RuntimeType = "docker"
	RuntimeTypeSSH        RuntimeType = "ssh"
	RuntimeTypeKubernetes RuntimeType = "kubernetes"
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
//
// Lifecycle:
//
//	Prepare(ctx, req) → PreparedApplication  (validate, allocate resources)
//	Start(ctx, app)    → RunningInstance      (launch the process)
//	Stop(ctx, id)      → error               (graceful shutdown)
//	Restart(ctx, id)   → error               (stop + start)
//	Destroy(ctx, id)   → error               (stop + release all resources)
type Runtime interface {
	// Identity
	Name() string
	Type() RuntimeType

	// ── Lifecycle ────────────────────────────────────────────────────────

	// Prepare validates the request, allocates resources (ports, directories),
	// and returns a PreparedApplication — an immutable descriptor of what
	// will run. The PreparedApplication is then passed to Start().
	Prepare(ctx context.Context, req *PrepareRequest) (*PreparedApplication, error)

	// Start launches an application instance from a PreparedApplication and
	// returns immediately once the process has been started (not necessarily
	// when it's healthy).
	Start(ctx context.Context, app *PreparedApplication) (*RunningInstance, error)

	// Stop terminates a running instance by ID. It should attempt a
	// graceful shutdown before force-killing. Returns an error if the
	// instance is not found.
	Stop(ctx context.Context, instanceID string) error

	// Restart stops and re-starts an instance.
	Restart(ctx context.Context, instanceID string) error

	// Destroy stops the instance and releases all associated resources
	// (ports, directories, network connections). After Destroy, the
	// instance ID is no longer valid.
	Destroy(ctx context.Context, instanceID string) error

	// ── Observability ───────────────────────────────────────────────────

	// Health returns the health status of a specific instance.
	// Returns a HealthReport with the current state, or an error if
	// the instance is not found.
	Health(ctx context.Context, instanceID string) (*HealthReport, error)

	// Logs returns a LogStream for streaming logs from an instance.
	// Use LogOptions to control tail count, follow behavior, and
	// source filtering.
	Logs(ctx context.Context, instanceID string, opts LogOptions) (LogStream, error)

	// Metrics returns performance metrics for a running instance.
	Metrics(ctx context.Context, instanceID string) (*Metrics, error)
}

// ── PrepareRequest ─────────────────────────────────────────────────────────

// PrepareRequest contains all parameters needed to prepare an application
// for execution. This is the input to Runtime.Prepare().
type PrepareRequest struct {
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

	// HealthCheck defines how health checks should be performed.
	// If nil, a default policy is used.
	HealthCheck *HealthPolicy

	// Labels are arbitrary key-value metadata attached to the instance.
	Labels map[string]string

	// Artifact references an optional build artifact. When set, the runtime
	// uses this artifact instead of the WorkDir for execution.
	Artifact *ArtifactRef
}

// ── PreparedApplication ────────────────────────────────────────────────────

// PreparedApplication is the immutable result of Runtime.Prepare().
// It describes everything needed to start the application process.
//
// A PreparedApplication is passed to Runtime.Start(). It should be treated
// as immutable once created.
type PreparedApplication struct {
	// ID is a unique identifier for this prepared application instance.
	ID string

	// AppID is the CloudOS Application ID this instance belongs to.
	AppID string

	// WorkDir is the working directory for the application process.
	WorkDir string

	// Command is the command to execute.
	Command string

	// Args are additional arguments to the command.
	Args []string

	// Port is the allocated port number (0 if not applicable).
	Port int

	// EnvVars are environment variables for the process.
	EnvVars map[string]string

	// Labels are arbitrary key-value metadata.
	Labels map[string]string

	// Artifact references the optional build artifact used for execution.
	Artifact *ArtifactRef
}

// ── ArtifactRef ────────────────────────────────────────────────────────────

// ArtifactRef references a build artifact. This is how Buildpacks pass
// their output to Runtimes without the Runtime needing to know how the
// artifact was produced.
type ArtifactRef struct {
	// Type is the artifact type (e.g. "binary", "static", "container").
	Type string `json:"type"`

	// Path is the filesystem path to the artifact.
	Path string `json:"path"`

	// Hash is a content hash for integrity verification.
	Hash string `json:"hash,omitempty"`
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

// ── Metrics ────────────────────────────────────────────────────────────────

// Metrics contains performance metrics for a running instance.
type Metrics struct {
	// CPUUsage is CPU usage as a percentage (0-100).
	CPUUsage float64 `json:"cpuUsage"`

	// MemoryUsage is memory usage in bytes.
	MemoryUsage int64 `json:"memoryUsage"`

	// Uptime is how long the instance has been running.
	Uptime time.Duration `json:"uptime"`

	// Timestamp is when these metrics were collected.
	Timestamp time.Time `json:"timestamp"`
}

// ── Log Options ────────────────────────────────────────────────────────────

// LogOptions configures how logs are retrieved from a running instance.
type LogOptions struct {
	// Tail is the number of recent log lines to include in the initial
	// response. If <= 0, no historical lines are returned.
	Tail int

	// Follow, when true, streams new log lines as they are produced.
	Follow bool

	// Source filters by log source ("stdout", "stderr", or "" for both).
	Source string
}

// ── LogStream ──────────────────────────────────────────────────────────────

// LogStream provides access to streaming application logs.
// Create one via Runtime.Logs().
type LogStream interface {
	// Lines returns a channel that receives log entries as they become
	// available. The channel is closed when the log source is exhausted
	// or the stream is closed via Close().
	Lines() <-chan LogEntry

	// Close releases any resources held by the stream.
	Close() error
}

// ── Log Reader (Internal) ──────────────────────────────────────────────────

// LogReader provides access to application logs with read, historical,
// and streaming access. It is used internally by runtimes and the
// LogManager. The public API uses LogStream.
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
