package buildpack

import (
	"context"
	"strings"
)

// ReactBuildpack detects React projects (CRA, Vite, or custom React with build scripts).
type ReactBuildpack struct{}

func (bp *ReactBuildpack) Name() string { return "react" }

func (bp *ReactBuildpack) Version() string { return "1.0.0" }

func (bp *ReactBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	if !fileExists(src, "package.json") {
		return false, nil
	}

	pkg, err := readPackageJSON(src)
	if err != nil {
		return false, nil
	}

	return isReact(pkg), nil
}

func (bp *ReactBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	pkg, _ := readPackageJSON(src)
	buildCmd := ""
	outputDir := "build"
	version := ""
	if pkg != nil {
		if pkg.Engines.Node != "" {
			version = "React + Node " + pkg.Engines.Node
		} else {
			version = pkg.Version
		}
		if pkg.Scripts.Build != "" {
			buildCmd = pkg.Scripts.Build
		}
		// CRA uses "build", Vite uses "dist"
		if _, hasVite := pkg.DevDependencies["vite"]; hasVite {
			outputDir = "dist"
		}
	}

	return &BuildPlan{
		BuildpackName: "react",
		RuntimeType:   RuntimeReact,
		Name:          "React",
		Version:       version,
		ArtifactType:  ArtifactTypeStatic,
		InstallCmd:    "npm install",
		BuildCmd:      buildCmd,
		StartCmd:      "npx serve -s " + outputDir + " -l {port}",
		OutputDir:     outputDir,
		DevPort:       3000,
		Source:        src,
		EnvVars: map[string]string{
			"NODE_ENV": "production",
		},
	}, nil
}

func (bp *ReactBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
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
			"framework":   "react",
			"output_dir": plan.OutputDir,
		},
	}, nil
}

// isReact checks if a package.json indicates a React project.
func isReact(pkg *PackageJSON) bool {
	hasReact := false
	if _, ok := pkg.Dependencies["react"]; ok {
		hasReact = true
	}
	if _, ok := pkg.DevDependencies["react"]; ok {
		hasReact = true
	}
	if !hasReact {
		return false
	}

	// Don't match Next.js projects (they have their own buildpack).
	if _, ok := pkg.Dependencies["next"]; ok {
		return false
	}
	if _, ok := pkg.DevDependencies["next"]; ok {
		return false
	}

	// Common CRA or Vite React projects have react-scripts or vite.
	if _, ok := pkg.Dependencies["react-scripts"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["vite"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["@vitejs/plugin-react"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["react-scripts"]; ok {
		return true
	}

	// If react is a dependency but no build tool is detected, check for a build script.
	if pkg.Scripts.Build != "" {
		return true
	}

	return false
}
