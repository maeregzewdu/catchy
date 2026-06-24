package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

// GET /api/v1/trap/messages
func (h *Handler) listTrapMessages(w http.ResponseWriter, r *http.Request) {
	filters := store.TrapFilter{
		From:    r.URL.Query().Get("from"),
		To:      r.URL.Query().Get("to"),
		Subject: r.URL.Query().Get("subject"),
	}

	msgs, err := h.store.ListTrapMessages(filters)
	if err != nil {
		slog.Error("list trap messages", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list messages", "LIST_FAILED")
		return
	}

	// Return an empty array rather than null when there are no messages.
	if msgs == nil {
		msgs = []*model.Message{}
	}
	writeJSON(w, http.StatusOK, msgs)
}

// GET /api/v1/trap/messages/:id
func (h *Handler) getTrapMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	msg, err := h.store.GetTrapMessage(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found", "NOT_FOUND")
		return
	}
	if err != nil {
		slog.Error("get trap message", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get message", "GET_FAILED")
		return
	}

	atts, err := h.store.ListAttachments(id)
	if err != nil {
		slog.Warn("listing attachments", "message_id", id, "err", err)
		atts = nil
	}
	if atts == nil {
		atts = []model.Attachment{}
	}

	writeJSON(w, http.StatusOK, struct {
		*model.Message
		Attachments []model.Attachment `json:"attachments"`
	}{msg, atts})
}

// DELETE /api/v1/trap/messages/:id
func (h *Handler) deleteTrapMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.store.DeleteTrapMessage(id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found", "NOT_FOUND")
		return
	} else if err != nil {
		slog.Error("delete trap message", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete message", "DELETE_FAILED")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/trap/messages
func (h *Handler) clearTrapMessages(w http.ResponseWriter, r *http.Request) {
	if err := h.store.ClearTrapMessages(); err != nil {
		slog.Error("clear trap messages", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to clear messages", "CLEAR_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/v1/trap/messages/:id
func (h *Handler) patchTrapMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		IsRead    *bool `json:"is_read"`
		IsStarred *bool `json:"is_starred"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if body.IsRead == nil && body.IsStarred == nil {
		writeError(w, http.StatusUnprocessableEntity, "is_read or is_starred required", "VALIDATION_FAILED")
		return
	}
	if err := h.store.PatchMessage(id, body.IsRead, body.IsStarred); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "message not found", "NOT_FOUND")
			return
		}
		slog.Error("patch trap message", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update message", "PATCH_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/trap/messages/:id/attachments/:attId
func (h *Handler) getTrapAttachment(w http.ResponseWriter, r *http.Request) {
	attID := chi.URLParam(r, "attId")

	att, err := h.store.GetAttachment(attID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "attachment not found", "NOT_FOUND")
		return
	}
	if err != nil {
		slog.Error("get attachment", "id", attID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get attachment", "GET_FAILED")
		return
	}

	// Security: ensure the resolved path is under the configured data directory.
	dataDir := filepath.Clean(h.cfg.Data.Dir)
	attPath := filepath.Clean(att.Path)
	if !strings.HasPrefix(attPath, dataDir) {
		writeError(w, http.StatusForbidden, "invalid attachment path", "FORBIDDEN")
		return
	}

	w.Header().Set("Content-Disposition", `attachment; filename="`+att.Filename+`"`)
	w.Header().Set("Content-Type", att.MIMEType)
	http.ServeFile(w, r, attPath)
}
