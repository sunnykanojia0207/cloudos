package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// ═════════════════════════════════════════════════════════════════════════════
// cloudosctl open — Browser Launch
// ═════════════════════════════════════════════════════════════════════════════

// runOpen resolves an application and opens its URL in the default browser.
//
// Usage:
//
//	cloudosctl open <application>
func runOpen(apiAddr string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: application name is required")
		fmt.Fprintln(os.Stderr, "Usage: cloudosctl open <application>")
		os.Exit(1)
	}

	appID := args[0]

	if err := sanitizeAppID(appID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Fetch application details from the API.
	url := fmt.Sprintf("%s/api/v1/resources/Application/%s", apiAddr, appID)
	resp, err := httpGet(nil, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to connect to CloudOS API at %s.\n", apiAddr)
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the CloudOS kernel is running:\n")
		fmt.Fprintf(os.Stderr, "  go run ./tools/cloudos\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Fprintf(os.Stderr, "Application %q not found.\n", appID)
		fmt.Fprintf(os.Stderr, "Check the application name: cloudosctl ps\n")
		os.Exit(1)
	}

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

	// Parse the application resource.
	var app struct {
		Metadata struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			Phase  string `json:"phase"`
			Health string `json:"health"`
			URL    string `json:"url"`
		} `json:"status"`
	}
	if err := json.Unmarshal(apiResp.Data, &app); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing application data: %v\n", err)
		os.Exit(1)
	}

	// Validate we have a URL.
	if app.Status.URL == "" {
		fmt.Fprintf(os.Stderr, "Application %q has no URL yet. It may still be deploying.\n", appID)
		fmt.Fprintf(os.Stderr, "Run: cloudosctl logs %s -f\n", appID)
		os.Exit(1)
	}

	// Check health.
	if app.Status.Health != "Healthy" || app.Status.Phase != "Running" {
		fmt.Println("Application is not healthy.")
		fmt.Println()
		fmt.Printf("Current status: %s / %s\n", app.Status.Phase, app.Status.Health)
		fmt.Println()
		fmt.Printf("Run: cloudosctl status %s\n", appID)
		fmt.Println("for more information.")
		os.Exit(1)
	}

	// Open the URL in the default browser.
	fmt.Printf("✓ Opening application...\n\n")
	fmt.Println(app.Status.URL)

	if err := openURL(app.Status.URL); err != nil {
		fmt.Println()
		fmt.Fprintf(os.Stderr, "Unable to launch browser.\n")
		fmt.Fprintf(os.Stderr, "Application URL:\n  %s\n", app.Status.URL)
		fmt.Fprintf(os.Stderr, "Open this URL manually.\n")
		os.Exit(1)
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// Platform-Specific Browser Launch
// ═════════════════════════════════════════════════════════════════════════════

// openURL opens the given URL in the user's default browser.
// It auto-detects the host platform.
func openURL(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		// Linux and everything else.
		return exec.Command("xdg-open", url).Start()
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// Post-Deployment Experience
// ═════════════════════════════════════════════════════════════════════════════

// waitForDeployment polls the application status until the deployment completes
// or a timeout is reached. It returns the application's URL and health status.
// Supports Ctrl+C to cancel the wait (B28 fix).
func waitForDeployment(apiAddr, appID string) (url string, healthy bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Handle Ctrl+C during wait.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", false
		case <-ticker.C:
			appURL := fmt.Sprintf("%s/api/v1/resources/Application/%s", apiAddr, appID)
			resp, err := httpGet(ctx, appURL)
			if err != nil {
				continue
			}

			apiResp, err := decodeAPIResponse(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			if !apiResp.Success {
				continue
			}

			var app struct {
				Status struct {
					Phase  string `json:"phase"`
					Health string `json:"health"`
					URL    string `json:"url"`
				} `json:"status"`
			}
			if err := json.Unmarshal(apiResp.Data, &app); err != nil {
				continue
			}

			if app.Status.URL != "" && app.Status.Phase == "Running" {
				return app.Status.URL, app.Status.Health == "Healthy"
			}

			// Check for failure.
			if app.Status.Phase == "Failed" || app.Status.Phase == "Error" {
				return app.Status.URL, false
			}
		}
	}
}

// promptOpenAfterDeploy waits for deployment completion and asks the user
// if they want to open the application in their browser.
func promptOpenAfterDeploy(apiAddr, appID string) {
	fmt.Println()
	fmt.Print("Waiting for deployment to complete...")

	url, healthy := waitForDeployment(apiAddr, appID)

	fmt.Println()

	if url == "" {
		fmt.Println()
		fmt.Println("Deployment is taking longer than expected.")
		fmt.Printf("Watch logs: cloudosctl logs %s -f\n", appID)
		return
	}

	if !healthy {
		fmt.Println()
		fmt.Println("⚠ Deployment completed with issues.")
		fmt.Printf("Run: cloudosctl status %s\n", appID)
		return
	}

	fmt.Println()
	fmt.Println("✓ Deployment completed.")
	fmt.Println()
	fmt.Printf("  Application: %s\n", appID)
	fmt.Printf("  URL:         %s\n", url)
	fmt.Println()

	// Interactive prompt.
	fmt.Print("  Open in browser? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" {
		fmt.Println()
		if err := openURL(url); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to launch browser.\n")
			fmt.Fprintf(os.Stderr, "Application URL:\n  %s\n", url)
			fmt.Fprintf(os.Stderr, "Open this URL manually.\n")
		}
	} else {
		fmt.Println()
		fmt.Printf("  Open manually: cloudosctl open %s\n", appID)
	}
}
