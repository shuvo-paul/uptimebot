package service

import (
	"fmt"
	"time"

	monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
	"github.com/shuvo-paul/uptimebot/internal/monitor/repository"
	notifCore "github.com/shuvo-paul/uptimebot/internal/notification/core"
	alertService "github.com/shuvo-paul/uptimebot/internal/notification/service"
)

type TargetServiceInterface interface {
	Create(userID int, url string, interval time.Duration) (*monitor.Target, error)
	GetByID(id int) (*monitor.Target, error)
	GetAll() ([]*monitor.Target, error)
	GetAllByUserID(userID int) ([]*monitor.Target, error)
	Update(*monitor.Target) (*monitor.Target, error)
	Delete(id int) error
	InitializeMonitoring() error
}

var _ TargetServiceInterface = (*TargetService)(nil)

type TargetService struct {
	repo            repository.TargetRepositoryInterface
	manager         *monitor.Manager
	notifierService alertService.NotifierServiceInterface
}

func NewTargetService(repo repository.TargetRepositoryInterface, notifierService alertService.NotifierServiceInterface) *TargetService {
	return &TargetService{
		repo:            repo,
		manager:         monitor.NewManager(),
		notifierService: notifierService,
	}
}

func (s *TargetService) handleStatusUpdate(target *monitor.Target, status string) error {
	if err := s.repo.UpdateStatus(target, status); err != nil {
		return err
	}

	if err := s.notifierService.ConfigureObservers(target.ID); err != nil {
		return fmt.Errorf("failed to configure observers: %w", err)
	}

	state := notifCore.State{
		Name:      target.URL,
		Status:    status,
		UpdatedAt: time.Now(),
		Message:   fmt.Sprintf("Target %s is %s", target.URL, status),
	}

	s.notifierService.GetSubject().Notify(state)

	return nil
}

func (s *TargetService) Create(userID int, url string, interval time.Duration) (*monitor.Target, error) {
	userTarget := model.UserTarget{
		UserID: userID,
		Target: &monitor.Target{
			URL:      url,
			Interval: interval,
			Enabled:  true,
			Status:   "pending",
		},
	}

	userTarget.Target.OnStatusUpdate = s.handleStatusUpdate

	newUserTarget, err := s.repo.Create(userTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to create target: %w", err)
	}

	// Create a new targets monitor
	if err := s.manager.RegisterTarget(newUserTarget.Target); err != nil {
		return nil, fmt.Errorf("failed to register target monitor: %w", err)
	}

	return newUserTarget.Target, nil
}

func (s *TargetService) GetByID(id int) (*monitor.Target, error) {
	return s.repo.GetByID(id)
}

func (s *TargetService) GetAll() ([]*monitor.Target, error) {
	return s.repo.GetAll()
}

func (s *TargetService) GetAllByUserID(userID int) ([]*monitor.Target, error) {
	return s.repo.GetAllByUserID(userID)
}

func (s *TargetService) Update(target *monitor.Target) (*monitor.Target, error) {
	target.OnStatusUpdate = s.handleStatusUpdate

	// First update the target in the database
	updatedTarget, err := s.repo.Update(target)
	if err != nil {
		return nil, fmt.Errorf("failed to update target: %w", err)
	}

	if existingTarget, exists := s.manager.Targets[target.ID]; exists {
		existingTarget.Update(updatedTarget)
	} else {
		// Register new monitor if it doesn't exist
		if err := s.manager.RegisterTarget(updatedTarget); err != nil {
			return nil, fmt.Errorf("failed to register target monitor: %w", err)
		}
	}

	return updatedTarget, nil
}

func (s *TargetService) Delete(id int) error {
	s.manager.RevokeTarget(id)
	return s.repo.Delete(id)
}

func (s *TargetService) InitializeMonitoring() error {
	targets, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to fetch targets: %w", err)
	}

	for _, target := range targets {
		target.OnStatusUpdate = s.handleStatusUpdate

		if err := s.manager.RegisterTarget(target); err != nil {
			return fmt.Errorf("failed to register target %s: %w", target.URL, err)
		}
	}

	return nil
}
