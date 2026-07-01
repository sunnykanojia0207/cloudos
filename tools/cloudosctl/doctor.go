package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cloudos/cloudos/packages/version"
)

// ═════════════════════════════════════════════════════════════════════════════
// Checker Interface
// ═════════════════════════════════════════════════════════════════════════════

// CheckResult is the outcome of a single environment check.
type CheckResult struct {
	// Name is the human-readable check name (e.g. "Git", "Docker Daemon").
	Name string

	// Passed is true when the check succeeded.
	Passed bool

	// Value is the version or status string (e.g. "2.51.0", "Running").
	Value string

	// Error is what specifically failed (empty when Passed is true).
	Error string

	// Reason explains why this check matters to the user.
	Reason string

	// Fix describes how to resolve the failure.
	Fix string
}

// Checker is the interface every environment check implements.
// Checks are independent, read-only, and must never modify the system.
type Checker interface {
	// Name returns the display name of this check.
	Name() string

	// Run performs the check and returns a result.
	// The context can be used for timeouts on slow checks.
	Run(ctx context.Context) *CheckResult
}

// ═════════════════════════════════════════════════════════════════════════════
// Doctor Runner
// ═════════════════════════════════════════════════════════════════════════════

// doctor runs a sequence of environment checks and prints results.
// It is read-only and must never modify any system state.
func doctor() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	checks := []Checker{
		&cloudOSCheck{},
		&gitCheck{},
		&dockerInstalledCheck{},
		&dockerRunningCheck{},
		&goCheck{},
		&nodeCheck{},
		&npmCheck{},
		&pythonCheck{},
		&phpCheck{},
		&composerCheck{},
		&portCheck{},
		&runtimeCheck{},
		&buildpackCheck{},
		&dirCheck{},
	}

	fmt.Println()
	fmt.Println("Checking CloudOS Environment...")
	fmt.Println()

	allPassed := true
	for _, c := range checks {
		result := c.Run(ctx)
		if result.Passed {
			fmt.Printf("  ✓ %s\n", result.Name)
			if result.Value != "" {
				fmt.Printf("    %s\n", result.Value)
			}
		} else {
			allPassed = false
			fmt.Printf("  ✗ %s\n", result.Name)
			if result.Value != "" {
				fmt.Printf("    %s\n", result.Value)
			}
			if result.Error != "" {
				fmt.Printf("    %s\n", result.Error)
			}
			if result.Reason != "" {
				fmt.Printf("    → %s\n", result.Reason)
			}
			if result.Fix != "" {
				lines := strings.Split(result.Fix, "\n")
				for i, line := range lines {
					if i == 0 {
						fmt.Printf("    Fix: %s\n", line)
					} else {
						fmt.Printf("         %s\n", line)
					}
				}
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("Environment Status")
	fmt.Println()
	if allPassed {
		fmt.Println("  ✓ Ready to deploy applications")
	} else {
		fmt.Printf("  ✗ %d issue(s) found. Fix the items above and run again:\n", countFailed(checks))
		fmt.Println()
		fmt.Println("    cloudosctl doctor")
	}
	fmt.Println()

	return allPassed
}

// countFailed counts how many checks failed.
func countFailed(checks []Checker) int {
	count := 0
	for _, c := range checks {
		if !c.Run(context.Background()).Passed {
			count++
		}
	}
	return count
}

// ═════════════════════════════════════════════════════════════════════════════
// Individual Checks
// ═════════════════════════════════════════════════════════════════════════════

// --- CloudOS CLI -----------------------------------------------------------

type cloudOSCheck struct{}

func (c *cloudOSCheck) Name() string { return "CloudOS CLI" }

func (c *cloudOSCheck) Run(ctx context.Context) *CheckResult {
	return &CheckResult{
		Name:   "CloudOS CLI",
		Passed: true,
		Value:  fmt.Sprintf("Version %s", version.Number),
	}
}

// --- Git -------------------------------------------------------------------

type gitCheck struct{}

func (c *gitCheck) Name() string { return "Git" }

func (c *gitCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "git", "--version")
	if err != nil {
		return &CheckResult{
			Name:   "Git",
			Passed: false,
			Error:  "Git is not installed or not found in PATH.",
			Reason: "Git is required to clone source repositories for deployment.",
			Fix:    installGuide("Git", "https://git-scm.com/downloads", "brew install git", "sudo apt install git"),
		}
	}
	return &CheckResult{
		Name:   "Git",
		Passed: true,
		Value:  strings.TrimSpace(v),
	}
}

// --- Docker Installed ------------------------------------------------------

type dockerInstalledCheck struct{}

func (c *dockerInstalledCheck) Name() string { return "Docker" }

func (c *dockerInstalledCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "docker", "--version")
	if err != nil {
		return &CheckResult{
			Name:   "Docker",
			Passed: false,
			Error:  "Docker is not installed or not found in PATH.",
			Reason: "Docker provides container isolation for deployed applications. Without Docker, applications run without isolation.",
			Fix:    installGuide("Docker Desktop", "https://docs.docker.com/desktop/", "brew install --cask docker", "sudo apt install docker.io"),
		}
	}
	return &CheckResult{
		Name:   "Docker",
		Passed: true,
		Value:  strings.TrimSpace(v),
	}
}

// --- Docker Daemon Running -------------------------------------------------

type dockerRunningCheck struct{}

func (c *dockerRunningCheck) Name() string { return "Docker Daemon" }

func (c *dockerRunningCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "docker", "info", "--format", "{{.ServerVersion}}")
	if err != nil {
		fixMsg := "Open Docker Desktop and wait for the daemon to start. Then run:\n    cloudosctl doctor"
		if runtime.GOOS == "linux" {
			fixMsg = "Start the Docker daemon:\n    sudo systemctl start docker\n  Then run:\n    cloudosctl doctor"
		}
		return &CheckResult{
			Name:   "Docker Daemon",
			Passed: false,
			Error:  "Docker is installed but the daemon is not running.",
			Reason: "The Docker daemon must be running to deploy applications in containers.",
			Fix:    fixMsg,
		}
	}
	return &CheckResult{
		Name:   "Docker Daemon",
		Passed: true,
		Value:  fmt.Sprintf("Running (v%s)", strings.TrimSpace(v)),
	}
}

// --- Go --------------------------------------------------------------------

type goCheck struct{}

func (c *goCheck) Name() string { return "Go" }

func (c *goCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "go", "version")
	if err != nil {
		return &CheckResult{
			Name:   "Go",
			Passed: false,
			Error:  "Go is not installed or not found in PATH.",
			Reason: "Go is required to build and deploy Go applications. It is optional for other stacks.",
			Fix:    installGuide("Go", "https://go.dev/dl/", "brew install go", "sudo apt install golang-go"),
		}
	}
	return &CheckResult{
		Name:   "Go",
		Passed: true,
		Value:  parseVersion(v),
	}
}

// --- Node.js ---------------------------------------------------------------

type nodeCheck struct{}

func (c *nodeCheck) Name() string { return "Node.js" }

func (c *nodeCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "node", "--version")
	if err != nil {
		return &CheckResult{
			Name:   "Node.js",
			Passed: false,
			Error:  "Node.js is not installed or not found in PATH.",
			Reason: "Node.js is required to build JavaScript, TypeScript, React, and Next.js applications. It is optional for other stacks.",
			Fix:    installGuide("Node.js", "https://nodejs.org/", "brew install node", "sudo apt install nodejs npm"),
		}
	}
	return &CheckResult{
		Name:   "Node.js",
		Passed: true,
		Value:  strings.TrimSpace(v),
	}
}

// --- npm -------------------------------------------------------------------

type npmCheck struct{}

func (c *npmCheck) Name() string { return "npm" }

func (c *npmCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "npm", "--version")
	if err != nil {
		return &CheckResult{
			Name:   "npm",
			Passed: false,
			Error:  "npm is not installed.",
			Reason: "npm is required to install dependencies for Node.js, React, and Next.js applications.",
			Fix:    "npm is installed with Node.js. If Node.js is present but npm is missing, reinstall Node.js from https://nodejs.org/",
		}
	}
	return &CheckResult{
		Name:   "npm",
		Passed: true,
		Value:  fmt.Sprintf("v%s", strings.TrimSpace(v)),
	}
}

// --- Python ----------------------------------------------------------------

type pythonCheck struct{}

func (c *pythonCheck) Name() string { return "Python" }

func (c *pythonCheck) Run(ctx context.Context) *CheckResult {
	// Try python3 first, fall back to python
	v, err := execCmd(ctx, "python3", "--version")
	if err != nil {
		v2, err2 := execCmd(ctx, "python", "--version")
		if err2 != nil {
			return &CheckResult{
				Name:   "Python",
				Passed: false,
				Error:  "Python is not installed or not found in PATH.",
				Reason: "Python is required to build Python/Flask applications. It is optional for other stacks.",
				Fix:    installGuide("Python", "https://www.python.org/downloads/", "brew install python@3.12", "sudo apt install python3 python3-pip python3-venv"),
			}
		}
		v = v2
	}
	return &CheckResult{
		Name:   "Python",
		Passed: true,
		Value:  strings.TrimSpace(v),
	}
}

// --- PHP -------------------------------------------------------------------

type phpCheck struct{}

func (c *phpCheck) Name() string { return "PHP" }

func (c *phpCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "php", "--version")
	if err != nil {
		return &CheckResult{
			Name:   "PHP",
			Passed: false,
			Error:  "PHP is not installed or not found in PATH.",
			Reason: "PHP is required to build Laravel/PHP applications. It is optional for other stacks.",
			Fix:    installGuide("PHP", "https://www.php.net/downloads", "brew install php", "sudo apt install php-cli php-mbstring php-xml php-curl"),
		}
	}
	return &CheckResult{
		Name:   "PHP",
		Passed: true,
		Value:  parsePHPVersion(v),
	}
}

// --- Composer --------------------------------------------------------------

type composerCheck struct{}

func (c *composerCheck) Name() string { return "Composer" }

func (c *composerCheck) Run(ctx context.Context) *CheckResult {
	v, err := execCmd(ctx, "composer", "--version")
	if err != nil {
		return &CheckResult{
			Name:   "Composer",
			Passed: false,
			Error:  "Composer is not installed or not found in PATH.",
			Reason: "Composer is the PHP dependency manager, required for Laravel applications.",
			Fix:    "Install Composer from https://getcomposer.org/download/\n  macOS: brew install composer\n  Linux: sudo apt install composer",
		}
	}
	return &CheckResult{
		Name:   "Composer",
		Passed: true,
		Value:  parseComposerVersion(v),
	}
}

// --- Ports -----------------------------------------------------------------

type portCheck struct{}

func (c *portCheck) Name() string { return "Ports" }

func (c *portCheck) Run(ctx context.Context) *CheckResult {
	// Check the default CloudOS API port.
	apiPort := 8080
	if addr := os.Getenv("CLOUDOS_API"); addr != "" {
		// Extract port from http://host:port
		if parts := strings.Split(addr, ":"); len(parts) > 0 {
			if p := parts[len(parts)-1]; p != "" {
				fmt.Sscanf(p, "%d", &apiPort)
			}
		}
	}

	ports := []int{apiPort, 3000, 8000, 8080, 9090}
	var busy []int

	for _, port := range ports {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			busy = append(busy, port)
			continue
		}
		ln.Close()
	}

	if len(busy) > 0 {
		portStrs := make([]string, len(busy))
		for i, p := range busy {
			portStrs[i] = fmt.Sprintf("%d", p)
		}
		return &CheckResult{
			Name:   "Ports",
			Passed: false,
			Error:  fmt.Sprintf("Port(s) %s are already in use.", strings.Join(portStrs, ", ")),
			Reason: "CloudOS needs available ports to run the API server and deployed applications.",
			Fix:    "Stop the services using those ports, or set a different API port:\n  export CLOUDOS_API=http://localhost:9090",
		}
	}

	return &CheckResult{
		Name:   "Ports",
		Passed: true,
		Value:  "Required ports are available",
	}
}

// --- Runtime ---------------------------------------------------------------

type runtimeCheck struct{}

func (c *runtimeCheck) Name() string { return "Runtime" }

func (c *runtimeCheck) Run(ctx context.Context) *CheckResult {
	// Check if Docker is available (OCI Runtime)
	dockerOK := false
	if _, err := execCmd(ctx, "docker", "info", "--format", "{{.ServerVersion}}"); err == nil {
		dockerOK = true
	}

	if dockerOK {
		return &CheckResult{
			Name:   "Runtime",
			Passed: true,
			Value:  "OCI Runtime available (Docker)",
		}
	}

	// LocalRuntime is always available (no external dependency).
	return &CheckResult{
		Name:   "Runtime",
		Passed: true,
		Value:  "LocalRuntime available (fallback — no Docker detected)",
	}
}

// --- Buildpacks ------------------------------------------------------------

type buildpackCheck struct{}

func (c *buildpackCheck) Name() string { return "Buildpacks" }

func (c *buildpackCheck) Run(ctx context.Context) *CheckResult {
	var detected []string
	if _, err := execCmd(ctx, "go", "version"); err == nil {
		detected = append(detected, "Go")
	}
	if _, err := execCmd(ctx, "node", "--version"); err == nil {
		detected = append(detected, "Node.js")
	}
	if _, err := execCmd(ctx, "python3", "--version"); err == nil {
		detected = append(detected, "Python")
	}
	if _, err := execCmd(ctx, "php", "--version"); err == nil {
		detected = append(detected, "Laravel")
	}
	detected = append(detected, "Static")
	available := strings.Join(detected, ", ")

	return &CheckResult{
		Name:   "Buildpacks",
		Passed: true,
		Value:  available,
	}
}

// --- Directories -----------------------------------------------------------

type dirCheck struct{}

func (c *dirCheck) Name() string { return "Working Directory" }

func (c *dirCheck) Run(ctx context.Context) *CheckResult {
	// Check current directory is writable.
	cwd, err := os.Getwd()
	if err != nil {
		return &CheckResult{
			Name:   "Working Directory",
			Passed: false,
			Error:  fmt.Sprintf("Cannot determine current directory: %v", err),
			Reason: "CloudOS needs to write build artifacts to the current working directory.",
			Fix:    "Change to a directory where you have write permissions:\n  cd ~/projects",
		}
	}

	tmpFile := filepath.Join(cwd, ".cloudos-tmp")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		return &CheckResult{
			Name:   "Working Directory",
			Passed: false,
			Error:  fmt.Sprintf("Current directory is not writable: %v", err),
			Reason: "CloudOS needs to write build artifacts to the current working directory.",
			Fix:    fmt.Sprintf("Grant write permission:\n  chmod +w %s\nOr change to a writable directory:\n  cd ~/projects", cwd),
		}
	}
	os.Remove(tmpFile)

	// Check temp directory is writable.
	tmpDir := os.TempDir()
	tmpFile2 := filepath.Join(tmpDir, ".cloudos-tmp")
	if err := os.WriteFile(tmpFile2, []byte("test"), 0644); err != nil {
		return &CheckResult{
			Name:   "Working Directory",
			Passed: false,
			Error:  fmt.Sprintf("System temp directory is not writable: %v", err),
			Reason: "CloudOS uses the system temp directory for intermediate build files.",
			Fix:    fmt.Sprintf("Check permissions on %s and ensure it is writable.", tmpDir),
		}
	}
	os.Remove(tmpFile2)

	return &CheckResult{
		Name:   "Working Directory",
		Passed: true,
		Value:  "Directory is writable",
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// Helpers
// ═════════════════════════════════════════════════════════════════════════════

// execCmd runs a command and returns its stdout as a string.
func execCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// parseVersion extracts a short version string from tool output.
func parseVersion(out string) string {
	// go version go1.24.4 linux/amd64 → 1.24.4
	parts := strings.Fields(out)
	if len(parts) >= 3 {
		return strings.TrimPrefix(parts[2], "go")
	}
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "go")
	}
	return out
}

// parsePHPVersion extracts the PHP version from php --version output.
func parsePHPVersion(out string) string {
	// PHP 8.3.0 (cli) ... → 8.3.0
	lines := strings.SplitN(out, "\n", 2)
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return out
}

// parseComposerVersion extracts the Composer version.
func parseComposerVersion(out string) string {
	// Composer version 2.7.0 ... → 2.7.0
	parts := strings.Fields(out)
	if len(parts) >= 3 {
		return parts[2]
	}
	return out
}

// installGuide returns a platform-appropriate installation guide message.
func installGuide(tool, url, brewCmd, aptCmd string) string {
	var guide string
	switch runtime.GOOS {
	case "darwin":
		guide = fmt.Sprintf("  macOS: %s", brewCmd)
	case "linux":
		guide = fmt.Sprintf("  Linux: %s", aptCmd)
	default:
		guide = fmt.Sprintf("  Download from: %s", url)
	}

	return fmt.Sprintf("Install %s:\n%s\n  Then run:\n    cloudosctl doctor", tool, guide)
}

// Ensure doctor is not elided by the compiler.
var _ = doctor
