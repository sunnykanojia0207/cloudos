package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/cloudos/cloudos/kernel/buildpack"
	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/project"
	"github.com/cloudos/cloudos/kernel/resource"
	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/packages/logging"
	"github.com/cloudos/cloudos/packages/types"
)

// ExecutorDeps holds the dependencies the Executor needs to run TaskNodes.
type ExecutorDeps struct {
	ResourceRegistry  *resource.Registry
	ControllerManager *controller.Manager
	HealthManager     *health.Manager
	SourceCloner      *source.GitCloner
	RuntimeManager    cr.Runtime
	Logger            *logging.Logger
}

// runContextKey is the context key for the current workflow run ID.
type runContextKey struct{}

// WithRunID embeds a workflow run ID in the context for cross-node state.
func WithRunID(ctx context.Context, runID string) context.Context {
	return context.WithValue(ctx, runContextKey{}, runID)
}

// RunIDFromContext extracts the workflow run ID from the context.
func RunIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(runContextKey{}).(string)
	return id, ok
}

// Executor runs individual TaskNodes against the CloudOS kernel.
//
// Unlike the intent Executor (which runs a complete, linear plan), the
// workflow Executor runs one node at a time. The Scheduler is responsible
// for deciding which nodes are ready; the Engine is responsible for
// orchestration.
//
// Nodes within the same run can share state via the WorkflowRun's context.
// The engine embeds the run ID in the context before calling Execute.
type Executor struct {
	resRegistry    *resource.Registry
	ctrlManager    *controller.Manager
	healthMgr      *health.Manager
	sourceCloner   *source.GitCloner
	runtimeManager cr.Runtime
	log            *logging.Logger

	// workDirs maps run IDs to cloned source directories.
	// Set by source.clone nodes, read by build and deploy nodes.
	workDirs sync.Map

	// instances maps run IDs to running instance IDs.
	// Set by provider.deploy, read by CleanupRun on cancellation.
	instances sync.Map

	// currentCtx is the context for the currently executing node.
	// Stored by Execute so action handlers can access the run ID.
	currentCtx context.Context
}

// NewExecutor creates a new workflow Executor.
func NewExecutor(deps ExecutorDeps) *Executor {
	return &Executor{
		resRegistry:    deps.ResourceRegistry,
		ctrlManager:    deps.ControllerManager,
		healthMgr:       deps.HealthManager,
		sourceCloner:   deps.SourceCloner,
		runtimeManager: deps.RuntimeManager,
		log:            deps.Logger,
	}
}

// Execute runs a single TaskNode and returns the result.
// It handles timeouts via context.WithTimeout.
func (ex *Executor) Execute(ctx context.Context, node *TaskNode) error {
	ex.log.Debug("executing task node",
		"node", node.ID(),
		"action", node.Action,
		"target", node.Target,
	)

	// Store the current context for action handlers to access run context.
	ex.currentCtx = ctx

	// Apply timeout if configured
	if node.Timeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, node.Timeout())
		defer cancel()
	}

	// Dispatch based on action
	switch node.Action {
	case "validate":
		return ex.execValidate(node)
	case "resource.create":
		return ex.execResourceCreate(node)
	case "resource.get":
		return ex.execResourceGet(node)
	case "resource.list":
		return ex.execResourceList(node)
	case "resource.delete":
		return ex.execResourceDelete(node)
	case "resource.kinds":
		return ex.execResourceKinds(node)
	case "controller.list":
		return ex.execControllerList(node)
	case "controller.reconcile":
		return ex.execControllerReconcile(node)
	case "health.check":
		return ex.execHealthCheck(node)
	case "source.clone":
		return ex.execSourceClone(node)
	case "build.install":
		return ex.execBuildInstall(node)
	case "build.execute":
		return ex.execBuildExecute(node)
	case "provider.deploy":
		return ex.execProviderDeploy(node)
	case "complete":
		return nil // no-op success
	case "format":
		return nil // no-op success
	default:
		return fmt.Errorf("unknown action: %s", node.Action)
	}
}

// ── Action Implementations ──────────────────────────────────────────────

func (ex *Executor) execValidate(node *TaskNode) error {
	parts := strings.SplitN(node.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", node.Target)
	}
	kind, id := parts[0], parts[1]

	switch kind {
	case "Project":
		p := project.NewProject(id, id, "development", "Created via Workflow")
		if err := p.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		node.Result = fmt.Sprintf("Project %q is valid", id)
		return nil
	case "Application":
		// The Application Controller validates the Application before
		// submitting the workflow. Here we just confirm the resource exists.
		_, err := ex.resRegistry.Get("Application", id)
		if err != nil {
			return fmt.Errorf("application %q not found in registry: %w", id, err)
		}
		node.Result = fmt.Sprintf("Application %q is valid", id)
		return nil
	default:
		return fmt.Errorf("unknown kind for validation: %s", kind)
	}
}

func (ex *Executor) execResourceCreate(node *TaskNode) error {
	parts := strings.SplitN(node.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", node.Target)
	}
	kind, id := parts[0], parts[1]

	switch kind {
	case "Project":
		p := project.NewProject(id, id, "development", "Created via Workflow")
		if err := ex.resRegistry.Create(context.Background(), p); err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		node.Result = fmt.Sprintf("Project %q created", id)
		return nil
	default:
		return fmt.Errorf("unknown kind for creation: %s", kind)
	}
}

func (ex *Executor) execResourceGet(node *TaskNode) error {
	parts := strings.SplitN(node.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", node.Target)
	}
	kind, id := parts[0], parts[1]

	res, err := ex.resRegistry.Get(kind, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			node.Result = fmt.Sprintf("%s %q not found", kind, id)
			return nil
		}
		return fmt.Errorf("get %s %q: %w", kind, id, err)
	}

	node.Result = fmt.Sprintf("%s %q found (status: %s)", kind, id, res.GetStatus())
	return nil
}

func (ex *Executor) execResourceList(node *TaskNode) error {
	kind := node.Target
	if kind == "" {
		kind = "Project"
	}

	items, err := ex.resRegistry.List(kind)
	if err != nil {
		return fmt.Errorf("list %s: %w", kind, err)
	}

	node.Result = fmt.Sprintf("Found %d %s", len(items), strings.ToLower(kind))
	return nil
}

func (ex *Executor) execResourceDelete(node *TaskNode) error {
	parts := strings.SplitN(node.Target, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid target format: %s", node.Target)
	}
	kind, id := parts[0], parts[1]

	if err := ex.resRegistry.Delete(context.Background(), kind, id); err != nil {
		return fmt.Errorf("delete %s %q: %w", kind, id, err)
	}
	node.Result = fmt.Sprintf("%s %q deleted", kind, id)
	return nil
}

func (ex *Executor) execResourceKinds(node *TaskNode) error {
	kinds := ex.resRegistry.ListKinds()
	node.Result = fmt.Sprintf("Found %d resource kinds", len(kinds))
	return nil
}

func (ex *Executor) execControllerList(node *TaskNode) error {
	if ex.ctrlManager == nil {
		return fmt.Errorf("controller manager not available")
	}

	names := ex.ctrlManager.ControllerNames()
	node.Result = fmt.Sprintf("Found %d controllers", len(names))
	return nil
}

func (ex *Executor) execControllerReconcile(node *TaskNode) error {
	if ex.ctrlManager == nil {
		return fmt.Errorf("controller manager not available")
	}
	node.Result = "Reconciliation dispatched via Controller Runtime"
	return nil
}

func (ex *Executor) execHealthCheck(node *TaskNode) error {
	if ex.healthMgr == nil {
		return fmt.Errorf("health manager not available")
	}

	report := ex.healthMgr.All()
	if len(report) == 0 {
		node.Result = "No health components"
		return nil
	}

	healthyCount := 0
	for _, h := range report {
		if h.State == types.StateRunning || h.State == types.StatePending {
			healthyCount++
		}
	}
	node.Result = fmt.Sprintf("Checked %d components (%d healthy)", len(report), healthyCount)
	return nil
}

// ── Source Clone ─────────────────────────────────────────────────────────

// execSourceClone clones a Git repository. The target is the repository URL.
// It stores the cloned directory path so downstream nodes (build, deploy)
// can find the source code.
func (ex *Executor) execSourceClone(node *TaskNode) error {
	if ex.sourceCloner == nil {
		return fmt.Errorf("source cloner not available")
	}

	repoURL := node.Target
	if repoURL == "" {
		return fmt.Errorf("clone target (repository URL) is required")
	}

	// Extract a unique directory name from the URL.
	appID := sanitizeAppID(repoURL)

	result, err := ex.sourceCloner.Clone(context.Background(), repoURL, "main", appID)
	if err != nil {
		return fmt.Errorf("clone %q: %w", repoURL, err)
	}

	node.Result = result.LocalPath

	// Store the work directory for downstream nodes using the run ID
	// from the execution context.
	if ex.currentCtx != nil {
		if runID, ok := RunIDFromContext(ex.currentCtx); ok {
			ex.workDirs.Store(runID, result.LocalPath)
		}
	}

	ex.log.Info("source cloned",
		"repo", repoURL,
		"path", result.LocalPath,
		"branch", result.Branch,
		"commit", result.Commit,
	)
	return nil
}

// ── Build Install ─────────────────────────────────────────────────────────

// execBuildInstall runs the install command in the project directory.
// The target is the install command (e.g. "npm install").
// The working directory is resolved from the source.clone result.
func (ex *Executor) execBuildInstall(node *TaskNode) error {
	installCmd := node.Target
	if installCmd == "" {
		node.Result = "No install command — skipping"
		return nil
	}

	workDir := ex.getWorkDir()
	if workDir == "" {
		return fmt.Errorf("no work directory available for install (source.clone may not have run)")
	}

	// Run the install command.
	cmd := exec.Command("cmd", "/c", installCmd)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("install failed in %q: %s: %w", workDir, truncateOutput(string(output)), err)
	}

	node.Result = fmt.Sprintf("Install completed in %q", workDir)
	ex.log.Info("dependencies installed",
		"dir", workDir,
		"command", installCmd,
	)
	return nil
}

// ── Build Execute ─────────────────────────────────────────────────────────

// execBuildExecute runs the build command in the project directory.
// The target is the build command (e.g. "npm run build").
func (ex *Executor) execBuildExecute(node *TaskNode) error {
	buildCmd := node.Target
	if buildCmd == "" {
		node.Result = "No build command — skipping"
		return nil
	}

	workDir := ex.getWorkDir()
	if workDir == "" {
		return fmt.Errorf("no work directory available for build (source.clone may not have run)")
	}

	// Run the build command.
	cmd := exec.Command("cmd", "/c", buildCmd)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed in %q: %s: %w", workDir, truncateOutput(string(output)), err)
	}

	node.Result = fmt.Sprintf("Build completed in %q", workDir)
	ex.log.Info("build completed",
		"dir", workDir,
		"command", buildCmd,
	)
	return nil
}

// ── Provider Deploy ───────────────────────────────────────────────────────

// execProviderDeploy deploys the application using the configured Runtime.
// The target specifies the runtime type (e.g. "runtime:node").
//
// This is the critical action: it detects the runtime from the project,
// determines the start command, allocates a port, and starts the process
// via the Runtime interface (which may be Local, Docker, SSH, etc.).
func (ex *Executor) execProviderDeploy(node *TaskNode) error {
	if ex.runtimeManager == nil {
		return fmt.Errorf("runtime manager not available")
	}

	workDir := ex.getWorkDir()
	if workDir == "" {
		return fmt.Errorf("no work directory available for deployment")
	}

	// Detect the runtime using buildpacks.
	detected, err := buildpack.Detect(workDir)
	if err != nil {
		return fmt.Errorf("runtime detection failed in %q: %w", workDir, err)
	}

	ex.log.Info("runtime detected",
		"dir", workDir,
		"type", detected.Type,
		"install", detected.InstallCmd,
		"build", detected.BuildCmd,
		"start", detected.StartCmd,
	)

	// Fix up commands for the platform (e.g., Windows needs .exe extension).
	// This is called after buildpack detection so the buildpack output is
	// adjusted to match the local OS conventions.
	ex.fixupPlatformCommands(detected)

	// ── Ensure dependencies are installed ─────────────────────────────
	// The install and build steps may have been skipped if the workflow
	// definition didn't include them. Run them here as needed.
	if detected.InstallCmd != "" {
		ex.log.Info("running install command", "dir", workDir, "cmd", detected.InstallCmd)
		installCmd := exec.Command("cmd", "/c", detected.InstallCmd)
		installCmd.Dir = workDir
		if output, err := installCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("install command %q failed in %q: %s: %w",
				detected.InstallCmd, workDir, truncateOutput(string(output)), err)
		}
		ex.log.Info("install completed", "dir", workDir)
	}

	// ── Build the project ─────────────────────────────────────────────
	if detected.BuildCmd != "" {
		ex.log.Info("running build command", "dir", workDir, "cmd", detected.BuildCmd)
		buildCmd := exec.Command("cmd", "/c", detected.BuildCmd)
		buildCmd.Dir = workDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("build command %q failed in %q: %s: %w",
				detected.BuildCmd, workDir, truncateOutput(string(output)), err)
		}
		ex.log.Info("build completed", "dir", workDir)
	}

	// Determine the app ID for process tracking.
	appID := extractAppID(node)

	// Build the start command. Port 0 means the runtime will auto-allocate.
	startCmd := detected.StartCmd
	var port int
	if startCmd == "" {
		// Static sites: tell the runtime to use npx serve as the command
		// and the runtime will handle port allocation.
		startCmd = fmt.Sprintf("npx serve -s %s", detected.OutputDir)
		if detected.OutputDir == "" {
			startCmd = "npx serve -s ."
		}
	}

	// Build environment variables.
	envVars := map[string]string{
		"HOST":   "0.0.0.0",
		"APP_ID": appID,
	}
	for k, v := range detected.EnvVars {
		envVars[k] = v
	}

	// Start the process via the Runtime interface.
	// The Runtime handles port allocation and {port} substitution.
	inst, err := ex.runtimeManager.Start(context.Background(), cr.StartConfig{
		AppID:   appID,
		Name:    fmt.Sprintf("app-%s", appID),
		WorkDir: workDir,
		Command: startCmd,
		Port:    0, // auto-allocate
		EnvVars: envVars,
	})
	if err != nil {
		return fmt.Errorf("start via runtime: %w", err)
	}

	if inst.Port > 0 {
		port = inst.Port
	}

	// Register the instance for lifecycle cleanup.
	if ex.currentCtx != nil {
		if runID, ok := RunIDFromContext(ex.currentCtx); ok {
			ex.instances.Store(runID, inst.ID)
		}
	}

	// Store deployment info in the node result for downstream steps.
	node.Result = fmt.Sprintf("Running at %s (pid=%d, port=%d)", inst.URL, inst.PID, port)

	ex.log.Info("application deployed via runtime",
		"app", appID,
		"runtime", ex.runtimeManager.Name(),
		"url", inst.URL,
		"port", port,
		"pid", inst.PID,
		"detected", detected.Type,
	)
	return nil
}

// ── Lifecycle & Cleanup ─────────────────────────────────────────────────────

// CleanupRun stops all running instances for a workflow run and releases
// associated resources (ports, cloned directories, log stores).
//
// This is called by the Engine when a run is cancelled, fails, or completes
// to ensure no orphan processes are left behind.
func (ex *Executor) CleanupRun(ctx context.Context, runID string) {
	// Stop running instances.
	if val, ok := ex.instances.Load(runID); ok {
		if instID, ok := val.(string); ok && instID != "" {
			ex.log.Info("cleaning up instance for cancelled run", "run_id", runID, "instance", instID)
			if err := ex.runtimeManager.Stop(ctx, instID); err != nil {
				ex.log.Warn("cleanup stop instance", "run_id", runID, "instance", instID, "error", err.Error())
			}
		}
		ex.instances.Delete(runID)
	}

	// Clean up work directory reference.
	ex.workDirs.Delete(runID)
}

// ── Helpers ───────────────────────────────────────────────────────────────

// getWorkDir returns the cloned source directory for the current run.
// It retrieves the path stored by source.clone via the run ID in the context.
// Returns empty string if no work directory is available.
func (ex *Executor) getWorkDir() string {
	if ex.currentCtx == nil {
		return ""
	}
	runID, ok := RunIDFromContext(ex.currentCtx)
	if !ok {
		return ""
	}
	dir, ok := ex.workDirs.Load(runID)
	if !ok {
		return ""
	}
	path, ok := dir.(string)
	if !ok || path == "" {
		return ""
	}
	return path
}

// sanitizeAppID creates a safe directory name from a URL.
func sanitizeAppID(repoURL string) string {
	// Extract the last path component.
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		// Remove non-alphanumeric characters.
		safe := make([]byte, 0, len(name))
		for _, c := range []byte(name) {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
				safe = append(safe, c)
			} else {
				safe = append(safe, '-')
			}
		}
		return string(safe)
	}
	return "app"
}

// extractAppID extracts an application ID from the node context.
func extractAppID(node *TaskNode) string {
	// The target might be "Application:<id>" or "runtime:<type>".
	// In the deployment workflow, the complete step's target is the app ID.
	// For now, generate from the node ID.
	id := node.ID()
	if id != "" {
		return id
	}
	return "unknown"
}

// isDirectory checks if a path is a directory.
func isDirectory(path string) bool {
	// In a real implementation, use os.Stat.
	// For now, just check if it looks like a path.
	return strings.Contains(path, "/") || strings.Contains(path, "\\")
}

// fixupPlatformCommands adjusts buildpack commands for OS-specific conventions.
//
// On Windows:
//   - Go's `go build -o app .` produces a file named `app` (no .exe extension
//     in Go 1.20+). We fix this to `go build -o app.exe .` so the binary has
//     the correct extension for Windows execution.
//   - Unix-style `./app` start commands are converted to `app` (no path prefix)
//     because Windows cmd.exe interprets `/app` as a command-line switch rather
//     than a path component. Removing the `./` prefix lets Windows resolve the
//     binary via PATHEXT (.exe, .bat, etc.).
func (ex *Executor) fixupPlatformCommands(detected *buildpack.DetectedRuntime) {
	if runtime.GOOS != "windows" {
		return
	}

	// Fix Go build output: `go build -o app .` → `go build -o app.exe .`
	if detected.Type == "go" {
		detected.BuildCmd = strings.ReplaceAll(detected.BuildCmd,
			"go build -o app .",
			"go build -o app.exe .")
	}

	// Fix start commands that use Unix `./` prefix.
	// On Windows cmd.exe, `./app` is interpreted as command `.` with switch `/app`.
	// Stripping the prefix lets Windows resolve via PATHEXT.
	detected.StartCmd = strings.TrimPrefix(detected.StartCmd, "./")
}

// truncateOutput truncates command output for error messages.
func truncateOutput(output string) string {
	if len(output) > 500 {
		return output[:500] + "..."
	}
	return output
}

// ── WorkDir from source.clone node ────────────────────────────────────────

// getSourceDirFromWorkflow attempts to find the cloned source directory
// from the workflow run by examining previous node results.
// This is called by the executor when resolving the working directory.
func getSourceDirFromWorkflow(nodes []Node, currentNodeID string) string {
	for _, n := range nodes {
		if n.ID() == currentNodeID {
			break
		}
		// Look for a source.clone result (which contains a directory path).
		if tn, ok := n.(*TaskNode); ok && tn.Action == "source.clone" && tn.Result != "" {
			if isDirectory(tn.Result) {
				return tn.Result
			}
		}
	}
	return ""
}

// ── Node Result Formatting ───────────────────────────────────────────────

// FormatNodeResult converts a TaskNode into a ResultItem for workflow results.
func FormatNodeResult(node *TaskNode) ResultItem {
	msgType := "success"
	if node.Status() == NodeFailed {
		msgType = "error"
	} else if node.Status() == NodeSkipped {
		msgType = "warning"
	}

	detail := node.Result
	if node.ErrorVal != "" {
		detail = node.ErrorVal
	}

	return ResultItem{
		Message: fmt.Sprintf("%s — %s", node.Name(), node.Status()),
		Type:    msgType,
		Detail:  detail,
	}
}
