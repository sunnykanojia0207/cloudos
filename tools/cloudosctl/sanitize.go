package main

import (
	"fmt"
	"regexp"
	"strings"
)

// ── App ID Validation ───────────────────────────────────────────────────────
//
// Application IDs are user-provided values that appear in URLs, file names,
// and API calls. They must be validated to prevent:
//   - Path traversal attacks (B11 fix)
//   - Command injection
//   - URL manipulation

// validAppIDPattern matches safe application identifiers.
// Only lowercase letters, digits, hyphens, underscores, and dots are allowed.
// This prevents path traversal (../..\), encoded traversal (%2e%2e),
// absolute paths (/etc/passwd), and Windows drive prefixes (C:\).
var validAppIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// sanitizeAppID validates an application ID and returns an error if it
// contains path traversal characters or other unsafe patterns.
// This MUST be called before using appID in file paths, URLs, or API calls.
func sanitizeAppID(appID string) error {
	if appID == "" {
		return fmt.Errorf("application ID cannot be empty")
	}

	// Reject path traversal patterns.
	if strings.Contains(appID, "..") {
		return fmt.Errorf("invalid application ID %q: path traversal detected", appID)
	}

	// Reject absolute paths on Unix and Windows.
	if strings.HasPrefix(appID, "/") {
		return fmt.Errorf("invalid application ID %q: cannot start with '/'", appID)
	}
	if strings.HasPrefix(appID, "\\") {
		return fmt.Errorf("invalid application ID %q: cannot start with '\\'", appID)
	}

	// Reject Windows drive letters (C:, D:, etc.).
	if len(appID) >= 2 && appID[1] == ':' {
		return fmt.Errorf("invalid application ID %q: drive letter prefix not allowed", appID)
	}

	// Reject encoded traversal (basic cases).
	lower := strings.ToLower(appID)
	if strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") || strings.Contains(lower, "%5c") {
		return fmt.Errorf("invalid application ID %q: encoded path traversal detected", appID)
	}

	// Validate character whitelist.
	if !validAppIDPattern.MatchString(appID) {
		return fmt.Errorf("invalid application ID %q: only letters, digits, hyphens, underscores, and dots allowed", appID)
	}

	return nil
}

// sanitizeLogFilename creates a safe log filename from an app ID.
// It replaces unsafe characters and prevents path traversal.
func sanitizeLogFilename(appID string) string {
	// Replace path separators and drive letters with safe alternatives.
	safe := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"..", "_",
		"%", "_",
	).Replace(appID)

	// Remove any remaining non-alphanumeric non-dot non-hyphen characters.
	result := make([]byte, 0, len(safe))
	for _, c := range []byte(safe) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}

	return string(result) + ".log"
}
