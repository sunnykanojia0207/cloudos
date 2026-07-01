package certification

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
	"github.com/cloudos/cloudos/kernel/runtime/oci"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/packages/logging"
)

// TestCertify_OCI_Docker validates the OCI Runtime with Docker through
// the full CloudOS deployment pipeline.
//
// This is the critical validation of ADR-0009: the exact same Go application
// that was deployed through LocalRuntime should deploy through the OCI Runtime
// without changing Workflow, Controller, Buildpack, or certification logic.
//
//	Go Source → GoBuildpack → Artifact → OCIRuntime → Docker → Container
//
// Requirements:
//   - Docker Engine installed and running
//   - Go toolchain installed
//   - git installed
//
// Run: go test ./tests/certification/ -run TestCertify_OCI_Docker -v -count=1
func TestCertify_OCI_Docker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OCI certification test in short mode")
	}

	// ── Check Requirements ──────────────────────────────────────────────
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found on PATH, skipping OCI certification")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found on PATH, skipping OCI certification")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH, skipping OCI certification")
	}

	// Verify Docker daemon is running.
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon not running, skipping OCI certification")
	}

	// ── 1. Create a minimal Go HTTP server project ──────────────────────
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
		fmt.Fprintf(w, "Hello from CloudOS OCI! Port=%s", port)
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

	goMod := `module github.com/cloudos-certify/oci-server

go 1.21
`
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo.
	if err := gitInit(repoDir, "Initial commit: Go HTTP server for OCI"); err != nil {
		t.Fatal(err)
	}

	// ── 2. Set up CloudOS Kernel Components ──────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	log := logging.NewSubsystemLogger("oci-cert", logging.LevelDebug)

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

	// ── OCI Runtime with Docker Engine ──────────────────────────────────
	dockerEngine := oci.NewDockerEngine()
	if err := dockerEngine.Available(ctx); err != nil {
		t.Skipf("Docker engine not available: %v", err)
	}

	ociRuntime := oci.NewOCIRuntime(dockerEngine, log)
	defer ociRuntime.StopAll()

	// Source Cloner.
	sourceCloner := source.NewGitCloner(t.TempDir())

	// Workflow Service.
	workflowSvc := workflow.NewService(workflow.ServiceDeps{
		ResourceRegistry:  reg,
		ControllerManager: ctrlMgr,
		HealthManager:     healthMgr,
		EventBus:          bus,
		SourceCloner:      sourceCloner,
		RuntimeManager:    ociRuntime,
		Logger:            log,
	})

	// Application Controller.
	appCtrl := application.NewApplicationController(reg, bus, workflowSvc, log)
	if err := ctrlMgr.Register(appCtrl); err != nil {
		t.Fatal(err)
	}

	// ── 3. Start the Controller Runtime and Workflow Engine ─────────────
	if err := ctrlMgr.Start(ctx); err != nil {
		t.Fatal(err)
	}
	go workflowSvc.Engine().Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// ── 4. Create the Application ───────────────────────────────────────
	repoURL := "file:///" + strings.ReplaceAll(repoDir, "\\", "/")
	if !strings.HasPrefix(repoURL, "file:///") {
		repoURL = "file:///" + repoURL
	}

	app := application.NewApplication("oci-go-server", "OCI Go Server", application.ApplicationSpec{
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

	if err := reg.Create(ctx, app); err != nil {
		t.Fatal(err)
	}
	t.Logf("Created Application: %s (source: %s)", app.Metadata_.ID, repoURL)

	// ── 5. Wait for Deployment ──────────────────────────────────────────
	var deployedApp *application.Application
	var lastStatus string

	deadline := time.Now().Add(90 * time.Second)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		res, err := reg.Get(application.Kind, "oci-go-server")
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
			t.Logf("Application is RUNNING via OCI Runtime: %s", lastStatus)
			break
		}

		if phase == application.PhaseFailed {
			t.Fatalf("Application deployment FAILED via OCI: %s", lastStatus)
		}

		time.Sleep(pollInterval)
	}

	if deployedApp == nil {
		t.Fatalf("Application was never created (last: %s)", lastStatus)
	}
	if deployedApp.Status_.Phase != application.PhaseRunning {
		t.Fatalf("Application did not reach Running phase via OCI (last: %s)", lastStatus)
	}

	// ── 6. Verify HTTP Endpoint ─────────────────────────────────────────
	appURL := deployedApp.Status_.URL
	if appURL == "" {
		t.Fatal("Application URL was not set in status")
	}
	t.Logf("Application URL (OCI): %s", appURL)

	// Perform HTTP health check with retries.
	var resp *http.Response
	var httpErr error
	for i := 0; i < 15; i++ {
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
		t.Fatalf("Application health check via OCI failed: %v", httpErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Health endpoint returned HTTP %d via OCI, want 200", resp.StatusCode)
	}
	t.Logf("✓ OCI Health check passed: HTTP %d", resp.StatusCode)

	// Verify root endpoint.
	resp2, err := http.Get(appURL + "/")
	if err != nil {
		t.Fatalf("Root endpoint request via OCI failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Root endpoint returned HTTP %d via OCI, want 200", resp2.StatusCode)
	}
	t.Logf("✓ OCI Root endpoint returned HTTP %d", resp2.StatusCode)

	t.Logf("✓ OCI Runtime certification PASSED")
	t.Logf("✓ ADR-0009 validated: Go app through OCI Runtime with zero Workflow/Controller/Buildpack changes")
}

// gitInit initializes a git repository in dir with an initial commit.
func gitInit(dir, msg string) error {
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "test@cloudos.io"},
		{"config", "user.name", "CloudOS Test"},
		{"add", "-A"},
		{"commit", "-m", msg},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %v failed: %w\n%s", args, err, string(out))
		}
	}
	return nil
}
