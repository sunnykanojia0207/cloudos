// Package source provides source code retrieval for CloudOS applications.
// It supports cloning from Git repositories (GitHub, GitLab, etc.) and
// preparing local directories for build and deployment.
package source

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// DefaultCloneTimeout is the maximum time allowed for a git clone operation.
	DefaultCloneTimeout = 2 * time.Minute

	// DefaultWorkDir is the base directory for cloned repositories.
	DefaultWorkDir = "work"
)

// ── Supported Source Types ─────────────────────────────────────────────────

// Type constants for source providers.
const (
	TypeGit   = "git"
	TypeLocal = "local"
	TypeDocker = "docker"
)

// ── GitCloner ──────────────────────────────────────────────────────────────

// GitCloner clones Git repositories to a local work directory.
// It uses the system `git` command rather than a library to keep
// dependencies minimal and leverage the user's existing Git configuration.
type GitCloner struct {
	// WorkDir is the base directory where repositories are cloned.
	WorkDir string

	// Timeout is the maximum duration for a clone operation.
	Timeout time.Duration
}

// NewGitCloner creates a GitCloner with the given work directory.
func NewGitCloner(workDir string) *GitCloner {
	if workDir == "" {
		workDir = DefaultWorkDir
	}
	return &GitCloner{
		WorkDir: workDir,
		Timeout: DefaultCloneTimeout,
	}
}

// CloneResult contains information about a successful clone operation.
type CloneResult struct {
	// LocalPath is the absolute path to the cloned repository.
	LocalPath string `json:"localPath"`

	// RepoName is the extracted repository name (e.g. "my-app").
	RepoName string `json:"repoName"`

	// Branch is the checked-out branch.
	Branch string `json:"branch"`

	// Commit is the HEAD commit hash (empty if not available).
	Commit string `json:"commit,omitempty"`
}

// validateGitURL checks that a git URL uses an allowed scheme and does not
// contain command injection characters. This prevents argument injection
// attacks via malicious repository URLs passed to exec.Command.
//
// Allowed schemes:
//   - https://  (remote repositories)
//   - git@     (SSH-style remote repositories)
//   - file://  (local repositories, for testing)
func validateGitURL(url string) error {
	// Reject URLs containing shell metacharacters or Git flag injection.
	disallowed := []string{"--", ";", "|", "`", "$", "(", ")", "\n", "\r"}
	for _, ch := range disallowed {
		if strings.Contains(url, ch) {
			return fmt.Errorf("git URL contains disallowed characters (%q)", ch)
		}
	}

	// Allow only known schemes.
	switch {
	case strings.HasPrefix(url, "https://"):
		return nil
	case strings.HasPrefix(url, "git@"):
		return nil
	case strings.HasPrefix(url, "file://"):
		return nil
	default:
		return fmt.Errorf("unsupported git URL scheme: %q (must be https://, git@, or file://)", url)
	}
}

// Clone clones a git repository to a local directory within the work directory.
//
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - url: the repository URL (e.g. "https://github.com/user/repo.git")
//   - branch: the branch to clone (empty defaults to "main")
//   - appID: a unique identifier for the application (used as subdirectory name)
//
// Returns a CloneResult with the local path and metadata, or an error.
func (g *GitCloner) Clone(ctx context.Context, url, branch, appID string) (*CloneResult, error) {
	if url == "" {
		return nil, fmt.Errorf("git URL is required")
	}
	if appID == "" {
		return nil, fmt.Errorf("app ID is required")
	}
	if err := validateGitURL(url); err != nil {
		return nil, fmt.Errorf("invalid git URL: %w", err)
	}
	if branch == "" {
		branch = "main"
	}

	// Determine the destination directory.
	destDir := filepath.Join(g.WorkDir, appID)

	// Ensure the work directory exists.
	if err := os.MkdirAll(g.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("create work directory %q: %w", g.WorkDir, err)
	}

	// Remove existing directory if present (for re-deploys).
	if _, err := os.Stat(destDir); err == nil {
		if err := os.RemoveAll(destDir); err != nil {
			return nil, fmt.Errorf("remove existing directory %q: %w", destDir, err)
		}
	}

	// Create a context with timeout.
	cloneCtx, cancel := context.WithTimeout(ctx, g.Timeout)
	defer cancel()

	// Build the git clone command.
	args := []string{"clone", "--depth", "1", "--branch", branch, url, destDir}
	cmd := exec.CommandContext(cloneCtx, "git", args...)

	// Capture stderr for error reporting.
	output, err := cmd.CombinedOutput()
	if err != nil {
		stderr := string(output)
		if len(stderr) > 500 {
			stderr = stderr[:500] + "..."
		}
		if cloneCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git clone timed out after %v: %s", g.Timeout, url)
		}
		return nil, fmt.Errorf("git clone failed: %s: %s", stderr, err.Error())
	}

	// Extract the repository name from the URL.
	repoName := extractRepoName(url)

	// Try to get the HEAD commit hash.
	commit := getCommitHash(destDir)

	return &CloneResult{
		LocalPath: destDir,
		RepoName:  repoName,
		Branch:    branch,
		Commit:    commit,
	}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────

// extractRepoName extracts the repository name from a git URL.
//
// Examples:
//
//	"https://github.com/user/my-app.git"  → "my-app"
//	"git@github.com:user/my-app.git"       → "my-app"
//	"https://github.com/user/my-app"       → "my-app"
func extractRepoName(url string) string {
	// Remove trailing .git
	url = strings.TrimSuffix(url, ".git")

	// Handle git@ URLs (SSH format)
	if idx := strings.LastIndex(url, "/"); idx >= 0 {
		url = url[idx+1:]
	} else if idx := strings.LastIndex(url, ":"); idx >= 0 {
		url = url[idx+1:]
	}

	return url
}

// getCommitHash returns the HEAD commit hash of a git repository.
func getCommitHash(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// EnsureWorkDir creates the work directory if it doesn't exist.
func EnsureWorkDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// RemoveAppDir removes an application's cloned directory.
func RemoveAppDir(workDir, appID string) error {
	dir := filepath.Join(workDir, appID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}
