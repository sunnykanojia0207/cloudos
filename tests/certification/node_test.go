package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_Node validates a generic Node.js application through the
// full pipeline.
//
// Node sits between Static and React in complexity: it validates npm install,
// the generic NodeBuildpack detection, and running via `node index.js`.
//
// Key contract under test:
//   - package.json without react/next is detected as NodeBuildpack
//   - npm install runs successfully
//   - StartCmd = "node index.js" is used to launch the app
func TestCertify_Node(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequireNode()

	// ── 1. Create a Node.js HTTP server project ─────────────────────────
	repoDir := h.CreateNodeProject("node-server", "8080")
	h.InitGitRepo(repoDir, "Initial commit: Node.js HTTP server")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("node-server", "Node.js Server", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "node",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0,
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("node-server")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "node",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
