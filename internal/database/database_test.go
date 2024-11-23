package database

import (
	"testing"

	// Import the godotenv package
	"github.com/joho/godotenv"
)

func TestInitDatabase(t *testing.T) {
	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Call the function to test
	db, err := InitDatabase()
	if err != nil {
		t.Fatalf("Error initializing database: %v", err)
	}

	// Check if the DB variable is not nil
	if db == nil {
		t.Fatal("Expected DB to be initialized, but it was nil")
	}

	// Check if the database connection is valid
	if err := db.Ping(); err != nil {
		t.Fatalf("Expected to connect to the database, but got error: %v", err)
	}

	// Clean up
	db.Close()
}
