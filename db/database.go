package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	DefaultDBPath = "/var/lib/lightweight-php/lightweight-php.db"
)

type Database struct {
	*sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	if dbPath == "" {
		dbPath = DefaultDBPath
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{DB: db}
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

func (db *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS php_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT NOT NULL UNIQUE,
		installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active',
		package_manager TEXT,
		os_family TEXT
	);

	CREATE TABLE IF NOT EXISTS pools (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		php_version TEXT NOT NULL,
		socket_path TEXT NOT NULL,
		config_path TEXT NOT NULL,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (php_version) REFERENCES php_versions(version),
		UNIQUE(username, php_version)
	);

	CREATE INDEX IF NOT EXISTS idx_pools_username ON pools(username);
	CREATE INDEX IF NOT EXISTS idx_pools_php_version ON pools(php_version);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *Database) Close() error {
	return db.DB.Close()
}
