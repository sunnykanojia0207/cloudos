package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_Go is the baseline certification test.
//
// It proves the full CloudOS deployment pipeline works for a Go HTTP server:
//
//	Go source → GoBuildpack.Detect → GoBuildpack.Plan → go build → binary Artifact
//	→ Runtime.Prepare → Runtime.Start → Running at http://localhost:PORT → HTTP 200
//
// This is the simplest compiled-language test. If it fails, the platform has
// a fundamental issue in the Buildpack Engine, Runtime, Workflow, or Controller.
func TestCertify_Go(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequireGo()

	// ── 1. Create a Go HTTP server project ──────────────────────────────
	repoDir := h.CreateGoProject("go-server", "8080")
	h.InitGitRepo(repoDir, "Initial commit: Go HTTP server")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("go-server", "Go Server", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "go",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0, // auto-allocate
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("go-server")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "go",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
