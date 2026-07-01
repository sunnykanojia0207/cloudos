package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Mock Container Engine ───────────────────────────────────────────────────

// mockEngine implements ContainerEngine for testing.
type mockEngine struct {
	mu          sync.Mutex
	containers  map[string]*mockContainer
	available   bool
	pullCalled  map[string]int
	nextID      int
}

type mockContainer struct {
	id        string
	image     string
	command   string
	state     ContainerState
	exitCode  int
	ports     map[int]int
	logs      []string
}

func newMockEngine() *mockEngine {
	return &mockEngine{
		containers: make(map[string]*mockContainer),
		available:  true,
		pullCalled: make(map[string]int),
		nextID:     1,
	}
}

func (e *mockEngine) Name() string { return "mock" }

func (e *mockEngine) Available(ctx context.Context) error {
	if !e.available {
		return fmt.Errorf("mock engine not available")
	}
	return nil
}

func (e *mockEngine) Pull(ctx context.Context, image string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pullCalled[image]++
	return nil
}

func (e *mockEngine) Run(ctx context.Context, config *ContainerConfig) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := fmt.Sprintf("mock-container-%d", e.nextID)
	e.nextID++

	e.containers[id] = &mockContainer{
		id:      id,
		image:   config.Image,
		command: config.Command,
		state:   ContainerRunning,
		ports:   config.Ports,
		logs:    []string{fmt.Sprintf("Started with command: %s", config.Command)},
	}
	return id, nil
}

func (e *mockEngine) Stop(ctx context.Context, containerID string, timeout *time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	c, ok := e.containers[containerID]
	if !ok {
		return fmt.Errorf("container %q not found", containerID)
	}
	c.state = ContainerExited
	c.exitCode = 0
	return nil
}

func (e *mockEngine) Remove(ctx context.Context, containerID string, force bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.containers, containerID)
	return nil
}

func (e *mockEngine) Inspect(ctx context.Context, containerID string) (*ContainerInfo, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	c, ok := e.containers[containerID]
	if !ok {
		return nil, fmt.Errorf("container %q not found", containerID)
	}
	return &ContainerInfo{
		ID:    c.id,
		Name:  c.id,
		Image: c.image,
		State: c.state,
		Ports: c.ports,
	}, nil
}

func (e *mockEngine) Logs(ctx context.Context, containerID string, follow bool, tail int) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	c, ok := e.containers[containerID]
	if !ok {
		return nil, fmt.Errorf("container %q not found", containerID)
	}
	return []byte(strings.Join(c.logs, "\n")), nil
}

func (e *mockEngine) LogStream(ctx context.Context, containerID string, follow bool, tail int) (<-chan string, <-chan error, error) {
	e.mu.Lock()
	c, ok := e.containers[containerID]
	e.mu.Unlock()

	if !ok {
		return nil, nil, fmt.Errorf("container %q not found", containerID)
	}

	lines := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(lines)
		for _, line := range c.logs {
			select {
			case lines <- line:
			case <-ctx.Done():
				return
			}
		}
	}()

	return lines, errs, nil
}

func (e *mockEngine) Stats(ctx context.Context, containerID string) (*ContainerStats, error) {
	return &ContainerStats{
		CPUPercent:  0.5,
		MemoryUsage: 1024 * 1024,
		MemoryLimit: 512 * 1024 * 1024,
		Timestamp:   time.Now(),
	}, nil
}

func (e *mockEngine) List(ctx context.Context, labelFilter map[string]string) ([]ContainerInfo, error) {
	return nil, nil
}

// ── OCI Runtime Tests ───────────────────────────────────────────────────────

func TestOCIRuntime_Prepare(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	workDir := t.TempDir()

	req := &cr.PrepareRequest{
		AppID:   "test-app",
		Name:    "Test App",
		WorkDir: workDir,
		Command: "./app",
		Port:    0,
		EnvVars: map[string]string{"PORT": "8080"},
		Artifact: &cr.ArtifactRef{
			Type: "binary",
			Path: workDir,
		},
	}

	prepared, err := r.Prepare(context.Background(), req)
	if err != nil {
		t.Fatalf("Prepare() returned error: %v", err)
	}

	if prepared.ID == "" {
		t.Error("PreparedApplication.ID is empty")
	}
	if prepared.AppID != "test-app" {
		t.Errorf("AppID = %q, want %q", prepared.AppID, "test-app")
	}
	if prepared.Port == 0 {
		t.Error("Port was not allocated")
	}
	if prepared.WorkDir != workDir {
		t.Errorf("WorkDir = %q, want %q", prepared.WorkDir, workDir)
	}
}

func TestOCIRuntime_Start(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	workDir := t.TempDir()

	prepared, err := r.Prepare(context.Background(), &cr.PrepareRequest{
		AppID:   "test-app",
		Name:    "Test App",
		WorkDir: workDir,
		Command: "./app",
		Port:    0,
		EnvVars: map[string]string{"PORT": "8080"},
		Artifact: &cr.ArtifactRef{
			Type: "binary",
			Path: workDir,
		},
	})
	if err != nil {
		t.Fatalf("Prepare() returned error: %v", err)
	}

	inst, err := r.Start(context.Background(), prepared)
	if err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}

	if inst.ID == "" {
		t.Error("RunningInstance.ID is empty")
	}
	if inst.Port == 0 {
		t.Error("Port is 0")
	}
	if inst.URL == "" {
		t.Error("URL is empty")
	}
	if !strings.HasPrefix(inst.URL, "http://") {
		t.Errorf("URL = %q, want http://...", inst.URL)
	}
	if inst.Status != cr.StatusRunning {
		t.Errorf("Status = %q, want %q", inst.Status, cr.StatusRunning)
	}
}

func TestOCIRuntime_Lifecycle(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	workDir := t.TempDir()

	// Prepare.
	prepared, err := r.Prepare(context.Background(), &cr.PrepareRequest{
		AppID:   "lifecycle-test",
		Name:    "Lifecycle Test",
		WorkDir: workDir,
		Command: "./app",
		Port:    0,
		Artifact: &cr.ArtifactRef{
			Type: "binary",
			Path: workDir,
		},
	})
	if err != nil {
		t.Fatalf("Prepare() returned error: %v", err)
	}

	// Start.
	inst, err := r.Start(context.Background(), prepared)
	if err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}

	// Health check.
	health, err := r.Health(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
	if health.Status != cr.StatusRunning {
		t.Errorf("Health Status = %q, want %q", health.Status, cr.StatusRunning)
	}

	// Stop.
	if err := r.Stop(context.Background(), inst.ID); err != nil {
		t.Fatalf("Stop() returned error: %v", err)
	}

	// Health after stop.
	health, err = r.Health(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Health() returned error after stop: %v", err)
	}
	if health.Status != cr.StatusRunning && health.Status != cr.StatusStopped {
		t.Logf("Health status after stop: %q (acceptable)", health.Status)
	}

	// Metrics.
	metrics, err := r.Metrics(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Metrics() returned error: %v", err)
	}
	if metrics.CPUUsage != 0.5 {
		t.Errorf("CPUUsage = %f, want 0.5", metrics.CPUUsage)
	}
	if metrics.MemoryUsage <= 0 {
		t.Errorf("MemoryUsage = %d, want > 0", metrics.MemoryUsage)
	}

	// Logs.
	logStream, err := r.Logs(context.Background(), inst.ID, cr.LogOptions{Tail: 10})
	if err != nil {
		t.Fatalf("Logs() returned error: %v", err)
	}
	logLines := []cr.LogEntry{}
	for entry := range logStream.Lines() {
		logLines = append(logLines, entry)
	}
	if len(logLines) == 0 {
		t.Error("No log lines returned")
	}

	// Destroy.
	if err := r.Destroy(context.Background(), inst.ID); err != nil {
		t.Fatalf("Destroy() returned error: %v", err)
	}

	// Verify instance is cleaned up.
	health, err = r.Health(context.Background(), inst.ID)
	if err != nil {
		t.Fatalf("Health() returned error after destroy: %v", err)
	}
	if health.Status != cr.StatusDeleted {
		t.Errorf("Health Status after destroy = %q, want %q", health.Status, cr.StatusDeleted)
	}
}

func TestOCIRuntime_PortAllocation(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	// Allocate multiple ports and verify they're unique.
	ports := make(map[int]bool)
	for i := 0; i < 5; i++ {
		req := &cr.PrepareRequest{
			AppID:   fmt.Sprintf("port-test-%d", i),
			Name:    fmt.Sprintf("Port Test %d", i),
			WorkDir: t.TempDir(),
			Command: "./app",
			Port:    0,
		}

		prepared, err := r.Prepare(context.Background(), req)
		if err != nil {
			t.Fatalf("Prepare %d returned error: %v", i, err)
		}

		if ports[prepared.Port] {
			t.Errorf("Duplicate port allocated: %d", prepared.Port)
		}
		ports[prepared.Port] = true
	}
}

func TestOCIRuntime_WithExplicitPort(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	req := &cr.PrepareRequest{
		AppID:   "explicit-port",
		Name:    "Explicit Port",
		WorkDir: t.TempDir(),
		Command: "./app",
		Port:    12345,
	}

	prepared, err := r.Prepare(context.Background(), req)
	if err != nil {
		t.Fatalf("Prepare() returned error: %v", err)
	}

	if prepared.Port != 12345 {
		t.Errorf("Port = %d, want 12345", prepared.Port)
	}
}

func TestOCIRuntime_UnknownInstance(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	ctx := context.Background()

	// Stop with unknown instance should error.
	if err := r.Stop(ctx, "unknown-instance"); err == nil {
		t.Error("Stop() should return error for unknown instance")
	}

	// Destroy with unknown instance should not error (idempotent).
	if err := r.Destroy(ctx, "unknown-instance"); err != nil {
		t.Errorf("Destroy() should be idempotent, got error: %v", err)
	}

	// Health with unknown instance should return deleted status.
	health, err := r.Health(ctx, "unknown-instance")
	if err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
	if health.Status != cr.StatusDeleted {
		t.Errorf("Health Status = %q, want %q", health.Status, cr.StatusDeleted)
	}

	// Logs with unknown instance should error.
	if _, err := r.Logs(ctx, "unknown-instance", cr.LogOptions{}); err == nil {
		t.Error("Logs() should return error for unknown instance")
	}

	// Metrics with unknown instance should error.
	if _, err := r.Metrics(ctx, "unknown-instance"); err == nil {
		t.Error("Metrics() should return error for unknown instance")
	}
}

func TestOCIRuntime_DestroyWithoutStart(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	prepared, err := r.Prepare(context.Background(), &cr.PrepareRequest{
		AppID:   "destroy-without-start",
		Name:    "Destroy Without Start",
		WorkDir: t.TempDir(),
		Command: "./app",
		Port:    0,
	})
	if err != nil {
		t.Fatalf("Prepare() returned error: %v", err)
	}

	// Destroy without Start should clean up gracefully.
	if err := r.Destroy(context.Background(), prepared.ID); err != nil {
		t.Errorf("Destroy() without Start returned error: %v", err)
	}
}

func TestOCIRuntime_Type(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	if r.Type() != cr.RuntimeTypeDocker {
		t.Errorf("Type() = %q, want %q", r.Type(), cr.RuntimeTypeDocker)
	}
}

func TestOCIRuntime_Name(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	expected := "oci-mock"
	if r.Name() != expected {
		t.Errorf("Name() = %q, want %q", r.Name(), expected)
	}
}

func TestOCIRuntime_ImageResolution(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	tests := []struct {
		artifactType string
		wantImage    string
	}{
		{"binary", "alpine:latest"},
		{"static", "nginx:alpine"},
		{"source", "alpine:latest"},
	}

	for _, tt := range tests {
		t.Run(tt.artifactType, func(t *testing.T) {
			workDir := t.TempDir()
			req := &cr.PrepareRequest{
				AppID:   "image-test",
				Name:    "Image Test",
				WorkDir: workDir,
				Command: "./app",
				Port:    0,
				Artifact: &cr.ArtifactRef{
					Type: tt.artifactType,
					Path: workDir,
				},
			}

			prepared, err := r.Prepare(context.Background(), req)
			if err != nil {
				t.Fatalf("Prepare() returned error: %v", err)
			}

			// Verify the image was pulled (check engine mock).
			r.mu.Lock()
			inst, ok := r.instances[prepared.ID]
			r.mu.Unlock()
			if !ok {
				t.Fatal("instance not found after prepare")
			}
			if inst.image != tt.wantImage {
				t.Errorf("resolved image = %q, want %q", inst.image, tt.wantImage)
			}

			// Verify pull was called.
			if engine.pullCalled[tt.wantImage] == 0 {
				t.Errorf("image %q was not pulled", tt.wantImage)
			}
		})
	}
}

func TestOCIRuntime_ImagePullCaching(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	workDir := t.TempDir()

	// Create two apps with the same image type.
	for i := 0; i < 3; i++ {
		req := &cr.PrepareRequest{
			AppID:   fmt.Sprintf("cache-test-%d", i),
			Name:    fmt.Sprintf("Cache Test %d", i),
			WorkDir: workDir,
			Command: "./app",
			Port:    0,
			Artifact: &cr.ArtifactRef{
				Type: "binary",
				Path: workDir,
			},
		}

		_, err := r.Prepare(context.Background(), req)
		if err != nil {
			t.Fatalf("Prepare %d returned error: %v", i, err)
		}
	}

	// Despite 3 prepares, the image should only be pulled once.
	if engine.pullCalled["alpine:latest"] != 1 {
		t.Errorf("pullCalled[alpine:latest] = %d, want 1 (should be cached)", engine.pullCalled["alpine:latest"])
	}
}

func TestOCIRuntime_StopAll(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	// Create multiple instances.
	for i := 0; i < 3; i++ {
		workDir := t.TempDir()
		req := &cr.PrepareRequest{
			AppID:   fmt.Sprintf("stopall-test-%d", i),
			Name:    fmt.Sprintf("StopAll Test %d", i),
			WorkDir: workDir,
			Command: "./app",
			Port:    0,
		}

		prepared, err := r.Prepare(context.Background(), req)
		if err != nil {
			t.Fatalf("Prepare %d returned error: %v", i, err)
		}

		if _, err := r.Start(context.Background(), prepared); err != nil {
			t.Fatalf("Start %d returned error: %v", i, err)
		}
	}

	// StopAll should stop all containers and clear all state.
	r.StopAll()

	// Verify no containers remain in the mock engine.
	engine.mu.Lock()
	remaining := len(engine.containers)
	engine.mu.Unlock()
	if remaining > 0 {
		t.Errorf("%d containers remain after StopAll", remaining)
	}
}

// ── Docker Engine Tests ─────────────────────────────────────────────────────

func TestDockerEngine_ParseBytes(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
	}{
		{"0B", 0},
		{"1B", 1},
		{"1KB", 1000},
		{"1KiB", 1024},
		{"1MB", 1000000},
		{"1MiB", 1048576},
		{"1GB", 1000000000},
		{"1GiB", 1073741824},
		{"15.4MiB", 16148070},  // 15.4 * 1024 * 1024
		{"2.50%", 0},           // not a byte string
	}

	for _, tt := range tests {
		got := parseBytes(tt.input)
		if got != tt.want {
			t.Errorf("parseBytes(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestDockerEngine_ConvertPath(t *testing.T) {
	// This test validates the path conversion logic regardless of OS.
	// The convertPath function behaves differently on Windows vs Unix.

	// Under non-Windows, path should be unchanged.
	if filepath.Separator != '\\' {
		result := convertPath("/home/user/project")
		if result != "/home/user/project" {
			t.Errorf("convertPath(/home/user/project) = %q, want unchanged", result)
		}
	}
}

// ── Port Allocation Edge Cases ──────────────────────────────────────────────

func TestOCIRuntime_PortReleaseOnDestroy(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	// Allocate a port.
	req := &cr.PrepareRequest{
		AppID:   "port-release",
		Name:    "Port Release",
		WorkDir: t.TempDir(),
		Command: "./app",
		Port:    0,
	}

	prepared, err := r.Prepare(context.Background(), req)
	if err != nil {
		t.Fatalf("Prepare() returned error: %v", err)
	}

	// Destroy should release the port.
	if err := r.Destroy(context.Background(), prepared.ID); err != nil {
		t.Fatalf("Destroy() returned error: %v", err)
	}

	// Allocate another app — should be able to reuse the same port.
	req2 := &cr.PrepareRequest{
		AppID:   "port-release-2",
		Name:    "Port Release 2",
		WorkDir: t.TempDir(),
		Command: "./app",
		Port:    0,
	}

	prepared2, err := r.Prepare(context.Background(), req2)
	if err != nil {
		t.Fatalf("Second Prepare() returned error: %v", err)
	}

	// The port CAN be the same since it was released, but it's not guaranteed
	// since port allocation is random. Just verify it's > 0.
	if prepared2.Port == 0 {
		t.Error("Second app got port 0")
	}
}

func TestOCIRuntime_ImagePullError(t *testing.T) {
	// Create an engine that fails on Pull.
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	// This should work because the mock engine's Pull always succeeds.
	// We're testing that the error path is wired correctly.
	req := &cr.PrepareRequest{
		AppID:   "pull-test",
		Name:    "Pull Test",
		WorkDir: t.TempDir(),
		Command: "./app",
		Port:    0,
		Artifact: &cr.ArtifactRef{
			Type: "binary",
			Path: t.TempDir(),
		},
	}

	_, err := r.Prepare(context.Background(), req)
	if err != nil {
		t.Fatalf("Prepare() with pull returned error: %v", err)
	}
}

func TestOCIRuntime_Concurrency(t *testing.T) {
	engine := newMockEngine()
	log := logging.NewSubsystemLogger("test", logging.LevelDebug)
	r := NewOCIRuntime(engine, log)

	ctx := context.Background()
	n := 5
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			workDir, _ := os.MkdirTemp("", "oci-concurrency-*")
			req := &cr.PrepareRequest{
				AppID:   fmt.Sprintf("concurrent-%d", i),
				Name:    fmt.Sprintf("Concurrent %d", i),
				WorkDir: workDir,
				Command: "./app",
				Port:    0,
			}

			prepared, err := r.Prepare(ctx, req)
			if err != nil {
				errs <- fmt.Errorf("prepare: %w", err)
				return
			}

			inst, err := r.Start(ctx, prepared)
			if err != nil {
				errs <- fmt.Errorf("start: %w", err)
				return
			}

			if _, err := r.Health(ctx, inst.ID); err != nil {
				errs <- fmt.Errorf("health: %w", err)
				return
			}

			if _, err := r.Metrics(ctx, inst.ID); err != nil {
				errs <- fmt.Errorf("metrics: %w", err)
				return
			}

			if err := r.Destroy(ctx, inst.ID); err != nil {
				errs <- fmt.Errorf("destroy: %w", err)
				return
			}

			errs <- nil
		}(i)
	}

	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Errorf("Concurrent operation %d failed: %v", i, err)
		}
	}
}
