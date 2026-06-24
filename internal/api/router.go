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

	// ── JSON API ──────────────────────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.handleHealth)
		r.Get("/search", h.searchMessages)

		// Trap inbox
		r.Route("/trap/messages", func(r chi.Router) {
			r.Get("/", h.listTrapMessages)
			r.Delete("/", h.clearTrapMessages)
			r.Get("/{id}", h.getTrapMessage)
			r.Patch("/{id}", h.patchTrapMessage)
			r.Delete("/{id}", h.deleteTrapMessage)
			r.Get("/{id}/raw", h.getTrapRawMIME)
			r.Get("/{id}/attachments/{attId}", h.getTrapAttachment)
		})

		// Real accounts + messages
		r.Route("/accounts", func(r chi.Router) {
			r.Get("/", h.listAccounts)
			r.Post("/", h.createAccount)
			r.Get("/{id}", h.getAccount)
			r.Put("/{id}", h.updateAccount)
			r.Delete("/{id}", h.deleteAccount)
			r.Post("/{id}/verify", h.verifyAccount)
			r.Post("/{id}/sync", h.syncAccount)
			r.Route("/{id}/messages", func(r chi.Router) {
				r.Get("/", h.listMessages)
				r.Get("/{msgId}", h.getMessage)
				r.Patch("/{msgId}", h.patchMessage)
				r.Delete("/{msgId}", h.deleteMessage)
				r.Get("/{msgId}/raw", h.getMessageRawMIME)
				r.Get("/{msgId}/attachments", h.listMessageAttachments)
				r.Get("/{msgId}/attachments/{attId}", h.getMessageAttachment)
			})
		})
	})

	// ── SPA catch-all (must be last) ──────────────────────────────────────────
	h.registerSPA(r)

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
