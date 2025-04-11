package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	migrate "github.com/rubenv/sql-migrate"
)

var templateContent = `
-- +migrate Up

-- +migrate Down
`

func migrateTool(db *sql.DB) {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("No migration command provided.")
	}

	switch args[2] {
	case "down":
		// Code for down migration
		n, err := migrate.Exec(db, "postgres", migrations, migrate.Down)
		if err != nil {
			fmt.Printf("down migration failed: %v", err)
		}
		fmt.Printf("Rolled back %d migrations!\n", n)
	case "new":
		if len(args) < 4 {
			fmt.Println("No migration name provided.")
		}
		name := args[3]
		timestamp := time.Now().Format("20060102150405")
		migrationFile := fmt.Sprintf("%s/%s_%s.sql", "./internal/database/migrations", timestamp, strings.TrimSpace(name))

		if err := os.WriteFile(migrationFile, []byte(templateContent), 0644); err != nil {
			fmt.Printf("Failed to create migration file: %v\n", err)
		}
		fmt.Printf("Created migration file: %s\n", migrationFile)
	case "redo":
		// Rollback the last migration
		n, err := migrate.ExecMax(db, "postgres", migrations, migrate.Down, 1)
		if err != nil {
			fmt.Printf("redo down migration failed: %v\n", err)
		}
		fmt.Printf("Rolled back %d migration(s)!\n", n)

		// Apply the last migration again
		n, err = migrate.ExecMax(db, "postgres", migrations, migrate.Up, 1)
		if err != nil {
			fmt.Printf("redo up migration failed: %v\n", err)
		}
		fmt.Printf("Applied %d migration(s)!\n", n)

	case "status":

	case "skip":

	case "up":
		n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
		if err != nil {
			fmt.Printf("migrations failed: %v", err)
		}
		fmt.Printf("Applied %d migrations!\n", n)

	case "fresh":
		n, err := migrate.ExecMax(db, "postgres", migrations, migrate.Down, -1)
		if err != nil {
			fmt.Printf("Failed to drop all migrations: %v\n", err)
		}
		fmt.Printf("Dropped %d migrations!\n", n)

		n, err = migrate.ExecMax(db, "postgres", migrations, migrate.Up, 0)
		if err != nil {
			fmt.Printf("Failed to apply all migrations: %v\n", err)
		}
		fmt.Printf("Applied %d migrations!\n", n)
	default:
		fmt.Println("Invalid migration command")
	}
}
