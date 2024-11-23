package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_HashPassword(t *testing.T) {
	user := &User{
		Name:     "testuser",
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
		Name:     "testuser",
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

func TestValidatePassword(t *testing.T) {
	user := &User{
		Name:  "test",
		Email: "test@example.org",
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "too short password",
			password: "abcd",
			wantErr:  true,
			errMsg:   "password must be between 6 and 12 characters",
		},
		{
			name:     "too long password",
			password: "abcdefg123456789",
			wantErr:  true,
			errMsg:   "password must be between 6 and 12 characters",
		},
		{
			name:     "missing number",
			password: "abcdefg!@",
			wantErr:  true,
			errMsg:   "password must contain at least one number",
		},
		{
			name:     "missing symbol",
			password: "abcdefg1234",
			wantErr:  true,
			errMsg:   "password must contain at least one symbol",
		},
		{
			name:     "valid password",
			password: "abcdefg1234@",
			wantErr:  false,
			errMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user.Password = tt.password
			err := user.ValidatePassword()

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
