package service

import (
	"fmt"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/repository"
)

type AuthServiceInterface interface {
	CreateUser(*model.User) (*model.User, error)
	Authenticate(string, string) (*model.User, error)
	GetUserByID(int) (*model.User, error)
	VerifyEmail(token string) error
	SendToken(userID int, email string) error
}

var _ AuthServiceInterface = (*AuthService)(nil)

type AuthService struct {
	repo         repository.UserRepositoryInterface
	tokenService TokenServiceInterface
}

func NewAuthService(
	repo repository.UserRepositoryInterface,
	tokenService TokenServiceInterface,
) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: tokenService,
	}
}

func (s *AuthService) CreateUser(user *model.User) (*model.User, error) {
	if exists, err := s.repo.EmailExists(user.Email); err != nil {
		return nil, fmt.Errorf("error checking email existence: %w", err)
	} else if exists {
		return nil, fmt.Errorf("email already exists")
	}

	if err := user.ValidatePassword(); err != nil {
		return nil, err
	}

	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	user.Verified = false

	createdUser, err := s.repo.SaveUser(user)
	if err != nil {
		return nil, fmt.Errorf("error saving user: %w", err)
	}

	if err := s.tokenService.SendToken(createdUser.ID, createdUser.Email, model.TokenTypeEmailVerification, "Verify Your Email Address", "verify-email", 24*time.Hour); err != nil {
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return createdUser, nil
}

func (s *AuthService) Authenticate(email, password string) (*model.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.VerifyPassword(password) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (s *AuthService) GetUserByID(id int) (*model.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *AuthService) VerifyEmail(token string) error {
	accountToken, err := s.tokenService.ValidateToken(token, model.TokenTypeEmailVerification)
	if err != nil {
		return fmt.Errorf("invalid verification token: %w", err)
	}

	user, err := s.repo.GetUserByID(accountToken.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	user.Verified = true
	_, err = s.repo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}

	return nil
}

func (s *AuthService) SendToken(userID int, email string) error {
	return s.tokenService.SendToken(userID, email, model.TokenTypeEmailVerification, "Verify Your Email Address", "verify-email", 24*time.Hour)
}
