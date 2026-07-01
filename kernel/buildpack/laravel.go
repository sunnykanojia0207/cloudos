package buildpack

import (
	"context"
)

// LaravelBuildpack detects PHP/Laravel projects by checking for composer.json.
type LaravelBuildpack struct{}

func (bp *LaravelBuildpack) Name() string { return "laravel" }

func (bp *LaravelBuildpack) Version() string { return "1.0.0" }

func (bp *LaravelBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	return fileExists(src, "composer.json"), nil
}

func (bp *LaravelBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	installCmd := "composer install"
	startCmd := "php artisan serve --host=0.0.0.0 --port={port}"
	outputDir := ""

	if _, err := readComposerJSON(src); err != nil {
		return defaultLaravelPlan(src), nil
	}

	if fileExists(src, "artisan") {
		outputDir = "public"
	}

	return &BuildPlan{
		BuildpackName: "laravel",
		RuntimeType:   RuntimeLaravel,
		Name:          "Laravel",
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    installCmd,
		BuildCmd:      "",
		StartCmd:      startCmd,
		OutputDir:     outputDir,
		DevPort:       8000,
		Source:        src,
		EnvVars: map[string]string{
			"APP_ENV": "local",
		},
	}, nil
}

func (bp *LaravelBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	outputPath := plan.Source.Path
	if plan.OutputDir != "" {
		outputPath = plan.Source.Path + "/" + plan.OutputDir
	}
	artifact := ArtifactFromPlan(plan, outputPath)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"language": "php",
			"framework": "laravel",
		},
	}, nil
}

func defaultLaravelPlan(src Source) *BuildPlan {
	return &BuildPlan{
		BuildpackName: "laravel",
		RuntimeType:   RuntimeLaravel,
		Name:          "PHP",
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    "composer install",
		BuildCmd:      "",
		StartCmd:      "php artisan serve --host=0.0.0.0 --port={port}",
		DevPort:       8000,
		Source:        src,
	}
}
