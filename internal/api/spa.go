package api

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed all:static
var staticFS embed.FS

func (h *Handler) registerSPA(r chi.Router) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("spa: sub static: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(sub))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Serve files that have an extension (JS, CSS, images, etc.) directly.
		// If the file doesn't exist the FileServer returns a proper 404.
		if strings.Contains(filepath.Base(path), ".") {
			fileServer.ServeHTTP(w, r)
			return
		}

		// For all other paths (SPA client-side routes) serve index.html.
		if _, err := sub.Open("index.html"); err == nil {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		// Frontend not built yet — show a developer-friendly page.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>catchy — frontend not built</title>
<style>body{font-family:monospace;padding:2rem;background:#0f172a;color:#94a3b8}
code{background:#1e293b;padding:2px 6px;border-radius:4px;color:#7dd3fc}
h1{color:#e2e8f0}a{color:#60a5fa}</style></head><body>
<h1>Frontend not built</h1>
<p>The Vue.js frontend has not been compiled yet.</p>
<p>Run the following then restart catchy:</p>
<pre><code>cd web
npm install
npm run build</code></pre>
<p>Or for development, run <code>npm run dev</code> in the <code>web/</code> directory
and open <a href="http://localhost:5173">localhost:5173</a>.</p>
<p>The JSON API is still available at
<a href="/api/v1/health">/api/v1/health</a>.</p>
</body></html>`)
	})
}
