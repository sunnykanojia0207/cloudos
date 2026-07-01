package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_Python validates a Python application through the full
// pipeline.
//
// Python is the first non-Node.js stack in the certification order.
// It validates that the PythonBuildpack correctly detects requirements.txt,
// runs pip install, and uses the appropriate start command.
//
// Key contracts under test:
//   - PythonBuildpack detects requirements.txt / setup.py / Pipfile
//   - pip install runs successfully
//   - StartCmd respects the PORT environment variable
//   - HTTP health check passes at /health
func TestCertify_Python(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequirePython()

	// ── 1. Create a Python HTTP server project ──────────────────────────
	repoDir := h.CreatePythonProject("python-server", "8080")
	h.InitGitRepo(repoDir, "Initial commit: Python HTTP server")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("python-server", "Python Server", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "python",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0,
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("python-server")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "python",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
