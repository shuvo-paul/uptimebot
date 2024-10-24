package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var DB *sql.DB

func InitDatabase() error {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	dbToken := os.Getenv("TURSO_AUTH_TOKEN")
	if dbToken == "" {
		return fmt.Errorf("TURSO_AUTH_TOKEN environment variable is not set")
	}
	if dbURL == "" {
		return fmt.Errorf("TURSO_DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = sql.Open("libsql", dbURL+"?authToken="+dbToken)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		DB.Close() // Close the connection if Ping fails
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to Turso database successfully")
	return nil
}
