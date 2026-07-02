package buildpack

import "context"

// NodeBuildpack detects generic Node.js projects by checking for package.json.
// It is checked after NextJSBuildpack and ReactBuildpack in the detection chain,
// so it only matches projects that aren't Next.js or React.
type NodeBuildpack struct{}

func (bp *NodeBuildpack) Name() string { return "node" }

func (bp *NodeBuildpack) Version() string { return "1.0.0" }

func (bp *NodeBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	return fileExists(src, "package.json"), nil
}

func (bp *NodeBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	pkg, err := readPackageJSON(src)
	if err != nil {
		return defaultNodePlan(src), nil
	}

	buildCmd := pkg.Scripts.Build
	startCmd := pkg.Scripts.Start
	if startCmd == "" {
		startCmd = "npm start"
	}

	// Detect Node.js runtime version from engines.node in package.json
	nodeVersion := pkg.Engines.Node
	if nodeVersion == "" {
		nodeVersion = pkg.Version // fallback to package version
	}

	return &BuildPlan{
		BuildpackName: "node",
		RuntimeType:   RuntimeNode,
		Name:          "Node.js",
		Version:       nodeVersion,
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    "npm install",
		BuildCmd:      buildCmd,
		StartCmd:      startCmd,
		OutputDir:     "",
		DevPort:       3000,
		Source:        src,
		EnvVars: map[string]string{
			"NODE_ENV": "production",
		},
	}, nil
}

func (bp *NodeBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	outputPath := plan.Source.Path
	if plan.OutputDir != "" {
		outputPath = plan.Source.Path + "/" + plan.OutputDir
	}
	artifact := ArtifactFromPlan(plan, outputPath)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"language": "node.js",
		},
	}, nil
}

func defaultNodePlan(src Source) *BuildPlan {
	return &BuildPlan{
		BuildpackName: "node",
		RuntimeType:   RuntimeNode,
		Name:          "Node.js",
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    "npm install",
		BuildCmd:      "",
		StartCmd:      "npm start",
		DevPort:       3000,
		Source:        src,
		EnvVars: map[string]string{
			"NODE_ENV": "production",
		},
	}
}
