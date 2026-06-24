package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

// GET /api/v1/search?q=&account=&source=
func (h *Handler) searchMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []*model.Message{})
		return
	}
	accountID := r.URL.Query().Get("account")
	source := r.URL.Query().Get("source")

	msgs, err := h.store.SearchMessages(q, accountID, source)
	if errors.Is(err, store.ErrNotImplemented) {
		writeError(w, http.StatusServiceUnavailable, "search not available", "NOT_IMPLEMENTED")
		return
	}
	if err != nil {
		slog.Error("search messages", "q", q, "err", err)
		writeError(w, http.StatusInternalServerError, "search failed", "SEARCH_FAILED")
		return
	}
	if msgs == nil {
		msgs = []*model.Message{}
	}
	writeJSON(w, http.StatusOK, msgs)
}
