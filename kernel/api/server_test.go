package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/kernel/api"
	"github.com/cloudos/cloudos/packages/config"
	"github.com/cloudos/cloudos/packages/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nopWriter discards log output during tests.
type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testKernel boots a real Kernel for integration-style testing.
func testKernel(t *testing.T) *kernel.Kernel {
	t.Helper()

	cfg := config.Config{
		Kernel: config.KernelConfig{
			LogLevel: "debug",
		},
	}
	k, err := kernel.New(cfg)
	require.NoError(t, err)
	require.NoError(t, k.Boot(context.Background()))
	t.Cleanup(func() {
		_ = k.Shutdown(context.Background())
	})
	return k
}

// testServer creates an API server bound to a kernel and wraps it in an
// httptest.Server for convenient HTTP testing.
func testServer(t *testing.T, k *kernel.Kernel) *httptest.Server {
	t.Helper()
	srv := api.NewServer(k, ":0")
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		ts.Close()
		_ = srv.Shutdown(context.Background())
	})
	return ts
}

// httpGet performs an HTTP GET and decodes the JSON response.
func httpGet(t *testing.T, url string) (int, api.Response) {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	var envelope api.Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
	return resp.StatusCode, envelope
}

// ---------------------------------------------------------------------------
// Health endpoint
// ---------------------------------------------------------------------------

func TestHealth_Success(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/health")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data, ok := env.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "overall")
	assert.Contains(t, data, "components")
}

func TestHealth_OverallIsRunning(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/health")

	assert.Equal(t, 200, code)
	data := env.Data.(map[string]interface{})
	overall := data["overall"].(map[string]interface{})
	assert.Equal(t, "running", overall["state"])
}

func TestHealth_HasComponents(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/health")

	data := env.Data.(map[string]interface{})
	components := data["components"].(map[string]interface{})
	assert.NotEmpty(t, components)
	assert.Contains(t, components, "kernel")
}

// ---------------------------------------------------------------------------
// Ready endpoint
// ---------------------------------------------------------------------------

func TestReady_WhenRunning(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/ready")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	data := env.Data.(map[string]interface{})
	assert.Equal(t, true, data["ready"])
	assert.Equal(t, "running", data["state"])
}

// ---------------------------------------------------------------------------
// Live endpoint
// ---------------------------------------------------------------------------

func TestLive_AlwaysAlive(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/live")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	data := env.Data.(map[string]interface{})
	assert.Equal(t, true, data["alive"])
	assert.Equal(t, "serving", data["state"])
}

// ---------------------------------------------------------------------------
// Version endpoint
// ---------------------------------------------------------------------------

func TestVersion_ReturnsBuildInfo(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/version")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Contains(t, data, "number")
	assert.Contains(t, data, "commit")
	assert.Contains(t, data, "date")
	assert.Contains(t, data, "build")
}

func TestVersion_NumberIsSet(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/version")

	data := env.Data.(map[string]interface{})
	number, ok := data["number"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, number)
}

// ---------------------------------------------------------------------------
// Kernel endpoint
// ---------------------------------------------------------------------------

func TestKernel_ReturnsState(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/kernel")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "running", data["state"])
	assert.Contains(t, data, "uptime")
	assert.Contains(t, data, "startedAt")
}

func TestKernel_UptimeIsPositive(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/kernel")

	data := env.Data.(map[string]interface{})
	uptimeNs := data["uptimeNs"].(float64)
	assert.Greater(t, uptimeNs, float64(0))
}

// ---------------------------------------------------------------------------
// System endpoint
// ---------------------------------------------------------------------------

func TestSystem_ReturnsRuntimeInfo(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/system")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Contains(t, data, "os")
	assert.Contains(t, data, "arch")
	assert.Contains(t, data, "goVersion")
	assert.Contains(t, data, "numCpu")
	assert.Contains(t, data, "numGoroutine")
	assert.Contains(t, data, "compiler")
}

func TestSystem_OSIsNotEmpty(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/system")

	data := env.Data.(map[string]interface{})
	os, ok := data["os"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, os)
}

func TestSystem_NumCPUIsPositive(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/system")

	data := env.Data.(map[string]interface{})
	numCPU := data["numCpu"].(float64)
	assert.GreaterOrEqual(t, numCPU, float64(1))
}

// ---------------------------------------------------------------------------
// Capabilities: List
// ---------------------------------------------------------------------------

func TestCapabilities_List(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	data := env.Data.(map[string]interface{})
	assert.Equal(t, "CapabilityList", data["kind"])
	assert.Equal(t, "cloudos.io/v1", data["apiVersion"])

	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(items), 5) // compute, storage, database, ai, network

	listMeta := data["metadata"].(map[string]interface{})
	assert.GreaterOrEqual(t, int(listMeta["total"].(float64)), 5)
}

func TestCapabilities_List_HasAllBuiltins(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/capabilities")

	data := env.Data.(map[string]interface{})
	items := data["items"].([]interface{})

	ids := make(map[string]bool)
	for _, item := range items {
		obj := item.(map[string]interface{})
		meta := obj["metadata"].(map[string]interface{})
		ids[meta["id"].(string)] = true
	}

	assert.True(t, ids["compute"], "should have compute capability")
	assert.True(t, ids["storage"], "should have storage capability")
	assert.True(t, ids["database"], "should have database capability")
	assert.True(t, ids["ai"], "should have ai capability")
	assert.True(t, ids["network"], "should have network capability")
}

func TestCapabilities_List_ItemsAreResourceObjects(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/capabilities")

	data := env.Data.(map[string]interface{})
	items := data["items"].([]interface{})
	require.Greater(t, len(items), 0)

	item := items[0].(map[string]interface{})
	// Every item must be a valid ResourceObject.
	assert.Contains(t, item, "kind")
	assert.Contains(t, item, "apiVersion")
	assert.Contains(t, item, "metadata")
	assert.Contains(t, item, "spec")
	assert.Contains(t, item, "status")
	assert.Equal(t, "Capability", item["kind"])
	assert.Equal(t, "cloudos.io/v1", item["apiVersion"])
	meta := item["metadata"].(map[string]interface{})
	assert.Contains(t, meta, "id")
	assert.Contains(t, meta, "name")
}

// ---------------------------------------------------------------------------
// Capabilities: Get by ID
// ---------------------------------------------------------------------------

func TestCapabilities_GetCompute(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities/compute")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)
	require.NotNil(t, env.Data)

	obj := env.Data.(map[string]interface{})
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "compute", meta["id"])
	assert.Equal(t, "Compute", meta["name"])
	assert.Equal(t, "Capability", obj["kind"])
	assert.Equal(t, "cloudos.io/v1", obj["apiVersion"])

	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "Compute Engine", spec["displayName"])
	assert.Equal(t, "compute", spec["category"])

	status := obj["status"].(map[string]interface{})
	assert.Equal(t, "stable", status["status"])
	assert.Equal(t, false, status["available"]) // no providers registered in test
	assert.Equal(t, float64(0), status["providerCount"])

	labels := meta["labels"].(map[string]interface{})
	assert.Equal(t, "compute", labels["category"])
	assert.Equal(t, "stable", labels["status"])
}

func TestCapabilities_GetStorage(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities/storage")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	obj := env.Data.(map[string]interface{})
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "storage", meta["id"])
	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "Object Storage", spec["displayName"])
}

func TestCapabilities_GetDatabase(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities/database")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	obj := env.Data.(map[string]interface{})
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "database", meta["id"])
	spec := obj["spec"].(map[string]interface{})
	assert.Equal(t, "SQL Database", spec["displayName"])
}

func TestCapabilities_GetAI(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities/ai")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	obj := env.Data.(map[string]interface{})
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "ai", meta["id"])
	status := obj["status"].(map[string]interface{})
	assert.Equal(t, "experimental", status["status"])
}

func TestCapabilities_GetNetwork(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities/network")

	assert.Equal(t, 200, code)
	assert.True(t, env.Success)

	obj := env.Data.(map[string]interface{})
	meta := obj["metadata"].(map[string]interface{})
	assert.Equal(t, "network", meta["id"])
}

func TestCapabilities_GetNotFound(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/capabilities/nonexistent")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "CAPABILITY_NOT_FOUND", env.Error.Code)
}

func TestCapabilities_GetHasOperations(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	_, env := httpGet(t, ts.URL+"/api/v1/capabilities/compute")

	obj := env.Data.(map[string]interface{})
	spec := obj["spec"].(map[string]interface{})
	ops := spec["operations"].([]interface{})
	assert.Greater(t, len(ops), 0)

	op := ops[0].(map[string]interface{})
	assert.Contains(t, op, "name")
	assert.Contains(t, op, "description")
}

// ---------------------------------------------------------------------------
// 404 handling
// ---------------------------------------------------------------------------

func TestUnknownEndpoint_Returns404(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	code, env := httpGet(t, ts.URL+"/api/v1/nonexistent")

	assert.Equal(t, 404, code)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "NOT_FOUND", env.Error.Code)
}

// ---------------------------------------------------------------------------
// Middleware: Request ID
// ---------------------------------------------------------------------------

func TestMiddleware_RequestIDAdded(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	resp, err := http.Get(ts.URL + "/api/v1/live")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEmpty(t, resp.Header.Get("X-Request-Id"))
}

func TestMiddleware_RequestIDPassthrough(t *testing.T) {
	k := testKernel(t)
	ts := testServer(t, k)

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/live", nil)
	req.Header.Set("X-Request-Id", "my-trace-id")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "my-trace-id", resp.Header.Get("X-Request-Id"))
}

// ---------------------------------------------------------------------------
// Middleware: Panic recovery
// ---------------------------------------------------------------------------

func TestMiddleware_Recovery(t *testing.T) {
	// Use an in-memory logger to avoid nil panics.
	log := logging.NewSubsystemLoggerWithWriter("test", logging.LevelDebug, &nopWriter{})

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	recovered := api.RecoveryMiddleware(log)(panicHandler)
	ts := httptest.NewServer(recovered)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 500, resp.StatusCode)

	var env api.Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
}

// ---------------------------------------------------------------------------
// Response helpers
// ---------------------------------------------------------------------------

func TestResponse_OK(t *testing.T) {
	w := httptest.NewRecorder()
	api.OK(w, map[string]string{"hello": "world"})

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var env api.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	data := env.Data.(map[string]interface{})
	assert.Equal(t, "world", data["hello"])
}

func TestResponse_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	api.NotFound(w, "NOT_FOUND", "resource not found")

	assert.Equal(t, 404, w.Code)

	var env api.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "NOT_FOUND", env.Error.Code)
	assert.Equal(t, "resource not found", env.Error.Message)
}

func TestResourceObject_HasRequiredFields(t *testing.T) {
	obj := api.NewObject("TestKind", "test-id", "Test Name", map[string]string{"foo": "bar"}, nil)

	assert.Equal(t, "test-id", obj.Metadata.ID)
	assert.Equal(t, "Test Name", obj.Metadata.Name)
	assert.Equal(t, "TestKind", obj.Kind)
	assert.Equal(t, "cloudos.io/v1", obj.APIVersion)
	assert.NotNil(t, obj.Metadata)
	assert.NotNil(t, obj.Spec)
	assert.Nil(t, obj.Status)
}

func TestResourceObject_JSONRoundTrip(t *testing.T) {
	spec := map[string]string{"hello": "world"}
	obj := api.NewObject("Widget", "w-1", "Widget One", spec, nil)

	w := httptest.NewRecorder()
	api.OK(w, obj)

	var envelope api.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))

	data := envelope.Data.(map[string]interface{})
	meta := data["metadata"].(map[string]interface{})
	assert.Equal(t, "w-1", meta["id"])
	assert.Equal(t, "Widget", data["kind"])
	assert.Equal(t, "cloudos.io/v1", data["apiVersion"])
	assert.Contains(t, data, "metadata")
	assert.Contains(t, data, "spec")
}

func TestResourceObjectList_HasRequiredFields(t *testing.T) {
	items := []api.Object{
		api.NewObject("Item", "a", "Item A", nil, nil),
		api.NewObject("Item", "b", "Item B", nil, nil),
	}
	list := api.NewObjectList("Item", items)

	assert.Equal(t, "ItemList", list.Kind)
	assert.Equal(t, "cloudos.io/v1", list.APIVersion)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, 2, list.Metadata.Total)
}

func TestResourceObjectList_JSONRoundTrip(t *testing.T) {
	items := []api.Object{
		api.NewObject("Capability", "compute", "Compute", nil, nil),
	}
	list := api.NewObjectList("Capability", items)

	w := httptest.NewRecorder()
	api.OK(w, list)

	var envelope api.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))

	data := envelope.Data.(map[string]interface{})
	assert.Equal(t, "CapabilityList", data["kind"])
	assert.Equal(t, "cloudos.io/v1", data["apiVersion"])

	itemsArr := data["items"].([]interface{})
	assert.Len(t, itemsArr, 1)
}

func TestResponse_ServiceUnavailable(t *testing.T) {
	w := httptest.NewRecorder()
	api.ServiceUnavailable(w, "NOT_READY", "not ready")

	assert.Equal(t, 503, w.Code)

	var env api.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.False(t, env.Success)
	assert.Equal(t, "NOT_READY", env.Error.Code)
	assert.Equal(t, "not ready", env.Error.Message)
}
