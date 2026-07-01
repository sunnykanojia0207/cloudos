package oci

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── OCIRuntime ──────────────────────────────────────────────────────────────

// OCIRuntime implements the CloudOS Runtime interface using OCI-compatible
// container engines (Docker, Podman, etc.). It is a drop-in replacement for
// LocalRuntime — no Workflow, Controller, Buildpack, or certification test
// changes needed.
//
// Architecture:
//
//	Runtime.Prepare → determine image + ports + mounts
//	Runtime.Start   → ContainerEngine.Run (docker run)
//	Runtime.Stop    → ContainerEngine.Stop (docker stop)
//	Runtime.Destroy → ContainerEngine.Remove (docker rm -f)
//	Runtime.Health  → ContainerEngine.Inspect
//	Runtime.Logs    → ContainerEngine.Logs / LogStream
//	Runtime.Metrics → ContainerEngine.Stats
type OCIRuntime struct {
	mu       sync.Mutex
	log      *logging.Logger
	engine   ContainerEngine
	instances map[string]*ociInstance

	// Port allocation tracking.
	portMin  int
	portMax  int
	usedPorts map[int]bool

	// baseWorkDir is the host directory for artifact viewing.
	baseWorkDir string

	// imagePullCache prevents repeated pulls of the same image.
	imagePullCache map[string]bool
}

// ociInstance tracks a managed OCI container instance.
type ociInstance struct {
	id          string
	appID       string
	containerID string
	port        int
	workDir     string
	command     string
	image       string
	createdAt   time.Time
}

// NewOCIRuntime creates a new OCI Runtime backed by the given ContainerEngine.
func NewOCIRuntime(engine ContainerEngine, log *logging.Logger) *OCIRuntime {
	return &OCIRuntime{
		log:            logging.NewSubsystemLogger("oci-runtime", logging.LevelInfo),
		engine:         engine,
		instances:      make(map[string]*ociInstance),
		portMin:        9000,
		portMax:        9999,
		usedPorts:      make(map[int]bool),
		baseWorkDir:    filepath.Join(os.TempDir(), "cloudos-oci"),
		imagePullCache: make(map[string]bool),
	}
}

// ── Runtime Interface ───────────────────────────────────────────────────────

func (r *OCIRuntime) Name() string {
	return fmt.Sprintf("oci-%s", r.engine.Name())
}

func (r *OCIRuntime) Type() cr.RuntimeType {
	return cr.RuntimeTypeDocker // OCI is the docker runtime type
}

// Prepare validates the request, pulls the base image, allocates a port,
// and returns a PreparedApplication.
func (r *OCIRuntime) Prepare(ctx context.Context, req *cr.PrepareRequest) (*cr.PreparedApplication, error) {
	r.log.Debug("preparing OCI deployment",
		"app", req.AppID,
		"command", req.Command,
		"workdir", req.WorkDir,
	)

	// Determine the base image.
	image := r.resolveImage(req)

	// Pull the image if not already cached.
	if err := r.ensureImage(ctx, image); err != nil {
		return nil, fmt.Errorf("ensure image %q: %w", image, err)
	}

	// Allocate a port (if not explicitly provided).
	port := req.Port
	if port == 0 {
		var err error
		port, err = r.allocatePort()
		if err != nil {
			return nil, fmt.Errorf("allocate port: %w", err)
		}
	}
	r.log.Debug("port allocated", "app", req.AppID, "port", port)

	// Create a unique instance ID.
	instanceID := fmt.Sprintf("oci-%s-%d", req.AppID, time.Now().UnixNano())

	// Register the instance.
	r.mu.Lock()
	r.instances[instanceID] = &ociInstance{
		id:        instanceID,
		appID:     req.AppID,
		port:      port,
		workDir:   req.WorkDir,
		command:   req.Command,
		image:     image,
		createdAt: time.Now(),
	}
	r.mu.Unlock()

	// Build the command.
	command := req.Command
	args := req.Args
	if args == nil {
		args = []string{}
	}

	// If the command is a path (e.g., "./app" or "/app/server"), ensure
	// it will work inside the container. For binary artifacts, the binary
	// is at the workdir path inside the container.
	if command != "" && !strings.Contains(command, " ") {
		// Absolute path in container or relative to workdir.
		if !strings.HasPrefix(command, "/") {
			// If the workdir has the binary, prepend the workdir.
			command = filepath.Join(req.WorkDir, command)
		}
	}

	return &cr.PreparedApplication{
		ID:      instanceID,
		AppID:   req.AppID,
		WorkDir: req.WorkDir,
		Command: command,
		Args:    args,
		Port:    port,
		EnvVars: req.EnvVars,
		Labels:  req.Labels,
		Artifact: req.Artifact,
	}, nil
}

// Start launches the application in a container via the ContainerEngine.
func (r *OCIRuntime) Start(ctx context.Context, app *cr.PreparedApplication) (*cr.RunningInstance, error) {
	r.mu.Lock()
	inst, exists := r.instances[app.ID]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("instance %q not found (was Prepare called?)", app.ID)
	}

	// Determine the container workdir path.
	// Inside the container, the application lives at /app.
	containerWorkDir := "/app"

	// Build the container config.
	config := &ContainerConfig{
		Image:   inst.image,
		Command: app.Command,
		Args:    app.Args,
		WorkDir: containerWorkDir,
		Env:     app.EnvVars,
		Ports:   map[int]int{app.Port: app.Port},
		Volumes: map[string]string{inst.workDir: containerWorkDir},
		Name:    fmt.Sprintf("cloudos-%s", inst.appID),
		Labels: map[string]string{
			"cloudos.app":    inst.appID,
			"cloudos.runtime": "oci",
			"cloudos.instance": inst.id,
		},
		NetworkMode: "bridge",
		AutoRemove:  false,
	}

	// Run the container.
	r.log.Info("starting container",
		"app", inst.appID,
		"image", inst.image,
		"port", app.Port,
		"command", app.Command,
		"workdir", inst.workDir,
	)

	containerID, err := r.engine.Run(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("container run: %w", err)
	}

	// Store the container ID.
	r.mu.Lock()
	inst.containerID = containerID
	r.mu.Unlock()

	r.log.Info("container started",
		"app", inst.appID,
		"container", containerID,
		"port", app.Port,
	)

	// Build the running instance response.
	runInst := &cr.RunningInstance{
		ID:        inst.id,
		AppID:     inst.appID,
		Name:      fmt.Sprintf("oci-%s", inst.appID),
		Status:    cr.StatusRunning,
		Port:      app.Port,
		URL:       fmt.Sprintf("http://localhost:%d", app.Port),
		StartTime: time.Now(),
		Labels:    app.Labels,
	}

	return runInst, nil
}

// Stop stops a running container gracefully.
func (r *OCIRuntime) Stop(ctx context.Context, instanceID string) error {
	r.mu.Lock()
	inst, exists := r.instances[instanceID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("instance %q not found", instanceID)
	}

	if inst.containerID == "" {
		return fmt.Errorf("instance %q has no container ID (was Start called?)", instanceID)
	}

	r.log.Info("stopping container", "app", inst.appID, "container", inst.containerID)
	return r.engine.Stop(ctx, inst.containerID, durationPtr(10*time.Second))
}

// Restart stops and re-starts a container.
func (r *OCIRuntime) Restart(ctx context.Context, instanceID string) error {
	if err := r.Stop(ctx, instanceID); err != nil {
		return err
	}

	// Re-start: we need the PreparedApplication to re-create the container.
	// For now, this is a limitation — restart requires the app to be prepared again.
	// In a full implementation, we'd cache the PreparedApplication or re-create
	// from stored state.
	r.mu.Lock()
	inst, exists := r.instances[instanceID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("instance %q not found", instanceID)
	}

	// Remove the old container.
	if inst.containerID != "" {
		if err := r.engine.Remove(ctx, inst.containerID, true); err != nil {
			r.log.Warn("remove old container on restart", "error", err.Error())
		}
	}

	// The caller must call Start again with a new PreparedApplication.
	return fmt.Errorf("restart requires re-Prepare and re-Start — call Prepare then Start")
}

// Destroy stops the container and releases all resources.
func (r *OCIRuntime) Destroy(ctx context.Context, instanceID string) error {
	r.mu.Lock()
	inst, exists := r.instances[instanceID]
	if exists {
		delete(r.instances, instanceID)
		// Release the port.
		if inst.port > 0 {
			delete(r.usedPorts, inst.port)
		}
	}
	r.mu.Unlock()

	if !exists {
		return nil // already cleaned up
	}

	if inst.containerID != "" {
		r.log.Info("destroying container", "app", inst.appID, "container", inst.containerID)
		if err := r.engine.Stop(ctx, inst.containerID, durationPtr(5*time.Second)); err != nil {
			r.log.Warn("stop container on destroy", "error", err.Error())
		}
		if err := r.engine.Remove(ctx, inst.containerID, true); err != nil {
			r.log.Warn("remove container on destroy", "error", err.Error())
		}
	}

	return nil
}

// Health returns the health status of a container.
func (r *OCIRuntime) Health(ctx context.Context, instanceID string) (*cr.HealthReport, error) {
	r.mu.Lock()
	inst, exists := r.instances[instanceID]
	r.mu.Unlock()

	if !exists {
		return &cr.HealthReport{
			Status:    cr.StatusDeleted,
			Message:   "instance not found",
			LastChecked: time.Now(),
		}, nil
	}

	if inst.containerID == "" {
		return &cr.HealthReport{
			Status:    cr.StatusPending,
			Message:   "container not yet started",
			LastChecked: time.Now(),
		}, nil
	}

	info, err := r.engine.Inspect(ctx, inst.containerID)
	if err != nil {
		return &cr.HealthReport{
			Status:    cr.StatusFailed,
			Message:   fmt.Sprintf("inspect error: %v", err),
			LastChecked: time.Now(),
		}, nil
	}

	status := cr.StatusRunning
	healthMsg := "running"

	switch info.State {
	case ContainerCreated:
		status = cr.StatusPending
		healthMsg = "created"
	case ContainerRunning:
		status = cr.StatusRunning
		healthMsg = "running"
	case ContainerPaused:
		status = cr.StatusStopped
		healthMsg = "paused"
	case ContainerRestarting:
		status = cr.StatusStarting
		healthMsg = "restarting"
	case ContainerExited:
		status = cr.StatusStopped
		healthMsg = fmt.Sprintf("exited (code: %d)", info.ExitCode)
		if info.ExitCode != 0 {
			status = cr.StatusFailed
		}
	case ContainerDead:
		status = cr.StatusFailed
		healthMsg = "dead"
	default:
		status = cr.StatusPending
		healthMsg = "unknown"
	}

	return &cr.HealthReport{
		Status:    status,
		Message:   healthMsg,
		LastChecked: time.Now(),
	}, nil
}

// Logs returns a LogStream for streaming container logs.
func (r *OCIRuntime) Logs(ctx context.Context, instanceID string, opts cr.LogOptions) (cr.LogStream, error) {
	r.mu.Lock()
	inst, exists := r.instances[instanceID]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("instance %q not found", instanceID)
	}

	if inst.containerID == "" {
		return nil, fmt.Errorf("instance %q has no container ID", instanceID)
	}

	lines, errs, err := r.engine.LogStream(ctx, inst.containerID, opts.Follow, opts.Tail)
	if err != nil {
		return nil, err
	}

	return &ociLogStream{
		lines: lines,
		errs:  errs,
	}, nil
}

// Metrics returns container resource usage statistics.
func (r *OCIRuntime) Metrics(ctx context.Context, instanceID string) (*cr.Metrics, error) {
	r.mu.Lock()
	inst, exists := r.instances[instanceID]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("instance %q not found", instanceID)
	}

	if inst.containerID == "" {
		return nil, fmt.Errorf("instance %q has no container ID", instanceID)
	}

	stats, err := r.engine.Stats(ctx, inst.containerID)
	if err != nil {
		return nil, err
	}

	return &cr.Metrics{
		CPUUsage:    stats.CPUPercent,
		MemoryUsage: int64(stats.MemoryUsage),
		Uptime:      time.Since(inst.createdAt),
		Timestamp:   time.Now(),
	}, nil
}

// StopAll stops all managed containers. Used for test cleanup.
func (r *OCIRuntime) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, inst := range r.instances {
		if inst.containerID != "" {
			r.log.Info("stopping all containers", "instance", id, "container", inst.containerID)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := r.engine.Stop(ctx, inst.containerID, durationPtr(5*time.Second)); err != nil {
				r.log.Warn("stop container on StopAll", "error", err.Error())
			}
			if err := r.engine.Remove(ctx, inst.containerID, true); err != nil {
				r.log.Warn("remove container on StopAll", "error", err.Error())
			}
			cancel()
		}
		delete(r.instances, id)
	}

	// Clear used ports.
	r.usedPorts = make(map[int]bool)
}

// WithLogManager is a no-op for OCI Runtime (logs are managed by Docker).
func (r *OCIRuntime) WithLogManager(lm *cr.LogManager) {
	// OCI Runtime uses docker logs, not the local LogManager.
}

// ── Internal ────────────────────────────────────────────────────────────────

// resolveImage determines the base OCI image from the PrepareRequest.
func (r *OCIRuntime) resolveImage(req *cr.PrepareRequest) string {
	// If an artifact is provided, use its type to determine the image.
	if req.Artifact != nil {
		if img, ok := ArtifactToImage[req.Artifact.Type]; ok && img != "" {
			return img
		}
	}

	// Fallback to generic Alpine image.
	return "alpine:latest"
}

// ensureImage pulls an image if it hasn't been pulled before.
func (r *OCIRuntime) ensureImage(ctx context.Context, image string) error {
	r.mu.Lock()
	cached := r.imagePullCache[image]
	r.mu.Unlock()

	if cached {
		return nil
	}

	r.log.Info("pulling image", "image", image)
	if err := r.engine.Pull(ctx, image); err != nil {
		return err
	}

	r.mu.Lock()
	r.imagePullCache[image] = true
	r.mu.Unlock()

	return nil
}

// allocatePort finds and reserves a free port in the configured range.
func (r *OCIRuntime) allocatePort() (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Try up to 100 random ports in the range.
	for i := 0; i < 100; i++ {
		port := r.portMin + rand.Intn(r.portMax-r.portMin+1)
		if r.usedPorts[port] {
			continue
		}

		// Verify the port is actually free by trying to listen on it.
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		ln.Close()

		r.usedPorts[port] = true
		return port, nil
	}

	return 0, fmt.Errorf("no free ports in range %d-%d", r.portMin, r.portMax)
}

// durationPtr returns a pointer to a time.Duration.
func durationPtr(d time.Duration) *time.Duration {
	return &d
}

// ── LogStream ───────────────────────────────────────────────────────────────

// ociLogStream implements cr.LogStream for container logs.
type ociLogStream struct {
	lines <-chan string
	errs  <-chan error
}

func (s *ociLogStream) Lines() <-chan cr.LogEntry {
	entryCh := make(chan cr.LogEntry)
	go func() {
		defer close(entryCh)
		for line := range s.lines {
			entryCh <- cr.LogEntry{
				Line:      line,
				Source:    "stdout",
				Timestamp: time.Now(),
			}
		}
	}()
	return entryCh
}

func (s *ociLogStream) Close() error {
	return nil
}
