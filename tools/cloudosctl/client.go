package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ── Shared HTTP Client ──────────────────────────────────────────────────────
//
// All CLI commands use this shared client so that:
//   - A 30-second timeout applies to every request (B18 fix)
//   - A consistent User-Agent identifies CLI requests
//   - Retry/backoff can be added centrally in the future

// defaultHTTPTimeout is the maximum wait time for any single HTTP request.
const defaultHTTPTimeout = 30 * time.Second

// httpClient is the shared HTTP client used by all CLI commands.
// It has a 30-second timeout to prevent hangs when the API is unreachable.
var httpClient = &http.Client{
	Timeout: defaultHTTPTimeout,
	Transport: &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

// httpGet performs an HTTP GET with the shared client and context.
// If the context is nil, context.Background() is used.
func httpGet(ctx context.Context, url string) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "cloudosctl/"+versionFull())
	return httpClient.Do(req)
}

// httpPost performs an HTTP POST with the shared client.
func httpPost(ctx context.Context, url, contentType, body string) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "cloudosctl/"+versionFull())
	req.Header.Set("Content-Type", contentType)
	return httpClient.Do(req)
}

// ── Safe API Response Parsing ───────────────────────────────────────────────

// decodeAPIResponse decodes an API response from a reader.
// Returns a friendly error on malformed JSON or unexpected structure.
func decodeAPIResponse(r io.Reader) (*APIResponse, error) {
	var apiResp APIResponse
	if err := json.NewDecoder(r).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("unable to parse server response: %w\n"+
			"The server returned an unexpected format. Try:\n"+
			"  cloudosctl doctor", err)
	}
	return &apiResp, nil
}

// checkAPIError validates the API response and returns a user-friendly error.
func checkAPIError(apiResp *APIResponse, statusCode int) error {
	if !apiResp.Success {
		if apiResp.Error != nil {
			return fmt.Errorf("server error [%d]: %s — %s\n"+
				"This may indicate a temporary issue. Try:\n"+
				"  cloudosctl doctor",
				statusCode, apiResp.Error.Code, apiResp.Error.Message)
		}
		return fmt.Errorf("server error [%d]: request was not successful\n"+
			"Try running cloudosctl doctor to check system health.", statusCode)
	}
	return nil
}

// versionFull returns the version string for User-Agent header.
func versionFull() string {
	return "0.6.0-rc1"
}
