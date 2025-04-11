package repository

import (
	"testing"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/database"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func createTestUser(t *testing.T, tx database.Querier) *model.User {
	userRepo := NewUserRepository(tx)
	user := &model.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	savedUser, err := userRepo.SaveUser(user)
	assert.NoError(t, err)
	return savedUser
}

func TestSessionRepository_Create(t *testing.T) {
	tx := testutil.GetTestTx(t)
	user := createTestUser(t, tx)

	sessionRepo := NewSessionRepository(tx)
	session := &model.Session{
		UserID:    user.ID,
		Token:     "test-token",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := sessionRepo.Create(session)
	assert.NoError(t, err)
}

func TestSessionRepository_GetByToken(t *testing.T) {
	tx := testutil.GetTestTx(t)
	user := createTestUser(t, tx)

	sessionRepo := NewSessionRepository(tx)

	now := time.Now()
	expectedSession := &model.Session{
		UserID:    user.ID,
		Token:     "test-token",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err := sessionRepo.Create(expectedSession)
	assert.NoError(t, err)

	session, err := sessionRepo.GetByToken("test-token")
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.UserID, session.UserID)
	assert.Equal(t, expectedSession.Token, session.Token)
}

func TestSessionRepository_Delete(t *testing.T) {
	tx := testutil.GetTestTx(t)
	user := createTestUser(t, tx)

	sessionRepo := NewSessionRepository(tx)

	session := &model.Session{
		UserID:    user.ID,
		Token:     "token",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := sessionRepo.Create(session)
	assert.NoError(t, err)

	err = sessionRepo.Delete("token")
	assert.NoError(t, err)

	_, err = sessionRepo.GetByToken("token")
	assert.Error(t, err)
}

func TestSessionRepository_Errors(t *testing.T) {
	tx := testutil.GetTestTx(t)
	sessionRepo := NewSessionRepository(tx)

	t.Run("GetByToken Error - Non-existent Token", func(t *testing.T) {
		session, err := sessionRepo.GetByToken("non-existent-token")
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Delete Error - Non-existent Token", func(t *testing.T) {
		err := sessionRepo.Delete("non-existent-token")
		assert.Error(t, err)
	})
}
