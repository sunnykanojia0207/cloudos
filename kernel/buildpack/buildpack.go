// Package buildpack provides automatic detection, planning, and building of
// CloudOS applications. It inspects a project directory for well-known files
// (package.json, go.mod, requirements.txt, etc.), produces a build plan, and
// executes the build to produce an Artifact.
//
// A Buildpack Engine orchestrates the pipeline:
//
//	Source
//	   │
//	   ▼
//	Engine.Detect()   ←  asks each buildpack: "Can you build this?"
//	   │
//	   ▼
//	Engine.Plan()     ←  produces a BuildPlan (commands, output dir, env)
//	   │
//	   ▼
//	Engine.Build()    ←  executes the plan, produces an Artifact
//	   │
//	   ▼
//	Artifact
//	   │
//	   ▼
//	Runtime.Prepare() → Runtime.Start()
//
// The Engine knows nothing about individual buildpacks. It just iterates
// through registered buildpacks in priority order until one says "yes".
//
// Detection priority (first match wins):
//
//	1. Go          (go.mod)
//	2. Next.js     (package.json + "next" dependency)
//	3. React       (package.json + "react" + build tool)
//	4. Node.js     (package.json — generic)
//	5. Laravel     (composer.json)
//	6. Python      (requirements.txt / setup.py / Pipfile)
//	7. Docker      (Dockerfile)
//	8. Static      (always matches — fallback)
package buildpack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── API Version ─────────────────────────────────────────────────────────────
//
// BuildpackAPIVersion is the frozen version of the Buildpack interface.
// Per ADR-0011, this contract is declared v1.0 and will only receive
// additive extensions via optional interfaces.
//
// Every Buildpack implementation declares which API version it implements.
const BuildpackAPIVersion = "buildpack.cloudos.io/v1"

// ── Source ──────────────────────────────────────────────────────────────────

// Source describes where the application source code lives.
type Source struct {
	// Path is the local filesystem path to the source code.
	Path string

	// GitURL is the optional git repository URL the source was cloned from.
	GitURL string

	// Branch is the git branch (if applicable).
	Branch string

	// Commit is the specific commit hash (if applicable).
	Commit string
}

// ── Diagnostic ──────────────────────────────────────────────────────────────

// Diagnostic provides detailed information about a build step.
type Diagnostic struct {
	// Step identifies the build step (e.g. "detect", "install", "build").
	Step string `json:"step"`

	// Message is a human-readable description.
	Message string `json:"message"`

	// Duration is how long the step took.
	Duration string `json:"duration,omitempty"`

	// Level indicates severity: "info", "warning", "error".
	Level string `json:"level,omitempty"`
}

// ── BuildResult ─────────────────────────────────────────────────────────────

// BuildResult wraps the output of a Buildpack.Build() call with metadata,
// warnings, and diagnostics for observability and AI analysis.
type BuildResult struct {
	// Artifact is the built application artifact.
	Artifact *Artifact `json:"artifact"`

	// RuntimeType indicates which runtime type this artifact needs.
	RuntimeType string `json:"runtimeType"`

	// Metadata contains structured key-value data about the build.
	// Examples: {"build_time": "8.2s", "bundle_size": "420KB", "node_version": "22"}
	Metadata map[string]string `json:"metadata,omitempty"`

	// Warnings are non-fatal issues discovered during the build.
	Warnings []string `json:"warnings,omitempty"`

	// Diagnostics provide detailed per-step build information.
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// ── Buildpack Interface ─────────────────────────────────────────────────────

// Buildpack is the interface for automatic detection, planning, and building
// of CloudOS applications.
//
// Each Buildpack knows three things:
//  1. Can it build this source?           → Detect()
//  2. How should it be built?             → Plan()
//  3. Execute the build, produce output   → Build()
//
// Architecture:
//
//	Engine
//	  ├── StaticBuildpack  (fallback — always matches)
//	  ├── GoBuildpack      (go.mod)
//	  ├── NextJSBuildpack  (package.json + "next")
//	  ├── ReactBuildpack   (package.json + "react" + build tool)
//	  ├── NodeBuildpack    (package.json — generic)
//	  ├── LaravelBuildpack (composer.json)
//	  ├── PythonBuildpack  (requirements.txt / setup.py / Pipfile)
//	  └── DockerBuildpack  (Dockerfile)
type Buildpack interface {
	// Name returns the buildpack identifier (e.g. "go", "node", "python").
	Name() string

	// Version returns the buildpack version.
	Version() string

	// Detect checks if this buildpack applies to the source.
	// Returns true if the source matches this buildpack's criteria.
	Detect(ctx context.Context, src Source) (bool, error)

	// Plan generates the build and runtime configuration for the source.
	// Called only if Detect returned true.
	Plan(ctx context.Context, src Source) (*BuildPlan, error)

	// Build executes the build plan and produces a BuildResult containing
	// the artifact plus metadata, warnings, and diagnostics.
	Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error)
}

// ── BuildPlan ───────────────────────────────────────────────────────────────

// BuildPlan contains the build and runtime configuration produced by a Buildpack.
type BuildPlan struct {
	// BuildpackName identifies which buildpack produced this plan.
	BuildpackName string `json:"buildpackName"`

	// RuntimeType is the runtime type string (e.g. "node", "go", "python").
	RuntimeType string `json:"runtimeType"`

	// Name is a human-readable name (e.g. "Node.js", "Go", "Python").
	Name string `json:"name"`

	// Version is the detected version (if available).
	Version string `json:"version,omitempty"`

	// ArtifactType indicates what kind of artifact this build produces.
	ArtifactType ArtifactType `json:"artifactType"`

	// InstallCmd is the command to install dependencies.
	InstallCmd string `json:"installCmd,omitempty"`

	// BuildCmd is the command to build the application.
	// Empty means no build step is needed.
	BuildCmd string `json:"buildCmd,omitempty"`

	// OutputDir is the directory containing build output (e.g. "build", "dist").
	// Empty means the source root is the output.
	OutputDir string `json:"outputDir,omitempty"`

	// StartCmd is the command to start the application.
	StartCmd string `json:"startCmd,omitempty"`

	// DevPort is the default port the application listens on.
	DevPort int `json:"devPort,omitempty"`

	// EnvVars are recommended environment variables for the runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Source is the source this plan was created from.
	Source Source `json:"-"`
}

// ── Artifact ────────────────────────────────────────────────────────────────

// ArtifactType categorizes build artifacts.
type ArtifactType string

const (
	ArtifactTypeBinary  ArtifactType = "binary"  // Compiled binary (Go, Rust, C)
	ArtifactTypeStatic  ArtifactType = "static"  // Static files (HTML, JS, CSS)
	ArtifactTypeSource  ArtifactType = "source"  // Source directory (Python, PHP, Node)
	ArtifactTypeImage   ArtifactType = "image"   // Container image (Docker)
	ArtifactTypeArchive ArtifactType = "archive" // Tarball / zip
)

// Artifact is the output of a Buildpack.Build() call.
// It represents a built application ready for Runtime execution.
//
// Artifact is designed to parallel the Resource pattern:
//
//	kind: Artifact
//	metadata:
//	  id: "artifact-..."
//	  labels: ...
//	spec:
//	  type: "static"
//	  path: "/path/to/dist"
//	  startCmd: "npx serve -s ."
type Artifact struct {
	// ID is a unique identifier for this artifact.
	ID string `json:"id"`

	// Type indicates the artifact type.
	Type ArtifactType `json:"type"`

	// BuildpackName identifies which buildpack produced this artifact.
	BuildpackName string `json:"buildpackName"`

	// RuntimeType indicates which runtime should execute this artifact.
	RuntimeType string `json:"runtimeType"`

	// Path is the filesystem path to the artifact.
	Path string `json:"path"`

	// StartCmd is the command to start the application from this artifact.
	StartCmd string `json:"startCmd"`

	// DevPort is the default port the application listens on.
	DevPort int `json:"devPort"`

	// EnvVars are recommended environment variables.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Source describes the source code this artifact was built from.
	Source Source `json:"source,omitempty"`

	// CreatedAt is when this artifact was produced.
	CreatedAt time.Time `json:"createdAt"`

	// Metadata contains additional metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ── Buildpack Engine ────────────────────────────────────────────────────────

// Engine orchestrates the buildpack pipeline: Detect → Plan → Build.
//
// Usage:
//
//	engine := buildpack.NewEngine()
//	bp, err := engine.Detect(ctx, src)      // which buildpack?
//	plan, err := engine.Plan(ctx, src, bp)  // how to build?
//	artifact, err := engine.Build(ctx, plan) // do it!
//
// Or in one call:
//
//	artifact, err := engine.Run(ctx, src) // detect + plan + build
type Engine struct {
	buildpacks []Buildpack
}

// NewEngine creates a buildpack Engine with the default buildpacks.
func NewEngine() *Engine {
	return &Engine{
		buildpacks: DefaultBuildpacks(),
	}
}

// Register adds a buildpack to the end of the detection chain.
func (e *Engine) Register(bp Buildpack) {
	e.buildpacks = append(e.buildpacks, bp)
}

// Buildpacks returns the registered buildpacks.
func (e *Engine) Buildpacks() []Buildpack {
	result := make([]Buildpack, len(e.buildpacks))
	copy(result, e.buildpacks)
	return result
}

// Detect iterates registered buildpacks in priority order and returns the
// first one that matches the source. Returns nil if no buildpack matches.
//
// This is a lightweight operation — each buildpack just checks for the
// presence of well-known files.
func (e *Engine) Detect(ctx context.Context, src Source) (Buildpack, error) {
	for _, bp := range e.buildpacks {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		ok, err := bp.Detect(ctx, src)
		if err != nil {
			continue // skip broken buildpacks
		}
		if ok {
			return bp, nil
		}
	}
	// StaticBuildpack should always match, but if it's not registered,
	// return nil instead of panicking.
	return nil, fmt.Errorf("no buildpack matched for source %q", src.Path)
}

// Plan generates a BuildPlan from a source using the given buildpack.
// The buildpack must have been returned by Detect().
func (e *Engine) Plan(ctx context.Context, src Source, bp Buildpack) (*BuildPlan, error) {
	plan, err := bp.Plan(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("buildpack %q plan: %w", bp.Name(), err)
	}
	plan.Source = src
	plan.BuildpackName = bp.Name()
	return plan, nil
}

// Build executes a BuildPlan and returns a BuildResult containing the
// produced Artifact plus metadata, warnings, and diagnostics.
func (e *Engine) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	// Find the buildpack that created this plan.
	var bp Buildpack
	for _, candidate := range e.buildpacks {
		if candidate.Name() == plan.BuildpackName {
			bp = candidate
			break
		}
	}
	if bp == nil {
		return nil, fmt.Errorf("buildpack %q not found for build", plan.BuildpackName)
	}

	result, err := bp.Build(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("buildpack %q build: %w", bp.Name(), err)
	}

	// Ensure artifact has all the metadata from the plan.
	if result.Artifact == nil {
		return nil, fmt.Errorf("buildpack %q returned nil artifact", bp.Name())
	}
	if result.Artifact.ID == "" {
		result.Artifact.ID = fmt.Sprintf("artifact-%d", time.Now().UnixNano())
	}
	if result.Artifact.BuildpackName == "" {
		result.Artifact.BuildpackName = bp.Name()
	}
	if result.Artifact.RuntimeType == "" {
		result.Artifact.RuntimeType = plan.RuntimeType
	}
	if result.Artifact.CreatedAt.IsZero() {
		result.Artifact.CreatedAt = time.Now()
	}
	result.Artifact.Source = plan.Source
	if result.RuntimeType == "" {
		result.RuntimeType = plan.RuntimeType
	}

	return result, nil
}

// Run is a convenience method that combines Detect + Plan + Build into
// a single call. Returns the BuildResult or an error at any stage.
func (e *Engine) Run(ctx context.Context, src Source) (*BuildResult, error) {
	bp, err := e.Detect(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("detect: %w", err)
	}

	plan, err := e.Plan(ctx, src, bp)
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}

	result, err := e.Build(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}

	return result, nil
}

// ── Default Buildpacks ─────────────────────────────────────────────────────

// DefaultBuildpacks returns the standard buildpack chain in priority order.
//
// Note: StaticBuildpack is the universal fallback and must be last —
// its Detect() always returns true.
func DefaultBuildpacks() []Buildpack {
	return []Buildpack{
		&GoBuildpack{},      // 1. go.mod
		&NextJSBuildpack{},  // 2. package.json + "next"
		&ReactBuildpack{},   // 3. package.json + react + build tool
		&NodeBuildpack{},    // 4. package.json (generic Node.js)
		&LaravelBuildpack{}, // 5. composer.json
		&PythonBuildpack{},  // 6. requirements.txt / setup.py / Pipfile
		&DockerBuildpack{},  // 7. Dockerfile
		&StaticBuildpack{},  // 8. Always matches (fallback)
	}
}

// ── Backward Compat ─────────────────────────────────────────────────────────

// DetectedRuntime describes the runtime detected from a project directory.
// Deprecated: Use Engine.Detect() + Engine.Plan() instead.
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

// Runtime type constants (kept for backward compat).
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

// Detect inspects a project directory and returns the detected runtime.
// Deprecated: Use NewEngine().Run() or Engine.Detect() + Plan().
func Detect(projectDir string) (*DetectedRuntime, error) {
	engine := NewEngine()
	src := Source{Path: projectDir}
	bp, err := engine.Detect(context.Background(), src)
	if err != nil {
		return nil, err
	}
	plan, err := engine.Plan(context.Background(), src, bp)
	if err != nil {
		return nil, err
	}
	return planToDetected(plan), nil
}

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

// ── Runtime type constants (new) ───────────────────────────────────────────

// ArtifactFromPlan creates an Artifact from a BuildPlan and a build output path.
// This is a helper for Buildpack implementations.
func ArtifactFromPlan(plan *BuildPlan, outputPath string) *Artifact {
	return &Artifact{
		ID:            fmt.Sprintf("artifact-%d", time.Now().UnixNano()),
		Type:          plan.ArtifactType,
		BuildpackName: plan.BuildpackName,
		RuntimeType:   plan.RuntimeType,
		Path:          outputPath,
		StartCmd:      plan.StartCmd,
		DevPort:       plan.DevPort,
		EnvVars:       plan.EnvVars,
		Source:        plan.Source,
		CreatedAt:     time.Now(),
	}
}

// ── Shared Helpers ─────────────────────────────────────────────────────────

// hasFile checks if a file exists in the directory.
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

// readFile reads a file from the source directory.
func readFile(src Source, name string) ([]byte, error) {
	path := filepath.Join(src.Path, name)
	return os.ReadFile(path)
}

// fileExists checks if a file exists in the source directory.
func fileExists(src Source, name string) bool {
	return hasFile(src.Path, name)
}

// getVersion safely extracts a version string from a PackageJSON pointer.
func getVersion(pkg *PackageJSON) string {
	if pkg != nil {
		return pkg.Version
	}
	return ""
}

// StartCommandWithPort replaces {port} placeholders in the start command
// with the actual port number.
func StartCommandWithPort(cmd string, port int) string {
	portStr := fmt.Sprintf("%d", port)
	return strings.ReplaceAll(cmd, "{port}", portStr)
}

// ── Package.json structures (shared by Node, React, Next.js) ────────────────

// PackageJSON represents the structure of a Node.js package.json file.
type PackageJSON struct {
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

// readPackageJSON reads and parses a package.json file.
func readPackageJSON(src Source) (*PackageJSON, error) {
	data, err := readFile(src, "package.json")
	if err != nil {
		return nil, err
	}
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// ── Composer.json structures (shared by Laravel) ────────────────────────────

// ComposerJSON represents the structure of a PHP composer.json file.
type ComposerJSON struct {
	Name    string            `json:"name"`
	Require map[string]string `json:"require,omitempty"`
}

// readComposerJSON reads and parses a composer.json file.
func readComposerJSON(src Source) (*ComposerJSON, error) {
	data, err := readFile(src, "composer.json")
	if err != nil {
		return nil, err
	}
	var composer ComposerJSON
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}
	return &composer, nil
}
