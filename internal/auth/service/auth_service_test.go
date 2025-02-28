package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/stretchr/testify/assert"
)

// MockUserRepository is a mock implementation of UserRepository
type mockUserRepository struct {
	saveUserFunc       func(user *model.User) (*model.User, error)
	emailExistsFunc    func(email string) (bool, error)
	getUserByEmailFunc func(email string) (*model.User, error)
	getUserByIdFunc    func(id int) (*model.User, error)
	updateUserFunc     func(user *model.User) (*model.User, error)
	updatePasswordFunc func(userID int, hashedPassword string) error
}

func (m *mockUserRepository) SaveUser(user *model.User) (*model.User, error) {
	return m.saveUserFunc(user)
}

func (m *mockUserRepository) EmailExists(email string) (bool, error) {
	return m.emailExistsFunc(email)
}

func (m *mockUserRepository) GetUserByEmail(email string) (*model.User, error) {
	return m.getUserByEmailFunc(email)
}

func (m *mockUserRepository) GetUserByID(id int) (*model.User, error) {
	return m.getUserByIdFunc(id)
}

func (m *mockUserRepository) UpdateUser(user *model.User) (*model.User, error) {
	return m.updateUserFunc(user)
}

func (m *mockUserRepository) UpdatePassword(userID int, hashedPassword string) error {
	return m.updatePasswordFunc(userID, hashedPassword)
}

// MockTokenService is a mock implementation of TokenServiceInterface
type mockTokenService struct {
	validateTokenFunc   func(token string, tokenType model.TokenType) (*model.Token, error)
	sendTokenFunc       func(userID int, email string, tokenType model.TokenType, subject string, path string, expiresIn time.Duration) error
	markTokenAsUsedFunc func(tokenID int) error
}

func (m *mockTokenService) createToken(userID int, tokenType model.TokenType, expiresIn time.Duration) (*model.Token, error) {
	return nil, nil
}

func (m *mockTokenService) ValidateToken(token string, tokenType model.TokenType) (*model.Token, error) {
	return m.validateTokenFunc(token, tokenType)
}

func (m *mockTokenService) invalidateAndCreateNewToken(userID int, tokenType model.TokenType, expiresIn time.Duration) (*model.Token, error) {
	return nil, nil
}

func (m *mockTokenService) SendToken(userID int, email string, tokenType model.TokenType, subject string, path string, expiresIn time.Duration) error {
	if m.sendTokenFunc != nil {
		return m.sendTokenFunc(userID, email, tokenType, subject, path, expiresIn)
	}
	return nil
}

func (m *mockTokenService) MarkTokenAsUsed(tokenID int) error {
	if m.markTokenAsUsedFunc != nil {
		return m.markTokenAsUsedFunc(tokenID)
	}
	return nil
}

func TestCreateUser(t *testing.T) {
	mockRepo := &mockUserRepository{
		saveUserFunc: func(user *model.User) (*model.User, error) {
			user.ID = 1
			return user, nil
		},
	}
	mockTokenService := &mockTokenService{
		sendTokenFunc: func(userID int, email string, tokenType model.TokenType, subject string, path string, expiresIn time.Duration) error {
			return nil
		},
	}
	userService := NewAuthService(mockRepo, mockTokenService)
	t.Run("User created successfully", func(t *testing.T) {
		mockRepo.emailExistsFunc = func(email string) (bool, error) {
			return false, nil
		}

		user := &model.User{
			Name:     "testuser",
			Email:    "test@example.com",
			Password: "password123@",
		}
		savedUser, err := userService.CreateUser(user)
		assert.NoError(t, err)
		assert.Equal(t, savedUser.ID, 1)
	})

	t.Run("Email Exist", func(t *testing.T) {
		mockRepo.emailExistsFunc = func(email string) (bool, error) {
			return true, fmt.Errorf("email already exists")
		}
		user := &model.User{
			Name:     "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		_, err := userService.CreateUser(user)
		assert.Error(t, err)
	})

}

func TestVerifyEmail(t *testing.T) {
	mockRepo := &mockUserRepository{
		getUserByIdFunc: func(id int) (*model.User, error) {
			return &model.User{ID: 1, Email: "test@example.com"}, nil
		},
		saveUserFunc: func(user *model.User) (*model.User, error) {
			return user, nil
		},
		updateUserFunc: func(user *model.User) (*model.User, error) {
			return user, nil
		},
	}

	mockTokenService := &mockTokenService{
		validateTokenFunc: func(token string, tokenType model.TokenType) (*model.Token, error) {
			if token == "valid_token" {
				return &model.Token{UserID: 1, Token: token}, nil
			}
			return nil, fmt.Errorf("invalid token")
		},
	}

	userService := NewAuthService(mockRepo, mockTokenService)

	t.Run("Email verified successfully", func(t *testing.T) {
		err := userService.VerifyEmail("valid_token")
		assert.NoError(t, err)
	})

	t.Run("Invalid token", func(t *testing.T) {
		err := userService.VerifyEmail("invalid_token")
		assert.Error(t, err)
	})
}

func TestAuthentication(t *testing.T) {
	email := "test@example.com"
	password := "password123"
	wrongPassword := "wrongpassword123"

	user := &model.User{
		Name:     "testuser",
		Email:    email,
		Password: password,
		Verified: true,
	}

	user.HashPassword()

	mockRepo := &mockUserRepository{
		getUserByEmailFunc: func(email string) (*model.User, error) {
			return user, nil
		},
	}

	mockTokenService := &mockTokenService{}
	userService := NewAuthService(mockRepo, mockTokenService)

	t.Run("Logged in succesfully", func(t *testing.T) {
		user, err := userService.Authenticate(email, password)
		assert.NoError(t, err)
		assert.Equal(t, user.Email, email)
	})

	t.Run("Login failed", func(t *testing.T) {
		_, err := userService.Authenticate(email, wrongPassword)
		assert.Error(t, err)
	})
}

func TestValidateToken(t *testing.T) {
	mockRepo := &mockUserRepository{}

	t.Run("Valid token", func(t *testing.T) {
		mockTokenService := &mockTokenService{
			validateTokenFunc: func(token string, tokenType model.TokenType) (*model.Token, error) {
				return &model.Token{ID: 1, UserID: 1, Token: token}, nil
			},
		}
		userService := NewAuthService(mockRepo, mockTokenService)

		token, err := userService.ValidateToken("valid_token", model.TokenTypeEmailVerification)
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "valid_token", token.Token)
	})

	t.Run("Invalid token", func(t *testing.T) {
		mockTokenService := &mockTokenService{
			validateTokenFunc: func(token string, tokenType model.TokenType) (*model.Token, error) {
				return nil, fmt.Errorf("invalid token")
			},
		}
		userService := NewAuthService(mockRepo, mockTokenService)

		token, err := userService.ValidateToken("invalid_token", model.TokenTypeEmailVerification)
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("Expired token", func(t *testing.T) {
		mockTokenService := &mockTokenService{
			validateTokenFunc: func(token string, tokenType model.TokenType) (*model.Token, error) {
				return nil, fmt.Errorf("token has expired")
			},
		}
		userService := NewAuthService(mockRepo, mockTokenService)

		token, err := userService.ValidateToken("expired_token", model.TokenTypeEmailVerification)
		assert.Error(t, err)
		assert.Nil(t, token)
	})
}
