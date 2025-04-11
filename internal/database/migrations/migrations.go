package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed *.sql
var sqlFs embed.FS

var migrations = &migrate.EmbedFileSystemMigrationSource{
	FileSystem: sqlFs,
	Root:       ".",
}

func SetupMigration(db *sql.DB) error {
	args := os.Args
	if len(args) > 1 && args[1] == "migrate" {
		migrateTool(db)
		os.Exit(0)
	}

	// Ensure database connection is valid
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection error: %v", err)
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("migrations failed: %v", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)
	return nil
}
