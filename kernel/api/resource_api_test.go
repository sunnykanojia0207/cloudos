package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Resource Engine API endpoints
// ---------------------------------------------------------------------------

func TestResourceAPI_ListKinds(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/resources")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	kinds, ok := data["kinds"].([]interface{})
	require.True(t, ok, "expected 'kinds' array in response")

	// At minimum, the "Namespace" kind should be registered at boot.
	found := false
	for _, k := range kinds {
		entry := k.(map[string]interface{})
		if entry["name"] == "Namespace" {
			found = true
			assert.False(t, entry["namespaced"].(bool))
		}
	}
	assert.True(t, found, "expected Namespace kind to be registered")
}

func TestResourceAPI_ListKindsTotal(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/resources")

	data := env.Data.(map[string]interface{})
	total, ok := data["total"].(float64)
	require.True(t, ok)
	assert.GreaterOrEqual(t, total, float64(1))
}

func TestResourceAPI_ListResources(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/resources/Namespace")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	obj := env.Data.(map[string]interface{})
	assert.Equal(t, "NamespaceList", obj["kind"])
	assert.Equal(t, "cloudos.io/v1", obj["apiVersion"])

	items, ok := obj["items"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 1)

	// Check the default namespace.
	first := items[0].(map[string]interface{})
	meta := first["metadata"].(map[string]interface{})
	assert.Equal(t, "default", meta["id"])
	assert.Equal(t, "Default", meta["name"])
}

func TestResourceAPI_GetResource(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/resources/Namespace/default")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	obj := env.Data.(map[string]interface{})
	assert.Equal(t, "Namespace", obj["kind"])
	assert.Equal(t, "cloudos.io/v1", obj["apiVersion"])

	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "default", meta["id"])
	assert.Equal(t, "Default", meta["name"])
	assert.NotEmpty(t, meta["resourceVersion"])

	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "Default", spec["displayName"])

	status := obj["status"].(map[string]interface{})
	assert.Equal(t, "Active", status["phase"])
}

func TestResourceAPI_GetResource_NotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/resources/Namespace/nonexistent")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "RESOURCE_NOT_FOUND", env.Error.Code)
}

func TestResourceAPI_GetResource_KindNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/resources/UnknownKind/x")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "KIND_NOT_FOUND", env.Error.Code)
}

func TestResourceAPI_ListResources_KindNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/resources/UnknownKind")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "KIND_NOT_FOUND", env.Error.Code)
}

// Ensure the resource endpoint is properly isolated from capabilities.
func TestResourceAPI_DoesNotConflictWithCapabilities(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	// Capabilities should still work.
	code, env := httpGet(t, ts.URL+"/api/v1/capabilities")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	// Resources should also work.
	code, env = httpGet(t, ts.URL+"/api/v1/resources/Namespace")
	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
}

// Verify the ResourceObject envelope structure.
func TestResourceAPI_GetResource_HasEnvelope(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/resources/Namespace/default")

	obj := env.Data.(map[string]interface{})
	assert.Contains(t, obj, "apiVersion")
	assert.Contains(t, obj, "kind")
	assert.Contains(t, obj, "metadata")
	assert.Contains(t, obj, "spec")
	assert.Contains(t, obj, "status")

	meta := obj["metadata"].(map[string]interface{})
	assert.Contains(t, meta, "id")
	assert.Contains(t, meta, "name")
	assert.Contains(t, meta, "createdAt")
	assert.Contains(t, meta, "resourceVersion")
}
