package buildpack

import "context"

// StaticBuildpack detects static HTML sites.
// It is the fallback buildpack — always matches.
type StaticBuildpack struct{}

func (bp *StaticBuildpack) Name() string { return "static" }

func (bp *StaticBuildpack) Version() string { return "1.0.0" }

func (bp *StaticBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	return true, nil
}

func (bp *StaticBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	return &BuildPlan{
		BuildpackName: "static",
		RuntimeType:   RuntimeStatic,
		Name:          "Static Website",
		ArtifactType:  ArtifactTypeStatic,
		InstallCmd:    "",
		BuildCmd:      "",
		StartCmd:      "",
		OutputDir:     "",
		DevPort:       80,
		Source:        src,
	}, nil
}

func (bp *StaticBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	artifact := ArtifactFromPlan(plan, plan.Source.Path)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"type": "static",
		},
	}, nil
}
