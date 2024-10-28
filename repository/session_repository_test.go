package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/shuvo-paul/sitemonitor/utils"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepository_Create(t *testing.T) {
	db, mock := utils.SetupTestDB(t)
	defer db.Close()

	sessionRepo := NewSessionRepository(db)
	session := &models.Session{
		UserID:    1,
		Token:     "test-token",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(session.UserID, session.Token, session.CreatedAt, session.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := sessionRepo.Create(session)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_GetByToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sessionRepo := NewSessionRepository(db)
	now := time.Now()
	expectedSession := &models.Session{
		ID:        1,
		UserID:    1,
		Token:     "test-token",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "token", "created_at", "expires_at"}).
		AddRow(expectedSession.ID, expectedSession.UserID, expectedSession.Token,
			expectedSession.CreatedAt, expectedSession.ExpiresAt)

	mock.ExpectQuery("SELECT (.+) FROM sessions").
		WithArgs("test-token").
		WillReturnRows(rows)

	session, err := sessionRepo.GetByToken("test-token")
	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sessionRepo := NewSessionRepository(db)

	mock.ExpectExec("DELETE FROM sessions").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = sessionRepo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sessionRepo := NewSessionRepository(db)

	t.Run("Create Error", func(t *testing.T) {
		session := &models.Session{
			UserID:    1,
			Token:     "test-token",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		mock.ExpectExec("INSERT INTO sessions").
			WithArgs(session.UserID, session.Token, session.CreatedAt, session.ExpiresAt).
			WillReturnError(sqlmock.ErrCancelled)

		err = sessionRepo.Create(session)
		assert.Error(t, err)
	})

	t.Run("GetByToken Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("non-existent-token").
			WillReturnError(sqlmock.ErrCancelled)

		session, err := sessionRepo.GetByToken("non-existent-token")
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Delete Error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM sessions").
			WithArgs(999).
			WillReturnError(sqlmock.ErrCancelled)

		err = sessionRepo.Delete(999)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
