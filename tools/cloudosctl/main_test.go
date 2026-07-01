package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ── App ID Sanitization Tests ───────────────────────────────────────────────
//
// These tests verify that sanitizeAppID rejects path traversal and
// other unsafe input patterns (B11 fix).

func TestSanitizeAppID_ValidIDs(t *testing.T) {
	valid := []string{
		"my-app",
		"my_app",
		"myapp",
		"my-app-123",
		"my.app",
		"a",
		"12345",
	}
	for _, id := range valid {
		if err := sanitizeAppID(id); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", id, err)
		}
	}
}

func TestSanitizeAppID_Empty(t *testing.T) {
	if err := sanitizeAppID(""); err == nil {
		t.Error("expected error for empty app ID")
	}
}

func TestSanitizeAppID_PathTraversal(t *testing.T) {
	traversal := []string{
		"../etc/passwd",
		"../../etc/passwd",
		"..\\windows\\system32",
		"..%2f..%2fetc",
	}
	for _, id := range traversal {
		if err := sanitizeAppID(id); err == nil {
			t.Errorf("expected error for path traversal %q", id)
		}
	}
}

func TestSanitizeAppID_AbsolutePaths(t *testing.T) {
	absolute := []string{
		"/etc/passwd",
		"/var/log",
		"\\windows\\system32",
	}
	for _, id := range absolute {
		if err := sanitizeAppID(id); err == nil {
			t.Errorf("expected error for absolute path %q", id)
		}
	}
}

func TestSanitizeAppID_WindowsDriveLetter(t *testing.T) {
	drives := []string{
		"C:",
		"D:",
		"C:foo",
		"D:bar",
	}
	for _, id := range drives {
		if err := sanitizeAppID(id); err == nil {
			t.Errorf("expected error for drive letter %q", id)
		}
	}
}

func TestSanitizeAppID_EncodedTraversal(t *testing.T) {
	encoded := []string{
		"%2e%2e%2f",
		"%2e%2e/",
		"foo%2fbar",
		"%5c..%5c",
	}
	for _, id := range encoded {
		if err := sanitizeAppID(id); err == nil {
			t.Errorf("expected error for encoded traversal %q", id)
		}
	}
}

func TestSanitizeAppID_SpecialChars(t *testing.T) {
	invalid := []string{
		"my app",
		"my;app",
		"my|app",
		"my$app",
		"my`app",
		"-myapp",
		".myapp",
	}
	for _, id := range invalid {
		if err := sanitizeAppID(id); err == nil {
			t.Errorf("expected error for special chars %q", id)
		}
	}
}

// ── Log Filename Sanitization Tests ─────────────────────────────────────────

func TestSanitizeLogFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-app", "my-app.log"},
		{"hello", "hello.log"},
		{"../etc/passwd", "__etc_passwd.log"},
		{"C:foo", "C_foo.log"},
		{"a/b/c", "a_b_c.log"},
	}
	for _, tc := range tests {
		result := sanitizeLogFilename(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeLogFilename(%q) = %q, want %q", tc.input, result, tc.expected)
		}
		// Verify no path separators in result.
		if strings.Contains(result, "/") || strings.Contains(result, "\\") || strings.Contains(result, "..") {
			t.Errorf("sanitizeLogFilename(%q) = %q still contains path traversal", tc.input, result)
		}
	}
}

// ── Deployment Number Validation Tests ──────────────────────────────────────
//
// These tests verify that parseDeploymentNumber rejects invalid input (B49 fix).

func TestParseDeploymentNumber_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},
		{"42", 42},
		{"100", 100},
		{"999999", 999999},
	}
	for _, tc := range tests {
		got, err := parseDeploymentNumber(tc.input)
		if err != nil {
			t.Errorf("parseDeploymentNumber(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseDeploymentNumber(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseDeploymentNumber_Invalid(t *testing.T) {
	invalid := []string{
		"",     // empty
		"0",    // zero
		"-1",   // negative
		"+1",   // positive sign
		"abc",  // letters
		"1.5",  // decimal
		"0xFF", // hex
		" 1",   // leading space
		"1 ",   // trailing space
	}
	for _, input := range invalid {
		_, err := parseDeploymentNumber(input)
		if err == nil {
			t.Errorf("expected error for parseDeploymentNumber(%q)", input)
		}
	}
}

// ── Safe Type Assertion Tests ───────────────────────────────────────────────
//
// These tests verify that the ps command doesn't panic on malformed API
// responses (B10 fix).

func TestGetStringField_NilMap(t *testing.T) {
	result := getStringField(nil, "key")
	if result != "" {
		t.Errorf("getStringField(nil, \"key\") = %q, want empty string", result)
	}
}

func TestGetStringField_MissingKey(t *testing.T) {
	m := map[string]interface{}{"other": "value"}
	result := getStringField(m, "key")
	if result != "" {
		t.Errorf("getStringField(m, \"key\") = %q, want empty string", result)
	}
}

func TestGetStringField_WrongType(t *testing.T) {
	m := map[string]interface{}{"key": 42}
	result := getStringField(m, "key")
	if result != "" {
		t.Errorf("getStringField(m, \"key\") = %q, want empty string", result)
	}
}

func TestGetStringField_Valid(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	result := getStringField(m, "key")
	if result != "value" {
		t.Errorf("getStringField(m, \"key\") = %q, want \"value\"", result)
	}
}

// ── API Response Parsing Tests ──────────────────────────────────────────────

func TestDecodeAPIResponse_Valid(t *testing.T) {
	body := strings.NewReader(`{"success":true,"data":{"foo":"bar"}}`)
	resp, err := decodeAPIResponse(body)
	if err != nil {
		t.Fatalf("decodeAPIResponse unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestDecodeAPIResponse_MalformedJSON(t *testing.T) {
	body := strings.NewReader(`{invalid json`)
	_, err := decodeAPIResponse(body)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestDecodeAPIResponse_MissingFields(t *testing.T) {
	body := strings.NewReader(`{}`)
	resp, err := decodeAPIResponse(body)
	if err != nil {
		t.Fatalf("decodeAPIResponse unexpected error: %v", err)
	}
	if resp.Success {
		t.Error("expected success to be false for empty object")
	}
}

func TestDecodeAPIResponse_EmptyBody(t *testing.T) {
	body := strings.NewReader(``)
	_, err := decodeAPIResponse(body)
	if err == nil {
		t.Error("expected error for empty body")
	}
}

// ── API Error Check Tests ───────────────────────────────────────────────────

func TestCheckAPIError_Success(t *testing.T) {
	resp := &APIResponse{Success: true}
	err := checkAPIError(resp, 200)
	if err != nil {
		t.Errorf("unexpected error for success response: %v", err)
	}
}

func TestCheckAPIError_WithError(t *testing.T) {
	resp := &APIResponse{
		Success: false,
		Error: &struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: "NOT_FOUND", Message: "resource not found"},
	}
	err := checkAPIError(resp, 404)
	if err == nil {
		t.Error("expected error for failed response")
	}
	if !strings.Contains(err.Error(), "NOT_FOUND") {
		t.Errorf("error should contain error code, got: %v", err)
	}
}

func TestCheckAPIError_NoErrorDetails(t *testing.T) {
	resp := &APIResponse{Success: false}
	err := checkAPIError(resp, 500)
	if err == nil {
		t.Error("expected error for failed response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain HTTP status code, got: %v", err)
	}
}

// ── HTTP Client Tests ───────────────────────────────────────────────────────

func TestHTTPClient_TimeoutSet(t *testing.T) {
	if httpClient.Timeout != defaultHTTPTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", httpClient.Timeout, defaultHTTPTimeout)
	}
}

func TestHTTPClient_UserAgent(t *testing.T) {
	// Verify that the versionFull function returns something sensible.
	ua := versionFull()
	if ua == "" {
		t.Error("versionFull() should not return empty string")
	}
}

// ── Parse Deployment Number ─────────────────────────────────────────────────

func TestParseDeploymentNumber_EdgeCases(t *testing.T) {
	_, err := parseDeploymentNumber("999999999999999999999999999")
	if err == nil {
		t.Error("expected error for overflow-sized number")
	}
}

// ── Extract App Name Tests ──────────────────────────────────────────────────

func TestExtractAppName(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/my-app.git", "my-app"},
		{"https://github.com/user/my-app", "my-app"},
		{"git@github.com:user/my-app.git", "my-app"},
		{"", ""},
	}
	for _, tc := range tests {
		got := extractAppName(tc.url)
		if got != tc.want {
			t.Errorf("extractAppName(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

// ── JSON Encoding/Decoding Consistency ──────────────────────────────────────

func TestDeployRequestJSON(t *testing.T) {
	// Verify the deploy request serializes correctly (B03 fix).
	req := deployRequest{
		Metadata: deployMetadata{
			ID:   "test-app",
			Name: "test-app",
			Kind: "Application",
		},
		Spec: deploySpec{
			Source: deploySource{
				Type: "git",
				URL:  "https://github.com/user/repo.git",
			},
			Settings: map[string]string{
				"autoDeploy": "true",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var decoded deployRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if decoded.Metadata.ID != "test-app" {
		t.Errorf("decoded ID = %q, want %q", decoded.Metadata.ID, "test-app")
	}
	if decoded.Spec.Source.URL != "https://github.com/user/repo.git" {
		t.Errorf("decoded URL = %q, want %q", decoded.Spec.Source.URL, "https://github.com/user/repo.git")
	}
}

// ── Truncate Tests ──────────────────────────────────────────────────────────

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}
	for _, tc := range tests {
		got := truncate(tc.input, tc.max)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
		}
	}
}

// ── Ensure no unused variables ──────────────────────────────────────────────

func TestConstants(t *testing.T) {
	if defaultAPIAddr != "http://localhost:8080" {
		t.Errorf("defaultAPIAddr = %q, want %q", defaultAPIAddr, "http://localhost:8080")
	}
	if defaultHTTPTimeout != 30e9 {
		t.Errorf("defaultHTTPTimeout = %v, want 30s", defaultHTTPTimeout)
	}
}

// Compile-time check that APIResponse can be decoded.
var _ = json.Unmarshal
var _ *http.Response
