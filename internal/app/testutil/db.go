package testutil

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuvo-paul/sitemonitor/internal/database/migrations"
)

// NewInMemoryDB creates a new in-memory database connection for testing
func NewInMemoryDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(fmt.Sprintf("Error setting up test database: %v", err))
	}

	migrations.SetupMigration(db)

	return db
}
