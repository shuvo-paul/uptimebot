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

func SetupMigration(db *sql.DB) {
	args := os.Args
	if len(args) > 1 && args[1] == "migrate" {
		migrateTool(db)
		os.Exit(0)
	}

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		fmt.Printf("migrations failed: %v", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)
}
