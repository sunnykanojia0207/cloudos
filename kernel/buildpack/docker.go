package buildpack

import "context"

// DockerBuildpack detects Dockerfile-based projects.
type DockerBuildpack struct{}

func (bp *DockerBuildpack) Name() string { return "docker" }

func (bp *DockerBuildpack) Version() string { return "1.0.0" }

func (bp *DockerBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	return fileExists(src, "Dockerfile"), nil
}

func (bp *DockerBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	return &BuildPlan{
		BuildpackName: "docker",
		RuntimeType:   RuntimeDocker,
		Name:          "Docker",
		ArtifactType:  ArtifactTypeImage,
		InstallCmd:    "",
		BuildCmd:      "docker build -t {app} .",
		StartCmd:      "docker run -p {port}:{port} {app}",
		OutputDir:     "",
		DevPort:       0, // Port defined by Dockerfile EXPOSE
		Source:        src,
	}, nil
}

func (bp *DockerBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	artifact := ArtifactFromPlan(plan, plan.Source.Path)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"type": "docker",
		},
	}, nil
}
