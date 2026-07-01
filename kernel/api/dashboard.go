package api

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dashboard-dist/*
var dashboardFS embed.FS

// DashboardHandler serves the CloudOS web dashboard (React SPA).
// All routes under the dashboard path serve prebuilt static assets.
// Unknown paths fall back to index.html for client-side routing.
type DashboardHandler struct {
	handler http.Handler
}

// NewDashboardHandler creates a handler that serves the dashboard SPA.
func NewDashboardHandler() *DashboardHandler {
	// Strip the leading "dashboard-dist" prefix from the embedded FS.
	subFS, err := fs.Sub(dashboardFS, "dashboard-dist")
	if err != nil {
		// Fallback: return a handler that shows an error.
		return &DashboardHandler{
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("CloudOS Dashboard\n\nBuild the dashboard first:\n  cd apps/dashboard && npm run build"))
			}),
		}
	}

	fileServer := http.FileServer(http.FS(subFS))

	return &DashboardHandler{
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For SPA routing: if the requested path doesn't have an extension
			// and isn't an API route, serve index.html.
			if !isStaticAsset(r.URL.Path) {
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		}),
	}
}

// ServeHTTP implements http.Handler.
func (d *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.handler.ServeHTTP(w, r)
}

// isStaticAsset returns true if the path looks like a static file request
// (has a file extension and is not an API route).
func isStaticAsset(urlPath string) bool {
	ext := path.Ext(strings.TrimSuffix(urlPath, "/"))
	return ext != "" || urlPath == "/"
}
