package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/repository"
)

type SessionServiceInterface interface {
	CreateSession(userID int) (*model.Session, string, error)
	ValidateSession(token string) (*model.Session, error)
	DeleteSession(sessionID string) error
}

var _ SessionServiceInterface = (*SessionService)(nil)

type SessionService struct {
	sessionRepo repository.SessionRepositoryInterface
}

func NewSessionService(sessionRepo repository.SessionRepositoryInterface) *SessionService {
	return &SessionService{sessionRepo: sessionRepo}
}

func (s *SessionService) CreateSession(userID int) (*model.Session, string, error) {
	// Generate a unique token
	plainToken := uuid.New().String()

	session := &model.Session{
		UserID:    userID,
		Token:     plainToken,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, "", err
	}

	return session, plainToken, nil
}

func (s *SessionService) ValidateSession(token string) (*model.Session, error) {

	session, err := s.sessionRepo.GetByToken(token)
	if err != nil {
		return nil, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.sessionRepo.Delete(session.Token)
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}

func (s *SessionService) DeleteSession(token string) error {
	return s.sessionRepo.Delete(token)
}
