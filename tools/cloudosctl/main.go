// CloudOS CLI — a command-line client for the CloudOS API.
//
// Commands:
//
//	doctor    Check environment for CloudOS requirements
//	logs      Stream, snapshot, or download application logs
//	deploy    Deploy an application from a git repository
//	ps        List running applications
//	open      Open an application in the default browser
//	status    Show application dashboard
//	timeline  Show deployment timeline
//	compare   Compare two deployments
//	version   Show CloudOS version information
//
// Usage:
//
//	cloudosctl doctor
//	cloudosctl logs <app-id> [--follow] [--tail N]
//	cloudosctl logs <app-id> --download
//	cloudosctl deploy <git-url>
//	cloudosctl ps
//	cloudosctl open <app-id>
//	cloudosctl status <app-id> [--json] [--watch]
//	cloudosctl timeline <app-id> [--number N]
//	cloudosctl compare <app-id> <from-number> <to-number>
//	cloudosctl version
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cloudos/cloudos/packages/version"
)

// ── Config ─────────────────────────────────────────────────────────────────

const defaultAPIAddr = "http://localhost:8080"

// ── API Types ──────────────────────────────────────────────────────────────

// LogEvent is the structured log entry from the API.
type LogEvent struct {
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"`
	Level     string `json:"level"`
	Step      string `json:"step,omitempty"`
	Message   string `json:"message"`
}

// APIResponse is the standard API envelope.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ── Timeline & Compare Types ─────────────────────────────────────────────────

// TimelineStep is a single step in the deployment timeline.
type TimelineStep struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Action string `json:"action"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// TimelineResponse is the full deployment timeline.
type TimelineResponse struct {
	Application      string         `json:"application"`
	DeploymentNumber int            `json:"deploymentNumber"`
	WorkflowID       string         `json:"workflowId"`
	OverallStatus    string         `json:"overallStatus"`
	StartedAt        string         `json:"startedAt,omitempty"`
	CompletedAt      string         `json:"completedAt,omitempty"`
	Duration         string         `json:"duration,omitempty"`
	Steps            []TimelineStep `json:"steps"`
}

// DeploymentSummary is a compact view of a single deployment.
type DeploymentSummary struct {
	DeploymentNumber int      `json:"deploymentNumber"`
	StartedAt        string   `json:"startedAt,omitempty"`
	CompletedAt      string   `json:"completedAt,omitempty"`
	Duration         string   `json:"duration,omitempty"`
	Repository       string   `json:"repository,omitempty"`
	Branch           string   `json:"branch,omitempty"`
	CommitSHA        string   `json:"commitSha,omitempty"`
	DetectedRuntime  string   `json:"detectedRuntime,omitempty"`
	Buildpack        string   `json:"buildpack,omitempty"`
	BuildSuccess     bool     `json:"buildSuccess"`
	RuntimeName      string   `json:"runtimeName,omitempty"`
	Environment      string   `json:"environment,omitempty"`
	ArtifactType     string   `json:"artifactType,omitempty"`
	HealthStatus     string   `json:"healthStatus"`
	Endpoint         string   `json:"endpoint,omitempty"`
	WorkflowSteps    int      `json:"workflowSteps"`
	Errors           []string `json:"errors,omitempty"`
}

// NodeComparison compares a single workflow step between two deployments.
type NodeComparison struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Action     string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	FromResult string `json:"fromResult,omitempty"`
	ToResult   string `json:"toResult,omitempty"`
	FromError  string `json:"fromError,omitempty"`
	ToError    string `json:"toError,omitempty"`
	Changed    bool   `json:"changed"`
}

// ComparisonSummary highlights what changed between two deployments.
type ComparisonSummary struct {
	StatusChanged    bool   `json:"statusChanged"`
	HealthChanged    bool   `json:"healthChanged"`
	DurationChanged  bool   `json:"durationChanged"`
	DurationDiff     string `json:"durationDiff,omitempty"`
	CommitChanged    bool   `json:"commitChanged"`
	BuildChanged     bool   `json:"buildChanged"`
	TotalStepsMatch  bool   `json:"totalStepsMatch"`
	ChangedNodeCount int    `json:"changedNodeCount"`
}

// ComparisonResponse is the full comparison between two deployments.
type ComparisonResponse struct {
	From           DeploymentSummary `json:"from"`
	To             DeploymentSummary `json:"to"`
	NodeComparison []NodeComparison  `json:"nodeComparison,omitempty"`
	Summary        ComparisonSummary `json:"summary"`
}

// ── Main ───────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	apiAddr := os.Getenv("CLOUDOS_API")
	if apiAddr == "" {
		apiAddr = defaultAPIAddr
	}

	cmd := os.Args[1]

	switch cmd {
	case "doctor":
		doctor()
	case "logs":
		runLogs(apiAddr, os.Args[2:])
	case "deploy":
		runDeploy(apiAddr, os.Args[2:])
	case "ps":
		runPS(apiAddr, os.Args[2:])
	case "status":
		runStatus(apiAddr, os.Args[2:])
	case "open":
		runOpen(apiAddr, os.Args[2:])
	case "timeline":
		runTimeline(apiAddr, os.Args[2:])
	case "compare":
		runCompare(apiAddr, os.Args[2:])
	case "version", "--version":
		fmt.Println(version.Full())
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`CloudOS CLI — manage applications on CloudOS

Usage:

  cloudosctl doctor                       Check environment readiness

  cloudosctl logs <app-id>               Show recent logs (tail 50)
  cloudosctl logs <app-id> -f            Stream logs in real-time
  cloudosctl logs <app-id> -n 100        Show last 100 lines
  cloudosctl logs <app-id> -d            Download all logs as text

  cloudosctl deploy <git-url>            Deploy an application from git

  cloudosctl ps                          List all applications

  cloudosctl open <app-id>               Open application in browser

  cloudosctl status <app-id>             Show application dashboard
  cloudosctl status <app-id> --json      Output as JSON
  cloudosctl status <app-id> --watch     Live-updating dashboard

  cloudosctl timeline <app-id>           Show latest deployment timeline
  cloudosctl timeline <app-id> -n 2      Show timeline for deployment #2

  cloudosctl compare <app> 41 42         Compare deployments #41 and #42

  cloudosctl version                     Show version information

Environment:

  CLOUDOS_API  API server address (default: http://localhost:8080)`)
}

// ── logs command ───────────────────────────────────────────────────────────

func runLogs(apiAddr string, args []string) {
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	follow := fs.Bool("f", false, "Follow log output (stream)")
	tail := fs.Int("n", 50, "Number of lines to show")
	download := fs.Bool("d", false, "Download all logs as text")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	appID := fs.Arg(0)
	if appID == "" {
		fmt.Fprintln(os.Stderr, "Error: application ID is required")
		fmt.Fprintln(os.Stderr, "Usage: cloudosctl logs <app-id> [-f] [-n N] [-d]")
		os.Exit(1)
	}

	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *download {
		downloadLogs(apiAddr, appID)
		return
	}

	if *follow {
		streamLogs(apiAddr, appID, *tail)
		return
	}

	snapshotLogs(apiAddr, appID, *tail)
}

// snapshotLogs fetches and displays recent logs.
func snapshotLogs(apiAddr, appID string, tail int) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/logs?tail=%d", apiAddr, appID, tail)

	resp, err := httpGet(nil, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		printAPIError(resp)
		return
	}

	apiResp, err := decodeAPIResponse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := checkAPIError(apiResp, resp.StatusCode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var events []LogEvent
	if err := json.Unmarshal(apiResp.Data, &events); err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to read log data from server response.\n")
		fmt.Fprintf(os.Stderr, "  The server returned unexpected data. Try: cloudosctl doctor\n")
		os.Exit(1)
	}

	if len(events) == 0 {
		fmt.Println("No logs available.")
		return
	}

	for _, event := range events {
		printLogEvent(event)
	}
}

// streamLogs streams logs via SSE.
func streamLogs(apiAddr, appID string, tail int) {
	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s/logs/stream?tail=%d", apiAddr, appID, tail)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to create request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("User-Agent", "cloudosctl/"+versionFull())

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		printAPIErrorRaw(resp)
		return
	}

	// Parse SSE events.
	scanner := bufio.NewScanner(resp.Body)
	var dataBuf string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			dataBuf = strings.TrimPrefix(line, "data: ")
			continue
		}

		if line == "" && dataBuf != "" {
			// Empty line means end of event.
			var event LogEvent
			if err := json.Unmarshal([]byte(dataBuf), &event); err == nil {
				printLogEvent(event)
			}
			dataBuf = ""
		}
	}

	// Check for scanner errors (e.g., connection closed unexpectedly).
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: log stream disconnected: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run: cloudosctl logs %s -f\n", appID)
	}
}

// downloadLogs downloads and saves logs to a file.
func downloadLogs(apiAddr, appID string) {
	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s/logs/download", apiAddr, appID)

	resp, err := httpGet(nil, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		printAPIError(resp)
		return
	}

	// Use sanitized filename to prevent path traversal (B11).
	filename := sanitizeLogFilename(appID)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file %q: %v\n", filename, err)
		os.Exit(1)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded %d bytes to %s\n", written, filename)
}

// ── deploy command ─────────────────────────────────────────────────────────

// deployMetadata is the metadata portion of the deploy request payload.
type deployMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// deploySource is the source specification in a deploy request.
type deploySource struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// deploySpec is the spec portion of the deploy request payload.
type deploySpec struct {
	Source   deploySource            `json:"source"`
	Settings map[string]string       `json:"settings"`
}

// deployRequest is the full deploy request payload.
type deployRequest struct {
	Metadata deployMetadata `json:"metadata"`
	Spec     deploySpec     `json:"spec"`
}

func runDeploy(apiAddr string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: git repository URL is required")
		fmt.Fprintln(os.Stderr, "Usage: cloudosctl deploy <git-url>")
		os.Exit(1)
	}

	repoURL := args[0]
	appName := extractAppName(repoURL)

	// Build the application payload using typed structs.
	req := deployRequest{
		Metadata: deployMetadata{
			ID:   appName,
			Name: appName,
			Kind: "Application",
		},
		Spec: deploySpec{
			Source: deploySource{
				Type: "git",
				URL:  repoURL,
			},
			Settings: map[string]string{
				"autoDeploy": "true",
			},
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building deploy request: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/v1/resources/Application", apiAddr)
	resp, err := httpPost(nil, url, "application/json", string(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Printf("Deploying %s from %s...\n", appName, repoURL)
		fmt.Printf("Watch logs: cloudosctl logs %s -f\n", appName)

		// Wait for deployment to complete, then offer to open the browser.
		promptOpenAfterDeploy(apiAddr, appName)
	} else {
		printAPIError(resp)
	}
}

// ── ps command ─────────────────────────────────────────────────────────────

func runPS(apiAddr string, args []string) {
	url := fmt.Sprintf("%s/api/v1/resources/Application", apiAddr)

	resp, err := httpGet(nil, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		printAPIError(resp)
		return
	}

	apiResp, err := decodeAPIResponse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := checkAPIError(apiResp, resp.StatusCode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Try to parse as a list of resources.
	// Use safe parsing — never panic on unexpected API responses (B10).
	var resources []map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &resources); err != nil {
		// Might be a single resource.
		var res map[string]interface{}
		if err2 := json.Unmarshal(apiResp.Data, &res); err2 == nil {
			resources = []map[string]interface{}{res}
		} else {
			fmt.Println("No applications found.")
			return
		}
	}

	if len(resources) == 0 {
		fmt.Println("No applications found.")
		return
	}

	// Print table header.
	fmt.Printf("%-24s %-12s %-12s %-24s\n", "ID", "PHASE", "HEALTH", "URL")
	fmt.Println(strings.Repeat("-", 80))

	for _, r := range resources {
		// Safe type assertions — never panic on unexpected data (B10 fix).
		// The comma-ok pattern returns nil map if the type doesn't match,
		// and getStringField handles nil maps safely.
		meta, _ := r["metadata"].(map[string]interface{})
		status, _ := r["status"].(map[string]interface{})

		id := getStringField(meta, "id")
		phase := getStringField(status, "phase")
		health := getStringField(status, "health")
		url := getStringField(status, "url")

		fmt.Printf("%-24s %-12s %-12s %-24s\n", id, phase, health, url)
	}
}

// ── timeline command ────────────────────────────────────────────────────────

func runTimeline(apiAddr string, args []string) {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	number := fs.Int("n", 0, "Deployment number (default: latest)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	appID := fs.Arg(0)
	if appID == "" {
		fmt.Fprintln(os.Stderr, "Error: application ID is required")
		fmt.Fprintln(os.Stderr, "Usage: cloudosctl timeline <app-id> [-n N]")
		os.Exit(1)
	}

	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Determine the URL. If no number specified, find the latest deployment first.
	var timelineURL string
	if *number > 0 {
		timelineURL = fmt.Sprintf("%s/api/v1/applications/%s/deployments/%d/timeline", apiAddr, appID, *number)
	} else {
		// Get the latest deployment number from the application.
		latestNum := getLatestDeploymentNumber(apiAddr, appID)
		if latestNum < 1 {
			fmt.Fprintln(os.Stderr, "Error: no deployments found for this application")
			fmt.Fprintln(os.Stderr, "This application has not been deployed yet.")
			os.Exit(1)
		}
		timelineURL = fmt.Sprintf("%s/api/v1/applications/%s/deployments/%d/timeline", apiAddr, appID, latestNum)
	}

	resp, err := httpGet(nil, timelineURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		printAPIError(resp)
		return
	}

	apiResp, err := decodeAPIResponse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := checkAPIError(apiResp, resp.StatusCode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var timeline TimelineResponse
	if err := json.Unmarshal(apiResp.Data, &timeline); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing timeline: %v\n", err)
		os.Exit(1)
	}

	// Print header.
	fmt.Printf("Timeline: Deployment #%d\n", timeline.DeploymentNumber)
	fmt.Printf("Status:   %s\n", timeline.OverallStatus)
	if timeline.Duration != "" {
		fmt.Printf("Duration: %s\n", timeline.Duration)
	}
	if timeline.StartedAt != "" {
		fmt.Printf("Started:  %s\n", timeline.StartedAt)
	}
	if timeline.CompletedAt != "" {
		fmt.Printf("Ended:    %s\n", timeline.CompletedAt)
	}
	fmt.Printf("Workflow: %s\n", timeline.WorkflowID)
	fmt.Println()

	if len(timeline.Steps) == 0 {
		fmt.Println("No workflow steps recorded.")
		return
	}

	fmt.Println("Steps:")
	for _, step := range timeline.Steps {
		icon := "✓"
		detail := step.Result
		switch step.Status {
		case "succeeded":
			icon = "✓"
			detail = step.Result
		case "failed":
			icon = "✗"
			if step.Error != "" {
				detail = step.Error
			} else {
				detail = "Failed"
			}
		case "running":
			icon = "◌"
			detail = "In progress..."
		case "skipped":
			icon = "→"
			detail = "Skipped"
		case "cancelled":
			icon = "⊘"
			detail = "Cancelled"
		default:
			icon = "?"
			detail = step.Status
		}
		if detail == "" {
			detail = step.Status
		}
		fmt.Printf("  %s %s\n", icon, step.Name)
		if detail != "" {
			fmt.Printf("    %s\n", detail)
		}
	}
}

// ── compare command ──────────────────────────────────────────────────────────

func runCompare(apiAddr string, args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: application ID, from-number, and to-number are required")
		fmt.Fprintln(os.Stderr, "Usage: cloudosctl compare <app-id> <from-number> <to-number>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  cloudosctl compare my-app 1 2")
		os.Exit(1)
	}

	appID := args[0]
	fromStr := args[1]
	toStr := args[2]

	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate integer arguments (B49 fix).
	fromNum, err := parseDeploymentNumber(fromStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid from-number %q — must be a positive integer.\n", fromStr)
		fmt.Fprintf(os.Stderr, "Usage: cloudosctl compare <app-id> <from-number> <to-number>\n")
		os.Exit(1)
	}
	toNum, err := parseDeploymentNumber(toStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid to-number %q — must be a positive integer.\n", toStr)
		fmt.Fprintf(os.Stderr, "Usage: cloudosctl compare <app-id> <from-number> <to-number>\n")
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s/deployments/compare?from=%d&to=%d", apiAddr, appID, fromNum, toNum)

	resp, err := httpGet(nil, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		printAPIError(resp)
		return
	}

	apiResp, err := decodeAPIResponse(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := checkAPIError(apiResp, resp.StatusCode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var comp ComparisonResponse
	if err := json.Unmarshal(apiResp.Data, &comp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing comparison: %v\n", err)
		os.Exit(1)
	}

	// ── Section header ──
	fmt.Printf("Compare #%d vs #%d\n", fromNum, toNum)
	fmt.Println()

	// ── Diff: only what changed ──
	fmt.Println("Changes:")
	anyChange := false

	if comp.Summary.CommitChanged {
		anyChange = true
		fmt.Printf("  Commit\n")
		fmt.Printf("    #%d  %s\n", fromNum, comp.From.CommitSHA)
		fmt.Printf("    ↓\n")
		fmt.Printf("    #%d  %s  ✓ Changed\n", toNum, comp.To.CommitSHA)
	}
	if comp.Summary.DurationChanged {
		anyChange = true
		fmt.Printf("  Duration\n")
		fmt.Printf("    #%d  %s\n", fromNum, comp.From.Duration)
		fmt.Printf("    ↓\n")
		fmt.Printf("    #%d  %s  ✓ Changed\n", toNum, comp.To.Duration)
	}
	if comp.Summary.HealthChanged {
		anyChange = true
		fmt.Printf("  Health\n")
		fmt.Printf("    #%d  %s\n", fromNum, comp.From.HealthStatus)
		fmt.Printf("    ↓\n")
		fmt.Printf("    #%d  %s  ✓ Changed\n", toNum, comp.To.HealthStatus)
	}
	if comp.Summary.BuildChanged {
		anyChange = true
		fmt.Printf("  Build\n")
		fmt.Printf("    #%d  %s\n", fromNum, comp.From.DetectedRuntime)
		fmt.Printf("    ↓\n")
		fmt.Printf("    #%d  %s  ✓ Changed\n", toNum, comp.To.DetectedRuntime)
	}
	if !comp.Summary.TotalStepsMatch {
		anyChange = true
		fmt.Printf("  Steps\n")
		fmt.Printf("    #%d  %d\n", fromNum, comp.From.WorkflowSteps)
		fmt.Printf("    ↓\n")
		fmt.Printf("    #%d  %d  ✓ Changed\n", toNum, comp.To.WorkflowSteps)
	}

	if !anyChange {
		fmt.Println("  No differences found between these deployments.")
	}
	fmt.Println()

	// ── Unchanged values (compact) ──
	fmt.Println("Unchanged:")
	if !comp.Summary.CommitChanged {
		fmt.Printf("  Commit   %s  (no change)\n", comp.From.CommitSHA)
	}
	if !comp.Summary.DurationChanged {
		fmt.Printf("  Duration %s  (no change)\n", comp.From.Duration)
	}
	if !comp.Summary.HealthChanged {
		fmt.Printf("  Health   %s  (no change)\n", comp.From.HealthStatus)
	}
	if !comp.Summary.BuildChanged {
		fmt.Printf("  Runtime  %s  (no change)\n", comp.From.DetectedRuntime)
	}
	if comp.Summary.TotalStepsMatch {
		fmt.Printf("  Steps    %d  (no change)\n", comp.From.WorkflowSteps)
	}
	fmt.Println()

	// ── Step comparison ──
	if len(comp.NodeComparison) > 0 {
		changedNodes := 0
		for _, nc := range comp.NodeComparison {
			if nc.Changed {
				changedNodes++
			}
		}

		if changedNodes > 0 {
			fmt.Printf("Step Changes (%d changed):\n", changedNodes)
			for _, nc := range comp.NodeComparison {
				if !nc.Changed {
					continue
				}
				fromLabel := nc.FromResult
				if fromLabel == "" {
					fromLabel = nc.FromStatus
				}
				toLabel := nc.ToResult
				if toLabel == "" {
					toLabel = nc.ToStatus
				}
				fmt.Printf("  %s (%s)\n", nc.Name, nc.Action)
				fmt.Printf("    #%d  %s\n", fromNum, truncate(fromLabel, 50))
				fmt.Printf("    ↓\n")
				fmt.Printf("    #%d  %s\n", toNum, truncate(toLabel, 50))
			}
		} else {
			fmt.Println("All steps identical between deployments.")
		}
		fmt.Println()
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

func printLogEvent(event LogEvent) {
	ts := event.Timestamp
	// Trim to time-only for readability.
	if len(ts) > 19 {
		ts = ts[11:19] // HH:MM:SS
	}

	var source string
	if event.Source == "runtime" {
		source = "App"
	} else {
		source = strings.ToUpper(event.Source[:1]) + event.Source[1:]
	}

	var levelIcon string
	switch event.Level {
	case "error":
		levelIcon = "✗"
	case "warn":
		levelIcon = "⚠"
	default:
		levelIcon = "•"
	}

	var step string
	if event.Step != "" {
		step = " [" + event.Step + "]"
	}

	fmt.Printf("%s %s %s%s %s\n", ts, levelIcon, source+step, "", event.Message)
}

func getStringField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func extractAppName(repoURL string) string {
	// Extract name from git URL: https://github.com/user/my-app.git → my-app
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		if name == "" && len(parts) > 1 {
			name = parts[len(parts)-2]
		}
		return name
	}
	return "app"
}

func printAPIError(resp *http.Response) {
	apiResp, err := decodeAPIResponse(resp.Body)
	if err == nil && apiResp.Error != nil {
		fmt.Fprintf(os.Stderr, "Server error [%d]: %s — %s\n", resp.StatusCode, apiResp.Error.Code, apiResp.Error.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Server error [%d]\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "The server returned an unexpected response. Try:\n")
		fmt.Fprintf(os.Stderr, "  cloudosctl doctor\n")
	}
	os.Exit(1)
}

func printAPIErrorRaw(resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	apiResp, err := decodeAPIErrorResponse(body)
	if err == nil && apiResp.Error != nil {
		fmt.Fprintf(os.Stderr, "Server error [%d]: %s — %s\n", resp.StatusCode, apiResp.Error.Code, apiResp.Error.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Server error [%d]\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "The server returned: %s\n", string(body))
		fmt.Fprintf(os.Stderr, "Try: cloudosctl doctor\n")
	}
	os.Exit(1)
}

// decodeAPIErrorResponse decodes an API error response from raw bytes.
func decodeAPIErrorResponse(body []byte) (*APIResponse, error) {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp, nil
}

// getLatestDeploymentNumber fetches the application resource and returns the
// latest deployment number from its status.
func getLatestDeploymentNumber(apiAddr, appID string) int {
	url := fmt.Sprintf("%s/api/v1/resources/Application/%s", apiAddr, appID)
	resp, err := httpGet(nil, url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0
	}

	apiResp, err := decodeAPIResponse(resp.Body)
	if err != nil {
		return 0
	}
	if !apiResp.Success {
		return 0
	}

	var obj struct {
		Status struct {
			DeploymentCount   int `json:"deploymentCount"`
		} `json:"status"`
	}
	if err := json.Unmarshal(apiResp.Data, &obj); err != nil {
		return 0
	}
	return obj.Status.DeploymentCount
}

// truncate truncates a string to the given max length, adding "..." if truncated.
func truncate(s string, max int) string {
	if max < 1 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// parseDeploymentNumber validates and parses a deployment number string.
// Returns an error if the string is not a positive integer.
func parseDeploymentNumber(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty deployment number")
	}
	// Reject negative signs, hex, octal, etc.
	if s[0] == '-' || s[0] == '+' {
		return 0, fmt.Errorf("not a positive integer")
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a positive integer")
		}
		n = n*10 + int(c-'0')
	}
	if n < 1 {
		return 0, fmt.Errorf("deployment number must be >= 1")
	}
	return n, nil
}

