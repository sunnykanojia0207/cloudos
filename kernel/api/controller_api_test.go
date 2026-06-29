package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Controller Runtime API endpoints
// ---------------------------------------------------------------------------

func TestControllerAPI_ListControllers(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/controllers")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	controllers, ok := data["controllers"].([]interface{})
	require.True(t, ok, "expected 'controllers' array in response")

	// At minimum, the "namespace" controller should be registered at boot.
	found := false
	for _, c := range controllers {
		entry := c.(map[string]interface{})
		if entry["name"] == "namespace" {
			found = true
			assert.Equal(t, "Namespace", entry["kind"])
			assert.Contains(t, []string{"running", "stopped"}, entry["state"])
		}
	}
	assert.True(t, found, "expected namespace controller to be registered")
}

func TestControllerAPI_ListControllersTotal(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/controllers")

	data := env.Data.(map[string]interface{})
	controllers := data["controllers"].([]interface{})
	assert.GreaterOrEqual(t, len(controllers), 1)

	total, ok := data["total"].(float64)
	require.True(t, ok)
	assert.GreaterOrEqual(t, total, float64(1))
}

func TestControllerAPI_ListControllers_HasHealth(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/controllers")

	data := env.Data.(map[string]interface{})
	controllers := data["controllers"].([]interface{})
	require.Greater(t, len(controllers), 0)

	entry := controllers[0].(map[string]interface{})
	health, ok := entry["health"].(map[string]interface{})
	require.True(t, ok, "expected 'health' in controller entry")
	assert.Contains(t, health, "name")
	assert.Contains(t, health, "kind")
	assert.Contains(t, health, "state")
	assert.Contains(t, health, "reconcileCount")
	assert.Contains(t, health, "errorCount")
}

func TestControllerAPI_GetController(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/controllers/namespace")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "namespace", data["name"])
	assert.Equal(t, "Namespace", data["kind"])
	assert.Contains(t, []string{"running", "stopped"}, data["state"])

	health, ok := data["health"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, health, "reconcileCount")
}

func TestControllerAPI_GetController_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/controllers/nonexistent")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "CONTROLLER_NOT_FOUND", env.Error.Code)
}

func TestControllerAPI_GetControllerHealth(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/controllers/namespace/health")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "namespace", data["name"])
	assert.Equal(t, "Namespace", data["kind"])
	assert.Contains(t, data, "reconcileCount")
	assert.Contains(t, data, "errorCount")
	assert.Contains(t, data, "state")
}

func TestControllerAPI_GetControllerHealth_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/controllers/nonexistent/health")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "CONTROLLER_NOT_FOUND", env.Error.Code)
}

func TestControllerAPI_NoConflictWithOtherEndpoints(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Controllers should work.
	code, env := httpGet(t, ts.URL+"/api/v1/controllers")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	controllers := env.Data.(map[string]interface{})["controllers"].([]interface{})
	assert.Greater(t, len(controllers), 0)

	// Capabilities should still work.
	code, env = httpGet(t, ts.URL+"/api/v1/capabilities")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	// Resources should still work.
	code, env = httpGet(t, ts.URL+"/api/v1/resources")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	// Providers should still work.
	code, env = httpGet(t, ts.URL+"/api/v1/providers")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	// Kernel should still work.
	code, env = httpGet(t, ts.URL+"/api/v1/kernel")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
}
