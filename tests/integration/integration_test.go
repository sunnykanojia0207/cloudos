// Package integration contains end-to-end integration tests for CloudOS.
//
// These tests exercise the full platform stack: Resource Engine → Controller
// Runtime → Workflow Engine → Executor → Runtime (local processes).
//
// They require:
//   - git installed and available on PATH
//   - go installed (for building Go test apps)
//   - npx/serve available (for static apps)
//   - Network access (for git clone)
//
// Run: go test ./tests/integration/ -v -count=1
// Skip: go test ./tests/integration/ -short
package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/resource"
	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/kernel/runtime/local"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/packages/logging"
)

// TestDeployGoAppFromGit is the golden integration test for CloudOS.
//
// It proves the full deployment pipeline works end-to-end:
//
//	Git Repository → Application Resource → Application Controller
//	→ Workflow Service → Workflow Engine → Executor → Local Runtime
//	→ Running Application → Reachable HTTP URL
//
// The test creates a minimal Go HTTP server, inits a git repo, deploys
// it through the CloudOS stack, and verifies the server responds to HTTP.
func TestDeployGoAppFromGit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// ── 1. Create a minimal Go HTTP server project ───────────────────────

	repoDir := t.TempDir()
	mainGo := `package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from CloudOS! Port=%s", port)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	addr := "0.0.0.0:" + port
	fmt.Printf("Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
`
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	// Also create a go.mod so the Go buildpack detects it.
	goMod := `module github.com/cloudos-test/go-server

go 1.21
`
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo and commit.
	initGitRepo(t, repoDir, "Initial commit: Go HTTP server")

	// ── 2. Set up CloudOS Kernel Components ──────────────────────────────

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	log := logging.NewSubsystemLogger("integration", logging.LevelDebug)

	// Event Bus.
	bus := events.NewBus(log)
	bus.Start()
	defer bus.Stop()

	// Resource Registry.
	reg := resource.NewRegistry(bus, log)
	if err := reg.RegisterKind(resource.Kind{
		Name:       application.Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	// Health Manager.
	healthMgr := health.NewManager(log)

	// Controller Manager.
	ctrlMgr := controller.NewManager(reg, bus, healthMgr, log)

	// Local Runtime (manages application processes).
	runtimeMgr := local.NewManager(t.TempDir(), log)
	defer runtimeMgr.StopAll()

	// Log Manager — central log aggregator.
	logMgr := cr.NewLogManager(1000)
	runtimeMgr.WithLogManager(logMgr)

	// Source Cloner — clones git repositories.
	sourceCloner := source.NewGitCloner(t.TempDir())

	// Workflow Service (creates the engine + executor internally).
	workflowSvc := workflow.NewService(workflow.ServiceDeps{
		ResourceRegistry:  reg,
		ControllerManager: ctrlMgr,
		HealthManager:     healthMgr,
		EventBus:          bus,
		SourceCloner:      sourceCloner,
		RuntimeManager:    runtimeMgr,
		Logger:            log,
	})

	// Application Controller.
	appCtrl := application.NewApplicationController(reg, bus, workflowSvc, log)
	if err := ctrlMgr.Register(appCtrl); err != nil {
		t.Fatal(err)
	}

	// ── 3. Start the Controller Runtime and Workflow Engine ─────────────

	// Start the controller runtime (starts reconcile loops).
	if err := ctrlMgr.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// Start the workflow engine (starts the scheduler loop in a goroutine).
	go workflowSvc.Engine().Start(ctx)
	time.Sleep(100 * time.Millisecond) // let the engine start

	// ── 4. Create the Application ───────────────────────────────────────

	// Use file:// URL pointing to our local test repo.
	// On Windows, git needs the path converted: file:///C:/path/to/repo
	repoURL := "file:///" + strings.ReplaceAll(repoDir, "\\", "/")
	if !strings.HasPrefix(repoURL, "file:///") {
		repoURL = "file:///" + repoURL
	}

	app := application.NewApplication("test-go-app", "Test Go App", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  repoURL,
		},
		Runtime: application.ApplicationRuntime{
			Type: "go",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0, // auto-allocate
		},
	})
	app.EnsureDefaults()

	// Register the Application in the resource registry.
	// This triggers the controller to reconcile it.
	if err := reg.Create(ctx, app); err != nil {
		t.Fatal(err)
	}

	t.Logf("Created Application: %s (source: %s)", app.Metadata_.ID, repoURL)

	// ── 5. Wait for Deployment ──────────────────────────────────────────

	// Poll for the Application status to reach Running phase.
	var deployedApp *application.Application
	var lastStatus string

	deadline := time.Now().Add(60 * time.Second)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		// Fetch the Application from the registry.
		res, err := reg.Get(application.Kind, "test-go-app")
		if err != nil {
			lastStatus = fmt.Sprintf("get error: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		appObj, ok := res.(*application.Application)
		if !ok {
			lastStatus = "unexpected resource type"
			time.Sleep(pollInterval)
			continue
		}

		deployedApp = appObj
		phase := deployedApp.Status_.Phase
		health := deployedApp.Status_.Health
		url := deployedApp.Status_.URL

		lastStatus = fmt.Sprintf("phase=%s health=%s url=%s", phase, health, url)

		if phase == application.PhaseRunning && health == application.HealthHealthy {
			t.Logf("Application is RUNNING: %s", lastStatus)
			break
		}

		if phase == application.PhaseFailed {
			t.Fatalf("Application deployment FAILED: %s", lastStatus)
		}

		time.Sleep(pollInterval)
	}

	if deployedApp == nil {
		t.Fatalf("Application was never created (last status: %s)", lastStatus)
	}

	if deployedApp.Status_.Phase != application.PhaseRunning {
		t.Fatalf("Application did not reach Running phase (last: %s)", lastStatus)
	}

	// ── 6. Verify HTTP Endpoint ─────────────────────────────────────────

	appURL := deployedApp.Status_.URL
	if appURL == "" {
		// If the URL wasn't set in the status, try to find the running process.
		t.Fatal("Application URL was not set in status")
	}

	t.Logf("Application URL: %s", appURL)

	// Perform an HTTP health check with retries.
	var resp *http.Response
	var httpErr error
	for i := 0; i < 10; i++ {
		resp, httpErr = http.Get(appURL + "/health")
		if httpErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if httpErr != nil {
			t.Logf("Health check attempt %d: %v", i+1, httpErr)
		} else {
			resp.Body.Close()
			t.Logf("Health check attempt %d: HTTP %d", i+1, resp.StatusCode)
		}
		time.Sleep(1 * time.Second)
	}

	if httpErr != nil {
		t.Fatalf("Application health check failed: %v", httpErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Health endpoint returned HTTP %d, want 200", resp.StatusCode)
	}

	t.Logf("✓ Health check passed: HTTP %d", resp.StatusCode)

	// Also verify the root endpoint returns a greeting.
	resp2, err := http.Get(appURL + "/")
	if err != nil {
		t.Fatalf("Root endpoint request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Root endpoint returned HTTP %d, want 200", resp2.StatusCode)
	}

	t.Logf("✓ Root endpoint returned HTTP %d", resp2.StatusCode)
	t.Logf("✓ End-to-end deployment test PASSED")
}

// initGitRepo creates a git repo in dir and makes an initial commit.
func initGitRepo(t *testing.T, dir, msg string) {
	t.Helper()

	// Check for git.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH, skipping integration test")
	}

	cmds := []struct {
		name string
		args []string
	}{
		{"git", []string{"init", "-b", "main"}},
		{"git", []string{"config", "user.email", "test@cloudos.io"}},
		{"git", []string{"config", "user.name", "CloudOS Test"}},
		{"git", []string{"add", "-A"}},
		{"git", []string{"commit", "-m", msg}},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd.name, cmd.args...)
		c.Dir = dir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Fatalf("git %s failed: %v", cmd.args[0], err)
		}
	}
	t.Logf("Git repo initialized at %s", dir)
}
