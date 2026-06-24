package store

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type sqliteStore struct {
	db      *sql.DB
	dataDir string
	key     []byte
}

// New opens (or creates) the SQLite database in dataDir and runs all pending migrations.
// key must be 32 bytes (AES-256); it is used to encrypt account passwords at rest.
func New(dataDir string, key []byte) (Store, error) {
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

	s := &sqliteStore{db: db, dataDir: dataDir, key: key}
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

// splitStatements splits a SQL string on semicolons, correctly handling
// string literals, comments, and BEGIN...END trigger bodies.
func splitStatements(sql string) []string {
	var out []string
	var buf strings.Builder
	depth := 0 // BEGIN...END nesting depth (for trigger bodies)
	i, n := 0, len(sql)

	for i < n {
		ch := sql[i]

		// Single-quoted string literal '...'
		if ch == '\'' {
			buf.WriteByte(ch)
			i++
			for i < n {
				c := sql[i]
				buf.WriteByte(c)
				i++
				if c == '\'' {
					if i < n && sql[i] == '\'' { // escaped ''
						buf.WriteByte(sql[i])
						i++
					} else {
						break
					}
				}
			}
			continue
		}

		// Double-quoted identifier "..."
		if ch == '"' {
			buf.WriteByte(ch)
			i++
			for i < n && sql[i] != '"' {
				buf.WriteByte(sql[i])
				i++
			}
			if i < n {
				buf.WriteByte(sql[i])
				i++
			}
			continue
		}

		// Line comment -- ...
		if ch == '-' && i+1 < n && sql[i+1] == '-' {
			buf.WriteByte(ch)
			i++
			for i < n && sql[i] != '\n' {
				buf.WriteByte(sql[i])
				i++
			}
			continue
		}

		// Block comment /* ... */
		if ch == '/' && i+1 < n && sql[i+1] == '*' {
			buf.WriteByte(ch)
			i++
			for i < n {
				c := sql[i]
				buf.WriteByte(c)
				i++
				if c == '*' && i < n && sql[i] == '/' {
					buf.WriteByte('/')
					i++
					break
				}
			}
			continue
		}

		// Track BEGIN...END nesting so semicolons inside trigger bodies are not split points.
		if i == 0 || !isIdentByte(sql[i-1]) {
			if sqlKeywordAt(sql, i, "BEGIN") {
				depth++
			} else if sqlKeywordAt(sql, i, "END") && depth > 0 {
				depth--
			}
		}

		if ch == ';' && depth == 0 {
			if s := strings.TrimSpace(buf.String()); s != "" {
				out = append(out, s)
			}
			buf.Reset()
			i++
			continue
		}

		buf.WriteByte(ch)
		i++
	}
	if s := strings.TrimSpace(buf.String()); s != "" {
		out = append(out, s)
	}
	return out
}

// sqlKeywordAt reports whether sql[i:] starts with keyword (case-insensitive),
// followed by a non-identifier character.
func sqlKeywordAt(sql string, i int, keyword string) bool {
	if i+len(keyword) > len(sql) {
		return false
	}
	for j := 0; j < len(keyword); j++ {
		c := sql[i+j]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		if c != keyword[j] {
			return false
		}
	}
	after := i + len(keyword)
	return after >= len(sql) || !isIdentByte(sql[after])
}

func isIdentByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// ── Health ────────────────────────────────────────────────────────────────────

func (s *sqliteStore) Ping() error {
	return s.db.QueryRow("SELECT 1").Scan(new(int))
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// ── Maintenance ───────────────────────────────────────────────────────────────

// CleanOrphanAttachments removes any attachment directories that have no
// corresponding message in the database (e.g. left over from crashes).
func (s *sqliteStore) CleanOrphanAttachments() error {
	dir := filepath.Join(s.dataDir, "attachments")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		var n int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE id = ?", id).Scan(&n); err != nil {
			continue
		}
		if n == 0 {
			os.RemoveAll(filepath.Join(dir, id)) //nolint:errcheck
		}
	}
	return nil
}
