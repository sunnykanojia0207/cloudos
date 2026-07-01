// Package oci implements the CloudOS Runtime interface for OCI-compatible
// container engines (Docker, Podman, containerd, nerdctl).
//
// Architecture:
//
//	Workflow → Runtime.Prepare → Runtime.Start → Runtime.Stop → Runtime.Destroy
//	                         ↓
//	                   OCI Runtime
//	                         ↓
//	              ContainerEngine interface
//	                         ├── DockerAdapter
//	                         ├── PodmanAdapter  (future)
//	                         └── ContainerdAdapter (future)
//
// The ContainerEngine abstraction ensures CloudOS is not coupled to any
// single container runtime. Every implementation of the Runtime interface
// through OCI is interchangeable with Local Runtime — no Workflow changes,
// no Controller changes, no Buildpack changes, no certification test changes.
package oci

import (
	"context"
	"time"
)

// ── Container Engine Interface ─────────────────────────────────────────────

// ContainerConfig describes how to create and run a container.
type ContainerConfig struct {
	// Image is the OCI image to use (e.g. "alpine:latest", "node:20-slim").
	Image string

	// Command is the command to run inside the container.
	Command string

	// Args are additional arguments to the command.
	Args []string

	// Env is a map of environment variables (name → value).
	Env map[string]string

	// Ports maps host ports to container ports (host:container).
	Ports map[int]int

	// Volumes maps host paths to container paths (host:container).
	Volumes map[string]string

	// WorkDir is the working directory inside the container.
	WorkDir string

	// Name is an optional container name.
	Name string

	// Labels are metadata labels applied to the container.
	Labels map[string]string

	// NetworkMode sets the container's network mode (e.g. "bridge", "host").
	// Defaults to "bridge" if empty.
	NetworkMode string

	// AutoRemove, when true, removes the container automatically when it stops.
	AutoRemove bool
}

// ContainerState represents the current state of a container.
type ContainerState string

const (
	ContainerCreated    ContainerState = "created"
	ContainerRunning    ContainerState = "running"
	ContainerPaused     ContainerState = "paused"
	ContainerRestarting ContainerState = "restarting"
	ContainerExited     ContainerState = "exited"
	ContainerRemoving   ContainerState = "removing"
	ContainerDead       ContainerState = "dead"
	ContainerUnknown    ContainerState = "unknown"
)

// ContainerInfo contains metadata about a container.
type ContainerInfo struct {
	// ID is the container identifier.
	ID string

	// Name is the container name.
	Name string

	// Image is the image the container was created from.
	Image string

	// State is the current container state.
	State ContainerState

	// Status is a human-readable status string.
	Status string

	// Ports maps host ports to container ports.
	Ports map[int]int

	// CreatedAt is when the container was created.
	CreatedAt time.Time

	// StartedAt is when the container was last started.
	StartedAt time.Time

	// ExitCode is the exit code (only valid when Exited).
	ExitCode int
}

// ContainerStats contains resource usage statistics for a container.
type ContainerStats struct {
	// CPUPercent is CPU usage as a percentage.
	CPUPercent float64

	// MemoryUsage is memory usage in bytes.
	MemoryUsage uint64

	// MemoryLimit is the memory limit in bytes (0 = unlimited).
	MemoryLimit uint64

	// NetworkRx is received network bytes.
	NetworkRx uint64

	// NetworkTx is transmitted network bytes.
	NetworkTx uint64

	// BlockRead is block I/O read bytes.
	BlockRead uint64

	// BlockWrite is block I/O write bytes.
	BlockWrite uint64

	// PIDs is the number of processes in the container.
	PIDs uint64

	// Timestamp is when these stats were collected.
	Timestamp time.Time
}

// ContainerEngine is the abstraction over OCI-compatible container runtimes.
//
// Implementations: DockerEngine, PodmanEngine, NerdctlEngine, etc.
type ContainerEngine interface {
	// Name returns the engine name ("docker", "podman", etc.).
	Name() string

	// Available checks if the container engine is installed and functional.
	Available(ctx context.Context) error

	// Pull pulls an OCI image from a registry.
	Pull(ctx context.Context, image string) error

	// Run creates and starts a container. Returns the container ID.
	Run(ctx context.Context, config *ContainerConfig) (string, error)

	// Stop stops a running container gracefully (SIGTERM, then SIGKILL after timeout).
	Stop(ctx context.Context, containerID string, timeout *time.Duration) error

	// Remove removes a container (must be stopped first unless force=true).
	Remove(ctx context.Context, containerID string, force bool) error

	// Inspect returns detailed information about a container.
	Inspect(ctx context.Context, containerID string) (*ContainerInfo, error)

	// Logs streams logs from a container.
	Logs(ctx context.Context, containerID string, follow bool, tail int) ([]byte, error)

	// LogStream returns a channel of log lines for streaming.
	LogStream(ctx context.Context, containerID string, follow bool, tail int) (<-chan string, <-chan error, error)

	// Stats returns resource usage statistics for a container.
	Stats(ctx context.Context, containerID string) (*ContainerStats, error)

	// List returns all containers matching the optional label filter.
	List(ctx context.Context, labelFilter map[string]string) ([]ContainerInfo, error)
}

// ── Base Image Mapping ──────────────────────────────────────────────────────

// ArtifactToImage maps CloudOS artifact types to OCI base images.
// These are minimal images that provide the runtime environment needed
// for each artifact type.
var ArtifactToImage = map[string]string{
	"binary": "alpine:latest",
	"static": "nginx:alpine",
	"source": "alpine:latest",
	"image":  "", // already a container image — use directly
}

// LanguageToImage maps detected languages to appropriate base images.
var LanguageToImage = map[string]string{
	"go":      "alpine:latest",
	"node":    "node:20-alpine",
	"python":  "python:3.12-alpine",
	"php":     "php:8.2-cli-alpine",
	"react":   "nginx:alpine",
	"nextjs":  "node:20-alpine",
	"laravel": "php:8.2-cli-alpine",
	"static":  "nginx:alpine",
	"generic": "alpine:latest",
}
