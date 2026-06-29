// Package version provides the current CloudOS version and component version information.
// Version values are set at build time via -ldflags.
package version

import "fmt"

// These values are overridden at build time using -ldflags.
// Example: go build -ldflags="-X github.com/cloudos/cloudos/packages/version.Number=v0.1.0"
var (
	// Number is the semantic version of CloudOS.
	Number = "0.1.0-dev"

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
