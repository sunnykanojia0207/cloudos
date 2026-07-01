package buildpack

import (
	"context"
	"strings"
)

// NextJSBuildpack detects Next.js projects by checking for the "next" dependency.
type NextJSBuildpack struct{}

func (bp *NextJSBuildpack) Name() string { return "nextjs" }

func (bp *NextJSBuildpack) Version() string { return "1.0.0" }

func (bp *NextJSBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	if !fileExists(src, "package.json") {
		return false, nil
	}

	pkg, err := readPackageJSON(src)
	if err != nil {
		return false, nil
	}

	return isNextJS(pkg), nil
}

func (bp *NextJSBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	pkg, _ := readPackageJSON(src)
	buildCmd := ""
	if pkg != nil && pkg.Scripts.Build != "" {
		buildCmd = pkg.Scripts.Build
	}

	return &BuildPlan{
		BuildpackName: "nextjs",
		RuntimeType:   RuntimeNextJS,
		Name:          "Next.js",
		Version:       getVersion(pkg),
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    "npm install",
		BuildCmd:      buildCmd,
		StartCmd:      "npm start",
		OutputDir:     ".next",
		DevPort:       3000,
		Source:        src,
	}, nil
}

func (bp *NextJSBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	outputPath := plan.Source.Path
	if plan.OutputDir != "" {
		outputPath = strings.TrimSuffix(plan.Source.Path, "/") + "/" + strings.TrimPrefix(plan.OutputDir, "/")
	}
	artifact := ArtifactFromPlan(plan, outputPath)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"language":    "node.js",
			"framework":   "next.js",
			"output_dir": plan.OutputDir,
		},
	}, nil
}

// isNextJS checks if a package.json indicates a Next.js project.
func isNextJS(pkg *PackageJSON) bool {
	if _, ok := pkg.Dependencies["next"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["next"]; ok {
		return true
	}
	return false
}
