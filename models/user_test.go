package models

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/shuvo-paul/sitemonitor/config"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_HashPassword(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := user.HashPassword()
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}

	// Check that the password was actually hashed
	if user.Password == "password123" {
		t.Error("Password was not hashed")
	}

	// Verify that the hashed password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
	if err != nil {
		t.Error("Hashed password does not match original password")
	}
}

func TestRegister(t *testing.T) {
	// Load environment variables from .env.test file
	if err := godotenv.Load("../.env.test"); err != nil {
		t.Fatalf("Error loading .env.test file: %v", err)
	}

	// Initialize the database connection
	if err := config.InitDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.DB.Close()

	// Clear the users table before the test
	_, err := config.DB.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clear users table: %v", err)
	}

	t.Run("RegisterNewUser", func(t *testing.T) {
		// Register a new user
		user, err := Register("newuser", "new@example.com", "password123")
		if err != nil {
			t.Errorf("Failed to register user: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user, got nil")
		}

		if user.Username != "newuser" {
			t.Errorf("Expected username 'newuser', got '%s'", user.Username)
		}

		if user.Email != "new@example.com" {
			t.Errorf("Expected email 'new@example.com', got '%s'", user.Email)
		}

		// Check that the password was hashed
		if user.Password == "password123" {
			t.Error("Password was not hashed")
		}
	})

	t.Run("RegisterDuplicateEmail", func(t *testing.T) {
		// Attempt to register another user with the same email
		_, err := Register("anotheruser", "new@example.com", "password123")
		if err == nil {
			t.Error("Expected error when registering with an existing email, got nil")
		} else {
			t.Logf("Correctly failed to register with existing email: %v", err)
		}
	})
}
