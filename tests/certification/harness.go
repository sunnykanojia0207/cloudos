// Package certification provides a test harness for CloudOS certification tests.
//
// A certification test validates that a specific stack (Go, React, Node, etc.)
// passes the full CloudOS deployment pipeline:
//
//	Git Repository → Application Resource → Controller → Workflow →
//	Executor → Buildpack Engine → Runtime → Running Instance → Reachable URL
//
// The harness abstracts all CloudOS kernel setup so each certification test
// is a short, readable script that proves the platform contracts work.
//
// Usage:
//
//	h := NewHarness(t)
//	defer h.Cleanup()
//
//	// Create a sample project (helper per stack)
//	repoDir := h.CreateGoProject("my-server", "8080")
//
//	// Initialize git repo
//	h.InitGitRepo(repoDir, "Initial commit")
//
//	// Create and deploy application
//	app := h.CreateApp(application.ApplicationSpec{
//	    Source: application.ApplicationSource{
//	        Type: application.SourceGit,
//	        URL:  h.FileURL(repoDir),
//	    },
//	    Runtime: application.ApplicationRuntime{Type: "go"},
//	    Deployment: application.ApplicationDeployment{Port: 0},
//	})
//
//	h.Deploy(app)
//
//	// Assertions
//	h.AssertAppRunning("my-server")
//	h.AssertURLReachable(app.Status_.URL)
//	h.AssertLogs(app.Status_.URL)
package certification

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudos/cloudos/kernel/application"
	"github.com/cloudos/cloudos/kernel/controller"
	"github.com/cloudos/cloudos/kernel/events"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/kernel/resource"
	cr "github.com/cloudos/cloudos/kernel/runtime"
	"github.com/cloudos/cloudos/kernel/runtime/local"
	"github.com/cloudos/cloudos/kernel/source"
	"github.com/cloudos/cloudos/kernel/workflow"
	"github.com/cloudos/cloudos/packages/logging"
)

// ── Builder ─────────────────────────────────────────────────────────────────

// TestHarness encapsulates all CloudOS kernel components needed for a
// certification test. Use NewHarness to create one, then call methods
// to set up projects, deploy apps, and assert outcomes.
type TestHarness struct {
	t       *testing.T
	ctx     context.Context
	cancel  context.CancelFunc
	log     *logging.Logger
	cleanup []func()

	// CloudOS kernel components.
	EventBus          *events.Bus
	Registry          *resource.Registry
	HealthMgr         *health.Manager
	CtrlMgr           *controller.Manager
	RuntimeMgr        *local.Manager
	LogMgr            *cr.LogManager
	SourceCloner      *source.GitCloner
	WorkflowSvc       *workflow.Service

	// Polling configuration.
	PollInterval time.Duration
	Deadline     time.Duration
}

// NewHarness creates a certification test harness.
// Call t.Cleanup(h.Cleanup) or defer h.Cleanup() to release resources.
func NewHarness(t *testing.T) *TestHarness {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	log := logging.NewSubsystemLogger("certification", logging.LevelDebug)

	// ── Event Bus ───────────────────────────────────────────────────────
	bus := events.NewBus(log)
	bus.Start()

	// ── Resource Registry ───────────────────────────────────────────────
	reg := resource.NewRegistry(bus, log)
	if err := reg.RegisterKind(resource.Kind{
		Name:       application.Kind,
		Namespaced: true,
		Versions:   []string{"v1"},
	}); err != nil {
		t.Fatal(err)
	}

	// ── Health Manager ──────────────────────────────────────────────────
	healthMgr := health.NewManager(log)

	// ── Controller Manager ──────────────────────────────────────────────
	ctrlMgr := controller.NewManager(reg, bus, healthMgr, log)

	// ── Runtime ─────────────────────────────────────────────────────────
	runtimeDir := t.TempDir()
	runtimeMgr := local.NewManager(runtimeDir, log)

	// ── Log Manager ─────────────────────────────────────────────────────
	logMgr := cr.NewLogManager(1000)
	runtimeMgr.WithLogManager(logMgr)

	// ── Source Cloner ───────────────────────────────────────────────────
	cloneDir := t.TempDir()
	sourceCloner := source.NewGitCloner(cloneDir)

	// ── Workflow Service ────────────────────────────────────────────────
	workflowSvc := workflow.NewService(workflow.ServiceDeps{
		ResourceRegistry:  reg,
		ControllerManager: ctrlMgr,
		HealthManager:     healthMgr,
		EventBus:          bus,
		SourceCloner:      sourceCloner,
		RuntimeManager:    runtimeMgr,
		Logger:            log,
	})

	// ── Application Controller ──────────────────────────────────────────
	appCtrl := application.NewApplicationController(reg, bus, workflowSvc, log)
	if err := ctrlMgr.Register(appCtrl); err != nil {
		t.Fatal(err)
	}

	h := &TestHarness{
		t:            t,
		ctx:          ctx,
		cancel:       cancel,
		log:          log,
		EventBus:     bus,
		Registry:     reg,
		HealthMgr:    healthMgr,
		CtrlMgr:      ctrlMgr,
		RuntimeMgr:   runtimeMgr,
		LogMgr:       logMgr,
		SourceCloner: sourceCloner,
		WorkflowSvc:  workflowSvc,
		PollInterval: 500 * time.Millisecond,
		Deadline:     60 * time.Second,
	}

	// Register cleanup.
	h.cleanup = append(h.cleanup, func() {
		runtimeMgr.StopAll()
		bus.Stop()
		cancel()
	})

	return h
}

// Cleanup releases all resources held by the harness.
func (h *TestHarness) Cleanup() {
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		h.cleanup[i]()
	}
}

// ── Lifecycle ───────────────────────────────────────────────────────────────

// StartKernel starts the controller runtime and workflow engine.
// Must be called before Deploy.
func (h *TestHarness) StartKernel() {
	h.t.Helper()

	if err := h.CtrlMgr.Start(h.ctx); err != nil {
		h.t.Fatalf("controller manager start: %v", err)
	}

	go h.WorkflowSvc.Engine().Start(h.ctx)
	time.Sleep(100 * time.Millisecond) // let the engine start
}

// ── Git Helpers ─────────────────────────────────────────────────────────────

// InitGitRepo creates a git repository in dir and makes an initial commit.
func (h *TestHarness) InitGitRepo(dir, msg string) {
	h.t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		h.t.Skip("git not found on PATH, skipping certification test")
	}

	cmds := []struct {
		name string
		args []string
	}{
		{"git", []string{"init", "-b", "main"}},
		{"git", []string{"config", "user.email", "certify@cloudos.io"}},
		{"git", []string{"config", "user.name", "CloudOS Certification"}},
		{"git", []string{"add", "-A"}},
		{"git", []string{"commit", "-m", msg}},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd.name, cmd.args...)
		c.Dir = dir
		if output, err := c.CombinedOutput(); err != nil {
			h.t.Fatalf("git %s failed: %s: %v", cmd.args[0], string(output), err)
		}
	}
	h.t.Logf("Git repo initialized at %s", dir)
}

// FileURL converts a local path to a file:// URL suitable for git cloning.
// On Windows, converts C:\path\to\dir → file:///C:/path/to/dir.
func (h *TestHarness) FileURL(localPath string) string {
	abs, err := filepath.Abs(localPath)
	if err != nil {
		h.t.Fatalf("filepath.Abs(%q): %v", localPath, err)
	}
	url := "file:///" + strings.ReplaceAll(abs, "\\", "/")
	if !strings.HasPrefix(url, "file:///") {
		url = "file:///" + url
	}
	return url
}

// ── Application Helpers ─────────────────────────────────────────────────────

// CreateApp registers an Application resource in the registry.
// Returns the Application with defaults applied.
func (h *TestHarness) CreateApp(id, name string, spec application.ApplicationSpec) *application.Application {
	h.t.Helper()

	app := application.NewApplication(id, name, spec)
	app.EnsureDefaults()

	if err := h.Registry.Create(h.ctx, app); err != nil {
		h.t.Fatalf("create application %q: %v", id, err)
	}
	h.t.Logf("Application %q created (source: %s, runtime: %s)", id, spec.Source.Type, spec.Runtime.Type)
	return app
}

// Deploy triggers deployment via the Application Controller and waits for
// the application to reach PhaseRunning with HealthHealthy.
// Returns the deployed Application with status populated.
func (h *TestHarness) Deploy(app *application.Application) *application.Application {
	h.t.Helper()

	id := app.GetMetadata().ID
	deadline := time.Now().Add(h.Deadline)

	// The controller reconciles asynchronously. Poll for completion.
	var lastStatus string
	var deployedApp *application.Application

	for time.Now().Before(deadline) {
		res, err := h.Registry.Get(application.Kind, id)
		if err != nil {
			lastStatus = fmt.Sprintf("get error: %v", err)
			time.Sleep(h.PollInterval)
			continue
		}

		appObj, ok := res.(*application.Application)
		if !ok {
			lastStatus = "unexpected resource type"
			time.Sleep(h.PollInterval)
			continue
		}

		deployedApp = appObj
		phase := deployedApp.Status_.Phase
		health := deployedApp.Status_.Health
		url := deployedApp.Status_.URL
		lastStatus = fmt.Sprintf("phase=%s health=%s url=%s", phase, health, url)

		if phase == application.PhaseRunning && health == application.HealthHealthy {
			h.t.Logf("Application %q is RUNNING: %s", id, lastStatus)
			return deployedApp
		}

		if phase == application.PhaseFailed {
			h.t.Fatalf("Application %q FAILED: %s", id, lastStatus)
		}

		time.Sleep(h.PollInterval)
	}

	if deployedApp == nil {
		h.t.Fatalf("Application %q was never created (last: %s)", id, lastStatus)
	}
	if deployedApp.Status_.Phase != application.PhaseRunning {
		h.t.Fatalf("Application %q did not reach Running (last: %s)", id, lastStatus)
	}
	return deployedApp
}

// ── Assertions ──────────────────────────────────────────────────────────────

// AssertAppRunning verifies the application is in PhaseRunning.
func (h *TestHarness) AssertAppRunning(id string) {
	h.t.Helper()
	res, err := h.Registry.Get(application.Kind, id)
	if err != nil {
		h.t.Fatalf("get application %q: %v", id, err)
	}
	app, ok := res.(*application.Application)
	if !ok {
		h.t.Fatalf("application %q is not an Application", id)
	}
	if app.Status_.Phase != application.PhaseRunning {
		h.t.Fatalf("application %q phase = %q, want %q", id, app.Status_.Phase, application.PhaseRunning)
	}
}

// AssertURLReachable verifies an HTTP URL responds with 200 OK.
func (h *TestHarness) AssertURLReachable(url string) {
	h.t.Helper()
	if url == "" {
		h.t.Fatal("url is empty")
	}

	var resp *http.Response
	var httpErr error
	for i := 0; i < 10; i++ {
		resp, httpErr = http.Get(url + "/health")
		if httpErr == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			h.t.Logf("Health check PASSED at %s/health (attempt %d)", url, i+1)
			return
		}
		if httpErr != nil {
			h.t.Logf("Health check attempt %d: %v", i+1, httpErr)
		} else {
			resp.Body.Close()
			h.t.Logf("Health check attempt %d: HTTP %d", i+1, resp.StatusCode)
		}
		time.Sleep(1 * time.Second)
	}

	if httpErr != nil {
		h.t.Fatalf("Health check failed after 10 attempts: %v", httpErr)
	}
}

// AssertRootEndpoint verifies the root URL responds with 200 OK.
func (h *TestHarness) AssertRootEndpoint(url string) {
	h.t.Helper()
	resp, err := http.Get(url + "/")
	if err != nil {
		h.t.Fatalf("Root endpoint request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		h.t.Fatalf("Root endpoint returned HTTP %d, want 200", resp.StatusCode)
	}
	h.t.Logf("Root endpoint PASSED at %s/", url)
}

// ── Sample Project Creators ─────────────────────────────────────────────────

// CreateGoProject creates a minimal Go HTTP server in a temp directory.
// Returns the directory path.
func (h *TestHarness) CreateGoProject(name, port string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	mainGo := fmt.Sprintf(`package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = %q
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from CloudOS Go! Port=%%s", port)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	addr := "0.0.0.0:" + port
	fmt.Printf("Listening on %%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %%v\n", err)
		os.Exit(1)
	}
}
`, port)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644); err != nil {
		h.t.Fatal(err)
	}

	goMod := fmt.Sprintf(`module github.com/cloudos-certify/%s

go 1.21
`, name)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// CreateStaticProject creates a minimal static HTML site.
// Returns the directory path.
func (h *TestHarness) CreateStaticProject(name string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	indexHTML := `<!DOCTYPE html>
<html>
<head><title>CloudOS Static</title></head>
<body>
<h1>Hello from CloudOS Static!</h1>
</body>
</html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// CreateNodeProject creates a minimal Node.js HTTP server.
// Returns the directory path.
func (h *TestHarness) CreateNodeProject(name, port string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	indexJS := fmt.Sprintf(`const http = require('http');

const port = process.env.PORT || %s;

const server = http.createServer((req, res) => {
  if (req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('OK');
    return;
  }
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('Hello from CloudOS Node.js! Port=' + port);
});

server.listen(port, '0.0.0.0', () => {
  console.log('Listening on port ' + port);
});
`, port)

	if err := os.WriteFile(filepath.Join(dir, "index.js"), []byte(indexJS), 0644); err != nil {
		h.t.Fatal(err)
	}

	pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  }
}`, name)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// CreateReactProject creates a minimal Vite React project.
// Returns the directory path.
func (h *TestHarness) CreateReactProject(name string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	// Create the directory structure
	dirs := []string{"src", "public"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			h.t.Fatal(err)
		}
	}

	// index.html
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>CloudOS React</title>
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/src/main.jsx"></script>
</body>
</html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644); err != nil {
		h.t.Fatal(err)
	}

	// vite.config.js
	viteConfig := `import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
});
`
	if err := os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(viteConfig), 0644); err != nil {
		h.t.Fatal(err)
	}

	// src/main.jsx
	mainJSX := `import React from 'react';
import ReactDOM from 'react-dom/client';

function App() {
  return (
    <div>
      <h1>Hello from CloudOS React!</h1>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />);
`
	if err := os.WriteFile(filepath.Join(dir, "src", "main.jsx"), []byte(mainJSX), 0644); err != nil {
		h.t.Fatal(err)
	}

	// package.json
	pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.2.0",
    "vite": "^5.0.0"
  }
}`, name)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// CreateNextJSProject creates a minimal Next.js project.
// Returns the directory path.
func (h *TestHarness) CreateNextJSProject(name string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	dirs := []string{"src", "public"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			h.t.Fatal(err)
		}
	}

	// next.config.js
	nextConfig := `/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
};
module.exports = nextConfig;
`
	if err := os.WriteFile(filepath.Join(dir, "next.config.js"), []byte(nextConfig), 0644); err != nil {
		h.t.Fatal(err)
	}

	// src/pages directory
	if err := os.MkdirAll(filepath.Join(dir, "src", "pages"), 0755); err != nil {
		h.t.Fatal(err)
	}

	// src/pages/index.jsx
	indexJSX := `export default function Home() {
  return (
    <div>
      <h1>Hello from CloudOS Next.js!</h1>
    </div>
  );
}
`
	if err := os.WriteFile(filepath.Join(dir, "src", "pages", "index.jsx"), []byte(indexJSX), 0644); err != nil {
		h.t.Fatal(err)
	}

	// src/pages/api/health.js
	if err := os.MkdirAll(filepath.Join(dir, "src", "pages", "api"), 0755); err != nil {
		h.t.Fatal(err)
	}
	healthJS := `export default function handler(req, res) {
  res.status(200).json({ status: 'ok' });
}
`
	if err := os.WriteFile(filepath.Join(dir, "src", "pages", "api", "health.js"), []byte(healthJS), 0644); err != nil {
		h.t.Fatal(err)
	}

	// next.config.mjs (we use .js above; also provide jsconfig)
	jsConfig := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "jsconfig.json"), []byte(jsConfig), 0644); err != nil {
		h.t.Fatal(err)
	}

	// package.json
	pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "dependencies": {
    "next": "^14.2.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  }
}`, name)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// CreatePythonProject creates a minimal Python Flask HTTP server.
// Returns the directory path.
func (h *TestHarness) CreatePythonProject(name, port string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	appPy := fmt.Sprintf(`import os
import sys

try:
    from flask import Flask
except ImportError:
    # Fallback: use http.server if Flask is not installed
    from http.server import HTTPServer, BaseHTTPRequestHandler

    class HealthHandler(BaseHTTPRequestHandler):
        def do_GET(self):
            if self.path == '/health':
                self.send_response(200)
                self.send_header('Content-Type', 'text/plain')
                self.end_headers()
                self.wfile.write(b'OK')
            else:
                self.send_response(200)
                self.send_header('Content-Type', 'text/plain')
                self.end_headers()
                self.wfile.write(b'Hello from CloudOS Python!')
    
    port = int(os.environ.get('PORT', %s))
    server = HTTPServer(('0.0.0.0', port), HealthHandler)
    print(f'Listening on port {port}')
    server.serve_forever()
else:
    app = Flask(__name__)

    @app.route('/')
    def hello():
        return 'Hello from CloudOS Python!'

    @app.route('/health')
    def health():
        return 'OK', 200

    if __name__ == '__main__':
        port = int(os.environ.get('PORT', %s))
        app.run(host='0.0.0.0', port=port)
`, port, port)

	if err := os.WriteFile(filepath.Join(dir, "app.py"), []byte(appPy), 0644); err != nil {
		h.t.Fatal(err)
	}

	reqTxt := "flask\n"
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(reqTxt), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// CreateLaravelProject creates a minimal Laravel-like PHP project.
// Returns the directory path.
func (h *TestHarness) CreateLaravelProject(name string) string {
	h.t.Helper()
	dir := h.t.TempDir()

	// Create directory structure
	dirs := []string{"public", "routes"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			h.t.Fatal(err)
		}
	}

	// artisan
	artisan := `<?php
// Laravel-like artisan script
$command = $argv[1] ?? '';

if ($command === 'serve') {
    $host = $argv[2] ?? '0.0.0.0';
    $port = $argv[3] ?? '8080';
    echo "Laravel development server started on http://$host:$port\n";
    
    // Simple PHP built-in server with a router
    $_SERVER['APP_PORT'] = $port;
    
    // Start PHP's built-in server with the public/router.php
    $routerFile = __DIR__ . '/public/router.php';
    if (!file_exists($routerFile)) {
        file_put_contents($routerFile, '<?php
// Simple router for Laravel-like projects
$uri = $_SERVER["REQUEST_URI"];

if ($uri === "/health") {
    http_response_code(200);
    echo "OK";
    return;
}

// Serve static files
$publicPath = __DIR__ . $uri;
if ($uri !== "/" && file_exists($publicPath) && !is_dir($publicPath)) {
    return false;
}

// Default response
header("Content-Type: text/html; charset=utf-8");
?>
<!DOCTYPE html>
<html>
<head><title>CloudOS Laravel</title></head>
<body>
<h1>Hello from CloudOS Laravel!</h1>
</body>
</html>
');
    }
    
    $cmd = sprintf("php -S %s:%s -t %s %s",
        escapeshellarg($host),
        escapeshellarg($port),
        escapeshellarg(__DIR__ . '/public'),
        escapeshellarg($routerFile)
    );
    passthru($cmd);
} else {
    echo "CloudOS Artisan\n";
    echo "  serve   Start the development server\n";
}
`
	if err := os.WriteFile(filepath.Join(dir, "artisan"), []byte(artisan), 0644); err != nil {
		h.t.Fatal(err)
	}

	// composer.json
	composerJSON := fmt.Sprintf(`{
  "name": "cloudos-certify/%s",
  "require": {
    "php": ">=8.0"
  },
  "scripts": {
    "serve": "php artisan serve"
  }
}`, name)
	if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644); err != nil {
		h.t.Fatal(err)
	}

	// public/index.php
	indexPHP := `<?php
// Minimal Laravel-like front controller
$uri = $_SERVER['REQUEST_URI'];

if ($uri === '/health') {
    http_response_code(200);
    echo 'OK';
    exit;
}

http_response_code(200);
?>
<!DOCTYPE html>
<html>
<head><title>CloudOS Laravel</title></head>
<body>
<h1>Hello from CloudOS Laravel!</h1>
</body>
</html>
`
	if err := os.WriteFile(filepath.Join(dir, "public", "index.php"), []byte(indexPHP), 0644); err != nil {
		h.t.Fatal(err)
	}

	return dir
}

// ── Toolchain Checks ────────────────────────────────────────────────────────

// RequireNode checks that Node.js is available. Skips the test if not found.
func (h *TestHarness) RequireNode() {
	h.t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		h.t.Skip("node not found on PATH, skipping test")
	}
	if _, err := exec.LookPath("npm"); err != nil {
		h.t.Skip("npm not found on PATH, skipping test")
	}
}

// RequireGo checks that Go is available. Skips the test if not found.
func (h *TestHarness) RequireGo() {
	h.t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		h.t.Skip("go not found on PATH, skipping test")
	}
}

// RequirePython checks that Python is available. Skips the test if not found.
func (h *TestHarness) RequirePython() {
	h.t.Helper()
	if _, err := exec.LookPath("python"); err != nil {
		if _, err2 := exec.LookPath("python3"); err2 != nil {
			h.t.Skip("python/python3 not found on PATH, skipping test")
		}
	}
}

// RequirePHP checks that PHP is available. Skips the test if not found.
func (h *TestHarness) RequirePHP() {
	h.t.Helper()
	if _, err := exec.LookPath("php"); err != nil {
		h.t.Skip("php not found on PATH, skipping test")
	}
}

// RequireComposer checks that Composer is available. Skips the test if not found.
func (h *TestHarness) RequireComposer() {
	h.t.Helper()
	if _, err := exec.LookPath("composer"); err != nil {
		h.t.Skip("composer not found on PATH, skipping test")
	}
}

// ── Emit Metadata ──────────────────────────────────────────────────────────

// CertResult represents the result of a certification test.
type CertResult struct {
	Stack   string `json:"stack"`
	Detect  bool   `json:"detect"`
	Plan    bool   `json:"plan"`
	Build   bool   `json:"build"`
	Runtime bool   `json:"runtime"`
	Health  bool   `json:"health"`
	Logs    bool   `json:"logs"`
	Metrics bool   `json:"metrics"`
}

// EmitResult logs a structured certification result that can be parsed
// for automated compatibility matrix generation.
func (h *TestHarness) EmitResult(result CertResult) {
	h.t.Logf("CERTRESULT: stack=%s detect=%v plan=%v build=%v runtime=%v health=%v logs=%v metrics=%v",
		result.Stack, result.Detect, result.Plan, result.Build,
		result.Runtime, result.Health, result.Logs, result.Metrics)
}
