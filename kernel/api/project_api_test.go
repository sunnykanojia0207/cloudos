package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudos/cloudos/kernel/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Project resource API endpoints
// ---------------------------------------------------------------------------

func TestProjectAPI_ListProjects(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/projects")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "ProjectList", data["kind"])
	assert.Equal(t, "cloudos.io/v1", data["apiVersion"])

	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	// No projects created yet, but the list should still succeed.
	assert.GreaterOrEqual(t, len(items), 0)
}

func TestProjectAPI_CreateProject(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	body := map[string]string{
		"id":          "test-project",
		"displayName": "Test Project",
		"environment": "development",
		"description": "A project created via API",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 201, resp.StatusCode)

	var env api.Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	obj := env.Data.(map[string]interface{})
	assert.Equal(t, "Project", obj["kind"])
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "test-project", meta["id"])
	assert.Equal(t, "Test Project", meta["name"])

	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "Test Project", spec["displayName"])
	assert.Equal(t, "development", spec["environment"])

	status := obj["status"].(map[string]interface{})
	assert.Equal(t, "Creating", status["phase"])

	// Verify the project is listable.
	code, _ := httpGet(t, ts.URL+"/api/v1/projects")
	assert.Equal(t, 200, code)
}

func TestProjectAPI_CreateProject_InvalidJSON(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader([]byte("{invalid")))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var env api.Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
}

func TestProjectAPI_CreateProject_MissingID(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	body := map[string]string{
		"displayName": "No ID",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestProjectAPI_CreateProject_MissingName(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	body := map[string]string{
		"id": "no-name",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestProjectAPI_CreateProject_Duplicate(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	body := map[string]string{
		"id":          "dup-project",
		"displayName": "Duplicate",
		"environment": "development",
	}
	jsonBody, _ := json.Marshal(body)

	// First create — should succeed.
	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 201, resp.StatusCode)

	// Second create — should fail with 400 (already exists).
	resp, err = http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestProjectAPI_GetProject(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Create a project first.
	body := map[string]string{
		"id":          "get-test",
		"displayName": "Get Test",
		"environment": "staging",
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Now get it.
	code, env := httpGet(t, ts.URL+"/api/v1/projects/get-test")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	obj := env.Data.(map[string]interface{})
	assert.Equal(t, "Project", obj["kind"])
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "get-test", meta["id"])
	assert.Equal(t, "Get Test", meta["name"])

	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "staging", spec["environment"])
}

func TestProjectAPI_GetProject_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/projects/nonexistent")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "RESOURCE_NOT_FOUND", env.Error.Code)
}

func TestProjectAPI_UpdateProject(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Create first.
	createBody := map[string]string{
		"id":          "update-test",
		"displayName": "Original Name",
		"environment": "development",
	}
	jsonBody, _ := json.Marshal(createBody)
	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Update.
	updateBody := map[string]string{
		"displayName": "Updated Name",
		"environment": "production",
		"description": "Updated description",
	}
	updateJSON, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/projects/update-test", bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var env api.Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.True(t, env.Success)

	obj := env.Data.(map[string]interface{})
	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "Updated Name", spec["displayName"])
	assert.Equal(t, "production", spec["environment"])
	assert.Equal(t, "Updated description", spec["description"])
}

func TestProjectAPI_UpdateProject_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	updateBody := map[string]string{"displayName": "Nope"}
	updateJSON, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/projects/nonexistent", bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestProjectAPI_DeleteProject(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Create first.
	body := map[string]string{
		"id":          "delete-test",
		"displayName": "Delete Me",
		"environment": "development",
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Delete.
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/projects/delete-test", nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 204, resp.StatusCode)

	// Verify it's gone.
	code, _ := httpGet(t, ts.URL+"/api/v1/projects/delete-test")
	assert.Equal(t, 404, code)
}

func TestProjectAPI_DeleteProject_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/projects/nonexistent", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestProjectAPI_NoConflictWithOtherEndpoints(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Projects work.
	code, _ := httpGet(t, ts.URL+"/api/v1/projects")
	assert.Equal(t, 200, code)

	// Controllers work.
	code, _ = httpGet(t, ts.URL+"/api/v1/controllers")
	assert.Equal(t, 200, code)

	// Capabilities work.
	code, _ = httpGet(t, ts.URL+"/api/v1/capabilities")
	assert.Equal(t, 200, code)

	// Resources work.
	code, _ = httpGet(t, ts.URL+"/api/v1/resources")
	assert.Equal(t, 200, code)

	// Providers work.
	code, _ = httpGet(t, ts.URL+"/api/v1/providers")
	assert.Equal(t, 200, code)
}

// Test full lifecycle: create → get → update → list → delete.
func TestProjectAPI_FullLifecycle(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// 1. Create.
	body := map[string]interface{}{
		"id":          "lifecycle-test",
		"displayName": "Lifecycle Project",
		"environment": "development",
		"description": "Testing full lifecycle",
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 201, resp.StatusCode)

	// 2. Get.
	code, env := httpGet(t, ts.URL+"/api/v1/projects/lifecycle-test")
	assert.Equal(t, 200, code)
	obj := env.Data.(map[string]interface{})
	assert.Equal(t, "Lifecycle Project", obj["spec"].(map[string]interface{})["displayName"])

	// 3. Update.
	updateBody := map[string]string{"displayName": "Updated Lifecycle"}
	updateJSON, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/projects/lifecycle-test", bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// 4. List (should include our project).
	code, env = httpGet(t, ts.URL+"/api/v1/projects")
	assert.Equal(t, 200, code)
	list := env.Data.(map[string]interface{})
	items := list["items"].([]interface{})
	found := false
	for _, item := range items {
		obj := item.(map[string]interface{})
		meta := obj["metadata"].(map[string]interface{})
		if meta["id"] == "lifecycle-test" {
			found = true
			assert.Equal(t, "Updated Lifecycle", obj["spec"].(map[string]interface{})["displayName"])
		}
	}
	assert.True(t, found, "project should appear in list")

	// 5. Delete.
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/projects/lifecycle-test", nil)
	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 204, resp.StatusCode)

	// Verify deleted.
	code, _ = httpGet(t, ts.URL+"/api/v1/projects/lifecycle-test")
	assert.Equal(t, 404, code)
}


