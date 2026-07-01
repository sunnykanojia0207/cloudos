// Package version provides the current CloudOS version and component version information.
// Version values are set at build time via -ldflags.
package version

import (
	"fmt"
	"runtime"
)

// These values are overridden at build time using -ldflags.
// Example: go build -ldflags="-X github.com/cloudos/cloudos/packages/version.Number=v0.6.0-rc1"
var (
	// Number is the semantic version of CloudOS.
	Number = "0.6.0-rc1"

	// Commit is the short git commit hash from which this build was produced.
	Commit = "unknown"

	// Date is the RFC3339 timestamp of the build.
	Date = "unknown"
)

// Info returns a formatted string with version details.
func Info() string {
	return fmt.Sprintf("CloudOS %s (commit: %s, built: %s)", Number, Commit, Date)
}

// Short returns a concise version string suitable for CLI --version flags.
func Short() string {
	return Number
}

// Full returns a detailed version string including platform and Go version.
func Full() string {
	return fmt.Sprintf("CloudOS v%s\nCommit:     %s\nBuilt:      %s\nPlatform:   %s/%s\nGo version: %s",
		Number, Commit, Date, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
