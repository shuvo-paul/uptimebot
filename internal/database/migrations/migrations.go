package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed *.sql
var sqlFs embed.FS

func SetupMigration(db *sql.DB) {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: sqlFs,
		Root:       ".",
	}

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		fmt.Printf("migrations failed: %v", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)
}
