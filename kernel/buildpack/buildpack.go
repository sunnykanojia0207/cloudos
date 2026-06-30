// Package buildpack provides automatic runtime detection for CloudOS
// applications. It inspects a project directory for well-known files
// (package.json, go.mod, requirements.txt, etc.) and determines the
// runtime type, build commands, install commands, and output directory.
//
// This enables CloudOS to deploy any project without manual configuration —
// just point it at a repository and it figures out how to build and run it.
//
// Detection hierarchy (first match wins):
//
//	1. Dockerfile          → RuntimeDocker
//	2. package.json        → RuntimeNode / RuntimeNextJS / RuntimeReact
//	3. go.mod              → RuntimeGo
//	4. requirements.txt    → RuntimePython
//	5. setup.py / setup.cfg → RuntimePython
//	6. composer.json       → RuntimeLaravel
//	7. index.html          → RuntimeStatic
//	8. fallback            → RuntimeStatic
package buildpack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ── Buildpack Interface ────────────────────────────────────────────────────

// Buildpack is the interface for automatic runtime detection and build planning.
//
// Each Buildpack knows how to detect if it applies to a project directory,
// and if so, how to plan the build and runtime configuration.
//
// Architecture:
//
//	BuildpackRegistry
//	    ├── DockerBuildpack
//	    ├── NodeBuildpack    (Next.js, React, generic Node)
//	    ├── GoBuildpack
//	    ├── PythonBuildpack
//	    ├── LaravelBuildpack
//	    └── StaticBuildpack (fallback)
//
// The registry iterates buildpacks in priority order and returns the first match.
type Buildpack interface {
	// Name returns the buildpack identifier (e.g. "node", "go", "python").
	Name() string

	// Detect checks if this buildpack applies to the project directory.
	// Returns true if the project matches this buildpack's criteria.
	Detect(dir string) (bool, error)

	// Plan generates the build and runtime configuration for the project.
	// Called only if Detect returned true.
	Plan(dir string) (*BuildPlan, error)
}

// BuildPlan contains the build and runtime configuration produced by a Buildpack.
type BuildPlan struct {
	// RuntimeType is the runtime type string (e.g. "node", "go", "python").
	RuntimeType string `json:"runtimeType"`

	// Name is a human-readable name (e.g. "Node.js", "Go", "Python").
	Name string `json:"name"`

	// Version is the detected version (if available).
	Version string `json:"version,omitempty"`

	// InstallCmd is the command to install dependencies.
	InstallCmd string `json:"installCmd,omitempty"`

	// BuildCmd is the command to build the application.
	// Empty means no build step is needed.
	BuildCmd string `json:"buildCmd,omitempty"`

	// OutputDir is the directory containing build output (e.g. "build", "dist").
	// Empty means the project root is the output.
	OutputDir string `json:"outputDir,omitempty"`

	// StartCmd is the command to start the application.
	StartCmd string `json:"startCmd,omitempty"`

	// DevPort is the default port the application listens on.
	DevPort int `json:"devPort,omitempty"`

	// EnvVars are recommended environment variables for the runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`
}

// ── Buildpack Registry ─────────────────────────────────────────────────────

// BuildpackRegistry holds ordered buildpacks and provides detection.
type BuildpackRegistry struct {
	buildpacks []Buildpack
}

// NewBuildpackRegistry creates a registry with the default buildpacks.
func NewBuildpackRegistry() *BuildpackRegistry {
	return &BuildpackRegistry{
		buildpacks: DefaultBuildpacks(),
	}
}

// Register adds a buildpack to the end of the registry.
func (r *BuildpackRegistry) Register(bp Buildpack) {
	r.buildpacks = append(r.buildpacks, bp)
}

// Detect iterates registered buildpacks in order and returns the first match.
func (r *BuildpackRegistry) Detect(dir string) (*BuildPlan, error) {
	for _, bp := range r.buildpacks {
		ok, err := bp.Detect(dir)
		if err != nil {
			continue // skip broken buildpacks
		}
		if ok {
			return bp.Plan(dir)
		}
	}
	// Fallback to static.
	return (&StaticBuildpack{}).Plan(dir)
}

// DefaultBuildpacks returns the standard buildpack chain in priority order.
func DefaultBuildpacks() []Buildpack {
	return []Buildpack{
		&DockerBuildpack{},
		&NodeBuildpack{},
		&GoBuildpack{},
		&PythonBuildpack{},
		&LaravelBuildpack{},
		&StaticBuildpack{},
	}
}

// ── DetectedRuntime (backward compat) ──────────────────────────────────────

// DetectedRuntime describes the runtime detected from a project directory.
// This type is kept for backward compatibility; new code should use BuildPlan.
type DetectedRuntime struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Version    string            `json:"version,omitempty"`
	InstallCmd string            `json:"installCmd,omitempty"`
	BuildCmd   string            `json:"buildCmd,omitempty"`
	OutputDir  string            `json:"outputDir,omitempty"`
	StartCmd   string            `json:"startCmd,omitempty"`
	DevPort    int               `json:"devPort,omitempty"`
	EnvVars    map[string]string `json:"envVars,omitempty"`
}

// ── Runtime type constants ────────────────────────────────────────────────

const (
	RuntimeNode    = "node"
	RuntimeReact   = "react"
	RuntimeNextJS  = "nextjs"
	RuntimeGo      = "go"
	RuntimePython  = "python"
	RuntimeLaravel = "laravel"
	RuntimeStatic  = "static"
	RuntimeDocker  = "docker"
	RuntimeGeneric = "generic"
)

// ── Package.json structures ────────────────────────────────────────────────

// packageJSON represents the structure of a Node.js package.json file.
type packageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Scripts struct {
		Build string `json:"build,omitempty"`
		Start string `json:"start,omitempty"`
		Dev   string `json:"dev,omitempty"`
	} `json:"scripts,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

// composerJSON represents the structure of a PHP composer.json file.
type composerJSON struct {
	Name    string `json:"name"`
	Require map[string]string `json:"require,omitempty"`
}

// Detect inspects a project directory and returns the detected runtime.
// It walks well-known files in order of specificity and returns the first match.
//
// Returns a DetectedRuntime with all commands and settings, or an error
// if the directory cannot be read.
func Detect(projectDir string) (*DetectedRuntime, error) {
	// Use the BuildpackRegistry for detection.
	registry := NewBuildpackRegistry()
	plan, err := registry.Detect(projectDir)
	if err != nil {
		return nil, err
	}
	return planToDetected(plan), nil
}

// planToDetected converts a BuildPlan to a DetectedRuntime for backward compat.
func planToDetected(plan *BuildPlan) *DetectedRuntime {
	return &DetectedRuntime{
		Type:       plan.RuntimeType,
		Name:       plan.Name,
		Version:    plan.Version,
		InstallCmd: plan.InstallCmd,
		BuildCmd:   plan.BuildCmd,
		OutputDir:  plan.OutputDir,
		StartCmd:   plan.StartCmd,
		DevPort:    plan.DevPort,
		EnvVars:    plan.EnvVars,
	}
}

// ── Buildpack Implementations ─────────────────────────────────────────────

// ── Node Buildpack ─────────────────────────────────────────────────────────

// NodeBuildpack detects Node.js, Next.js, and React projects.
type NodeBuildpack struct{}

func (bp *NodeBuildpack) Name() string { return "node" }

func (bp *NodeBuildpack) Detect(dir string) (bool, error) {
	return hasFile(dir, "package.json"), nil
}

func (bp *NodeBuildpack) Plan(dir string) (*BuildPlan, error) {
	return planNode(dir)
}

// ── Go Buildpack ───────────────────────────────────────────────────────────

// GoBuildpack detects Go projects.
type GoBuildpack struct{}

func (bp *GoBuildpack) Name() string { return "go" }

func (bp *GoBuildpack) Detect(dir string) (bool, error) {
	return hasFile(dir, "go.mod"), nil
}

func (bp *GoBuildpack) Plan(dir string) (*BuildPlan, error) {
	return planGo(dir), nil
}

// ── Python Buildpack ───────────────────────────────────────────────────────

// PythonBuildpack detects Python projects.
type PythonBuildpack struct{}

func (bp *PythonBuildpack) Name() string { return "python" }

func (bp *PythonBuildpack) Detect(dir string) (bool, error) {
	return hasFile(dir, "requirements.txt") ||
		hasFile(dir, "setup.py") ||
		hasFile(dir, "setup.cfg") ||
		hasFile(dir, "Pipfile"), nil
}

func (bp *PythonBuildpack) Plan(dir string) (*BuildPlan, error) {
	return planPython(dir), nil
}

// ── Laravel Buildpack ──────────────────────────────────────────────────────

// LaravelBuildpack detects PHP/Laravel projects.
type LaravelBuildpack struct{}

func (bp *LaravelBuildpack) Name() string { return "laravel" }

func (bp *LaravelBuildpack) Detect(dir string) (bool, error) {
	return hasFile(dir, "composer.json"), nil
}

func (bp *LaravelBuildpack) Plan(dir string) (*BuildPlan, error) {
	return planLaravel(dir)
}

// ── Docker Buildpack ───────────────────────────────────────────────────────

// DockerBuildpack detects Dockerfile-based projects.
type DockerBuildpack struct{}

func (bp *DockerBuildpack) Name() string { return "docker" }

func (bp *DockerBuildpack) Detect(dir string) (bool, error) {
	return hasFile(dir, "Dockerfile"), nil
}

func (bp *DockerBuildpack) Plan(dir string) (*BuildPlan, error) {
	return planDocker(dir), nil
}

// ── Static Buildpack ───────────────────────────────────────────────────────

// StaticBuildpack is the fallback buildpack for static HTML sites.
type StaticBuildpack struct{}

func (bp *StaticBuildpack) Name() string { return "static" }

func (bp *StaticBuildpack) Detect(dir string) (bool, error) {
	// Always returns true as the fallback.
	return true, nil
}

func (bp *StaticBuildpack) Plan(dir string) (*BuildPlan, error) {
	return planStatic(dir), nil
}

// ── Internal Planning Functions ─────────────────────────────────────────────

// planNode plans the build for Node.js-based projects (Next.js, React, generic).
func planNode(projectDir string) (*BuildPlan, error) {
	pkg, err := readPackageJSON(projectDir)
	if err != nil {
		return &BuildPlan{
			RuntimeType: RuntimeNode,
			Name:        "Node.js",
			InstallCmd:  "npm install",
			BuildCmd:    "",
			StartCmd:    "npm start",
			DevPort:     3000,
		}, nil
	}

	if isNextJS(pkg) {
		return &BuildPlan{
			RuntimeType: RuntimeNextJS,
			Name:        "Next.js",
			Version:     pkg.Version,
			InstallCmd:  "npm install",
			BuildCmd:    pkg.Scripts.Build,
			StartCmd:    "npm start",
			OutputDir:   ".next",
			DevPort:     3000,
		}, nil
	}

	if isReact(pkg) {
		return &BuildPlan{
			RuntimeType: RuntimeReact,
			Name:        "React",
			Version:     pkg.Version,
			InstallCmd:  "npm install",
			BuildCmd:    pkg.Scripts.Build,
			StartCmd:    "npx serve -s build -l {port}",
			OutputDir:   "build",
			DevPort:     3000,
		}, nil
	}

	buildCmd := pkg.Scripts.Build
	startCmd := pkg.Scripts.Start
	if startCmd == "" {
		startCmd = "npm start"
	}

	return &BuildPlan{
		RuntimeType: RuntimeNode,
		Name:        "Node.js",
		Version:     pkg.Version,
		InstallCmd:  "npm install",
		BuildCmd:    buildCmd,
		StartCmd:    startCmd,
		OutputDir:   "",
		DevPort:     defaultNodePort(pkg),
	}, nil
}

func planGo(projectDir string) *BuildPlan {
	return &BuildPlan{
		RuntimeType: RuntimeGo,
		Name:        "Go",
		InstallCmd:  "go mod download",
		BuildCmd:    "go build -o app .",
		StartCmd:    "./app",
		OutputDir:   "",
		DevPort:     8080,
	}
}

func planPython(projectDir string) *BuildPlan {
	installCmd := "pip install -r requirements.txt"
	if hasFile(projectDir, "Pipfile") {
		installCmd = "pipenv install"
	}

	startCmd := "python app.py"
	if hasFile(projectDir, "manage.py") {
		startCmd = "python manage.py runserver 0.0.0.0:{port}"
	} else if hasFile(projectDir, "wsgi.py") {
		startCmd = "gunicorn wsgi:app --bind 0.0.0.0:{port}"
	} else if hasFile(projectDir, "app.py") {
		startCmd = "python app.py"
	} else if hasFile(projectDir, "main.py") {
		startCmd = "python main.py"
	}

	return &BuildPlan{
		RuntimeType: RuntimePython,
		Name:        "Python",
		InstallCmd:  installCmd,
		BuildCmd:    "",
		StartCmd:    startCmd,
		OutputDir:   "",
		DevPort:     8000,
		EnvVars: map[string]string{
			"PYTHONUNBUFFERED": "1",
		},
	}
}

func planLaravel(projectDir string) (*BuildPlan, error) {
	if _, err := readComposerJSON(projectDir); err != nil {
		return &BuildPlan{
			RuntimeType: RuntimeLaravel,
			Name:        "PHP",
			InstallCmd:  "composer install",
			BuildCmd:    "",
			StartCmd:    "php artisan serve --host=0.0.0.0 --port={port}",
			DevPort:     8000,
		}, nil
	}

	installCmd := "composer install"
	startCmd := "php artisan serve --host=0.0.0.0 --port={port}"

	if hasFile(projectDir, "artisan") {
		return &BuildPlan{
			RuntimeType: RuntimeLaravel,
			Name:        "Laravel",
			InstallCmd:  installCmd,
			BuildCmd:    "",
			StartCmd:    startCmd,
			OutputDir:   "public",
			DevPort:     8000,
			EnvVars: map[string]string{
				"APP_ENV": "local",
			},
		}, nil
	}

	return &BuildPlan{
		RuntimeType: RuntimeLaravel,
		Name:        "PHP",
		InstallCmd:  installCmd,
		BuildCmd:    "",
		StartCmd:    startCmd,
		DevPort:     8000,
	}, nil
}

func planStatic(projectDir string) *BuildPlan {
	return &BuildPlan{
		RuntimeType: RuntimeStatic,
		Name:        "Static Website",
		InstallCmd:  "",
		BuildCmd:    "",
		StartCmd:    "",
		OutputDir:   "",
		DevPort:     80,
	}
}

func planDocker(projectDir string) *BuildPlan {
	return &BuildPlan{
		RuntimeType: RuntimeDocker,
		Name:        "Docker",
		InstallCmd:  "",
		BuildCmd:    "docker build -t {app} .",
		StartCmd:   "docker run -p {port}:{port} {app}",
		DevPort:    0, // Port defined by Dockerfile EXPOSE
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────

// hasFile checks if a file exists in the project directory.
func hasFile(dir, name string) bool {
	path := filepath.Join(dir, name)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// hasDir checks if a directory exists.
func hasDir(dir, name string) bool {
	path := filepath.Join(dir, name)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// readPackageJSON reads and parses a package.json file.
func readPackageJSON(dir string) (*packageJSON, error) {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// readComposerJSON reads and parses a composer.json file.
func readComposerJSON(dir string) (*composerJSON, error) {
	path := filepath.Join(dir, "composer.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var composer composerJSON
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}
	return &composer, nil
}

// isNextJS checks if a package.json indicates a Next.js project.
func isNextJS(pkg *packageJSON) bool {
	if _, ok := pkg.Dependencies["next"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["next"]; ok {
		return true
	}
	return false
}

// isReact checks if a package.json indicates a React project.
func isReact(pkg *packageJSON) bool {
	// Check for common React frameworks.
	hasReact := false
	if _, ok := pkg.Dependencies["react"]; ok {
		hasReact = true
	}
	if _, ok := pkg.DevDependencies["react"]; ok {
		hasReact = true
	}
	if !hasReact {
		return false
	}

	// Common CRA or Vite React projects have react-scripts or vite.
	if _, ok := pkg.Dependencies["react-scripts"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["vite"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["@vitejs/plugin-react"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["react-scripts"]; ok {
		return true
	}

	// If react is a dependency but no build tool is detected, it's likely
	// a custom React project. Check for a build script.
	if pkg.Scripts.Build != "" {
		return true
	}

	return false
}

// defaultNodePort returns the default port for a Node.js project based on
// common conventions.
func defaultNodePort(pkg *packageJSON) int {
	// Express projects often use a PORT env var, defaulting to 3000.
	return 3000
}

// StartCommandWithPort replaces {port} placeholders in the start command
// with the actual port number.
func StartCommandWithPort(cmd string, port int) string {
	portStr := fmtPort(port)
	return strings.ReplaceAll(cmd, "{port}", portStr)
}

// fmtPort converts a port number to a string.
func fmtPort(port int) string {
	if port <= 0 {
		return "8080"
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(
		func() string { return fmtInt(port) }(),
		"\n", ""), " ", ""))
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
