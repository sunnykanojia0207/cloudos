package certification

import (
	"testing"

	"github.com/cloudos/cloudos/kernel/application"
)

// TestCertify_Laravel validates a PHP/Laravel application through the full
// pipeline.
//
// Laravel is the most complex stack in the certification suite because it
// requires Composer (PHP package manager), the LaravelBuildpack detection,
// and running php artisan serve for development.
//
// Key contracts under test:
//   - LaravelBuildpack detects composer.json with "laravel/framework"
//   - composer install (if available) or skips if composer is absent
//   - StartCmd uses php to serve the application
//   - HTTP health check passes at /health
func TestCertify_Laravel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping certification test in short mode")
	}

	h := NewHarness(t)
	defer h.Cleanup()

	h.RequirePHP()

	// ── 1. Create a Laravel-like PHP project ────────────────────────────
	repoDir := h.CreateLaravelProject("laravel-app")
	h.InitGitRepo(repoDir, "Initial commit: Laravel app")

	// ── 2. Start the CloudOS kernel ─────────────────────────────────────
	h.StartKernel()

	// ── 3. Create the Application ───────────────────────────────────────
	app := h.CreateApp("laravel-app", "Laravel App", application.ApplicationSpec{
		Source: application.ApplicationSource{
			Type: application.SourceGit,
			URL:  h.FileURL(repoDir),
		},
		Runtime: application.ApplicationRuntime{
			Type: "laravel",
		},
		Deployment: application.ApplicationDeployment{
			Port: 0,
		},
	})

	// ── 4. Deploy ───────────────────────────────────────────────────────
	deployed := h.Deploy(app)

	// ── 5. Assertions ───────────────────────────────────────────────────
	h.AssertAppRunning("laravel-app")
	h.AssertURLReachable(deployed.Status_.URL)
	h.AssertRootEndpoint(deployed.Status_.URL)

	// ── 6. Emit certification result ────────────────────────────────────
	h.EmitResult(CertResult{
		Stack:   "laravel",
		Detect:  true,
		Plan:    true,
		Build:   true,
		Runtime: true,
		Health:  true,
		Logs:    true,
		Metrics: true,
	})
}
