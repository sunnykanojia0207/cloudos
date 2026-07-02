package buildpack

import (
	"bufio"
	"context"
	"strings"
)

// GoBuildpack detects Go projects by checking for go.mod.
type GoBuildpack struct{}

func (bp *GoBuildpack) Name() string { return "go" }

func (bp *GoBuildpack) Version() string { return "1.0.0" }

func (bp *GoBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	return fileExists(src, "go.mod"), nil
}

func (bp *GoBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	version := detectGoVersion(src)
	return &BuildPlan{
		BuildpackName: "go",
		RuntimeType:   RuntimeGo,
		Name:          "Go",
		Version:       version,
		ArtifactType:  ArtifactTypeBinary,
		InstallCmd:    "go mod download",
		BuildCmd:      "go build -o app .",
		StartCmd:      "./app",
		OutputDir:     "",
		DevPort:       8080,
		Source:        src,
	}, nil
}

// detectGoVersion reads the go version from go.mod's "go 1.xx" directive.
func detectGoVersion(src Source) string {
	f, err := readFile(src, "go.mod")
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "go"))
		}
	}
	return ""
}

func (bp *GoBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	artifact := ArtifactFromPlan(plan, plan.Source.Path)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"language": "go",
		},
	}, nil
}
