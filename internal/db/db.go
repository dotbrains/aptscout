package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB with aptscout-specific operations.
type DB struct {
	*sql.DB
}

// DefaultPath returns the default database path.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "aptscout", "aptscout.db")
}

// Open opens (or creates) the SQLite database at the given path and runs migrations.
func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", path+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	db := &DB{sqlDB}
	if err := db.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

// OpenMemory opens an in-memory database for testing.
func OpenMemory() (*DB, error) {
	sqlDB, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(on)")
	if err != nil {
		return nil, err
	}

	db := &DB{sqlDB}
	if err := db.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
