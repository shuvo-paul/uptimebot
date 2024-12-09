package repository

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shuvo-paul/sitemonitor/internal/database/migrations"
)

var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(fmt.Sprintf("Error setting up test database: %v", err))
	}
	defer db.Close()

	migrations.SetupMigration(db)

	code := m.Run()
	os.Exit(code)
}
