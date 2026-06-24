package api

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	catchymail "github.com/maeregzewdu/catchy/internal/mail"
	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

//go:embed templates
var templateFS embed.FS

// PageData is passed to every page template.
type PageData struct {
	Title       string
	Nav         string // "trap" | accountID | "settings" | "search"
	Accounts    []*model.Account
	TrapCount   int
	Messages    []*model.Message
	Message     *model.Message
	Attachments []model.Attachment
	Headers     []HeaderEntry
	Account     *model.Account
	Folder      string
	Error       string
	Query       string // search query
}

// HeaderEntry is a single parsed email header for the Headers tab.
type HeaderEntry struct {
	Key   string
	Value string
}

var webTmpl *template.Template

func init() {
	fm := template.FuncMap{
		"formatTime": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			d := time.Since(*t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh", int(d.Hours()))
			case d < 7*24*time.Hour:
				return t.Format("Mon")
			default:
				return t.Format("Jan 2")
			}
		},
		"formatDate": func(sent, received *time.Time) string {
			t := sent
			if t == nil {
				t = received
			}
			if t == nil {
				return ""
			}
			return t.Format("Mon, Jan 2 2006  3:04 PM")
		},
		"formatSize": func(n int64) string {
			switch {
			case n < 1024:
				return fmt.Sprintf("%d B", n)
			case n < 1024*1024:
				return fmt.Sprintf("%.1f KB", float64(n)/1024)
			default:
				return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
			}
		},
		"joinAddrs": func(addrs []string) string {
			return strings.Join(addrs, ", ")
		},
		"rawStr": func(b []byte) string {
			return string(b)
		},
		"msgPatchURL": func(m *model.Message) string {
			if m.AccountID != nil && *m.AccountID != "" {
				return "/api/v1/accounts/" + *m.AccountID + "/messages/" + m.ID
			}
			return "/api/v1/trap/messages/" + m.ID
		},
		"msgDeleteURL": func(m *model.Message) string {
			if m.AccountID != nil && *m.AccountID != "" {
				return "/api/v1/accounts/" + *m.AccountID + "/messages/" + m.ID
			}
			return "/api/v1/trap/messages/" + m.ID
		},
		"attURL": func(m *model.Message, attID string) string {
			if m.AccountID != nil && *m.AccountID != "" {
				return "/api/v1/accounts/" + *m.AccountID + "/messages/" + m.ID + "/attachments/" + attID
			}
			return "/api/v1/trap/messages/" + m.ID + "/attachments/" + attID
		},
	}
	var err error
	webTmpl, err = template.New("").Funcs(fm).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic("web: parse templates: " + err.Error())
	}
}

func (h *Handler) baseData(nav string) PageData {
	accounts, _ := h.store.ListAccounts()
	msgs, _ := h.store.ListTrapMessages(store.TrapFilter{})
	unread := 0
	for _, m := range msgs {
		if !m.IsRead {
			unread++
		}
	}
	return PageData{
		Nav:       nav,
		Accounts:  accounts,
		TrapCount: unread,
	}
}

func (h *Handler) render(w http.ResponseWriter, name string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := webTmpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("render template", "name", name, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GET /
func (h *Handler) trapInbox(w http.ResponseWriter, r *http.Request) {
	data := h.baseData("trap")
	data.Title = "Trap Inbox"
	msgs, _ := h.store.ListTrapMessages(store.TrapFilter{})
	data.Messages = msgs
	h.render(w, "inbox-page", data)
}

// GET /trap/list  (HTMX partial — refreshes message list every 3s)
func (h *Handler) trapList(w http.ResponseWriter, r *http.Request) {
	data := h.baseData("trap")
	msgs, _ := h.store.ListTrapMessages(store.TrapFilter{})
	data.Messages = msgs
	h.render(w, "inbox-list", data)
}

// GET /accounts/:id/folders/:folder
func (h *Handler) accountFolder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	folder := chi.URLParam(r, "folder")

	account, err := h.store.GetAccount(id)
	if errors.Is(err, store.ErrNotFound) {
		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	data := h.baseData(id)
	data.Title = account.Name + " — " + folder
	data.Account = account
	data.Folder = folder
	msgs, _ := h.store.ListMessages(id, folder, store.MessageFilter{Limit: 50})
	data.Messages = msgs
	h.render(w, "folder-page", data)
}

// GET /messages/:id/detail  (HTMX partial — loaded into #msg-pane)
func (h *Handler) messageDetailPartial(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "id")

	msg, err := h.loadMessage(msgID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Mark as read on open.
	if !msg.IsRead {
		t := true
		h.store.PatchMessage(msgID, &t, nil) //nolint:errcheck
	}

	data := h.baseData("")
	data.Message = msg
	data.Attachments, _ = h.store.ListAttachments(msgID)
	data.Headers = extractHeaders(msg.RawMIME)
	h.render(w, "message-detail", data)
}

// GET /messages/:id  (full page — message pre-selected in the correct pane)
func (h *Handler) messagePage(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "id")

	msg, err := h.loadMessage(msgID)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Mark as read on open.
	if !msg.IsRead {
		t := true
		h.store.PatchMessage(msgID, &t, nil) //nolint:errcheck
	}

	if msg.AccountID != nil {
		data := h.baseData(*msg.AccountID)
		data.Title = msg.Subject
		data.Account, _ = h.store.GetAccount(*msg.AccountID)
		data.Folder = msg.Folder
		msgs, _ := h.store.ListMessages(*msg.AccountID, msg.Folder, store.MessageFilter{Limit: 50})
		data.Messages = msgs
		data.Message = msg
		data.Attachments, _ = h.store.ListAttachments(msgID)
		data.Headers = extractHeaders(msg.RawMIME)
		h.render(w, "folder-page", data)
	} else {
		data := h.baseData("trap")
		data.Title = msg.Subject
		msgs, _ := h.store.ListTrapMessages(store.TrapFilter{})
		data.Messages = msgs
		data.Message = msg
		data.Attachments, _ = h.store.ListAttachments(msgID)
		data.Headers = extractHeaders(msg.RawMIME)
		h.render(w, "inbox-page", data)
	}
}

// GET /settings
func (h *Handler) settingsPage(w http.ResponseWriter, r *http.Request) {
	data := h.baseData("settings")
	data.Title = "Settings"
	if e := r.URL.Query().Get("error"); e != "" {
		data.Error = e
	}
	h.render(w, "settings-page", data)
}

// POST /settings/accounts  (HTML form → create account → redirect)
func (h *Handler) createAccountWeb(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/settings?error=invalid+form", http.StatusSeeOther)
		return
	}
	smtpPort, _ := strconv.Atoi(r.FormValue("smtp_port"))
	imapPort, _ := strconv.Atoi(r.FormValue("imap_port"))

	a := &model.Account{
		Name:     strings.TrimSpace(r.FormValue("name")),
		Email:    strings.TrimSpace(r.FormValue("email")),
		SMTPHost: strings.TrimSpace(r.FormValue("smtp_host")),
		SMTPPort: smtpPort,
		IMAPHost: strings.TrimSpace(r.FormValue("imap_host")),
		IMAPPort: imapPort,
		Username: strings.TrimSpace(r.FormValue("username")),
		Password: r.FormValue("password"),
	}

	if a.Name == "" || a.Email == "" || a.SMTPHost == "" || a.IMAPHost == "" || a.Username == "" || a.Password == "" {
		http.Redirect(w, r, "/settings?error=all+fields+required", http.StatusSeeOther)
		return
	}

	if err := h.store.CreateAccount(a); errors.Is(err, store.ErrDuplicate) {
		http.Redirect(w, r, "/settings?error=account+already+exists", http.StatusSeeOther)
		return
	} else if err != nil {
		slog.Error("create account web", "err", err)
		http.Redirect(w, r, "/settings?error=server+error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// DELETE /settings/accounts/:id  (HTMX → returns empty 200 to remove the card)
func (h *Handler) deleteAccountWeb(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.store.DeleteAccount(id); err != nil && !errors.Is(err, store.ErrNotFound) {
		slog.Error("delete account web", "id", id, "err", err)
	}
	w.WriteHeader(http.StatusOK) // empty body → HTMX swaps target with nothing
}

// GET /search
func (h *Handler) searchWeb(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	data := h.baseData("search")
	data.Title = "Search"
	data.Query = q

	if q != "" {
		msgs, err := h.store.SearchMessages(q, "", "")
		if err == nil {
			data.Messages = msgs
		} else {
			slog.Warn("search web", "q", q, "err", err)
		}
	}

	if r.Header.Get("HX-Request") != "" {
		h.render(w, "search-results", data)
		return
	}
	h.render(w, "search-page", data)
}

// POST /settings/accounts/:id/verify  (returns HTML for HTMX target)
func (h *Handler) verifyAccountWeb(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	a, err := h.store.GetAccount(id)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		fmt.Fprint(w, `<span class="text-red-400 text-xs">error loading account</span>`)
		return
	}

	smtpErr := catchymail.VerifySMTP(a)
	imapErr := catchymail.VerifyIMAP(a)

	smtpStatus := "SMTP ok"
	if smtpErr != nil {
		smtpStatus = "SMTP: " + smtpErr.Error()
	}
	imapStatus := "IMAP ok"
	if imapErr != nil {
		imapStatus = "IMAP: " + imapErr.Error()
	}

	color := "text-green-400"
	if smtpErr != nil || imapErr != nil {
		color = "text-red-400"
	}
	fmt.Fprintf(w, `<span class="%s text-xs">%s &middot; %s</span>`, color, smtpStatus, imapStatus)
}

// loadMessage tries trap first, then account messages.
func (h *Handler) loadMessage(id string) (*model.Message, error) {
	msg, err := h.store.GetTrapMessage(id)
	if errors.Is(err, store.ErrNotFound) {
		return h.store.GetMessage(id)
	}
	return msg, err
}

func extractHeaders(raw []byte) []HeaderEntry {
	if len(raw) == 0 {
		return nil
	}
	m, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	var out []HeaderEntry
	for k, vals := range m.Header {
		for _, v := range vals {
			out = append(out, HeaderEntry{Key: k, Value: v})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
