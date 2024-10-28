package models

import (
	"testing"

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
