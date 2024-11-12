package migrations

import (
	"database/sql"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
)

func SetupMigration(db *sql.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "./migrations",
	}

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		fmt.Printf("migrations failed: %v", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)
}
