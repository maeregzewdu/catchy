package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	catchymail "github.com/maeregzewdu/catchy/internal/mail"
	"github.com/maeregzewdu/catchy/internal/store"
)

// GET /api/v1/accounts/:id/messages
// Query params: folder (default INBOX), read, starred, limit (default 50), offset (default 0)
func (h *Handler) listMessages(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")

	if _, err := h.store.GetAccount(accountID); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found", "NOT_FOUND")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get account", "GET_FAILED")
		return
	}

	q := r.URL.Query()
	folder := q.Get("folder")
	if folder == "" {
		folder = "INBOX"
	}

	f := store.MessageFilter{
		Limit:  50,
		Offset: 0,
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			f.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			f.Offset = n
		}
	}
	if v := q.Get("read"); v != "" {
		b := v == "true"
		f.Read = &b
	}
	if v := q.Get("starred"); v != "" {
		b := v == "true"
		f.Starred = &b
	}

	msgs, err := h.store.ListMessages(accountID, folder, f)
	if err != nil {
		slog.Error("list messages", "account", accountID, "folder", folder, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list messages", "LIST_FAILED")
		return
	}
	if msgs == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}

// GET /api/v1/accounts/:id/messages/:msgId
func (h *Handler) getMessage(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")
	msg, err := h.store.GetMessage(msgID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found", "NOT_FOUND")
		return
	}
	if err != nil {
		slog.Error("get message", "id", msgID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get message", "GET_FAILED")
		return
	}
	writeJSON(w, http.StatusOK, msg)
}

// PATCH /api/v1/accounts/:id/messages/:msgId
// Body: {"is_read": bool, "is_starred": bool} — at least one required.
func (h *Handler) patchMessage(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

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

	if err := h.store.PatchMessage(msgID, body.IsRead, body.IsStarred); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "message not found", "NOT_FOUND")
			return
		}
		slog.Error("patch message", "id", msgID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update message", "PATCH_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/accounts/:id/messages/:msgId
func (h *Handler) deleteMessage(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")
	if err := h.store.DeleteMessage(msgID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "message not found", "NOT_FOUND")
			return
		}
		slog.Error("delete message", "id", msgID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete message", "DELETE_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/accounts/:id/messages/:msgId/attachments/:attId
func (h *Handler) getMessageAttachment(w http.ResponseWriter, r *http.Request) {
	attID := chi.URLParam(r, "attId")

	att, err := h.store.GetAttachment(attID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "attachment not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get attachment", "GET_FAILED")
		return
	}

	// Path traversal guard.
	dataDir := filepath.Clean(h.cfg.Data.Dir)
	attPath := filepath.Clean(att.Path)
	if !strings.HasPrefix(attPath, dataDir) {
		writeError(w, http.StatusForbidden, "forbidden", "FORBIDDEN")
		return
	}

	w.Header().Set("Content-Disposition", `attachment; filename="`+att.Filename+`"`)
	w.Header().Set("Content-Type", att.MIMEType)
	http.ServeFile(w, r, attPath)
}

// POST /api/v1/accounts/:id/sync
// Triggers an immediate sync for the account and returns 202 Accepted.
func (h *Handler) syncAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	a, err := h.store.GetAccount(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found", "NOT_FOUND")
		return
	}
	if err != nil {
		slog.Error("sync: get account", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load account", "GET_FAILED")
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := catchymail.SyncAccount(ctx, a, h.store, h.cfg.Data.Dir, h.cfg.Sync.DefaultFolders); err != nil {
			slog.Error("manual sync", "email", a.Email, "err", err)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "sync started"})
}
