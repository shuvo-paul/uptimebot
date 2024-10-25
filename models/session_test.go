package models

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/shuvo-paul/sitemonitor/config"
)

func setupTestDB(t *testing.T) {
	// Load environment variables from .env.test file
	if err := godotenv.Load("../.env.test"); err != nil {
		t.Fatalf("Error loading .env.test file: %v", err)
	}

	// Initialize the database connection
	if err := config.InitDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	_, err := config.DB.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatalf("Failed to clear sessions table: %v", err)
	}

	_, err = config.DB.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clear users table: %v", err)
	}
}

func TestSession(t *testing.T) {
	setupTestDB(t)
	defer config.DB.Close()

	user, err := Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("User ID is zero")
	}

	session := &Session{}

	t.Run("CreateSession", func(t *testing.T) {
		session, err = CreateSession(user.ID)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		if session == nil {
			t.Fatal("Session is nil")
		}

		if session.UserID != user.ID {
			t.Errorf("Session user ID does not match user ID: %d != %d", session.UserID, user.ID)
		}
		if session.Token == "" {
			t.Fatal("Session token is empty")
		}
		if session.CreatedAt.IsZero() {
			t.Fatal("Session created at is zero")
		}
		if session.ExpiresAt.IsZero() {
			t.Fatal("Session expires at is zero")
		}
	})

	t.Run("GetSessionByToken", func(t *testing.T) {
		session, err = GetSessionByToken(session.Token)
		if err != nil {
			t.Fatalf("Failed to get session by token: %v", err)
		}

		if session == nil {
			t.Fatal("Session is nil")
		}

		if session.UserID != user.ID {
			t.Errorf("Session user ID does not match user ID: %d != %d", session.UserID, user.ID)
		}
	})

	t.Run("DeleteSession", func(t *testing.T) {
		err = DeleteSession(session.ID)
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}
	})
}
