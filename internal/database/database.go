package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shuvo-paul/uptimebot/internal/config"
	"github.com/tursodatabase/go-libsql"
)

func InitDatabase(config config.DatabaseConfig, tempDBDir *TempDBDir) (*sql.DB, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	if config.Token == "" {
		return nil, fmt.Errorf("database token is empty")
	}

	// Create a new embedded replica connector
	connector, err := libsql.NewEmbeddedReplicaConnector(tempDBDir.dbPath, config.URL,
		libsql.WithAuthToken(config.Token),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connector: %w", err)
	}

	// Open the database with the connector
	db := sql.OpenDB(connector)

	// Verify the connection
	if err = db.Ping(); err != nil {
		db.Close()
		connector.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to Turso database successfully with local replica")
	return db, nil
}
