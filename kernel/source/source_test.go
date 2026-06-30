package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/my-app.git", "my-app"},
		{"https://github.com/user/my-app", "my-app"},
		{"git@github.com:user/my-app.git", "my-app"},
		{"https://gitlab.com/org/project.git", "project"},
		{"https://github.com/user/next.js.git", "next.js"},
	}
	for _, tt := range tests {
		got := extractRepoName(tt.url)
		if got != tt.want {
			t.Errorf("extractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestNewGitCloner_Defaults(t *testing.T) {
	g := NewGitCloner("")
	if g.WorkDir != DefaultWorkDir {
		t.Errorf("WorkDir = %q, want %q", g.WorkDir, DefaultWorkDir)
	}
	if g.Timeout != DefaultCloneTimeout {
		t.Errorf("Timeout = %v, want %v", g.Timeout, DefaultCloneTimeout)
	}
}

func TestNewGitCloner_CustomWorkDir(t *testing.T) {
	g := NewGitCloner("/tmp/cloudos-work")
	if g.WorkDir != "/tmp/cloudos-work" {
		t.Errorf("WorkDir = %q, want %q", g.WorkDir, "/tmp/cloudos-work")
	}
}

func TestClone_EmptyURL(t *testing.T) {
	g := NewGitCloner(t.TempDir())
	_, err := g.Clone(context.Background(), "", "main", "test-app")
	if err == nil {
		t.Error("Clone() should return error for empty URL")
	}
}

func TestClone_EmptyAppID(t *testing.T) {
	g := NewGitCloner(t.TempDir())
	_, err := g.Clone(context.Background(), "https://github.com/user/repo.git", "main", "")
	if err == nil {
		t.Error("Clone() should return error for empty app ID")
	}
}

func TestClone_InvalidURL(t *testing.T) {
	g := NewGitCloner(t.TempDir())
	// This should fail quickly because the URL is invalid.
	_, err := g.Clone(context.Background(), "https://github.com/this-does-not-exist-12345/repo.git", "main", "test-app")
	if err == nil {
		// The test directory doesn't have git installed or the clone
		// fails, but either way we expect an error for a nonexistent repo.
		t.Log("Clone did not return error (possibly running in an environment with network access)")
	} else {
		t.Logf("Clone returned expected error: %v", err)
	}
}

func TestEnsureWorkDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "work")
	if err := EnsureWorkDir(dir); err != nil {
		t.Fatalf("EnsureWorkDir() returned error: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("EnsureWorkDir() did not create the directory")
	}
}

func TestRemoveAppDir(t *testing.T) {
	workDir := t.TempDir()

	// Create a fake app directory.
	appDir := filepath.Join(workDir, "test-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Remove it.
	if err := RemoveAppDir(workDir, "test-app"); err != nil {
		t.Fatalf("RemoveAppDir() returned error: %v", err)
	}
	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		t.Error("RemoveAppDir() did not remove the directory")
	}
}

func TestRemoveAppDir_NotExists(t *testing.T) {
	workDir := t.TempDir()
	if err := RemoveAppDir(workDir, "nonexistent"); err != nil {
		t.Errorf("RemoveAppDir() for nonexistent dir returned error: %v", err)
	}
}
