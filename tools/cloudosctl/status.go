package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// ═════════════════════════════════════════════════════════════════════════════
// cloudosctl status — Terminal Dashboard
// ═════════════════════════════════════════════════════════════════════════════

// separator is a full-width line used to visually separate sections.
var separator = "  " + strings.Repeat("─", 74)

// runStatus fetches and displays the full status of an application.
//
// Usage:
//
//	cloudosctl status <application>
//	cloudosctl status <application> --json
//	cloudosctl status <application> --watch
func runStatus(apiAddr string, args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	watchMode := fs.Bool("watch", false, "Refresh every 2 seconds")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	appID := fs.Arg(0)
	if appID == "" {
		fmt.Fprintln(os.Stderr, "Error: application name is required")
		fmt.Fprintln(os.Stderr, "Usage: cloudosctl status <application> [--json] [--watch]")
		os.Exit(1)
	}

	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *watchMode {
		watchStatus(apiAddr, appID, *jsonOutput)
		return
	}

	// Single render.
	data, err := fetchAppResource(apiAddr, appID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		printJSON(data)
		return
	}

	renderStatus(data)
}

// ═════════════════════════════════════════════════════════════════════════════
// Watch Mode
// ═════════════════════════════════════════════════════════════════════════════

// watchStatus refreshes the status display every 2 seconds.
func watchStatus(apiAddr, appID string, jsonMode bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Clear on first render.
	clearScreen()

	for range ticker.C {
		data, err := fetchAppResource(apiAddr, appID)
		if err != nil {
			clearScreen()
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		clearScreen()

		if jsonMode {
			printJSON(data)
			fmt.Println()
			fmt.Println("(watching — Ctrl+C to stop)")
			continue
		}

		renderStatus(data)
		fmt.Println()
		fmt.Println("  (watching — Ctrl+C to stop)")
	}
}

// clearScreen sends an ANSI clear sequence to refresh the terminal.
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// ═════════════════════════════════════════════════════════════════════════════
// API Fetch
// ═════════════════════════════════════════════════════════════════════════════

// appResource is the parsed structure of the application resource response.
type appResource struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	Metadata   resourceMeta `json:"metadata"`
	Spec       appSpec     `json:"spec"`
	Status     appStatus   `json:"status"`
}

type resourceMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type appSpec struct {
	Source struct {
		Type   string `json:"type"`
		URL    string `json:"url"`
		Branch string `json:"branch,omitempty"`
	} `json:"source"`
	Runtime struct {
		Type string `json:"type"`
	} `json:"runtime"`
	Settings map[string]string `json:"settings,omitempty"`
}

type appStatus struct {
	Phase               string            `json:"phase"`
	Health              string            `json:"health"`
	URL                 string            `json:"url"`
	DeploymentCount     int               `json:"deploymentCount"`
	CurrentDeploymentID string            `json:"currentDeploymentId,omitempty"`
	LastReport          *deployReport     `json:"lastReport,omitempty"`
}

type deployReport struct {
	DeploymentNumber int      `json:"deploymentNumber"`
	StartedAt        string   `json:"startedAt"`
	CompletedAt      string   `json:"completedAt"`
	Duration         string   `json:"duration"`
	Repository       string   `json:"repository"`
	Branch           string   `json:"branch"`
	CommitSHA        string   `json:"commitSha,omitempty"`
	DetectedRuntime  string   `json:"detectedRuntime,omitempty"`
	Buildpack        string   `json:"buildpack,omitempty"`
	BuildSuccess     bool     `json:"buildSuccess"`
	RuntimeName      string   `json:"runtimeName,omitempty"`
	RuntimeVersion   string   `json:"runtimeVersion,omitempty"`
	Environment      string   `json:"environment,omitempty"`
	ArtifactType     string   `json:"artifactType,omitempty"`
	HealthStatus     string   `json:"healthStatus"`
	Endpoint         string   `json:"endpoint"`
	WorkflowID       string   `json:"workflowId"`
	WorkflowSteps    int      `json:"workflowSteps"`
	Warnings         []string `json:"warnings,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

// fetchAppResource calls the API and returns the parsed application resource.
func fetchAppResource(apiAddr, appID string) (*appResource, error) {
	url := fmt.Sprintf("%s/api/v1/resources/Application/%s", apiAddr, appID)
	resp, err := httpGet(nil, url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to CloudOS API at %s: %w", apiAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("application %q not found.\nCheck the application name: cloudosctl ps", appID)
	}

	if resp.StatusCode != 200 {
		apiResp, decodeErr := decodeAPIResponse(resp.Body)
		if decodeErr == nil && apiResp.Error != nil {
			return nil, fmt.Errorf("server error [%d]: %s — %s", resp.StatusCode, apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("server error [%d]", resp.StatusCode)
	}

	apiResp, err := decodeAPIResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read server response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("server error: %s — %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("server request was not successful")
	}

	var data appResource
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("unable to parse application data from server: %w", err)
	}

	return &data, nil
}

// ═════════════════════════════════════════════════════════════════════════════
// Output
// ═════════════════════════════════════════════════════════════════════════════

// printJSON serializes the data as indented JSON.
func printJSON(data *appResource) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

// renderStatus prints the human-readable status dashboard.
func renderStatus(data *appResource) {
	r := data.Status.LastReport

	// ── Section 1: Application Identity ─────────────────────────────────
	fmt.Println(separator)
	fmt.Println()
	fmt.Printf("  Application\n")
	fmt.Printf("  %s\n", data.Metadata.ID)
	fmt.Println()
	fmt.Println(separator)

	// ── Section 2: Status & Health ──────────────────────────────────────
	fmt.Println()
	fmt.Printf("  Status\n")
	printPhase(data.Status.Phase)
	fmt.Println()
	fmt.Printf("  Health\n")
	printHealth(data.Status.Health)
	fmt.Println()

	// ── Section 3: Runtime Info (from last report) ──────────────────────
	if r != nil {
		if r.RuntimeName != "" {
			fmt.Printf("  Runtime\n")
			fmt.Printf("  %s\n", r.RuntimeName)
			fmt.Println()
		}
		if r.RuntimeVersion != "" {
			fmt.Printf("  Runtime Version\n")
			fmt.Printf("  %s\n", r.RuntimeVersion)
			fmt.Println()
		}
	}

	if data.Status.URL != "" {
		fmt.Printf("  URL\n")
		fmt.Printf("  %s\n", data.Status.URL)
		fmt.Println()
	}

	// ── Section 4: Deployment Details (from last report) ────────────────
	if r != nil {
		fmt.Println(separator)
		fmt.Println()

		if r.Repository != "" {
			fmt.Printf("  Repository\n")
			fmt.Printf("  %s\n", r.Repository)
			fmt.Println()
		}
		if r.Branch != "" {
			fmt.Printf("  Branch\n")
			fmt.Printf("  %s\n", r.Branch)
			fmt.Println()
		}
		if r.CommitSHA != "" {
			fmt.Printf("  Commit\n")
			fmt.Printf("  %s\n", r.CommitSHA)
			fmt.Println()
		}
		if r.DetectedRuntime != "" {
			fmt.Printf("  Detected Runtime\n")
			fmt.Printf("  %s\n", r.DetectedRuntime)
			fmt.Println()
		}
		if r.Buildpack != "" {
			fmt.Printf("  Buildpack\n")
			fmt.Printf("  %s\n", r.Buildpack)
			fmt.Println()
		}

		fmt.Printf("  Deployment\n")
		fmt.Printf("  #%d\n", r.DeploymentNumber)
		fmt.Println()

		if r.WorkflowID != "" {
			fmt.Printf("  Workflow\n")
			fmt.Printf("  %s\n", r.WorkflowID)
			fmt.Println()
		}

		if r.Duration != "" {
			fmt.Printf("  Duration\n")
			fmt.Printf("  %s\n", r.Duration)
			fmt.Println()
		}

		// Format timestamps.
		started := formatTimestamp(r.StartedAt)
		completed := formatTimestamp(r.CompletedAt)
		if started != "" {
			fmt.Printf("  Started\n")
			fmt.Printf("  %s\n", started)
			fmt.Println()
		}
		if completed != "" {
			fmt.Printf("  Completed\n")
			fmt.Printf("  %s\n", completed)
			fmt.Println()
		}
	}

	// ── Section 5: Deployment Summary ───────────────────────────────────
	fmt.Println(separator)
	fmt.Println()
	fmt.Println("  Deployment Summary")
	fmt.Println()

	if r != nil {
		if r.BuildSuccess && r.HealthStatus == "Healthy" {
			fmt.Printf("    Latest deployment: ✓ Success\n")
		} else if !r.BuildSuccess {
			fmt.Printf("    Latest deployment: ✗ Failed\n")
		} else {
			fmt.Printf("    Latest deployment: ⚠ %s\n", r.HealthStatus)
		}

		fmt.Printf("    Health:           %s\n", healthLabel(data.Status.Health))
		if len(r.Warnings) > 0 {
			fmt.Printf("    Warnings:         %d\n", len(r.Warnings))
			for _, w := range r.Warnings {
				fmt.Printf("                       • %s\n", w)
			}
		} else {
			fmt.Printf("    Warnings:         0\n")
		}
		if len(r.Errors) > 0 {
			fmt.Printf("    Errors:           %d\n", len(r.Errors))
			for _, e := range r.Errors {
				fmt.Printf("                       • %s\n", e)
			}
		} else {
			fmt.Printf("    Errors:           0\n")
		}
		fmt.Printf("    Workflow Steps:    %d\n", r.WorkflowSteps)
		fmt.Printf("    Total Deployments: %d\n", data.Status.DeploymentCount)
	} else {
		fmt.Printf("    Latest deployment: %s\n", phaseLabel(data.Status.Phase))
		fmt.Printf("    Health:           %s\n", healthLabel(data.Status.Health))
		fmt.Printf("    Total Deployments: %d\n", data.Status.DeploymentCount)
	}

	fmt.Println()

	// ── Section 6: Failure Details ──────────────────────────────────────
	if r != nil && (!r.BuildSuccess || len(r.Errors) > 0) {
		fmt.Println(separator)
		fmt.Println()
		fmt.Println("  Deployment Failed")
		fmt.Println()

		if !r.BuildSuccess {
			fmt.Println("    Step: Build")
		}
		if len(r.Errors) > 0 {
			fmt.Println("    Reason:")
			for _, e := range r.Errors {
				fmt.Printf("      • %s\n", e)
			}
		}
		fmt.Println()
		fmt.Println("    Next Steps")
		fmt.Printf("      cloudosctl logs %s -f\n", data.Metadata.ID)
		fmt.Printf("      cloudosctl timeline %s\n", data.Metadata.ID)
		fmt.Println()
	}

	// ── Section 7: Available Commands ───────────────────────────────────
	fmt.Println(separator)
	fmt.Println()
	fmt.Println("  Available Commands")
	fmt.Println()
	fmt.Printf("    cloudosctl logs %s -f\n", data.Metadata.ID)
	fmt.Printf("    cloudosctl timeline %s\n", data.Metadata.ID)
	fmt.Printf("    cloudosctl open %s\n", data.Metadata.ID)
	if r != nil && r.DeploymentNumber > 1 {
		fmt.Printf("    cloudosctl compare %s %d %d\n",
			data.Metadata.ID, r.DeploymentNumber-1, r.DeploymentNumber)
	}
	fmt.Println()
	fmt.Println(separator)
}

// ═════════════════════════════════════════════════════════════════════════════
// Formatting Helpers
// ═════════════════════════════════════════════════════════════════════════════

// printPhase prints the phase with an appropriate icon.
func printPhase(phase string) {
	switch phase {
	case "Running":
		fmt.Printf("  ✓ %s\n", phase)
	case "Failed", "Error":
		fmt.Printf("  ✗ %s\n", phase)
	case "Creating", "Deploying":
		fmt.Printf("  ◌ %s\n", phase)
	default:
		fmt.Printf("  • %s\n", phase)
	}
}

// printHealth prints the health with an appropriate icon.
func printHealth(health string) {
	switch health {
	case "Healthy":
		fmt.Printf("  ✓ %s\n", health)
	case "Degraded":
		fmt.Printf("  ⚠ %s\n", health)
	case "Error":
		fmt.Printf("  ✗ %s\n", health)
	default:
		fmt.Printf("  • %s\n", health)
	}
}

// healthLabel returns a short label for the health status.
func healthLabel(health string) string {
	switch health {
	case "Healthy":
		return "✓ Healthy"
	case "Degraded":
		return "⚠ Degraded"
	case "Error":
		return "✗ Error"
	default:
		return health
	}
}

// phaseLabel returns a short label for the phase.
func phaseLabel(phase string) string {
	switch phase {
	case "Running":
		return "✓ Running"
	case "Failed", "Error":
		return "✗ Failed"
	case "Creating":
		return "◌ Creating"
	case "Deploying":
		return "◌ Deploying"
	default:
		return phase
	}
}

// formatTimestamp attempts to parse an RFC3339 timestamp and format it
// as a human-readable date-time string. Returns the original on failure.
func formatTimestamp(ts string) string {
	if ts == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Try other common formats.
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", ts)
		if err != nil {
			return ts
		}
	}
	return t.Format("2006-01-02 15:04:05")
}
