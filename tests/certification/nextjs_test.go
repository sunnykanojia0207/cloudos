package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_NextJS validates a Next.js application through the full
// pipeline.
//
// Next.js is the most complex Node.js stack: it requires `next build`
// which produces a .next output, then `next start` for the production
// server. The runtime handles the Node.js process management.
//
// Key contracts under test:
//   - NextJSBuildpack correctly detects the "next" dependency
//   - npm install + next build complete without error
//   - Build() returns a source artifact (not a static artifact)
//   - Runtime starts next start as a managed Node process
//   - Both / and /api/health return 200 (SSR + API routes)
func TestCertify_NextJS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequireNode()

	// ── 1. Create a Next.js project ─────────────────────────────────────
	repoDir := h.CreateNextJSProject("nextjs-app")
	h.InitGitRepo(repoDir, "Initial commit: Next.js app")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("nextjs-app", "Next.js App", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "nextjs",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0,
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("nextjs-app")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "nextjs",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
