package local

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// ── Test Logger ────────────────────────────────────────────────────────────

type testLogger struct {
	mu     sync.Mutex
	infos  []string
	errors []string
	debugs []string
}

func (l *testLogger) Info(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infos = append(l.infos, msg)
}

func (l *testLogger) Error(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, msg)
}

func (l *testLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugs = append(l.debugs, msg)
}

// ── Log Buffer Tests ───────────────────────────────────────────────────────

func TestLogBuffer_Write(t *testing.T) {
	buf := NewLogBuffer(10)

	n, err := buf.Write([]byte("hello\nworld\n"))
	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
	if n != 12 {
		t.Errorf("Write() returned %d, want 12", n)
	}

	lines := buf.Lines()
	if len(lines) != 2 {
		t.Fatalf("Lines() = %d, want 2", len(lines))
	}
	if lines[0] != "hello" {
		t.Errorf("Line[0] = %q, want %q", lines[0], "hello")
	}
	if lines[1] != "world" {
		t.Errorf("Line[1] = %q, want %q", lines[1], "world")
	}
}

func TestLogBuffer_RingWrapping(t *testing.T) {
	buf := NewLogBuffer(3)

	// Write 5 lines to a buffer of 3 → should keep last 3.
	for i := 0; i < 5; i++ {
		buf.Write([]byte(string(rune('A'+i)) + "\n"))
	}

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Lines() = %d, want 3", len(lines))
	}
	// Should be C, D, E (ASCII 67, 68, 69)
	expected := []string{"C", "D", "E"}
	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Line[%d] = %q, want %q", i, line, expected[i])
		}
	}
}

func TestLogBuffer_Empty(t *testing.T) {
	buf := NewLogBuffer(10)
	lines := buf.Lines()
	if lines != nil {
		t.Errorf("Lines() = %v, want nil", lines)
	}
}

func TestLogBuffer_Clear(t *testing.T) {
	buf := NewLogBuffer(10)
	buf.Write([]byte("hello\nworld\n"))
	buf.Clear()

	lines := buf.Lines()
	if lines != nil {
		t.Errorf("Lines() after Clear = %v, want nil", lines)
	}
}

func TestLogBuffer_WriteNoNewline(t *testing.T) {
	buf := NewLogBuffer(10)
	buf.Write([]byte("hello"))

	lines := buf.Lines()
	if len(lines) != 0 {
		t.Errorf("Lines() = %d, want 0 (no newline means no complete line)", len(lines))
	}
}

func TestLogBuffer_WriteMultiple(t *testing.T) {
	buf := NewLogBuffer(10)
	buf.Write([]byte("line1\nline2\nline3\n"))

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Lines() = %d, want 3", len(lines))
	}
}

// ── Port Pool Tests ────────────────────────────────────────────────────────

func TestPortPool_Get(t *testing.T) {
	pool := NewPortPool(9000, 9005)

	port, err := pool.Get()
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if port != 9000 {
		t.Errorf("Get() = %d, want %d", port, 9000)
	}
}

func TestPortPool_Sequential(t *testing.T) {
	pool := NewPortPool(9000, 9002)

	for i := 0; i < 3; i++ {
		port, err := pool.Get()
		if err != nil {
			t.Fatalf("Get() iteration %d returned error: %v", i, err)
		}
		if port != 9000+i {
			t.Errorf("Get() iteration %d = %d, want %d", i, port, 9000+i)
		}
	}
}

func TestPortPool_Exhausted(t *testing.T) {
	pool := NewPortPool(9000, 9000) // Only one port.

	_, _ = pool.Get()
	_, err := pool.Get()
	if err == nil {
		t.Error("Get() should return error when pool is exhausted")
	}
}

func TestPortPool_Release(t *testing.T) {
	pool := NewPortPool(9000, 9000)

	port, _ := pool.Get()
	pool.Release(port)

	// After release, Get should return the freed port.
	port2, err := pool.Get()
	if err != nil {
		t.Fatalf("Get() after release returned error: %v", err)
	}
	if port2 != port {
		t.Errorf("Get() after release = %d, want %d (freed port)", port2, port)
	}
}

func TestPortPool_ReleaseOutOfRange(t *testing.T) {
	pool := NewPortPool(9000, 9005)
	pool.Release(9999) // Should not panic.
}

// ── Manager Tests ──────────────────────────────────────────────────────────

func TestNewManager(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.ProcessCount() != 0 {
		t.Errorf("ProcessCount() = %d, want 0", m.ProcessCount())
	}
}

func TestManager_AllocatePort(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)

	port, err := m.AllocatePort()
	if err != nil {
		t.Fatalf("AllocatePort() returned error: %v", err)
	}
	if port < 9000 || port > 9999 {
		t.Errorf("AllocatePort() = %d, out of range", port)
	}
}

func TestManager_ListEmpty(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)

	procs := m.ListProcesses()
	if len(procs) != 0 {
		t.Errorf("ListProcesses() = %d, want 0", len(procs))
	}
}

func TestManager_GetNotFound(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)

	_, ok := m.GetProcess("nonexistent")
	if ok {
		t.Error("GetProcess() for nonexistent should return false")
	}
}

func TestManager_GetByAppIDNotFound(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)

	_, ok := m.GetProcessByAppID("nonexistent-app")
	if ok {
		t.Error("GetProcessByAppID() for nonexistent should return false")
	}
}

func TestManager_StartAndStopProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process test in short mode")
	}

	log := &testLogger{}
	workDir := t.TempDir()
	m := NewManager(workDir, log)

	// Create a simple batch script.
	batchContent := "@echo off\n:start\necho hello\ntimeout /t 1 /nobreak >nul\ngoto start\n"
	batchPath := filepath.Join(workDir, "test-server.bat")
	if err := os.WriteFile(batchPath, []byte(batchContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Start the process.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	proc, err := m.StartProcess(ctx, StartConfig{
		AppID:   "test-app",
		Name:    "Test App",
		WorkDir: workDir,
		Command: batchPath,
	})
	if err != nil {
		t.Fatalf("StartProcess() returned error: %v", err)
	}

	// Verify process was created.
	if proc.ID == "" {
		t.Error("Process ID should not be empty")
	}
	if proc.Status != StatusRunning {
		t.Errorf("Status = %q, want %q", proc.Status, StatusRunning)
	}
	if proc.Port < 9000 {
		t.Errorf("Port = %d, should be >= 9000", proc.Port)
	}
	if !strings.HasPrefix(proc.URL, "http://localhost:") {
		t.Errorf("URL = %q, should start with http://localhost:", proc.URL)
	}
	if proc.PID <= 0 {
		t.Errorf("PID = %d, should be > 0", proc.PID)
	}

	// Verify we can retrieve the process by ID.
	got, ok := m.GetProcess(proc.ID)
	if !ok {
		t.Fatal("GetProcess() returned false for running process")
	}
	if got.ID != proc.ID {
		t.Errorf("GetProcess() returned process with ID %q, want %q", got.ID, proc.ID)
	}

	// Verify we can retrieve by App ID.
	byApp, ok := m.GetProcessByAppID("test-app")
	if !ok {
		t.Fatal("GetProcessByAppID() returned false for running process")
	}
	if byApp.ID != proc.ID {
		t.Errorf("GetProcessByAppID() returned process with ID %q, want %q", byApp.ID, proc.ID)
	}

	// Stop the process.
	if err := m.StopProc(proc.ID); err != nil {
		t.Fatalf("StopProc() returned error: %v", err)
	}

	// Verify process is stopped.
	stopped, ok := m.GetProcess(proc.ID)
	if !ok {
		t.Fatal("GetProcess() returned false after stop")
	}
	if stopped.Status != StatusStopped {
		t.Errorf("Status after stop = %q, want %q", stopped.Status, StatusStopped)
	}
}

func TestManager_StopNonexistent(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)

	err := m.StopProc("nonexistent")
	if err == nil {
		t.Error("StopProc() for nonexistent process should return error")
	}
}

func TestManager_StopAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process test in short mode")
	}

	log := &testLogger{}
	workDir := t.TempDir()
	m := NewManager(workDir, log)

	// Start a couple of processes.
	for i := 0; i < 2; i++ {
		appID := "test-app-" + string(rune('A'+i))
		cmd := "@echo off\necho " + appID + "\ntimeout /t 10 /nobreak >nul\n"
		batchPath := filepath.Join(workDir, "proc-"+appID+".bat")
		os.WriteFile(batchPath, []byte(cmd), 0755)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := m.StartProcess(ctx, StartConfig{
			AppID:   appID,
			Name:    "Test " + string(rune('A'+i)),
			WorkDir: workDir,
			Command: batchPath,
		})
		if err != nil {
			t.Fatalf("StartProcess(%q) returned error: %v", appID, err)
		}
	}

	if m.ProcessCount() != 2 {
		t.Errorf("ProcessCount() = %d, want 2", m.ProcessCount())
	}

	// Stop all.
	m.StopAll()

	if m.ProcessCount() != 2 {
		t.Errorf("ProcessCount() after StopAll = %d, want 2", m.ProcessCount())
	}

	// All should be stopped.
	for _, proc := range m.ListProcesses() {
		if proc.Status != StatusStopped {
			t.Errorf("Process %q status = %q, want %q", proc.ID, proc.Status, StatusStopped)
		}
	}
}

func TestManager_GetLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping log test in short mode")
	}

	log := &testLogger{}
	workDir := t.TempDir()
	m := NewManager(workDir, log)

	batchContent := "@echo off\necho line1\necho line2\necho line3\n"
	batchPath := filepath.Join(workDir, "log-test.bat")
	if err := os.WriteFile(batchPath, []byte(batchContent), 0755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proc, err := m.StartProcess(ctx, StartConfig{
		AppID:   "log-test",
		Name:    "Log Test",
		WorkDir: workDir,
		Command: batchPath,
	})
	if err != nil {
		t.Fatalf("StartProcess() returned error: %v", err)
	}

	// Wait a bit for logs to be captured.
	time.Sleep(500 * time.Millisecond)

	// Get logs.
	logs, err := m.GetLogs(proc.ID, 10)
	if err != nil {
		t.Fatalf("GetLogs() returned error: %v", err)
	}
	if len(logs) == 0 {
		t.Log("No logs captured yet (might need more time)")
	}

	// Get logs with limit.
	limited, err := m.GetLogs(proc.ID, 2)
	if err != nil {
		t.Fatalf("GetLogs(limit=2) returned error: %v", err)
	}
	if len(limited) > 2 {
		t.Errorf("GetLogs(limit=2) returned %d lines, want <= 2", len(limited))
	}

	// Get logs for nonexistent process.
	_, err = m.GetLogs("nonexistent", 10)
	if err == nil {
		t.Error("GetLogs() for nonexistent process should return error")
	}

	m.StopAll()
}

func TestManager_GetLogsNonexistent(t *testing.T) {
	log := &testLogger{}
	m := NewManager(t.TempDir(), log)

	_, err := m.GetLogs("nonexistent", 10)
	if err == nil {
		t.Error("GetLogs() for nonexistent should return error")
	}
}

// ── Port Pool Concurrent Tests ─────────────────────────────────────────────

func TestPortPool_Concurrent(t *testing.T) {
	pool := NewPortPool(9000, 9099)

	var wg sync.WaitGroup
	ports := make(chan int, 100)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			port, err := pool.Get()
			if err == nil {
				ports <- port
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ports)
	}()

	collected := make(map[int]bool)
	for port := range ports {
		if collected[port] {
			t.Errorf("Port %d was allocated twice", port)
		}
		collected[port] = true
	}

	// Verify ports are sorted.
	uniquePorts := make([]int, 0, len(collected))
	for p := range collected {
		uniquePorts = append(uniquePorts, p)
	}
	sort.Ints(uniquePorts)
	for i := 1; i < len(uniquePorts); i++ {
		if uniquePorts[i] <= uniquePorts[i-1] {
			t.Errorf("Ports not sequential: %d after %d", uniquePorts[i], uniquePorts[i-1])
		}
	}
}

// ── Log Buffer Concurrent Tests ────────────────────────────────────────────

func TestLogBuffer_Concurrent(t *testing.T) {
	buf := NewLogBuffer(100)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				buf.Write([]byte(string(rune('A'+n)) + "\n"))
			}
		}(i)
	}
	wg.Wait()

	lines := buf.Lines()
	if len(lines) != 100 {
		t.Errorf("Lines() = %d, want 100", len(lines))
	}
}
