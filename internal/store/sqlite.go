package store

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/maeregzewdu/catchy/internal/model"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type sqliteStore struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database in dataDir and runs all pending migrations.
func New(dataDir string) (Store, error) {
	dbPath := filepath.Join(dataDir, "catchy.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting %s: %w", pragma, err)
		}
	}

	s := &sqliteStore{db: db}
	if err := s.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

func (s *sqliteStore) runMigrations() error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var applied int
		if err := s.db.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?", entry.Name(),
		).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}

		// Execute each statement separately (files may contain multiple statements).
		for _, stmt := range splitStatements(string(content)) {
			if _, err := s.db.Exec(stmt); err != nil {
				return fmt.Errorf("migration %s: %w", entry.Name(), err)
			}
		}

		if _, err := s.db.Exec(
			"INSERT INTO schema_migrations (version) VALUES (?)", entry.Name(),
		); err != nil {
			return err
		}
	}

	return nil
}

// splitStatements splits a SQL string on semicolons, dropping empty fragments.
func splitStatements(sql string) []string {
	var out []string
	for _, s := range strings.Split(sql, ";") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ── Health ────────────────────────────────────────────────────────────────────

func (s *sqliteStore) Ping() error {
	return s.db.QueryRow("SELECT 1").Scan(new(int))
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// ── Trap (Phase 2) ────────────────────────────────────────────────────────────

func (s *sqliteStore) StoreTrapMessage(_ *model.Message, _ []model.Attachment) error {
	return ErrNotImplemented
}

func (s *sqliteStore) ListTrapMessages(_ TrapFilter) ([]*model.Message, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) GetTrapMessage(_ string) (*model.Message, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) DeleteTrapMessage(_ string) error {
	return ErrNotImplemented
}

func (s *sqliteStore) ClearTrapMessages() error {
	return ErrNotImplemented
}

// ── Accounts (Phase 3) ────────────────────────────────────────────────────────

func (s *sqliteStore) CreateAccount(_ *model.Account) error {
	return ErrNotImplemented
}

func (s *sqliteStore) GetAccount(_ string) (*model.Account, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) ListAccounts() ([]*model.Account, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) UpdateAccount(_ *model.Account) error {
	return ErrNotImplemented
}

func (s *sqliteStore) DeleteAccount(_ string) error {
	return ErrNotImplemented
}

// ── Messages (Phase 4) ────────────────────────────────────────────────────────

func (s *sqliteStore) CreateMessage(_ *model.Message) error {
	return ErrNotImplemented
}

func (s *sqliteStore) GetMessage(_ string) (*model.Message, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) ListMessages(_, _ string, _ MessageFilter) ([]*model.Message, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) UpdateMessage(_ *model.Message) error {
	return ErrNotImplemented
}

func (s *sqliteStore) PatchMessage(_ string, _ *bool, _ *bool) error {
	return ErrNotImplemented
}

func (s *sqliteStore) DeleteMessage(_ string) error {
	return ErrNotImplemented
}

func (s *sqliteStore) GetAttachment(_ string) (*model.Attachment, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) ListAttachments(_ string) ([]model.Attachment, error) {
	return nil, ErrNotImplemented
}

// ── Sync state (Phase 4) ──────────────────────────────────────────────────────

func (s *sqliteStore) GetSyncState(_, _ string) (*model.SyncState, error) {
	return nil, ErrNotImplemented
}

func (s *sqliteStore) UpsertSyncState(_ *model.SyncState) error {
	return ErrNotImplemented
}

// ── Search (Phase 6) ──────────────────────────────────────────────────────────

func (s *sqliteStore) SearchMessages(_, _, _ string) ([]*model.Message, error) {
	return nil, ErrNotImplemented
}
