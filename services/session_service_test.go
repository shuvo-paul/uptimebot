package services

import (
	"testing"
	"time"

	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/stretchr/testify/assert"
)

type mockSessionRepository struct {
	createFunc     func(session *models.Session) error
	getByTokenFunc func(token string) (*models.Session, error)
	deleteFunc     func(token string) error
}

func (m *mockSessionRepository) Create(session *models.Session) error {
	return m.createFunc(session)
}

func (m *mockSessionRepository) GetByToken(token string) (*models.Session, error) {
	return m.getByTokenFunc(token)
}

func (m *mockSessionRepository) Delete(token string) error {
	return m.deleteFunc(token)
}

func TestCreateSession(t *testing.T) {
	mockRepo := &mockSessionRepository{
		createFunc: func(session *models.Session) error {
			return nil
		},
	}
	service := NewSessionService(mockRepo)

	// Test
	session, plainToken, err := service.CreateSession(1)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, plainToken)
	assert.NotNil(t, session)
	assert.Equal(t, 1, session.UserID)
	assert.NotEmpty(t, session.Token)
	assert.False(t, session.CreatedAt.IsZero())
	assert.False(t, session.ExpiresAt.IsZero())
}

func TestValidateSession(t *testing.T) {

	t.Run("Valid session", func(t *testing.T) {
		validSession := &models.Session{
			UserID:    1,
			Token:     "hashed_token",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockRepo := &mockSessionRepository{
			getByTokenFunc: func(token string) (*models.Session, error) {
				return validSession, nil
			},
		}
		service := NewSessionService(mockRepo)

		session, err := service.ValidateSession("token")
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, validSession, session)
	})

	t.Run("Expired session", func(t *testing.T) {
		expiredSession := &models.Session{
			UserID:    1,
			Token:     "test_token",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			ExpiresAt: time.Now().Add(-24 * time.Hour),
		}

		var sessionToken string

		mockRepo := &mockSessionRepository{
			getByTokenFunc: func(token string) (*models.Session, error) {
				return expiredSession, nil
			},
			deleteFunc: func(token string) error {
				sessionToken = token
				return nil
			},
		}

		service := NewSessionService(mockRepo)

		session, err := service.ValidateSession("expired_token")
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "session has expired")
		assert.Equal(t, "test_token", sessionToken)
	})
}

func TestDeleteSession(t *testing.T) {
	var capturedToken string
	mockRepo := &mockSessionRepository{
		deleteFunc: func(token string) error {
			capturedToken = token
			return nil
		},
	}
	service := NewSessionService(mockRepo)
	testToken := "test_token"
	err := service.DeleteSession(testToken)
	assert.NoError(t, err)
	assert.Equal(t, testToken, capturedToken)
}
