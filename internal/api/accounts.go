package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	catchymail "github.com/maeregzewdu/catchy/internal/mail"
	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

// POST /api/v1/accounts
func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var req accountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if err := req.validate(false); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error(), "VALIDATION_FAILED")
		return
	}

	a := &model.Account{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		SMTPHost:  req.SMTPHost,
		SMTPPort:  req.SMTPPort,
		IMAPHost:  req.IMAPHost,
		IMAPPort:  req.IMAPPort,
		Username:  req.Username,
		Password:  req.Password,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.store.CreateAccount(a); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "email address already exists", "DUPLICATE_EMAIL")
			return
		}
		slog.Error("create account", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create account", "CREATE_FAILED")
		return
	}

	writeJSON(w, http.StatusCreated, a) // Password has json:"-" so it is never included.
}

// GET /api/v1/accounts
func (h *Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.store.ListAccounts()
	if err != nil {
		slog.Error("list accounts", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list accounts", "LIST_FAILED")
		return
	}
	if accounts == nil {
		accounts = []*model.Account{}
	}
	writeJSON(w, http.StatusOK, accounts)
}

// GET /api/v1/accounts/:id
func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	a, err := h.store.GetAccount(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found", "NOT_FOUND")
		return
	}
	if err != nil {
		slog.Error("get account", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get account", "GET_FAILED")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// PUT /api/v1/accounts/:id
func (h *Handler) updateAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req accountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	// Password is optional on update — omitting it keeps the existing value.
	if err := req.validate(true); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error(), "VALIDATION_FAILED")
		return
	}

	a := &model.Account{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		SMTPHost: req.SMTPHost,
		SMTPPort: req.SMTPPort,
		IMAPHost: req.IMAPHost,
		IMAPPort: req.IMAPPort,
		Username: req.Username,
		Password: req.Password, // empty string → store keeps existing encrypted value
	}

	if err := h.store.UpdateAccount(a); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "account not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "email address already exists", "DUPLICATE_EMAIL")
			return
		}
		slog.Error("update account", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update account", "UPDATE_FAILED")
		return
	}

	updated, err := h.store.GetAccount(id)
	if err != nil {
		slog.Error("get updated account", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve updated account", "GET_FAILED")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DELETE /api/v1/accounts/:id
func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.store.DeleteAccount(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "account not found", "NOT_FOUND")
			return
		}
		slog.Error("delete account", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete account", "DELETE_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/accounts/:id/verify
func (h *Handler) verifyAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	a, err := h.store.GetAccount(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found", "NOT_FOUND")
		return
	}
	if err != nil {
		slog.Error("verify: get account", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to load account", "GET_FAILED")
		return
	}

	smtpErr := catchymail.VerifySMTP(a)
	imapErr := catchymail.VerifyIMAP(a)

	writeJSON(w, http.StatusOK, map[string]string{
		"smtp": errStr(smtpErr),
		"imap": errStr(imapErr),
	})
}

func errStr(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

// ── request / validation ──────────────────────────────────────────────────────

type accountRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	IMAPHost string `json:"imap_host"`
	IMAPPort int    `json:"imap_port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *accountRequest) validate(passwordOptional bool) error {
	switch {
	case r.Name == "":
		return errors.New("name is required")
	case r.Email == "":
		return errors.New("email is required")
	case r.SMTPHost == "":
		return errors.New("smtp_host is required")
	case r.SMTPPort == 0:
		return errors.New("smtp_port is required")
	case r.IMAPHost == "":
		return errors.New("imap_host is required")
	case r.IMAPPort == 0:
		return errors.New("imap_port is required")
	case r.Username == "":
		return errors.New("username is required")
	case r.Password == "" && !passwordOptional:
		return errors.New("password is required")
	}
	return nil
}
