package repository

import (
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSaveUser(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	userRepo := NewUserRepository(db)

	user := &model.User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	savedUser, err := userRepo.SaveUser(user)

	assert.NoError(t, err)
	assert.NotZero(t, savedUser.ID)
	assert.Equal(t, user.Name, savedUser.Name)
}

func TestEmailExists(t *testing.T) {
	db := testutil.NewInMemoryDB()
	userRepo := NewUserRepository(db)
	defer db.Close()

	// Create a test user first
	user := &model.User{
		Name:     "testuser",
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
	db := testutil.NewInMemoryDB()
	userRepo := NewUserRepository(db)
	defer db.Close()

	// Create a test user first
	expectedUser := &model.User{
		Name:     "testuser",
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
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Password, user.Password)
	})

	t.Run("By ID", func(t *testing.T) {
		user, err := userRepo.GetUserByID(expectedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Email, user.Email)
	})
}
