package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/maeregzewdu/catchy/internal/config"
	"github.com/maeregzewdu/catchy/internal/store"
)

// Handler holds shared dependencies for all API handlers.
type Handler struct {
	store   store.Store
	cfg     *config.Config
	version string
}

// NewRouter wires all routes and middleware and returns the root http.Handler.
func NewRouter(s store.Store, cfg *config.Config, version string) http.Handler {
	h := &Handler{store: s, cfg: cfg, version: version}

	r := chi.NewRouter()
	r.Use(requestLogger)
	r.Use(chimiddleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.handleHealth)
	})

	return r
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable", "DB_UNAVAILABLE")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   h.version,
		"trap_addr": fmt.Sprintf("%s:%d", h.cfg.Trap.Host, h.cfg.Trap.Port),
	})
}
