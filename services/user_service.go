package services

import (
	"fmt"

	"github.com/shuvo-paul/sitemonitor/models"
	"github.com/shuvo-paul/sitemonitor/repository"
)

type UserServiceInterface interface {
	CreateUser(*models.User) (*models.User, error)
	Authenticate(string, string) (*models.User, error)
	GetUserByID(int) (*models.User, error)
}

var _ UserServiceInterface = (*UserService)(nil)

type UserService struct {
	repo repository.UserRepositoryInterface
}

func NewUserService(repo repository.UserRepositoryInterface) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
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

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.VerifyPassword(password) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	return s.repo.GetUserByID(id)
}
