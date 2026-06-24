package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maeregzewdu/catchy/internal/model"
)

// ── CreateMessage ─────────────────────────────────────────────────────────────

func (s *sqliteStore) CreateMessage(msg *model.Message, attachments []model.Attachment) error {
	// Skip duplicates: same account + folder + message_id already stored.
	if msg.MessageID != "" && msg.AccountID != nil {
		var n int
		s.db.QueryRow( //nolint:errcheck
			"SELECT COUNT(*) FROM messages WHERE account_id=? AND folder=? AND message_id=?",
			*msg.AccountID, msg.Folder, msg.MessageID,
		).Scan(&n)
		if n > 0 {
			return nil
		}
	}

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
			 is_read, is_starred, sent_at, received_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.AccountID, msg.Source, msg.MessageID, msg.Folder, msg.Subject,
		msg.FromAddr,
		marshalAddrs(msg.ToAddrs),
		marshalAddrs(msg.CCAddrs),
		marshalAddrs(msg.BCCAddrs),
		msg.BodyText, msg.BodyHTML, msg.RawMIME,
		boolToInt(msg.IsRead), boolToInt(msg.IsStarred),
		msg.SentAt, msg.ReceivedAt, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting message: %w", err)
	}

	for _, att := range attachments {
		if _, err = tx.Exec(`
			INSERT INTO attachments (id, message_id, filename, mime_type, size, path)
			VALUES (?, ?, ?, ?, ?, ?)`,
			att.ID, att.MessageID, att.Filename, att.MIMEType, att.Size, att.Path,
		); err != nil {
			return fmt.Errorf("inserting attachment %s: %w", att.Filename, err)
		}
	}

	return tx.Commit()
}

// ── GetMessage ────────────────────────────────────────────────────────────────

func (s *sqliteStore) GetMessage(id string) (*model.Message, error) {
	var msg model.Message
	var accountID sql.NullString
	var toJSON, ccJSON, bccJSON string
	var bodyText, bodyHTML sql.NullString
	var sentAt, receivedAt sql.NullTime
	var isRead, isStarred int

	err := s.db.QueryRow(`
		SELECT id, account_id, source, message_id, folder, subject,
		       from_addr, to_addrs, cc_addrs, bcc_addrs,
		       body_text, body_html, raw_mime,
		       is_read, is_starred, sent_at, received_at, created_at
		FROM messages WHERE id = ?`, id).
		Scan(&msg.ID, &accountID, &msg.Source, &msg.MessageID, &msg.Folder, &msg.Subject,
			&msg.FromAddr, &toJSON, &ccJSON, &bccJSON,
			&bodyText, &bodyHTML, &msg.RawMIME,
			&isRead, &isStarred, &sentAt, &receivedAt, &msg.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if accountID.Valid {
		msg.AccountID = &accountID.String
	}
	msg.ToAddrs = unmarshalAddrs(toJSON)
	msg.CCAddrs = unmarshalAddrs(ccJSON)
	msg.BCCAddrs = unmarshalAddrs(bccJSON)
	msg.BodyText = bodyText.String
	msg.BodyHTML = bodyHTML.String
	msg.IsRead = isRead != 0
	msg.IsStarred = isStarred != 0
	if sentAt.Valid {
		t := sentAt.Time
		msg.SentAt = &t
	}
	if receivedAt.Valid {
		t := receivedAt.Time
		msg.ReceivedAt = &t
	}
	return &msg, nil
}

// ── ListMessages ──────────────────────────────────────────────────────────────

func (s *sqliteStore) ListMessages(accountID, folder string, f MessageFilter) ([]*model.Message, error) {
	q := `SELECT id, account_id, source, message_id, folder, subject,
	             from_addr, to_addrs, cc_addrs, bcc_addrs,
	             is_read, is_starred, sent_at, received_at, created_at
	      FROM messages
	      WHERE account_id = ? AND folder = ?`
	args := []any{accountID, folder}

	if f.Read != nil {
		q += " AND is_read = ?"
		args = append(args, boolToInt(*f.Read))
	}
	if f.Starred != nil {
		q += " AND is_starred = ?"
		args = append(args, boolToInt(*f.Starred))
	}
	q += " ORDER BY received_at DESC"
	if f.Limit > 0 {
		q += " LIMIT ? OFFSET ?"
		args = append(args, f.Limit, f.Offset)
	}

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		var msg model.Message
		var accID sql.NullString
		var toJSON, ccJSON, bccJSON string
		var sentAt, receivedAt sql.NullTime
		var isRead, isStarred int

		if err := rows.Scan(
			&msg.ID, &accID, &msg.Source, &msg.MessageID, &msg.Folder, &msg.Subject,
			&msg.FromAddr, &toJSON, &ccJSON, &bccJSON,
			&isRead, &isStarred, &sentAt, &receivedAt, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}

		if accID.Valid {
			msg.AccountID = &accID.String
		}
		msg.ToAddrs = unmarshalAddrs(toJSON)
		msg.CCAddrs = unmarshalAddrs(ccJSON)
		msg.BCCAddrs = unmarshalAddrs(bccJSON)
		msg.IsRead = isRead != 0
		msg.IsStarred = isStarred != 0
		if sentAt.Valid {
			t := sentAt.Time
			msg.SentAt = &t
		}
		if receivedAt.Valid {
			t := receivedAt.Time
			msg.ReceivedAt = &t
		}
		msgs = append(msgs, &msg)
	}
	return msgs, rows.Err()
}

// ── UpdateMessage ─────────────────────────────────────────────────────────────

func (s *sqliteStore) UpdateMessage(msg *model.Message) error {
	_, err := s.db.Exec(`
		UPDATE messages
		SET subject=?, from_addr=?, to_addrs=?, cc_addrs=?, bcc_addrs=?,
		    body_text=?, body_html=?, is_read=?, is_starred=?, sent_at=?, received_at=?
		WHERE id=?`,
		msg.Subject, msg.FromAddr,
		marshalAddrs(msg.ToAddrs),
		marshalAddrs(msg.CCAddrs),
		marshalAddrs(msg.BCCAddrs),
		msg.BodyText, msg.BodyHTML,
		boolToInt(msg.IsRead), boolToInt(msg.IsStarred),
		msg.SentAt, msg.ReceivedAt,
		msg.ID,
	)
	return err
}

// ── PatchMessage ──────────────────────────────────────────────────────────────

func (s *sqliteStore) PatchMessage(id string, read *bool, starred *bool) error {
	switch {
	case read != nil && starred != nil:
		_, err := s.db.Exec("UPDATE messages SET is_read=?, is_starred=? WHERE id=?",
			boolToInt(*read), boolToInt(*starred), id)
		return err
	case read != nil:
		_, err := s.db.Exec("UPDATE messages SET is_read=? WHERE id=?", boolToInt(*read), id)
		return err
	case starred != nil:
		_, err := s.db.Exec("UPDATE messages SET is_starred=? WHERE id=?", boolToInt(*starred), id)
		return err
	}
	return nil
}

// ── DeleteMessage ─────────────────────────────────────────────────────────────

func (s *sqliteStore) DeleteMessage(id string) error {
	res, err := s.db.Exec("DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	os.RemoveAll(filepath.Join(s.dataDir, "attachments", id))
	return nil
}

// ── Sync state ────────────────────────────────────────────────────────────────

func (s *sqliteStore) GetSyncState(accountID, folder string) (*model.SyncState, error) {
	var state model.SyncState
	var lastSync sql.NullTime

	err := s.db.QueryRow(`
		SELECT account_id, folder, uid_next, uid_validity, last_sync
		FROM sync_state WHERE account_id = ? AND folder = ?`,
		accountID, folder).
		Scan(&state.AccountID, &state.Folder, &state.UIDNext, &state.UIDValidity, &lastSync)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if lastSync.Valid {
		t := lastSync.Time
		state.LastSync = &t
	}
	return &state, nil
}

func (s *sqliteStore) UpsertSyncState(state *model.SyncState) error {
	_, err := s.db.Exec(`
		INSERT INTO sync_state (account_id, folder, uid_next, uid_validity, last_sync)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(account_id, folder) DO UPDATE SET
			uid_next     = excluded.uid_next,
			uid_validity = excluded.uid_validity,
			last_sync    = excluded.last_sync`,
		state.AccountID, state.Folder, state.UIDNext, state.UIDValidity, state.LastSync,
	)
	return err
}

// ── SearchMessages ────────────────────────────────────────────────────────────

func (s *sqliteStore) SearchMessages(q, accountID, source string) ([]*model.Message, error) {
	ftsQ := sanitizeFTSQuery(q)
	if ftsQ == "" {
		return nil, nil
	}

	query := `
		SELECT m.id, m.account_id, m.source, m.message_id, m.folder, m.subject,
		       m.from_addr, m.to_addrs, m.cc_addrs, m.bcc_addrs,
		       m.is_read, m.is_starred, m.sent_at, m.received_at, m.created_at
		FROM messages m
		JOIN messages_fts ON messages_fts.rowid = m.rowid
		WHERE messages_fts MATCH ?`
	args := []any{ftsQ}

	if accountID != "" {
		query += " AND m.account_id = ?"
		args = append(args, accountID)
	}
	if source != "" {
		query += " AND m.source = ?"
		args = append(args, source)
	}
	query += " ORDER BY messages_fts.rank LIMIT 50"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		var msg model.Message
		var accID sql.NullString
		var toJSON, ccJSON, bccJSON string
		var sentAt, receivedAt sql.NullTime
		var isRead, isStarred int

		if err := rows.Scan(
			&msg.ID, &accID, &msg.Source, &msg.MessageID, &msg.Folder, &msg.Subject,
			&msg.FromAddr, &toJSON, &ccJSON, &bccJSON,
			&isRead, &isStarred, &sentAt, &receivedAt, &msg.CreatedAt,
		); err != nil {
			return nil, err
		}

		if accID.Valid {
			msg.AccountID = &accID.String
		}
		msg.ToAddrs = unmarshalAddrs(toJSON)
		msg.CCAddrs = unmarshalAddrs(ccJSON)
		msg.BCCAddrs = unmarshalAddrs(bccJSON)
		msg.IsRead = isRead != 0
		msg.IsStarred = isStarred != 0
		if sentAt.Valid {
			t := sentAt.Time
			msg.SentAt = &t
		}
		if receivedAt.Valid {
			t := receivedAt.Time
			msg.ReceivedAt = &t
		}
		msgs = append(msgs, &msg)
	}
	return msgs, rows.Err()
}

// sanitizeFTSQuery strips characters that would break the FTS5 query parser
// and appends a wildcard to each word for prefix matching.
func sanitizeFTSQuery(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range q {
		switch r {
		case '"', '(', ')', '*', '^', '{', '}', '[', ']', '~', '+', ':', '-':
			// strip FTS5 special chars
		default:
			b.WriteRune(r)
		}
	}
	clean := strings.TrimSpace(b.String())
	if clean == "" {
		return ""
	}
	words := strings.Fields(clean)
	for i, w := range words {
		words[i] = w + "*"
	}
	return strings.Join(words, " ")
}

// ── helpers ───────────────────────────────────────────────────────────────────

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
