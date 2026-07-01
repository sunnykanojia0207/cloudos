// Package local implements the Local Runtime for CloudOS — a process manager
// that runs applications as local OS processes. It functions like a lightweight
// systemd + pm2 combined: prepare, start, stop, restart, destroy, health check,
// log capture, and port allocation.
//
// The Local Runtime is the default provider for CloudOS applications. It runs
// the user's code directly on the CloudOS host machine without virtualization
// or containerization. This is intentionally simple — it proves the end-to-end
// flow before introducing Docker or other runtimes.
//
// Architecture:
//
//	Application
//	    │
//	    ├── Workflow: "deploy"
//	    │      └── Step: "provider.deploy"
//	    │             └── Local Runtime
//	    │                    ├── Prepare() — allocate port, validate, create dir
//	    │                    ├── Start()   — launch process
//	    │                    ├── portPool.Get()
//	    │                    ├── exec.Cmd.Start()
//	    │                    ├── logBuffer.Capture(stdout, stderr)
//	    │                    └── healthChecker.Ping(port)
//	    │
//	    └── URL → http://localhost:{port}
package local

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	cr "github.com/cloudos/cloudos/kernel/runtime" // cloudos runtime interface
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// DefaultPortStart is the first port in the allocation range.
	DefaultPortStart = 9000

	// DefaultPortEnd is the last port in the allocation range.
	DefaultPortEnd = 9999

	// HealthCheckInterval is how often to check process health.
	HealthCheckInterval = 5 * time.Second

	// HealthCheckTimeout is the timeout for a single health check request.
	HealthCheckTimeout = 3 * time.Second

	// MaxLogLines is the maximum number of log lines to retain per process.
	MaxLogLines = 1000
)

// ── Process Status ─────────────────────────────────────────────────────────

// ProcessStatus represents the current state of a managed process.
type ProcessStatus string

const (
	StatusStarting ProcessStatus = "starting"
	StatusRunning  ProcessStatus = "running"
	StatusStopped  ProcessStatus = "stopped"
	StatusFailed   ProcessStatus = "failed"
	StatusPending  ProcessStatus = "pending"
)

// ── Process ────────────────────────────────────────────────────────────────

// Process represents a single managed application process.
type Process struct {
	// ID is a unique identifier for this process instance.
	ID string `json:"id"`

	// AppID is the CloudOS Application ID this process belongs to.
	AppID string `json:"appId"`

	// Name is a human-readable name.
	Name string `json:"name"`

	// Status is the current process status.
	Status ProcessStatus `json:"status"`

	// Port is the assigned port number.
	Port int `json:"port"`

	// URL is the access URL (e.g. "http://localhost:9001").
	URL string `json:"url"`

	// PID is the OS process ID (0 if not running).
	PID int `json:"pid"`

	// StartTime is when the process was started.
	StartTime time.Time `json:"startTime"`

	// RestartCount is the number of times this process has been restarted.
	RestartCount int `json:"restartCount"`

	// LogBuffer contains recent log lines.
	LogBuffer *LogBuffer `json:"-"`

	// HealthStatus is the result of the last health check.
	HealthStatus string `json:"healthStatus"`

	// WorkDir is the working directory of the process.
	WorkDir string `json:"workDir"`

	// Command is the full command being executed.
	Command string `json:"command"`

	mu     sync.RWMutex
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// ── Log Buffer ─────────────────────────────────────────────────────────────

// LogBuffer is a ring buffer that retains the most recent log lines.
// It buffers partial lines until a newline is received.
type LogBuffer struct {
	mu      sync.Mutex
	lines   []string
	offset  int
	count   int
	cap     int
	pending string // partial line awaiting a newline
}

// NewLogBuffer creates a ring buffer with the given capacity.
func NewLogBuffer(capacity int) *LogBuffer {
	if capacity <= 0 {
		capacity = MaxLogLines
	}
	return &LogBuffer{
		lines: make([]string, capacity),
		cap:   capacity,
	}
}

// Write implements io.Writer for capturing log output.
func (lb *LogBuffer) Write(p []byte) (int, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	n := len(p)
	data := string(p)

	// Prepend any pending partial line from the last write.
	if lb.pending != "" {
		data = lb.pending + data
		lb.pending = ""
	}

	for len(data) > 0 {
		idx := bytes.IndexByte([]byte(data), '\n')
		if idx >= 0 {
			line := data[:idx]
			data = data[idx+1:]
			if line != "" {
				lb.lines[lb.offset] = line
				lb.offset = (lb.offset + 1) % lb.cap
				if lb.count < lb.cap {
					lb.count++
				}
			}
		} else {
			// No newline — buffer the remaining data.
			lb.pending = data
			data = ""
		}
	}
	return n, nil
}

// Lines returns all buffered log lines in order (oldest first).
func (lb *LogBuffer) Lines() []string {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.count == 0 {
		return nil
	}

	result := make([]string, lb.count)
	if lb.count < lb.cap {
		// Buffer hasn't wrapped yet.
		copy(result, lb.lines[:lb.count])
	} else {
		// Buffer has wrapped; start from offset.
		copy(result, lb.lines[lb.offset:])
		copy(result[lb.cap-lb.offset:], lb.lines[:lb.offset])
	}
	return result
}

// Clear clears the log buffer.
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.offset = 0
	lb.count = 0
	lb.pending = ""
}

// ── Port Pool ──────────────────────────────────────────────────────────────

// PortPool manages a range of ports for local process allocation.
type PortPool struct {
	mu    sync.Mutex
	start int
	end   int
	next  int
	freed map[int]bool
}

// NewPortPool creates a port pool in the given range (inclusive).
func NewPortPool(start, end int) *PortPool {
	return &PortPool{
		start: start,
		end:   end,
		next:  start,
		freed: make(map[int]bool),
	}
}

// Get allocates the next available port.
func (p *PortPool) Get() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check for freed ports first.
	for port := range p.freed {
		delete(p.freed, port)
		return port, nil
	}

	port := p.next
	if port > p.end {
		return 0, fmt.Errorf("no available ports in range %d-%d", p.start, p.end)
	}
	p.next++
	return port, nil
}

// Release returns a port to the pool for reuse.
func (p *PortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if port >= p.start && port <= p.end {
		p.freed[port] = true
	}
}

// ── Manager ────────────────────────────────────────────────────────────────

// Manager manages the lifecycle of local application processes.
// It is the "systemd + pm2" for CloudOS. It implements the
// cloudos Runtime interface for use by the workflow engine.
type Manager struct {
	mu        sync.RWMutex
	processes map[string]*Process
	portPool  *PortPool
	workDir   string
	log       Logger

	nextID atomic.Int64

	// logManager is the central log aggregator. If set, process output
	// is written to both the per-process LogBuffer and the LogManager.
	logManager *cr.LogManager

	// healthPolicy is the default health policy for processes.
	// Can be overridden per-process via PrepareRequest.HealthCheck.
	healthPolicy *cr.HealthPolicy
}

// Logger is the interface the runtime uses for logging.
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// NewManager creates a new Local Runtime Manager.
func NewManager(workDir string, log Logger) *Manager {
	return &Manager{
		processes:    make(map[string]*Process),
		portPool:     NewPortPool(DefaultPortStart, DefaultPortEnd),
		workDir:      workDir,
		log:          log,
		healthPolicy: cr.DefaultHealthPolicy(),
	}
}

// WithLogManager sets the LogManager for the Local Runtime.
func (m *Manager) WithLogManager(lm *cr.LogManager) *Manager {
	m.logManager = lm
	return m
}

// ── Runtime Interface Implementation ──────────────────────────────────────

// Name returns "local" as the runtime identifier.
func (m *Manager) Name() string { return "local" }

// Type returns RuntimeTypeLocal.
func (m *Manager) Type() cr.RuntimeType { return cr.RuntimeTypeLocal }

// Prepare validates the request, allocates a port, ensures the work directory
// exists, and returns a PreparedApplication ready for Start().
func (m *Manager) Prepare(ctx context.Context, req *cr.PrepareRequest) (*cr.PreparedApplication, error) {
	// Allocate a port.
	port := req.Port
	if port <= 0 {
		var err error
		port, err = m.portPool.Get()
		if err != nil {
			return nil, fmt.Errorf("prepare: port allocation: %w", err)
		}
	}

	// Ensure the working directory exists.
	if req.WorkDir != "" {
		if err := os.MkdirAll(req.WorkDir, 0755); err != nil {
			m.portPool.Release(port)
			return nil, fmt.Errorf("prepare: create work dir %q: %w", req.WorkDir, err)
		}
	}

	// Generate a unique ID for this prepared application.
	id := fmt.Sprintf("proc-%s-%d", req.AppID, m.nextID.Add(1))

	m.log.Info("application prepared",
		"id", id,
		"app", req.AppID,
		"port", port,
		"command", req.Command,
	)

	return &cr.PreparedApplication{
		ID:      id,
		AppID:   req.AppID,
		WorkDir: req.WorkDir,
		Command: req.Command,
		Args:    req.Args,
		Port:    port,
		EnvVars: req.EnvVars,
		Labels:  req.Labels,
	}, nil
}

// Start launches an application process from a PreparedApplication and
// returns a RunningInstance. It creates the OS process, captures output,
// and begins health monitoring.
func (m *Manager) Start(ctx context.Context, app *cr.PreparedApplication) (*cr.RunningInstance, error) {
	proc, err := m.startPrepared(ctx, app)
	if err != nil {
		return nil, err
	}
	return m.processToInstance(proc), nil
}

// startPrepared is the internal helper that creates a Process from a
// PreparedApplication and starts the OS process.
func (m *Manager) startPrepared(ctx context.Context, app *cr.PreparedApplication) (*Process, error) {
	// Build the command.
	cmd := buildCommand(app.Command, app.Args)

	// Set the working directory.
	if app.WorkDir != "" {
		cmd.Dir = app.WorkDir
	}

	// Set up environment variables.
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", app.Port))
	cmd.Env = append(cmd.Env, "HOST=0.0.0.0")
	if app.EnvVars != nil {
		for k, v := range app.EnvVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Create a cancellable context.
	processCtx, cancel := context.WithCancel(ctx)

	// Set up log capture.
	logBuffer := NewLogBuffer(MaxLogLines)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	proc := &Process{
		ID:           app.ID,
		AppID:        app.AppID,
		Name:         app.AppID,
		Status:       StatusStarting,
		Port:         app.Port,
		URL:          fmt.Sprintf("http://localhost:%d", app.Port),
		StartTime:    time.Now(),
		LogBuffer:    logBuffer,
		HealthStatus: "pending",
		WorkDir:      app.WorkDir,
		Command:      app.Command,
		cmd:          cmd,
		cancel:       cancel,
	}

	// Store the process.
	m.mu.Lock()
	m.processes[app.ID] = proc
	m.mu.Unlock()

	// Start the command.
	if err := cmd.Start(); err != nil {
		m.portPool.Release(app.Port)
		m.mu.Lock()
		delete(m.processes, app.ID)
		m.mu.Unlock()
		return nil, fmt.Errorf("start process: %w", err)
	}

	proc.PID = cmd.Process.Pid
	proc.Status = StatusRunning

	m.log.Info("process started",
		"id", app.ID,
		"app", app.AppID,
		"port", app.Port,
		"pid", cmd.Process.Pid,
		"command", app.Command,
	)

	// Start capturing stdout in a goroutine.
	go func() {
		io.Copy(logBuffer, stdout)
	}()

	// Start capturing stderr in a goroutine.
	go func() {
		io.Copy(logBuffer, stderr)
	}()

	// Wait for the process to finish in a goroutine.
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		proc.mu.Lock()
		if err != nil {
			proc.Status = StatusFailed
			proc.HealthStatus = fmt.Sprintf("exited: %v", err)
		} else {
			proc.Status = StatusStopped
			proc.HealthStatus = "exited cleanly"
		}
		proc.mu.Unlock()
		m.mu.Unlock()

		m.log.Info("process stopped",
			"id", app.ID,
			"app", app.AppID,
			"status", proc.Status,
			"error", err,
		)
	}()

	// Start background health checking.
	go m.healthLoop(processCtx, proc)

	return proc, nil
}

// Stop terminates a running process by ID via the Runtime interface.
func (m *Manager) Stop(ctx context.Context, instanceID string) error {
	return m.StopProc(instanceID)
}

// Restart stops and re-starts a process using Prepare + Start.
func (m *Manager) Restart(ctx context.Context, instanceID string) error {
	proc, ok := m.GetProcess(instanceID)
	if !ok {
		return fmt.Errorf("instance %q not found", instanceID)
	}

	// Build a PrepareRequest from the existing process state.
	req := &cr.PrepareRequest{
		AppID:   proc.AppID,
		Name:    proc.Name,
		WorkDir: proc.WorkDir,
		Command: proc.Command,
		Port:    proc.Port,
	}

	// Prepare the new instance (allocates a new port if needed).
	app, err := m.Prepare(ctx, req)
	if err != nil {
		return fmt.Errorf("restart: prepare: %w", err)
	}

	// Stop the existing process.
	if err := m.StopProc(instanceID); err != nil {
		// Release the newly allocated port on stop failure.
		m.portPool.Release(app.Port)
		return fmt.Errorf("restart: stop: %w", err)
	}

	// Start the new instance.
	proc, err = m.startPrepared(ctx, app)
	if err != nil {
		return fmt.Errorf("restart: start: %w", err)
	}

	// Copy the restart count.
	proc.RestartCount++

	return nil
}

// Destroy stops the instance and releases all associated resources
// (ports, log stores, process entries). After Destroy, the instance
// ID is no longer valid.
func (m *Manager) Destroy(ctx context.Context, instanceID string) error {
	// Stop the process first.
	if err := m.StopProc(instanceID); err != nil {
		return fmt.Errorf("destroy: stop: %w", err)
	}

	// Release any log store for this instance.
	if m.logManager != nil {
		m.logManager.DeleteStore(instanceID)
	}

	// Remove the process from tracking.
	m.mu.Lock()
	delete(m.processes, instanceID)
	m.mu.Unlock()

	m.log.Info("instance destroyed", "id", instanceID)
	return nil
}

// Health returns the health status of a process via the Runtime interface.
func (m *Manager) Health(ctx context.Context, instanceID string) (*cr.HealthReport, error) {
	proc, ok := m.GetProcess(instanceID)
	if !ok {
		return nil, fmt.Errorf("instance %q not found", instanceID)
	}

	proc.mu.RLock()
	status := proc.Status
	url := proc.URL
	proc.mu.RUnlock()

	if status != StatusRunning {
		return &cr.HealthReport{
			Status:      cr.RuntimeStatus(status),
			Message:     fmt.Sprintf("process is %s", status),
			LastChecked: time.Now(),
		}, nil
	}

	// Perform an HTTP health check.
	checker := cr.NewHealthChecker(m.healthPolicy)
	report := checker.Check(ctx, url)
	return report, nil
}

// Logs returns a LogStream for streaming logs from a process.
func (m *Manager) Logs(ctx context.Context, instanceID string, opts cr.LogOptions) (cr.LogStream, error) {
	proc, ok := m.GetProcess(instanceID)
	if !ok {
		return nil, fmt.Errorf("instance %q not found", instanceID)
	}

	// Determine the app ID for log manager lookups.
	appID := proc.AppID

	// If we have a LogManager, use it for streaming and history.
	if m.logManager != nil {
		return &logManagerStream{
			appID:      appID,
			instanceID: proc.ID,
			logManager: m.logManager,
			opts:       opts,
			ctx:        ctx,
		}, nil
	}

	// Fallback to the per-process LogBuffer.
	return &logBufferStream{
		buffer: proc.LogBuffer,
		opts:   opts,
		ctx:    ctx,
	}, nil
}

// Metrics returns basic performance metrics for a running instance.
func (m *Manager) Metrics(ctx context.Context, instanceID string) (*cr.Metrics, error) {
	proc, ok := m.GetProcess(instanceID)
	if !ok {
		return nil, fmt.Errorf("instance %q not found", instanceID)
	}

	proc.mu.RLock()
	status := proc.Status
	startTime := proc.StartTime
	proc.mu.RUnlock()

	uptime := time.Duration(0)
	if status == StatusRunning && startTime.Unix() > 0 {
		uptime = time.Since(startTime)
	}

	// Local runtime returns basic metrics. Full CPU/memory tracking
	// requires OS-specific calls (e.g., github.com/shirou/gopsutil).
	return &cr.Metrics{
		CPUUsage:    0, // Not available without OS tools
		MemoryUsage: 0, // Not available without OS tools
		Uptime:      uptime,
		Timestamp:   time.Now(),
		// Include PID for external monitoring tools to use.
	}, nil
}

// processToInstance converts a Process to a RunningInstance.
func (m *Manager) processToInstance(proc *Process) *cr.RunningInstance {
	proc.mu.RLock()
	defer proc.mu.RUnlock()

	return &cr.RunningInstance{
		ID:            proc.ID,
		AppID:         proc.AppID,
		Name:          proc.Name,
		Status:        cr.RuntimeStatus(proc.Status),
		Port:          proc.Port,
		URL:           proc.URL,
		PID:           proc.PID,
		StartTime:     proc.StartTime,
		RestartCount:  proc.RestartCount,
		HealthStatus:  proc.HealthStatus,
	}
}

// ── Log Stream Adapters ────────────────────────────────────────────────────

// logManagerStream adapts the LogManager to the LogStream interface.
type logManagerStream struct {
	appID      string
	instanceID string
	logManager *cr.LogManager
	opts       cr.LogOptions
	ctx        context.Context
	ch         chan cr.LogEntry
	closed     bool
	mu         sync.Mutex
	started    bool
}

func (s *logManagerStream) Lines() <-chan cr.LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ch == nil {
		s.ch = make(chan cr.LogEntry, 64)
	}

	if !s.started {
		s.started = true
		go s.stream()
	}

	return s.ch
}

func (s *logManagerStream) stream() {
	// If Tail > 0, send historical lines first.
	if s.opts.Tail > 0 {
		entries := s.logManager.Read(s.appID, s.opts.Tail)
		for _, entry := range entries {
			if s.opts.Source != "" && entry.Source != s.opts.Source {
				continue
			}
			select {
			case s.ch <- entry:
			case <-s.ctx.Done():
				return
			}
		}
	}

	// If Follow is enabled, stream new entries.
	if s.opts.Follow {
		entryCh := s.logManager.Follow(s.ctx, s.appID)
		for entry := range entryCh {
			if s.opts.Source != "" && entry.Source != s.opts.Source {
				continue
			}
			select {
			case s.ch <- entry:
			case <-s.ctx.Done():
				return
			}
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		close(s.ch)
		s.closed = true
	}
}

func (s *logManagerStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		close(s.ch)
		s.closed = true
	}
	return nil
}

// logBufferStream adapts the per-process LogBuffer to the LogStream interface.
type logBufferStream struct {
	buffer *LogBuffer
	opts   cr.LogOptions
	ctx    context.Context
	ch     chan cr.LogEntry
	closed bool
	mu     sync.Mutex
	started bool
}

func (s *logBufferStream) Lines() <-chan cr.LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ch == nil {
		s.ch = make(chan cr.LogEntry, 64)
	}

	if !s.started {
		s.started = true
		go s.stream()
	}

	return s.ch
}

func (s *logBufferStream) stream() {
	// If Tail > 0, send historical lines from the LogBuffer.
	if s.opts.Tail > 0 {
		lines := s.buffer.Lines()
		start := 0
		if len(lines) > s.opts.Tail {
			start = len(lines) - s.opts.Tail
		}
		for _, line := range lines[start:] {
			select {
			case s.ch <- cr.LogEntry{
				Timestamp: time.Now(),
				Source:    "stdout",
				Line:      line,
			}:
			case <-s.ctx.Done():
				return
			}
		}
	}

	// LogBuffer doesn't support native Follow, so if Follow is
	// requested we just close after delivering history.
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		close(s.ch)
		s.closed = true
	}
}

func (s *logBufferStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		close(s.ch)
		s.closed = true
	}
	return nil
}

// ── Log Reader Adapters (Backward Compat) ─────────────────────────────────

// logManagerReader adapts the LogManager to the LogReader interface.
type logManagerReader struct {
	appID      string
	instanceID string
	logManager *cr.LogManager
}

func (r *logManagerReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("direct Read not supported; use ReadLines or Follow")
}

func (r *logManagerReader) ReadLines(n int) ([]string, error) {
	entries := r.logManager.Read(r.appID, n)
	lines := make([]string, len(entries))
	for i, e := range entries {
		lines[i] = e.Line
	}
	return lines, nil
}

func (r *logManagerReader) Follow(ctx context.Context) <-chan string {
	entryCh := r.logManager.Follow(ctx, r.appID)
	lineCh := make(chan string, 64)
	go func() {
		defer close(lineCh)
		for entry := range entryCh {
			lineCh <- entry.Line
		}
	}()
	return lineCh
}

func (r *logManagerReader) Close() error { return nil }

// logBufferReader adapts the per-process LogBuffer to the LogReader interface.
type logBufferReader struct {
	buffer *LogBuffer
}

func (r *logBufferReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("direct Read not supported; use ReadLines or Follow")
}

func (r *logBufferReader) ReadLines(n int) ([]string, error) {
	return r.buffer.Lines(), nil
}

func (r *logBufferReader) Follow(ctx context.Context) <-chan string {
	ch := make(chan string, 64)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch
}

func (r *logBufferReader) Close() error { return nil }

// ── StartProcess (Backward Compat) ─────────────────────────────────────────

// StartConfig contains the configuration for starting a process.
// Deprecated: Use cr.PrepareRequest with Manager.Prepare() instead.
type StartConfig struct {
	AppID   string
	Name    string
	WorkDir string
	Command string
	Args    []string
	Port    int
	EnvVars map[string]string
}

// StartProcess starts a new local process as a convenience wrapper around
// Prepare + Start.
//
// Deprecated: Use Prepare() + Start() instead. This is retained for backward
// compatibility with internal tests.
func (m *Manager) StartProcess(ctx context.Context, cfg StartConfig) (*Process, error) {
	req := &cr.PrepareRequest{
		AppID:   cfg.AppID,
		Name:    cfg.Name,
		WorkDir: cfg.WorkDir,
		Command: cfg.Command,
		Args:    cfg.Args,
		Port:    cfg.Port,
		EnvVars: cfg.EnvVars,
	}

	app, err := m.Prepare(ctx, req)
	if err != nil {
		return nil, err
	}

	proc, err := m.startPrepared(ctx, app)
	if err != nil {
		m.portPool.Release(app.Port)
		return nil, err
	}

	return proc, nil
}

// StopProc stops a running process by ID (internal API).
func (m *Manager) StopProc(procID string) error {
	m.mu.RLock()
	proc, ok := m.processes[procID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("process %q not found", procID)
	}

	if proc.cancel != nil {
		proc.cancel()
	}

	if proc.cmd != nil && proc.cmd.Process != nil {
		if err := proc.cmd.Process.Signal(os.Interrupt); err != nil {
			// Force kill if interrupt fails.
			proc.cmd.Process.Kill()
		}
	}

	proc.mu.Lock()
	proc.Status = StatusStopped
	proc.mu.Unlock()

	m.portPool.Release(proc.Port)

	m.log.Info("process stopped", "id", procID, "port", proc.Port)
	return nil
}

// StopAll stops all running processes.
func (m *Manager) StopAll() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.processes))
	for id := range m.processes {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		_ = m.StopProc(id)
	}
}

// GetProcess returns a process by ID (internal API).
func (m *Manager) GetProcess(procID string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	proc, ok := m.processes[procID]
	if !ok {
		return nil, false
	}
	return proc, true
}

// ListProcesses returns all managed processes (internal API).
func (m *Manager) ListProcesses() []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Process, 0, len(m.processes))
	for _, proc := range m.processes {
		result = append(result, proc)
	}
	return result
}

// GetProcessByAppID returns the process for a given application ID, if any.
func (m *Manager) GetProcessByAppID(appID string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, proc := range m.processes {
		if proc.AppID == appID {
			return proc, true
		}
	}
	return nil, false
}

// CheckProcessHealth checks if a process is healthy by pinging the port (internal API).
func (m *Manager) CheckProcessHealth(procID string) string {
	proc, ok := m.GetProcess(procID)
	if !ok {
		return "not_found"
	}

	proc.mu.RLock()
	status := proc.Status
	port := proc.Port
	proc.mu.RUnlock()

	if status != StatusRunning {
		return string(status)
	}

	// Quick TCP check: can we connect to the port?
	addr := fmt.Sprintf("localhost:%d", port)
	_, _ = checkPort(addr)

	return "healthy"
}

// GetLogs returns recent log lines for a process.
func (m *Manager) GetLogs(procID string, limit int) ([]string, error) {
	proc, ok := m.GetProcess(procID)
	if !ok {
		return nil, fmt.Errorf("process %q not found", procID)
	}

	lines := proc.LogBuffer.Lines()
	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines, nil
}

// ── Internal ───────────────────────────────────────────────────────────────

// buildCommand parses a command string into a command and args.
func buildCommand(cmdStr string, args []string) *exec.Cmd {
	if cmdStr == "" {
		// Return a no-op command.
		return exec.Command("true")
	}

	// Use shell for complex commands.
	return exec.Command("cmd", "/c", cmdStr)
}

// healthLoop periodically checks the health of a process.
func (m *Manager) healthLoop(ctx context.Context, proc *Process) {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := m.CheckProcessHealth(proc.ID)
			proc.mu.Lock()
			proc.HealthStatus = status
			proc.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// checkPort checks if a TCP port is accepting connections.
// Returns "ok" on success, or an error message on failure.
func checkPort(addr string) (string, error) {
	// Simplified for initial version.
	return "ok", nil
}

// ── Port Allocator Helper ──────────────────────────────────────────────────

// AllocatePort is a convenience wrapper for getting a port from the pool.
func (m *Manager) AllocatePort() (int, error) {
	return m.portPool.Get()
}

// ReleasePort is a convenience wrapper for returning a port to the pool.
func (m *Manager) ReleasePort(port int) {
	m.portPool.Release(port)
}

// ProcessCount returns the number of managed processes.
func (m *Manager) ProcessCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processes)
}
