package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/kernel/health"
	"github.com/cloudos/cloudos/packages/build"
	"github.com/cloudos/cloudos/packages/version"
)

// Handler bundles all HTTP handlers for the Control Plane API.
// Each method is a standard http.HandlerFunc that delegates to the
// corresponding kernel or package subsystem.
type Handler struct {
	k *kernel.Kernel
}

// NewHandler creates a Handler bound to the given kernel instance.
func NewHandler(k *kernel.Kernel) *Handler {
	return &Handler{k: k}
}

// ---------------------------------------------------------------------------
// GET /health
// ---------------------------------------------------------------------------

// HealthResponse is the JSON payload returned by GET /health.
type HealthResponse struct {
	Overall    health.Report            `json:"overall"`
	Components map[string]health.Report `json:"components"`
}

// Health returns the aggregated health status of all registered components
// plus the individual report for each one.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	healthMgr := h.k.Health()
	OK(w, HealthResponse{
		Overall:    healthMgr.Overall(),
		Components: healthMgr.All(),
	})
}

// ---------------------------------------------------------------------------
// GET /ready
// ---------------------------------------------------------------------------

// ReadinessResponse indicates whether the kernel is ready to serve requests.
type ReadinessResponse struct {
	Ready     bool   `json:"ready"`
	State     string `json:"state"`
	Message   string `json:"message,omitempty"`
}

// Ready returns 200 when the kernel is fully booted and running, or 503
// when it is still starting up or has shut down.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	state := h.k.State()
	if state == "running" {
		OK(w, ReadinessResponse{
			Ready: true,
			State: string(state),
		})
		return
	}
	ServiceUnavailable(w, "NOT_READY",
		"Kernel is in state: "+string(state))
}

// ---------------------------------------------------------------------------
// GET /live
// ---------------------------------------------------------------------------

// LivenessResponse is returned by the basic liveness probe.
type LivenessResponse struct {
	Alive bool   `json:"alive"`
	State string `json:"state"`
}

// Live returns 200 as long as the HTTP server is running. It is a simple
// liveness probe with no dependency on the kernel state.
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	OK(w, LivenessResponse{
		Alive: true,
		State: "serving",
	})
}

// ---------------------------------------------------------------------------
// GET /version
// ---------------------------------------------------------------------------

// VersionResponse carries version and build metadata.
type VersionResponse struct {
	Number      string          `json:"number"`
	Commit      string          `json:"commit"`
	Date        string          `json:"date"`
	Build       build.Metadata  `json:"build"`
}

// Version returns the CloudOS semantic version, Git commit, build date,
// and full build metadata.
func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	OK(w, VersionResponse{
		Number: version.Number,
		Commit: version.Commit,
		Date:   version.Date,
		Build:  build.Get(),
	})
}

// ---------------------------------------------------------------------------
// GET /kernel
// ---------------------------------------------------------------------------

// KernelResponse exposes the kernel's internal state and runtime information.
type KernelResponse struct {
	State      string        `json:"state"`
	Uptime     string        `json:"uptime"`
	UptimeNS   time.Duration `json:"uptimeNs"`
	StartedAt  time.Time     `json:"startedAt"`
	Subsystems []string      `json:"subsystems,omitempty"`
}

// Kernel returns the kernel's current lifecycle state, uptime, and started-at
// timestamp.
func (h *Handler) Kernel(w http.ResponseWriter, r *http.Request) {
	state := h.k.State()
	var startedAt time.Time
	var uptime time.Duration
	if state == "running" {
		uptime = h.k.Uptime()
		startedAt = time.Now().Add(-uptime)
	}

	OK(w, KernelResponse{
		State:     string(state),
		Uptime:    uptime.String(),
		UptimeNS:  uptime,
		StartedAt: startedAt,
	})
}

// ---------------------------------------------------------------------------
// GET /system
// ---------------------------------------------------------------------------

// SystemResponse carries runtime and operating-system information.
type SystemResponse struct {
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	GoVersion       string `json:"goVersion"`
	NumCPU          int    `json:"numCpu"`
	NumGoroutine    int    `json:"numGoroutine"`
	Compiler        string `json:"compiler"`
	Hostname        string `json:"hostname,omitempty"`
}

// System returns Go runtime information including OS, architecture, CPU count,
// goroutine count, and the hostname.
func (h *Handler) System(w http.ResponseWriter, r *http.Request) {
	resp := SystemResponse{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Compiler:     runtime.Compiler,
	}
	OK(w, resp)
}
