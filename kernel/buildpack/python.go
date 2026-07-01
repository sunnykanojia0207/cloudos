package buildpack

import "context"

// PythonBuildpack detects Python projects by checking for requirements.txt,
// setup.py, setup.cfg, or Pipfile.
type PythonBuildpack struct{}

func (bp *PythonBuildpack) Name() string { return "python" }

func (bp *PythonBuildpack) Version() string { return "1.0.0" }

func (bp *PythonBuildpack) Detect(ctx context.Context, src Source) (bool, error) {
	return fileExists(src, "requirements.txt") ||
		fileExists(src, "setup.py") ||
		fileExists(src, "setup.cfg") ||
		fileExists(src, "Pipfile"), nil
}

func (bp *PythonBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	installCmd := "pip install -r requirements.txt"
	if fileExists(src, "Pipfile") {
		installCmd = "pipenv install"
	}

	startCmd := "python app.py"
	if fileExists(src, "manage.py") {
		startCmd = "python manage.py runserver 0.0.0.0:{port}"
	} else if fileExists(src, "wsgi.py") {
		startCmd = "gunicorn wsgi:app --bind 0.0.0.0:{port}"
	} else if fileExists(src, "app.py") {
		startCmd = "python app.py"
	} else if fileExists(src, "main.py") {
		startCmd = "python main.py"
	}

	return &BuildPlan{
		BuildpackName: "python",
		RuntimeType:   RuntimePython,
		Name:          "Python",
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    installCmd,
		BuildCmd:      "",
		StartCmd:      startCmd,
		OutputDir:     "",
		DevPort:       8000,
		Source:        src,
		EnvVars: map[string]string{
			"PYTHONUNBUFFERED": "1",
		},
	}, nil
}

func (bp *PythonBuildpack) Build(ctx context.Context, plan *BuildPlan) (*BuildResult, error) {
	artifact := ArtifactFromPlan(plan, plan.Source.Path)
	return &BuildResult{
		Artifact:    artifact,
		RuntimeType: plan.RuntimeType,
		Metadata: map[string]string{
			"language": "python",
		},
	}, nil
}
