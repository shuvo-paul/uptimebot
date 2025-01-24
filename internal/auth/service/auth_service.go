package service

import (
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/repository"
)

type AuthServiceInterface interface {
	CreateUser(*model.User) (*model.User, error)
	Authenticate(string, string) (*model.User, error)
	GetUserByID(int) (*model.User, error)
}

var _ AuthServiceInterface = (*AuthService)(nil)

type AuthService struct {
	repo repository.UserRepositoryInterface
}

func NewAuthService(repo repository.UserRepositoryInterface) *AuthService {
	return &AuthService{repo: repo}
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

	return s.repo.SaveUser(user)
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
