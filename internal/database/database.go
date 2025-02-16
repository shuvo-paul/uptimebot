package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shuvo-paul/uptimebot/internal/config"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func InitDatabase(config config.DatabaseConfig) (*sql.DB, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	if config.Token == "" {
		return nil, fmt.Errorf("database token is empty")
	}
	db, err := sql.Open("libsql", config.URL+"?authToken="+config.Token)
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
