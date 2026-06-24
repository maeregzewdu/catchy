package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/maeregzewdu/catchy/internal/model"
)

func (s *sqliteStore) CreateAccount(a *model.Account) error {
	enc, err := encrypt(s.key, a.Password)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO accounts
			(id, name, email, smtp_host, smtp_port, imap_host, imap_port, username, password, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Email,
		a.SMTPHost, a.SMTPPort,
		a.IMAPHost, a.IMAPPort,
		a.Username, enc, a.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("inserting account: %w", err)
	}
	return nil
}

func (s *sqliteStore) GetAccount(id string) (*model.Account, error) {
	var a model.Account
	var enc string

	err := s.db.QueryRow(`
		SELECT id, name, email, smtp_host, smtp_port, imap_host, imap_port, username, password, created_at
		FROM accounts WHERE id = ?`, id).
		Scan(&a.ID, &a.Name, &a.Email,
			&a.SMTPHost, &a.SMTPPort,
			&a.IMAPHost, &a.IMAPPort,
			&a.Username, &enc, &a.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	plain, err := decrypt(s.key, enc)
	if err != nil {
		return nil, fmt.Errorf("decrypting password: %w", err)
	}
	a.Password = plain
	return &a, nil
}

func (s *sqliteStore) ListAccounts() ([]*model.Account, error) {
	rows, err := s.db.Query(`
		SELECT id, name, email, smtp_host, smtp_port, imap_host, imap_port, username, created_at
		FROM accounts ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		var a model.Account
		if err := rows.Scan(
			&a.ID, &a.Name, &a.Email,
			&a.SMTPHost, &a.SMTPPort,
			&a.IMAPHost, &a.IMAPPort,
			&a.Username, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		// Password intentionally not scanned for list views — never needed in responses.
		accounts = append(accounts, &a)
	}
	return accounts, rows.Err()
}

func (s *sqliteStore) UpdateAccount(a *model.Account) error {
	var err error

	if a.Password != "" {
		enc, encErr := encrypt(s.key, a.Password)
		if encErr != nil {
			return fmt.Errorf("encrypting password: %w", encErr)
		}
		_, err = s.db.Exec(`
			UPDATE accounts
			SET name=?, email=?, smtp_host=?, smtp_port=?, imap_host=?, imap_port=?, username=?, password=?
			WHERE id=?`,
			a.Name, a.Email, a.SMTPHost, a.SMTPPort, a.IMAPHost, a.IMAPPort, a.Username, enc, a.ID,
		)
	} else {
		_, err = s.db.Exec(`
			UPDATE accounts
			SET name=?, email=?, smtp_host=?, smtp_port=?, imap_host=?, imap_port=?, username=?
			WHERE id=?`,
			a.Name, a.Email, a.SMTPHost, a.SMTPPort, a.IMAPHost, a.IMAPPort, a.Username, a.ID,
		)
	}

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

func (s *sqliteStore) DeleteAccount(id string) error {
	res, err := s.db.Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
