package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func InitDatabase() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	dbToken := os.Getenv("TURSO_AUTH_TOKEN")
	if dbToken == "" {
		return nil, fmt.Errorf("TURSO_AUTH_TOKEN environment variable is not set")
	}
	if dbURL == "" {
		return nil, fmt.Errorf("TURSO_DATABASE_URL environment variable is not set")
	}

	var err error
	db, err := sql.Open("libsql", dbURL+"?authToken="+dbToken)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close() // Close the connection if Ping fails
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to Turso database successfully")
	return db, nil
}
