package utils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func SetupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	return db, mock
}
