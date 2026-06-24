package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/maeregzewdu/catchy/internal/model"
)

// ── StoreTrapMessage ──────────────────────────────────────────────────────────

func (s *sqliteStore) StoreTrapMessage(msg *model.Message, attachments []model.Attachment) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(`
		INSERT INTO messages
			(id, account_id, source, message_id, folder, subject,
			 from_addr, to_addrs, cc_addrs, bcc_addrs,
			 body_text, body_html, raw_mime,
			 is_read, is_starred, received_at, created_at)
		VALUES (?, NULL, 'trap', ?, 'trap', ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?)`,
		msg.ID, msg.MessageID, msg.Subject,
		msg.FromAddr,
		marshalAddrs(msg.ToAddrs),
		marshalAddrs(msg.CCAddrs),
		marshalAddrs(msg.BCCAddrs),
		msg.BodyText, msg.BodyHTML, msg.RawMIME,
		msg.ReceivedAt, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting message: %w", err)
	}

	for _, att := range attachments {
		_, err = tx.Exec(`
			INSERT INTO attachments (id, message_id, filename, mime_type, size, path)
			VALUES (?, ?, ?, ?, ?, ?)`,
			att.ID, att.MessageID, att.Filename, att.MIMEType, att.Size, att.Path,
		)
		if err != nil {
			return fmt.Errorf("inserting attachment %s: %w", att.Filename, err)
		}
	}

	return tx.Commit()
}

// ── ListTrapMessages ──────────────────────────────────────────────────────────

func (s *sqliteStore) ListTrapMessages(f TrapFilter) ([]*model.Message, error) {
	q := `SELECT id, message_id, subject, from_addr, to_addrs, cc_addrs, bcc_addrs,
	             is_read, is_starred, received_at, created_at
	      FROM messages
	      WHERE source = 'trap' AND folder = 'trap'`
	args := []any{}

	if f.From != "" {
		q += " AND from_addr LIKE ?"
		args = append(args, "%"+f.From+"%")
	}
	if f.To != "" {
		q += " AND to_addrs LIKE ?"
		args = append(args, "%"+f.To+"%")
	}
	if f.Subject != "" {
		q += " AND subject LIKE ?"
		args = append(args, "%"+f.Subject+"%")
	}
	q += " ORDER BY received_at DESC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		m, err := scanMessageRow(rows)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// ── GetTrapMessage ────────────────────────────────────────────────────────────

func (s *sqliteStore) GetTrapMessage(id string) (*model.Message, error) {
	row := s.db.QueryRow(`
		SELECT id, message_id, subject, from_addr, to_addrs, cc_addrs, bcc_addrs,
		       body_text, body_html, raw_mime, is_read, is_starred, received_at, created_at
		FROM messages
		WHERE id = ? AND source = 'trap'`, id)

	var (
		msg                         model.Message
		toJSON, ccJSON, bccJSON     string
		bodyText, bodyHTML          sql.NullString
		receivedAt                  sql.NullTime
		isRead, isStarred           int
	)
	err := row.Scan(
		&msg.ID, &msg.MessageID, &msg.Subject, &msg.FromAddr,
		&toJSON, &ccJSON, &bccJSON,
		&bodyText, &bodyHTML, &msg.RawMIME,
		&isRead, &isStarred, &receivedAt, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	msg.ToAddrs   = unmarshalAddrs(toJSON)
	msg.CCAddrs   = unmarshalAddrs(ccJSON)
	msg.BCCAddrs  = unmarshalAddrs(bccJSON)
	msg.BodyText  = bodyText.String
	msg.BodyHTML  = bodyHTML.String
	msg.IsRead    = isRead != 0
	msg.IsStarred = isStarred != 0
	if receivedAt.Valid {
		t := receivedAt.Time
		msg.ReceivedAt = &t
	}

	return &msg, nil
}

// ── DeleteTrapMessage ─────────────────────────────────────────────────────────

func (s *sqliteStore) DeleteTrapMessage(id string) error {
	res, err := s.db.Exec("DELETE FROM messages WHERE id = ? AND source = 'trap'", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	// Remove attachment files; non-fatal if they don't exist.
	os.RemoveAll(filepath.Join(s.dataDir, "attachments", id))
	return nil
}

// ── ClearTrapMessages ─────────────────────────────────────────────────────────

func (s *sqliteStore) ClearTrapMessages() error {
	// Collect IDs before deleting so we can clean up attachment directories.
	rows, err := s.db.Query("SELECT id FROM messages WHERE source = 'trap'")
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()

	if _, err := s.db.Exec("DELETE FROM messages WHERE source = 'trap'"); err != nil {
		return err
	}

	for _, id := range ids {
		os.RemoveAll(filepath.Join(s.dataDir, "attachments", id))
	}
	return nil
}

// ── Attachments ───────────────────────────────────────────────────────────────

func (s *sqliteStore) GetAttachment(id string) (*model.Attachment, error) {
	var att model.Attachment
	err := s.db.QueryRow(
		"SELECT id, message_id, filename, mime_type, size, path FROM attachments WHERE id = ?", id,
	).Scan(&att.ID, &att.MessageID, &att.Filename, &att.MIMEType, &att.Size, &att.Path)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &att, nil
}

func (s *sqliteStore) ListAttachments(messageID string) ([]model.Attachment, error) {
	rows, err := s.db.Query(
		"SELECT id, message_id, filename, mime_type, size, path FROM attachments WHERE message_id = ?",
		messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var atts []model.Attachment
	for rows.Next() {
		var att model.Attachment
		if err := rows.Scan(&att.ID, &att.MessageID, &att.Filename, &att.MIMEType, &att.Size, &att.Path); err != nil {
			return nil, err
		}
		atts = append(atts, att)
	}
	return atts, rows.Err()
}

// ── helpers ───────────────────────────────────────────────────────────────────

func marshalAddrs(addrs []string) string {
	if len(addrs) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(addrs)
	return string(b)
}

func unmarshalAddrs(s string) []string {
	if s == "" || s == "[]" {
		return nil
	}
	var addrs []string
	json.Unmarshal([]byte(s), &addrs) //nolint:errcheck
	return addrs
}

// scanMessageRow scans a list-view row (no body or raw_mime columns).
func scanMessageRow(rows *sql.Rows) (*model.Message, error) {
	var (
		msg               model.Message
		toJSON, ccJSON, bccJSON string
		receivedAt        sql.NullTime
		isRead, isStarred int
	)
	err := rows.Scan(
		&msg.ID, &msg.MessageID, &msg.Subject, &msg.FromAddr,
		&toJSON, &ccJSON, &bccJSON,
		&isRead, &isStarred, &receivedAt, &msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	msg.Source    = "trap"
	msg.Folder    = "trap"
	msg.ToAddrs   = unmarshalAddrs(toJSON)
	msg.CCAddrs   = unmarshalAddrs(ccJSON)
	msg.BCCAddrs  = unmarshalAddrs(bccJSON)
	msg.IsRead    = isRead != 0
	msg.IsStarred = isStarred != 0
	if receivedAt.Valid {
		t := receivedAt.Time
		msg.ReceivedAt = &t
	}
	return &msg, nil
}

