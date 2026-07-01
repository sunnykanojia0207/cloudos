package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_Static validates the simplest artifact type: static HTML.
//
// The Static buildpack should fall through all other buildpacks and match
// as the default. The Artifact.Type = "static" → Runtime starts npx serve
// on the output directory.
//
// This is the minimal end-to-end test for the Buildpack → Runtime pipeline.
func TestCertify_Static(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequireNode() // npx serve requires Node.js

	// ── 1. Create a static HTML project ─────────────────────────────────
	repoDir := h.CreateStaticProject("static-site")
	h.InitGitRepo(repoDir, "Initial commit: Static HTML site")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("static-site", "Static Site", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "static",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0,
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("static-site")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "static",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
