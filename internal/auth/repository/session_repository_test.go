package repository

import (
	"testing"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepository_Create(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	sessionRepo := NewSessionRepository(db)
	session := &model.Session{
		UserID:    1,
		Token:     "test-token",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := sessionRepo.Create(session)
	assert.NoError(t, err)
}

func TestSessionRepository_GetByToken(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	sessionRepo := NewSessionRepository(db)

	// First create a session
	now := time.Now()
	expectedSession := &model.Session{
		UserID:    1,
		Token:     "test-token",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err := sessionRepo.Create(expectedSession)
	assert.NoError(t, err)

	// Then try to get it
	session, err := sessionRepo.GetByToken("test-token")
	assert.NoError(t, err)
	assert.Equal(t, expectedSession.UserID, session.UserID)
	assert.Equal(t, expectedSession.Token, session.Token)
}

func TestSessionRepository_Delete(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	sessionRepo := NewSessionRepository(db)

	// First create a session
	session := &model.Session{
		UserID:    1,
		Token:     "token",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := sessionRepo.Create(session)
	assert.NoError(t, err)

	// Then delete it
	err = sessionRepo.Delete("token")
	assert.NoError(t, err)

	// Verify it's deleted
	_, err = sessionRepo.GetByToken("token")
	assert.Error(t, err)
}

func TestSessionRepository_Errors(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	sessionRepo := NewSessionRepository(db)

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
