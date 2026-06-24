package mail

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"

	"github.com/maeregzewdu/catchy/internal/model"
	"github.com/maeregzewdu/catchy/internal/store"
)

// trapBackend implements smtp.Backend. It accepts all connections without auth.
type trapBackend struct {
	store   store.Store
	dataDir string
}

func (b *trapBackend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &trapSession{backend: b}, nil
}

// trapSession handles one SMTP connection, collecting envelope fields before
// the DATA command delivers the full message.
type trapSession struct {
	backend  *trapBackend
	envelope struct {
		from string
		to   []string
	}
}

func (s *trapSession) AuthPlain(_, _ string) error { return nil }

func (s *trapSession) Mail(from string, _ *smtp.MailOptions) error {
	s.envelope.from = from
	return nil
}

func (s *trapSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	s.envelope.to = append(s.envelope.to, to)
	return nil
}

func (s *trapSession) Data(r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading message data: %w", err)
	}

	msgID := uuid.New().String()
	msg, attachments, parseErr := parseMIME(raw, s.backend.dataDir, msgID)
	if parseErr != nil {
		slog.Warn("mime parse partial failure", "err", parseErr)
	}

	// Fall back to SMTP envelope fields when headers are missing.
	if msg.FromAddr == "" {
		msg.FromAddr = s.envelope.from
	}
	if len(msg.ToAddrs) == 0 {
		msg.ToAddrs = s.envelope.to
	}

	return s.backend.store.StoreTrapMessage(msg, attachments)
}

func (s *trapSession) Reset() {
	s.envelope.from = ""
	s.envelope.to = nil
}

func (s *trapSession) Logout() error { return nil }

// StartTrapServer starts the SMTP trap server in a background goroutine.
// It waits up to 100 ms for the listener to bind, returning an error if the
// port is already in use or otherwise unreachable.
func StartTrapServer(addr string, s store.Store, dataDir string) (*smtp.Server, error) {
	be := &trapBackend{store: s, dataDir: dataDir}

	srv := smtp.NewServer(be)
	srv.Addr = addr
	srv.Domain = "localhost"
	srv.AllowInsecureAuth = true

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return nil, fmt.Errorf("smtp trap: %w", err)
	case <-time.After(100 * time.Millisecond):
		slog.Info("smtp trap listening", "addr", addr)
		return srv, nil
	}
}

// ── MIME parsing ─────────────────────────────────────────────────────────────

// parseMIME parses raw RFC 2822 bytes into a Message, writing any attachments
// to dataDir/attachments/<msgID>/. Parsing errors are non-fatal; the raw bytes
// are always stored so the Raw tab in the UI works even for malformed mail.
func parseMIME(raw []byte, dataDir, msgID string) (*model.Message, []model.Attachment, error) {
	now := time.Now().UTC()
	msg := &model.Message{
		ID:         msgID,
		Source:     "trap",
		Folder:     "trap",
		RawMIME:    raw,
		ReceivedAt: &now,
		CreatedAt:  now,
	}

	env, err := enmime.ReadEnvelope(bytes.NewReader(raw))
	if err != nil {
		return msg, nil, fmt.Errorf("parsing envelope: %w", err)
	}

	msg.Subject   = env.GetHeader("Subject")
	msg.MessageID = env.GetHeader("Message-Id")
	msg.FromAddr  = env.GetHeader("From")
	msg.BodyText  = env.Text
	msg.BodyHTML  = env.HTML
	msg.ToAddrs   = parseAddressList(env.GetHeader("To"))
	msg.CCAddrs   = parseAddressList(env.GetHeader("Cc"))
	msg.BCCAddrs  = parseAddressList(env.GetHeader("Bcc"))

	if dateStr := env.GetHeader("Date"); dateStr != "" {
		if t, err := mail.ParseDate(dateStr); err == nil {
			msg.SentAt = &t
		}
	}

	var attachments []model.Attachment
	for _, part := range env.Attachments {
		att, err := saveAttachment(part.FileName, part.ContentType, part.Content, dataDir, msgID)
		if err != nil {
			slog.Warn("saving attachment failed", "filename", part.FileName, "err", err)
			continue
		}
		attachments = append(attachments, att)
	}

	return msg, attachments, nil
}

func saveAttachment(filename, mimeType string, content []byte, dataDir, msgID string) (model.Attachment, error) {
	attDir := filepath.Join(dataDir, "attachments", msgID)
	if err := os.MkdirAll(attDir, 0755); err != nil {
		return model.Attachment{}, fmt.Errorf("creating attachment dir: %w", err)
	}

	safe := sanitizeFilename(filename)
	path := filepath.Join(attDir, safe)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return model.Attachment{}, fmt.Errorf("writing attachment file: %w", err)
	}

	return model.Attachment{
		ID:        uuid.New().String(),
		MessageID: msgID,
		Filename:  filename,
		MIMEType:  mimeType,
		Size:      int64(len(content)),
		Path:      path,
	}, nil
}

// parseAddressList extracts email addresses from a header like
// "Alice <alice@example.com>, bob@example.com".
func parseAddressList(header string) []string {
	if header == "" {
		return nil
	}
	addrs, err := mail.ParseAddressList(header)
	if err != nil {
		// Not parseable — return the raw header value as a single entry.
		return []string{header}
	}
	out := make([]string, len(addrs))
	for i, a := range addrs {
		out[i] = a.Address
	}
	return out
}

// sanitizeFilename removes path separators so an attachment filename cannot
// escape its storage directory.
func sanitizeFilename(name string) string {
	if name == "" {
		return "attachment"
	}
	safe := filepath.Base(name)
	for _, bad := range []string{"/", "\\", "..", ":"} {
		safe = strings.ReplaceAll(safe, bad, "_")
	}
	if safe == "" || safe == "." {
		return "attachment"
	}
	return safe
}
