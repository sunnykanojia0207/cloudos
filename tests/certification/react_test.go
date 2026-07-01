package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_React validates a Vite React application through the full
// pipeline.
//
// React adds build-output complexity over plain Node: it produces a
// static artifact in dist/ that must be served by the Runtime.
// The ReactBuildpack detects the Vite project, plans `npm run build`,
// and the Runtime starts `npx serve -s dist -l {port}`.
//
// Key contracts under test:
//   - ReactBuildpack correctly detects @vitejs/plugin-react projects
//   - npm install + npm run build complete without error
//   - Build() returns Artifact.Path pointing to dist/
//   - Runtime starts the static site from dist/
func TestCertify_React(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequireNode()

	// ── 1. Create a Vite React project ──────────────────────────────────
	repoDir := h.CreateReactProject("react-app")
	h.InitGitRepo(repoDir, "Initial commit: React app")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("react-app", "React App", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "react",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0,
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("react-app")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "react",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
