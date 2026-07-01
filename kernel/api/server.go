package api

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudos/cloudos/kernel"
	"github.com/cloudos/cloudos/packages/logging"
)

const (
	// DefaultAddr is the default listen address for the API server.
	DefaultAddr = ":8080"

	// shutdownTimeout is the maximum time to wait for in-flight requests to
	// complete during graceful shutdown.
	shutdownTimeout = 15 * time.Second

	// readHeaderTimeout is the time to wait for request headers.
	readHeaderTimeout = 10 * time.Second

	// idleTimeout is the maximum time to wait for the next request on a
	// keep-alive connection.
	idleTimeout = 60 * time.Second
)

// Server is the CloudOS Control Plane HTTP server. It exposes the kernel
// state, health, version, and system information over REST.
type Server struct {
	http   *http.Server
	log    *logging.Logger
	done   chan struct{}
}

// NewServer creates a new API server bound to the given kernel and address.
// The address defaults to ":8080" when addr is empty.
func NewServer(k *kernel.Kernel, addr string) *Server {
	if addr == "" {
		addr = DefaultAddr
	}

	log := logging.NewSubsystemLogger("api", logging.LevelInfo)
	handler := NewHandler(k)
	capHandler := NewCapabilityHandler(k)
	provHandler := NewProviderHandler(k)
	resHandler := NewResourceHandler(k.ResourceRegistry())
	ctrlHandler := NewControllerHandler(k)
	projHandler := NewProjectHandler(k)
	wfHandler := NewWorkflowHandler(k.ResourceRegistry())
	logHandler := NewLogHandler(k.ResourceRegistry(), k.RuntimeManager())
	deployHandler := NewDeploymentHandler(k.ResourceRegistry())

	mux := http.NewServeMux()
	registerRoutes(mux, handler, capHandler, provHandler, resHandler, ctrlHandler, projHandler, wfHandler, logHandler, deployHandler)

	// Build the middleware chain: outer → inner is RequestID → Recovery → Logging → mux.
	var h http.Handler = mux
	h = LoggingMiddleware(log)(h)
	h = RecoveryMiddleware(log)(h)
	h = RequestIDMiddleware(h)

	return &Server{
		http: &http.Server{
			Addr:              addr,
			Handler:           h,
			ReadHeaderTimeout: readHeaderTimeout,
			IdleTimeout:       idleTimeout,
		},
		log:  log,
		done: make(chan struct{}),
	}
}

// registerRoutes maps URL paths to handler methods.
func registerRoutes(mux *http.ServeMux, h *Handler, ch *CapabilityHandler, ph *ProviderHandler, rh *ResourceHandler, ctrlh *ControllerHandler, projh *ProjectHandler, wfh *WorkflowHandler, lh *LogHandler, dh *DeploymentHandler) {
	// --- System endpoints ---------------------------------------------------
	mux.HandleFunc("GET /api/v1/health", h.Health)
	mux.HandleFunc("GET /api/v1/ready", h.Ready)
	mux.HandleFunc("GET /api/v1/live", h.Live)
	mux.HandleFunc("GET /api/v1/version", h.Version)
	mux.HandleFunc("GET /api/v1/kernel", h.Kernel)
	mux.HandleFunc("GET /api/v1/system", h.System)

	// --- Capability discovery endpoints -------------------------------------
	mux.HandleFunc("GET /api/v1/capabilities", ch.ListCapabilities)
	mux.HandleFunc("GET /api/v1/capabilities/{id}", ch.GetCapability)

	// --- Provider discovery endpoints ---------------------------------------
	mux.HandleFunc("GET /api/v1/providers", ph.ListProviders)
	mux.HandleFunc("GET /api/v1/providers/{id}", ph.GetProvider)
	mux.HandleFunc("GET /api/v1/providers/{id}/health", ph.GetProviderHealth)
	mux.HandleFunc("GET /api/v1/providers/{id}/capabilities", ph.GetProviderCapabilities)

	// --- Resource Engine endpoints -----------------------------------------
	mux.HandleFunc("GET /api/v1/resources", rh.ListResourceKinds)
	mux.HandleFunc("GET /api/v1/resources/{kind}", rh.ListResources)
	mux.HandleFunc("GET /api/v1/resources/{kind}/{id}", rh.GetResource)

	// --- Controller Runtime endpoints --------------------------------------
	mux.HandleFunc("GET /api/v1/controllers", ctrlh.ListControllers)
	mux.HandleFunc("GET /api/v1/controllers/{id}", ctrlh.GetController)
	mux.HandleFunc("GET /api/v1/controllers/{id}/health", ctrlh.GetControllerHealth)

	// --- Project resource endpoints ---------------------------------------
	mux.HandleFunc("GET /api/v1/projects", projh.ListProjects)
	mux.HandleFunc("POST /api/v1/projects", projh.CreateProject)
	mux.HandleFunc("GET /api/v1/projects/{id}", projh.GetProject)
	mux.HandleFunc("PUT /api/v1/projects/{id}", projh.UpdateProject)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", projh.DeleteProject)

	// --- Workflow Execution endpoints ------------------------------------
	mux.HandleFunc("GET /api/v1/workflow-executions/{id}", wfh.GetExecution)
	mux.HandleFunc("GET /api/v1/workflow-executions/{id}/events", wfh.StreamExecutionEvents)

	// --- Application Log endpoints ---------------------------------------
	mux.HandleFunc("GET /api/v1/applications/{id}/logs", lh.SnapshotLogs)
	mux.HandleFunc("GET /api/v1/applications/{id}/logs/stream", lh.StreamLogs)
	mux.HandleFunc("GET /api/v1/applications/{id}/logs/download", lh.DownloadLogs)

	// --- Deployment Timeline / Comparison endpoints --------------------
	mux.HandleFunc("GET /api/v1/applications/{id}/deployments/{number}/timeline", dh.Timeline)
	mux.HandleFunc("GET /api/v1/applications/{id}/deployments/compare", dh.Compare)

	// Catch-all for unmatched paths — returns our standard JSON 404.
	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		NotFound(w, "NOT_FOUND", "The requested resource was not found")
	})
}

// Handler returns the fully-wrapped HTTP handler (middleware chain + mux).
// This is useful for testing with httptest.NewServer.
func (s *Server) Handler() http.Handler { return s.http.Handler }

// ListenAndServe starts the HTTP server in a goroutine and returns
// immediately. Call AwaitShutdown to block until the server stops.
func (s *Server) ListenAndServe() error {
	s.log.Info("api server starting", "addr", s.http.Addr)
	err := s.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server with a timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("api server shutting down")
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		s.log.Error("api server shutdown error", "error", err)
		return err
	}
	close(s.done)
	s.log.Info("api server stopped")
	return nil
}

// Done returns a channel that is closed when the server has fully shut down.
func (s *Server) Done() <-chan struct{} {
	return s.done
}
