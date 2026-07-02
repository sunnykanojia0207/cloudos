package buildpack

import (
	"context"
	"runtime"
	"strings"
)

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

// venvPython returns the relative path to the Python interpreter inside a venv,
// and the command to create the venv, based on the current platform.
// On Windows: .\venv\Scripts\python.exe
// On Unix:    ./venv/bin/python
func venvPaths() (pythonPath, pipPath, createVenvCmd string) {
	const venvDir = "venv"
	if runtime.GOOS == "windows" {
		pythonPath = ".\\" + venvDir + "\\Scripts\\python.exe"
		pipPath = ".\\" + venvDir + "\\Scripts\\pip"
		createVenvCmd = "python -m venv " + venvDir
	} else {
		pythonPath = "./" + venvDir + "/bin/python"
		pipPath = "./" + venvDir + "/bin/pip"
		createVenvCmd = "python3 -m venv " + venvDir
	}
	return
}

// hasDependency checks if a Python package name appears in a requirements file.
func hasDependency(src Source, filename, packageName string) bool {
	if !fileExists(src, filename) {
		return false
	}
	data, err := readFile(src, filename)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, packageName)
}

func (bp *PythonBuildpack) Plan(ctx context.Context, src Source) (*BuildPlan, error) {
	// Determine venv paths based on platform.
	pythonPath, pipPath, createVenvCmd := venvPaths()

	// Build install command: create venv, then install dependencies inside it.
	installCmd := createVenvCmd + " && " + pipPath + " install -r requirements.txt"
	if fileExists(src, "Pipfile") {
		installCmd = createVenvCmd + " && " + pipPath + " install pipenv && " + strings.Replace(pipPath, "pip", "pipenv", 1) + " install"
	}

	// Build start command using the venv Python interpreter.
	startCmd := pythonPath + " app.py"
	if fileExists(src, "manage.py") {
		startCmd = pythonPath + " manage.py runserver 0.0.0.0:{port}"
	} else if fileExists(src, "wsgi.py") {
		// Check for uvicorn or gunicorn in requirements.
		if hasDependency(src, "requirements.txt", "uvicorn") {
			startCmd = pythonPath + " -m uvicorn wsgi:app --host 0.0.0.0 --port {port}"
		} else {
			startCmd = pythonPath + " -m gunicorn wsgi:app --bind 0.0.0.0:{port}"
		}
	} else if fileExists(src, "app.py") {
		startCmd = pythonPath + " app.py"
	} else if fileExists(src, "main.py") {
		startCmd = pythonPath + " main.py"
	}

	// Detect Python version from runtime.txt or Pipfile
	version := detectPythonVersion(src)

	return &BuildPlan{
		BuildpackName: "python",
		RuntimeType:   RuntimePython,
		Name:          "Python",
		Version:       version,
		ArtifactType:  ArtifactTypeSource,
		InstallCmd:    installCmd,
		BuildCmd:      "",
		StartCmd:      startCmd,
		OutputDir:     "",
		DevPort:       8000,
		Source:        src,
		EnvVars: map[string]string{
			"PYTHONUNBUFFERED":      "1",
			"PYTHONDONTWRITEBYTECODE": "1",
		},
	}, nil
}

// detectPythonVersion reads the Python version from runtime.txt or Pipfile.
func detectPythonVersion(src Source) string {
	// Check runtime.txt (common on Heroku/deploy platforms)
	if fileExists(src, "runtime.txt") {
		data, err := readFile(src, "runtime.txt")
		if err == nil {
			content := strings.TrimSpace(string(data))
			if strings.HasPrefix(content, "python-") || strings.HasPrefix(content, "python") {
				return strings.TrimPrefix(strings.TrimPrefix(content, "python-"), "python")
			}
			return content
		}
	}

	// Check Pipfile for python_version
	if fileExists(src, "Pipfile") {
		data, err := readFile(src, "Pipfile")
		if err == nil {
			content := string(data)
			if idx := strings.Index(content, "python_version"); idx >= 0 {
				rest := content[idx:]
				if eqIdx := strings.Index(rest, "="); eqIdx >= 0 {
					ver := strings.TrimSpace(rest[eqIdx+1:])
					ver = strings.Trim(ver, "\"")
					return ver
				}
			}
		}
	}

	return ""
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
