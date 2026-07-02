package buildpack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Test Project Setup ─────────────────────────────────────────────────────

// setupProject creates a temporary directory with the given files.
// Files is a map of relative path → contents.
func setupProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// ── Node.js Detection ──────────────────────────────────────────────────────

func TestDetect_NodeJS(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{
			"name": "my-app",
			"version": "1.0.0",
			"scripts": {
				"start": "node index.js"
			}
		}`,
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeNode {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeNode)
	}

	expected := []struct {
		field string
		got   string
		want  string
	}{
		{"InstallCmd", r.InstallCmd, "npm install"},
		{"StartCmd", r.StartCmd, "node index.js"},
	}
	for _, e := range expected {
		if e.got != e.want {
			t.Errorf("%s = %q, want %q", e.field, e.got, e.want)
		}
	}
}

func TestDetect_NodeJSNoPackageJSON(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": "invalid json",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeNode {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeNode)
	}
	// Should have sensible defaults even with unparseable package.json.
	if r.InstallCmd != "npm install" {
		t.Errorf("InstallCmd = %q, want %q", r.InstallCmd, "npm install")
	}
	if r.StartCmd != "npm start" {
		t.Errorf("StartCmd = %q, want %q", r.StartCmd, "npm start")
	}
}

func TestDetect_React(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{
			"name": "my-react-app",
			"scripts": {
				"build": "react-scripts build",
				"start": "react-scripts start"
			},
			"dependencies": {
				"react": "^18.0.0",
				"react-scripts": "^5.0.0"
			}
		}`,
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeReact {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeReact)
	}
	if r.OutputDir != "build" {
		t.Errorf("OutputDir = %q, want %q", r.OutputDir, "build")
	}
}

func TestDetect_NextJS(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"package.json": `{
			"name": "my-next-app",
			"scripts": {
				"build": "next build",
				"start": "next start"
			},
			"dependencies": {
				"next": "^14.0.0",
				"react": "^18.0.0"
			}
		}`,
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeNextJS {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeNextJS)
	}
	if r.OutputDir != ".next" {
		t.Errorf("OutputDir = %q, want %q", r.OutputDir, ".next")
	}
}

// ── Go Detection ───────────────────────────────────────────────────────────

func TestDetect_Go(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"go.mod": `module github.com/user/my-app

go 1.22

require (
	example.com/foo v1.0.0
)`,
		"main.go": `package main

func main() {}`,
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeGo {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeGo)
	}
	if r.BuildCmd != "go build -o app ." {
		t.Errorf("BuildCmd = %q, want %q", r.BuildCmd, "go build -o app .")
	}
}

// ── Python Detection ───────────────────────────────────────────────────────

func TestDetect_Python(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"requirements.txt": "flask\nrequests\n",
		"app.py":           `print("hello")`,
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimePython {
		t.Errorf("Type = %q, want %q", r.Type, RuntimePython)
	}

	// Python buildpack now creates a virtualenv. Expected install command
	// is platform-dependent: "python -m venv venv && <venv-pip> install -r requirements.txt"
	if !strings.Contains(r.InstallCmd, "install -r requirements.txt") {
		t.Errorf("InstallCmd = %q, want it to contain %q", r.InstallCmd, "install -r requirements.txt")
	}
	if !strings.Contains(r.InstallCmd, "python -m venv") {
		t.Errorf("InstallCmd = %q, want it to contain %q", r.InstallCmd, "python -m venv")
	}

	// Start command should use the venv Python interpreter
	if !strings.Contains(r.StartCmd, "venv") && !strings.Contains(r.StartCmd, "python") {
		t.Errorf("StartCmd = %q, want it to use venv Python", r.StartCmd)
	}
}

// ── Laravel Detection ──────────────────────────────────────────────────────

func TestDetect_Laravel(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"composer.json": `{
			"name": "my-laravel-app",
			"require": {
				"laravel/framework": "^10.0"
			}
		}`,
		"artisan": "<?php // Laravel artisan",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeLaravel {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeLaravel)
	}
}

// ── Static Detection ───────────────────────────────────────────────────────

func TestDetect_Static(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"index.html": "<html><body>Hello</body></html>",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeStatic {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeStatic)
	}
}

// ── Docker Detection ───────────────────────────────────────────────────────

func TestDetect_Docker(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"Dockerfile": "FROM node:18\n",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeDocker {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeDocker)
	}
}

// ── Fallback Detection ─────────────────────────────────────────────────────

func TestDetect_Fallback(t *testing.T) {
	dir := setupProject(t, map[string]string{
		"README.md": "# My project",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeStatic {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeStatic)
	}
}

// ── Empty Directory ────────────────────────────────────────────────────────

func TestDetect_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeStatic {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeStatic)
	}
}

// ── StartCommandWithPort ───────────────────────────────────────────────────

func TestStartCommandWithPort(t *testing.T) {
	tests := []struct {
		cmd  string
		port int
		want string
	}{
		{"python manage.py runserver 0.0.0.0:{port}", 8000, "python manage.py runserver 0.0.0.0:8000"},
		{"npx serve -s build -l {port}", 3000, "npx serve -s build -l 3000"},
		{"npm start", 8080, "npm start"},
	}
	for _, tt := range tests {
		got := StartCommandWithPort(tt.cmd, tt.port)
		if got != tt.want {
			t.Errorf("StartCommandWithPort(%q, %d) = %q, want %q", tt.cmd, tt.port, got, tt.want)
		}
	}
}

// ── Detection Hierarchy ────────────────────────────────────────────────────

func TestDetect_Hierarchy_GoBeforeNode(t *testing.T) {
	// Go (go.mod) should be detected before Node (package.json).
	dir := setupProject(t, map[string]string{
		"go.mod":       "module test",
		"package.json": `{"name":"test","scripts":{"start":"node index.js"}}`,
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeGo {
		t.Errorf("Type = %q, want %q (go.mod should take priority over package.json)", r.Type, RuntimeGo)
	}
}

func TestDetect_Hierarchy_DockerIsLastNonFallback(t *testing.T) {
	// Dockerfile should be checked before Static fallback but after
	// all other buildpacks. Since Go is higher priority, a project
	// with both go.mod and Dockerfile will match Go first.
	dir := setupProject(t, map[string]string{
		"go.mod":      "module test",
		"Dockerfile":  "FROM node:18",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeGo {
		t.Errorf("Type = %q, want %q (go.mod should take priority over Dockerfile)", r.Type, RuntimeGo)
	}
}

func TestDetect_Hierarchy_DockerBeatsStatic(t *testing.T) {
	// Dockerfile should be detected when nothing else matches.
	dir := setupProject(t, map[string]string{
		"Dockerfile":  "FROM node:18",
	})

	r, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}
	if r.Type != RuntimeDocker {
		t.Errorf("Type = %q, want %q", r.Type, RuntimeDocker)
	}
}
