package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	log.Printf("database initialized at %s", path)
	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS content (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		type         TEXT NOT NULL CHECK(type IN ('article', 'project', 'process')),
		slug         TEXT NOT NULL UNIQUE,
		title        TEXT NOT NULL,
		summary      TEXT DEFAULT '',
		body         TEXT NOT NULL DEFAULT '',
		cover_image  TEXT DEFAULT '',
		status       TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft', 'published')),
		sort_order   INTEGER DEFAULT 0,
		published_at TIMESTAMP,
		created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS images (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		r2_key        TEXT NOT NULL UNIQUE,
		filename      TEXT NOT NULL,
		alt_text      TEXT DEFAULT '',
		width         INTEGER DEFAULT 0,
		height        INTEGER DEFAULT 0,
		size_bytes    INTEGER DEFAULT 0,
		content_type  TEXT DEFAULT '',
		public_url    TEXT NOT NULL,
		created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS content_images (
		content_id INTEGER REFERENCES content(id) ON DELETE CASCADE,
		image_id   INTEGER REFERENCES images(id) ON DELETE CASCADE,
		PRIMARY KEY (content_id, image_id)
	);

	CREATE TABLE IF NOT EXISTS deploys (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		status       TEXT NOT NULL DEFAULT 'pending',
		snapshot_key TEXT,
		gh_run_url   TEXT,
		created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_content_type_status ON content(type, status);
	CREATE INDEX IF NOT EXISTS idx_content_slug ON content(slug);
	CREATE INDEX IF NOT EXISTS idx_content_type_sort ON content(type, sort_order);
	`

	_, err := db.Exec(schema)
	return err
}
