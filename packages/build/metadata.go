// Package build carries build-time metadata injected by the compiler.
// Use -ldflags at build time to populate these values.
//
// Example:
//
//	go build -ldflags="
//	  -X github.com/cloudos/cloudos/packages/build.CommitHash=$(git rev-parse --short HEAD)
//	  -X github.com/cloudos/cloudos/packages/build.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
//	  -X github.com/cloudos/cloudos/packages/build.GoVersion=$(go version | sed 's/go version //')
//	" ./tools/cloudos
package build

// Build-time metadata populated via -ldflags.
var (
	// CommitHash is the short Git commit hash.
	CommitHash = "unknown"

	// BuildTime is the UTC timestamp of the build in RFC3339 format.
	BuildTime = "unknown"

	// GoVersion is the Go compiler version used for this build.
	GoVersion = "unknown"

	// BuildUser is the user who triggered the build.
	BuildUser = "unknown"
)

// Metadata aggregates all build-time values into a single struct.
type Metadata struct {
	CommitHash string `json:"commitHash"`
	BuildTime  string `json:"buildTime"`
	GoVersion  string `json:"goVersion"`
	BuildUser  string `json:"buildUser"`
}

// Get returns the current build metadata.
func Get() Metadata {
	return Metadata{
		CommitHash: CommitHash,
		BuildTime:  BuildTime,
		GoVersion:  GoVersion,
		BuildUser:  BuildUser,
	}
}
