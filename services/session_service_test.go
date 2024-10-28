package services

import (
	"testing"
	"time"

	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(session *models.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByToken(token string) (*models.Session, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) Delete(sessionID int) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func TestCreateSession(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	service := NewSessionService(mockRepo)

	// Setup mock expectation
	mockRepo.On("Create", mock.AnythingOfType("*models.Session")).Return(nil)

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
	mockRepo.AssertExpectations(t)
}

func TestValidateSession(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	service := NewSessionService(mockRepo)

	t.Run("Valid session", func(t *testing.T) {
		validSession := &models.Session{
			ID:        1,
			UserID:    1,
			Token:     "hashed_token",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}
		mockRepo.On("GetByToken", "token").Return(validSession, nil)

		session, err := service.ValidateSession("token")
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, validSession, session)
	})

	t.Run("Expired session", func(t *testing.T) {
		expiredSession := &models.Session{
			ID:        2,
			UserID:    1,
			Token:     "hashed_token",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			ExpiresAt: time.Now().Add(-24 * time.Hour),
		}
		mockRepo.On("GetByToken", "expired_token").Return(expiredSession, nil)
		mockRepo.On("Delete", 2).Return(nil)

		session, err := service.ValidateSession("expired_token")
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "session has expired")
	})
}

func TestDeleteSession(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	service := NewSessionService(mockRepo)

	mockRepo.On("Delete", 1).Return(nil)

	err := service.DeleteSession(1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
