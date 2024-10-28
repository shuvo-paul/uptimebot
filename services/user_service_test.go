package services

import (
	"testing"

	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) SaveUser(user *models.User) (*models.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo) // Store the service instance

	tests := []struct {
		name          string
		user          *models.User
		setupMock     func()
		expectedError bool
	}{
		{
			name: "Success",
			user: &models.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func() {
				mockRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockRepo.On("SaveUser", mock.Anything).Return(&models.User{
					Username: "testuser",
					Email:    "test@example.com",
					Password: "password123",
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}
			_, err := userService.CreateUser(tt.user)
			if (err != nil) != tt.expectedError {
				t.Errorf("CreateUser() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
