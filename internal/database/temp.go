package database

import (
	"fmt"
	"os"
	"path/filepath"
)

type TempDBDir struct {
	dir    string
	dbPath string
}

func (t *TempDBDir) Dir() string {
	return t.dir
}

func (t *TempDBDir) DBPath() string {
	return t.dbPath
}

func (t *TempDBDir) Cleanup() error {
	return os.RemoveAll(t.dir)
}

func NewTempDir() (*TempDBDir, error) {
	dbName := "local.db"
	// Create a temporary directory for the local database
	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Set up the local database path
	dbPath := filepath.Join(dir, dbName)

	return &TempDBDir{
		dir:    dir,
		dbPath: dbPath,
	}, nil
}
