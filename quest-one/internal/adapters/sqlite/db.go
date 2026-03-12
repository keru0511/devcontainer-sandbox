// Package sqlite implements all repository ports using SQLite (with optional
// SQLCipher encryption via mattn/go-sqlite3 and CGO).
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3" // CGO driver
)

const schema = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tasks (
	id              TEXT PRIMARY KEY,
	title           TEXT NOT NULL,
	description     TEXT NOT NULL DEFAULT '',
	status          TEXT NOT NULL DEFAULT 'todo',
	memo            TEXT NOT NULL DEFAULT '',
	project_id      TEXT,
	parent_id       TEXT,
	due_date        TEXT,              -- RFC3339 or NULL
	tags            TEXT NOT NULL DEFAULT '[]', -- JSON array
	priority_json   TEXT NOT NULL DEFAULT '{}', -- JSON PriorityScore
	created_at      TEXT NOT NULL,
	updated_at      TEXT NOT NULL,
	completed_at    TEXT
);

CREATE TABLE IF NOT EXISTS source_items (
	id              TEXT PRIMARY KEY,
	source_type     TEXT NOT NULL,
	external_id     TEXT NOT NULL,
	title           TEXT NOT NULL,
	description     TEXT NOT NULL DEFAULT '',
	url             TEXT NOT NULL DEFAULT '',
	priority        INTEGER NOT NULL DEFAULT 0,
	status          TEXT NOT NULL DEFAULT '',
	assignee_id     TEXT NOT NULL DEFAULT '',
	project_id      TEXT NOT NULL DEFAULT '',
	labels          TEXT NOT NULL DEFAULT '[]', -- JSON array
	due_date        TEXT,
	last_synced_at  TEXT NOT NULL,
	created_at      TEXT NOT NULL,
	updated_at      TEXT NOT NULL,
	raw_payload     BLOB,
	UNIQUE(source_type, external_id)
);

CREATE TABLE IF NOT EXISTS task_source_items (
	task_id         TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	source_item_id  TEXT NOT NULL REFERENCES source_items(id) ON DELETE CASCADE,
	source_type     TEXT NOT NULL,
	PRIMARY KEY (task_id, source_item_id)
);

CREATE TABLE IF NOT EXISTS integrations (
	id              TEXT PRIMARY KEY,
	provider        TEXT NOT NULL,
	name            TEXT NOT NULL,
	base_url        TEXT NOT NULL,
	enabled         INTEGER NOT NULL DEFAULT 0,
	sync_filters    TEXT NOT NULL DEFAULT '{}', -- JSON SyncFilters
	last_synced_at  TEXT,
	created_at      TEXT NOT NULL,
	updated_at      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS tasks_fts USING fts5(
	id UNINDEXED,
	title,
	description,
	memo,
	tags,
	content='tasks',
	content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS tasks_ai AFTER INSERT ON tasks BEGIN
	INSERT INTO tasks_fts(rowid, id, title, description, memo, tags)
	VALUES (new.rowid, new.id, new.title, new.description, new.memo, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS tasks_au AFTER UPDATE ON tasks BEGIN
	INSERT INTO tasks_fts(tasks_fts, rowid, id, title, description, memo, tags)
	VALUES ('delete', old.rowid, old.id, old.title, old.description, old.memo, old.tags);
	INSERT INTO tasks_fts(rowid, id, title, description, memo, tags)
	VALUES (new.rowid, new.id, new.title, new.description, new.memo, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS tasks_ad AFTER DELETE ON tasks BEGIN
	INSERT INTO tasks_fts(tasks_fts, rowid, id, title, description, memo, tags)
	VALUES ('delete', old.rowid, old.id, old.title, old.description, old.memo, old.tags);
END;
`

// DB wraps *sql.DB and exposes repository factory methods.
type DB struct {
	db  *sql.DB
	log *slog.Logger
}

// Open opens (or creates) a SQLite database at the given path.
// If passphrase is non-empty, the database is opened with SQLCipher encryption.
func Open(path, passphrase string, log *slog.Logger) (*DB, error) {
	dsn := path
	if passphrase != "" {
		// SQLCipher PRAGMA is injected via the _pragma query parameter.
		dsn = fmt.Sprintf("%s?_pragma_key=%s&_pragma_cipher_page_size=4096", path, passphrase)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}

	// Single-writer to avoid SQLITE_BUSY on WAL mode.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}

	d := &DB{db: db, log: log}
	if err := d.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}

	return d, nil
}

// Close releases the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// migrate runs the schema DDL (idempotent).
func (d *DB) migrate() error {
	_, err := d.db.ExecContext(context.Background(), schema)
	return err
}

// BeginTx starts a new transaction.
func (d *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, nil)
}

// Repositories returns factory instances for all ports.
func (d *DB) TaskRepository() *TaskRepository {
	return &TaskRepository{db: d.db, log: d.log}
}

func (d *DB) SourceItemRepository() *SourceItemRepository {
	return &SourceItemRepository{db: d.db, log: d.log}
}

func (d *DB) IntegrationRepository() *IntegrationRepository {
	return &IntegrationRepository{db: d.db, log: d.log}
}

func (d *DB) SettingsRepository() *SettingsRepository {
	return &SettingsRepository{db: d.db, log: d.log}
}

func (d *DB) SearchIndex() *SearchIndex {
	return &SearchIndex{db: d.db, log: d.log}
}
