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

		if user.ID == 0 {
			t.Fatal("User ID is zero")
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

func TestLogin(t *testing.T) {
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

	// Register a test user
	testUser, err := Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("LoginSuccess", func(t *testing.T) {
		user, err := Login("test@example.com", "password123")
		if err != nil {
			t.Errorf("Failed to login: %v", err)
		}

		if user == nil {
			t.Fatal("Expected user, got nil")
		}

		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
		}

		if user.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", user.Username)
		}

		if user.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
		}
	})

	t.Run("LoginInvalidEmail", func(t *testing.T) {
		_, err := Login("nonexistent@example.com", "password123")
		if err == nil {
			t.Error("Expected error when logging in with non-existent email, got nil")
		}
	})

	t.Run("LoginInvalidPassword", func(t *testing.T) {
		_, err := Login("test@example.com", "wrongpassword")
		if err == nil {
			t.Error("Expected error when logging in with incorrect password, got nil")
		}
	})
}

func TestUser_VerifyPassword(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := user.HashPassword()
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !user.VerifyPassword("password123") {
		t.Error("VerifyPassword failed for correct password")
	}

	if user.VerifyPassword("wrongpassword") {
		t.Error("VerifyPassword succeeded for incorrect password")
	}
}
