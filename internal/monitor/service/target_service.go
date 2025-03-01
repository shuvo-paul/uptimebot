// Package service provides target monitoring functionality for the uptime monitoring system.
// It handles target creation, updates, deletion, and status monitoring.
package service

import (
	"errors"
	"fmt"
	"time"

	monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
	"github.com/shuvo-paul/uptimebot/internal/monitor/repository"
	notifCore "github.com/shuvo-paul/uptimebot/internal/notification/core"
	alertService "github.com/shuvo-paul/uptimebot/internal/notification/service"
)

// Common errors returned by the target service.
var (
	// ErrUnauthorized is returned when a user attempts to access a target they don't own.
	ErrUnauthorized = errors.New("unauthorized access to target")
	// ErrTargetNotFound is returned when the requested target does not exist.
	ErrTargetNotFound = errors.New("target not found")
	// ErrInvalidInput is returned when the provided input parameters are invalid.
	ErrInvalidInput = errors.New("invalid input parameters")
	// ErrTargetLimitReached is returned when a user attempts to create more than the allowed number of targets.
	ErrTargetLimitReached = errors.New("maximum number of targets (5) reached")
)

// TargetServiceInterface defines the contract for managing monitoring targets.
// It provides methods for CRUD operations and monitoring initialization.
type TargetServiceInterface interface {
	// Create adds a new monitoring target for a user.
	// Returns the created target or an error if the operation fails.
	// Possible errors: ErrInvalidInput if parameters are invalid.
	Create(userID int, url string, interval time.Duration) (model.UserTarget, error)

	// GetByID retrieves a target by its ID and verifies user ownership.
	// Returns the target or an error if not found or unauthorized.
	// Possible errors: ErrTargetNotFound, ErrUnauthorized, ErrInvalidInput.
	GetByID(id int, userID int) (model.UserTarget, error)

	// GetAll retrieves all monitoring targets in the system.
	// Returns a slice of targets or an error if the operation fails.
	GetAll() ([]model.UserTarget, error)

	// GetAllByUserID retrieves all monitoring targets owned by a specific user.
	// Returns a slice of targets or an error if the operation fails.
	GetAllByUserID(userID int) ([]model.UserTarget, error)

	// Update modifies an existing target's properties after verifying ownership.
	// Returns the updated target or an error if the operation fails.
	// Possible errors: ErrTargetNotFound, ErrUnauthorized, ErrInvalidInput.
	Update(target model.UserTarget, userID int) (model.UserTarget, error)

	// Delete removes a target after verifying ownership.
	// Returns an error if the operation fails or user is not authorized.
	// Possible errors: ErrTargetNotFound, ErrUnauthorized, ErrInvalidInput.
	Delete(id int, userID int) error

	// InitializeMonitoring starts monitoring for all existing targets.
	// Returns an error if initialization fails.
	InitializeMonitoring() error

	// ToggleEnabled toggles the monitoring state for a target after verifying ownership.
	// Returns the updated target or an error if the operation fails.
	// Possible errors: ErrTargetNotFound, ErrUnauthorized, ErrInvalidInput.
	ToggleEnabled(id int, userID int) (model.UserTarget, error)
}

var _ TargetServiceInterface = (*TargetService)(nil)

// TargetService implements the TargetServiceInterface and manages monitoring targets.
type TargetService struct {
	// repo provides persistence operations for targets
	repo repository.TargetRepositoryInterface
	// manager handles the monitoring of targets
	manager *monitor.Manager
	// notifierService handles notifications when target status changes
	notifierService alertService.NotifierServiceInterface
}

// NewTargetService creates a new instance of TargetService with the provided dependencies.
// It initializes a new monitor manager and returns the service instance.
func NewTargetService(repo repository.TargetRepositoryInterface, notifierService alertService.NotifierServiceInterface) *TargetService {
	s := &TargetService{
		repo:            repo,
		notifierService: notifierService,
	}
	s.initializeManager()
	return s
}

// initializeManager initializes the monitor manager for the service.
func (s *TargetService) initializeManager() {
	s.manager = monitor.NewManager()
}

// validateTarget validates the target's basic properties.
func (s *TargetService) validateTarget(userID int, url string, interval time.Duration) error {
	if userID <= 0 {
		return fmt.Errorf("%w: invalid userID", ErrInvalidInput)
	}
	if url == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidInput)
	}
	if interval <= 0 {
		return fmt.Errorf("%w: interval must be positive", ErrInvalidInput)
	}
	return nil
}

// handleStatusUpdate processes status changes for a target.
// It updates the target's status in the repository and notifies observers of the change.
// Returns an error if the status update fails or if notification configuration fails.
func (s *TargetService) handleStatusUpdate(target *monitor.Target, status string) error {
	if target == nil || status == "" {
		return fmt.Errorf("%w: target or status is nil", ErrInvalidInput)
	}

	if err := s.repo.UpdateStatus(target, status); err != nil {
		if errors.Is(err, repository.ErrTargetNotFound) {
			return fmt.Errorf("%w: target %s not found", ErrTargetNotFound, target.URL)
		}
		return fmt.Errorf("failed to update target status: %w", err)
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

func (s *TargetService) Create(userID int, url string, interval time.Duration) (model.UserTarget, error) {
	if err := s.validateTarget(userID, url, interval); err != nil {
		return model.UserTarget{}, err
	}

	// Check if user has reached the target limit
	existingTargets, err := s.GetAllByUserID(userID)
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to check target limit: %w", err)
	}
	if len(existingTargets) >= 5 {
		return model.UserTarget{}, ErrTargetLimitReached
	}

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
		if errors.Is(err, repository.ErrTargetNotFound) {
			return model.UserTarget{}, fmt.Errorf("%w: %v", ErrTargetNotFound, err)
		}
		return model.UserTarget{}, fmt.Errorf("failed to create target: %w", err)
	}

	// Create a new targets monitor
	if err := s.manager.RegisterTarget(newUserTarget.Target); err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to register target monitor: %w", err)
	}

	return newUserTarget, nil
}

func (s *TargetService) GetByID(id int, userID int) (model.UserTarget, error) {
	if id <= 0 || userID <= 0 {
		return model.UserTarget{}, fmt.Errorf("%w: invalid id or userID", ErrInvalidInput)
	}

	userTarget, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrTargetNotFound) {
			return model.UserTarget{}, fmt.Errorf("%w: target with id %d not found", ErrTargetNotFound, id)
		}
		return model.UserTarget{}, fmt.Errorf("failed to fetch target: %w", err)
	}

	if userTarget.UserID != userID {
		return model.UserTarget{}, ErrUnauthorized
	}

	return userTarget, nil
}

func (s *TargetService) GetAll() ([]model.UserTarget, error) {
	userTargets, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch targets: %w", err)
	}
	return userTargets, nil
}

func (s *TargetService) GetAllByUserID(userID int) ([]model.UserTarget, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid userID", ErrInvalidInput)
	}
	return s.repo.GetAllByUserID(userID)
}

func (s *TargetService) Update(userTarget model.UserTarget, userID int) (model.UserTarget, error) {
	if err := s.validateTarget(userID, userTarget.URL, userTarget.Interval); err != nil {
		return model.UserTarget{}, err
	}

	if userTarget.UserID != userID {
		return model.UserTarget{}, fmt.Errorf("%w: user %d does not own target %d", ErrUnauthorized, userID, userTarget.ID)
	}

	userTarget.OnStatusUpdate = s.handleStatusUpdate

	// First update the target in the database
	updatedUserTarget, err := s.repo.Update(userTarget)
	if err != nil {
		if errors.Is(err, repository.ErrTargetNotFound) {
			return model.UserTarget{}, fmt.Errorf("%w: target with id %d not found", ErrTargetNotFound, userTarget.ID)
		}
		return model.UserTarget{}, fmt.Errorf("failed to update target: %w", err)
	}

	if existingTarget, exists := s.manager.Targets[userTarget.ID]; exists {
		existingTarget.Update(updatedUserTarget.Target)
	} else {
		// Register new monitor if it doesn't exist
		if err := s.manager.RegisterTarget(updatedUserTarget.Target); err != nil {
			return model.UserTarget{}, fmt.Errorf("failed to register target monitor: %w", err)
		}
	}

	return updatedUserTarget, nil
}

func (s *TargetService) Delete(id int, userID int) error {
	userTarget, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	s.manager.RevokeTarget(userTarget.ID)
	return s.repo.Delete(userTarget.ID)
}

func (s *TargetService) ToggleEnabled(id int, userID int) (model.UserTarget, error) {
	userTarget, err := s.GetByID(id, userID)
	if err != nil {
		return model.UserTarget{}, err
	}

	userTarget.Enabled = !userTarget.Enabled
	userTarget.OnStatusUpdate = s.handleStatusUpdate

	// Update the target in the database
	updatedUserTarget, err := s.repo.Update(userTarget)
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to update target: %w", err)
	}

	// Update the target in the monitor manager
	if existingTarget, exists := s.manager.Targets[userTarget.ID]; exists {
		existingTarget.Update(updatedUserTarget.Target)
	} else if updatedUserTarget.Enabled {
		// Register new monitor if it doesn't exist and is being enabled
		if err := s.manager.RegisterTarget(updatedUserTarget.Target); err != nil {
			return model.UserTarget{}, fmt.Errorf("failed to register target monitor: %w", err)
		}
	}

	return updatedUserTarget, nil
}

func (s *TargetService) InitializeMonitoring() error {
	userTargets, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to fetch targets: %w", err)
	}

	for _, target := range userTargets {
		target.OnStatusUpdate = s.handleStatusUpdate

		if err := s.manager.RegisterTarget(target.Target); err != nil {
			return fmt.Errorf("failed to register target %s: %w", target.URL, err)
		}
	}

	return nil
}
