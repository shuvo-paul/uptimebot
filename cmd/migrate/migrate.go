package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/shuvo-paul/sitemonitor/database"
)

func main() {
	// Define a flag for using the test environment
	useTestEnv := flag.Bool("test", false, "Use the test environment (.env.test)")
	flag.Parse()

	// Load the appropriate environment file
	if err := loadEnvFile(*useTestEnv); err != nil {
		log.Panicf("Failed to load environment file: %v", err)
	}

	if len(flag.Args()) < 1 {
		log.Panic("Please provide a command (up, down, create, redo, status, fresh)")
	}

	command := flag.Arg(0)

	// Initialize the database only if the command is not "new"
	var db *sql.DB
	if command != "new" {
		db, err := database.InitDatabase()
		if err != nil {
			log.Panicf("Failed to initialize database: %v", err)
		}
		defer db.Close() // Ensure the database connection is closed
	}

	migrationsDir := "./migrations"

	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	commands := map[string]func() error{
		"up": func() error {
			n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
			if err != nil {
				return err
			}
			fmt.Printf("Applied %d migrations!\n", n)
			return nil
		},
		"down": func() error {
			n, err := migrate.ExecMax(db, "sqlite3", migrations, migrate.Down, 1)
			if err != nil {
				return err
			}
			fmt.Printf("Rolled back %d migration!\n", n)
			return nil
		},
		"new": func() error {
			if len(flag.Args()) < 2 {
				log.Panic("Please provide a name for the migration")
			}
			name := flag.Arg(1)
			timestamp := time.Now().Format("20060102150405")
			migrationFile := fmt.Sprintf("%s/%s_%s.sql", migrationsDir, timestamp, name)
			content := `-- +migrate Up


-- +migrate Down
`
			if err := os.WriteFile(migrationFile, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Printf("Created migration file: %s\n", migrationFile)
			return nil
		},
		"redo": func() error {
			n, err := migrate.ExecMax(db, "sqlite3", migrations, migrate.Down, 1)
			if err != nil {
				return err
			}
			fmt.Printf("Rolled back %d migration!\n", n)

			n, err = migrate.ExecMax(db, "sqlite3", migrations, migrate.Up, 1)
			if err != nil {
				return err
			}
			fmt.Printf("Reapplied %d migration!\n", n)
			return nil
		},
		"status": func() error {
			records, err := migrate.GetMigrationRecords(db, "sqlite3")
			if err != nil {
				return err
			}

			if len(records) == 0 {
				fmt.Println("No migrations applied yet.")
				return nil
			}

			fmt.Println("Applied migrations:")
			for _, record := range records {
				fmt.Printf("- %s\n", record.Id)
			}
			return nil
		},
		"fresh": func() error {
			// Rollback all migrations
			n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Down)
			if err != nil {
				return err
			}
			fmt.Printf("Rolled back %d migrations!\n", n)

			// Apply all migrations
			n, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
			if err != nil {
				return err
			}
			fmt.Printf("Applied %d migrations!\n", n)
			return nil
		},
	}

	if cmdFunc, exists := commands[command]; exists {
		if err := cmdFunc(); err != nil {
			log.Panicf("Failed to execute command %s: %v", command, err)
		}
	} else {
		log.Panicf("Unknown command: %s", command)
	}

	fmt.Println("Command executed successfully!")
}

// loadEnvFile loads environment variables from a specified file
func loadEnvFile(useTestEnv bool) error {
	filename := ".env"
	if useTestEnv {
		filename = ".env.test"
	}
	return godotenv.Load(filename)
}
