package repository

import (
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSaveUser(t *testing.T) {
	tx := testutil.GetTestTx(t)

	userRepo := NewUserRepository(tx)

	user := &model.User{

		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	savedUser, err := userRepo.SaveUser(user)

	assert.NoError(t, err)
	assert.NotZero(t, savedUser.ID)
	// Removing Name field assertion

}

func TestEmailExists(t *testing.T) {
	tx := testutil.GetTestTx(t)

	userRepo := NewUserRepository(tx)

	// Create a test user first
	user := &model.User{

		Email:    "existing@example.com",
		Password: "hashedpassword",
	}
	_, err := userRepo.SaveUser(user)
	assert.NoError(t, err)

	t.Run("email exists", func(t *testing.T) {
		exists, err := userRepo.EmailExists("existing@example.com")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("email does not exist", func(t *testing.T) {
		exists, err := userRepo.EmailExists("nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestGetUser(t *testing.T) {
	tx := testutil.GetTestTx(t)
	userRepo := NewUserRepository(tx)

	// Create a test user first
	expectedUser := &model.User{

		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	savedUser, err := userRepo.SaveUser(expectedUser)
	assert.NoError(t, err)
	expectedUser.ID = savedUser.ID

	t.Run("By Email: user found", func(t *testing.T) {
		user, err := userRepo.GetUserByEmail(expectedUser.Email)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Password, user.Password)
	})

	t.Run("By ID", func(t *testing.T) {
		user, err := userRepo.GetUserByID(expectedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
	})
}

func TestUpdateUser(t *testing.T) {
	tx := testutil.GetTestTx(t)
	userRepo := NewUserRepository(tx)
	// Create a test user first
	expectedUser := &model.User{

		Email:    "email@example.org",
		Password: "hashedpassword",
		Verified: false,
	}

	savedUser, err := userRepo.SaveUser(expectedUser)
	assert.NoError(t, err)
	expectedUser.ID = savedUser.ID
	expectedUser.Email = "updated@example.org"
	expectedUser.Verified = true
	updatedUser, err := userRepo.UpdateUser(expectedUser)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, updatedUser.ID)
	assert.Equal(t, expectedUser.Email, updatedUser.Email)
	assert.Equal(t, expectedUser.Verified, updatedUser.Verified)
}

func TestUpdatePassword(t *testing.T) {
	tx := testutil.GetTestTx(t)
	userRepo := NewUserRepository(tx)
	// Create a test user first
	user := &model.User{

		Email:    "test@example.com",
		Password: "oldpassword",
	}
	savedUser, err := userRepo.SaveUser(user)
	assert.NoError(t, err)

	t.Run("successful password update", func(t *testing.T) {
		err := userRepo.UpdatePassword(savedUser.ID, "newhashedpassword")
		assert.NoError(t, err)

		// Verify password was updated
		updatedUser, err := userRepo.GetUserByEmail(user.Email)
		assert.NoError(t, err)
		assert.Equal(t, "newhashedpassword", updatedUser.Password)
	})

	t.Run("non-existent user", func(t *testing.T) {
		err := userRepo.UpdatePassword(9999, "newpassword")
		assert.Error(t, err)
	})
}
