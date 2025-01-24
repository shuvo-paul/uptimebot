package service

import (
	"fmt"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/notification/provider"
	alertService "github.com/shuvo-paul/uptimebot/internal/notification/service"
	"github.com/shuvo-paul/uptimebot/internal/uptime/model"
	"github.com/shuvo-paul/uptimebot/internal/uptime/monitor"
	"github.com/shuvo-paul/uptimebot/internal/uptime/repository"
)

type TargetServiceInterface interface {
	Create(userID int, url string, interval time.Duration) (*monitor.Target, error)
	GetByID(id int) (*monitor.Target, error)
	GetAll() ([]*monitor.Target, error)
	GetAllByUserID(userID int) ([]*monitor.Target, error)
	Update(site *monitor.Target) (*monitor.Target, error)
	Delete(id int) error
	InitializeMonitoring() error
}

var _ TargetServiceInterface = (*TargetService)(nil)

type TargetService struct {
	repo            repository.SiteRepositoryInterface
	manager         *monitor.Manager
	notifierService alertService.NotifierServiceInterface
}

func NewTargetService(repo repository.SiteRepositoryInterface, notifierService alertService.NotifierServiceInterface) *TargetService {
	return &TargetService{
		repo:            repo,
		manager:         monitor.NewManager(),
		notifierService: notifierService,
	}
}

func (s *TargetService) handleStatusUpdate(site *monitor.Target, status string) error {
	if err := s.repo.UpdateStatus(site, status); err != nil {
		return err
	}

	if err := s.notifierService.ConfigureObservers(site.ID); err != nil {
		return fmt.Errorf("failed to configure observers: %w", err)
	}

	state := provider.State{
		Name:      site.URL,
		Status:    status,
		UpdatedAt: time.Now(),
		Message:   fmt.Sprintf("Target %s is %s", site.URL, status),
	}

	s.notifierService.GetSubject().Notify(state)

	return nil
}

func (s *TargetService) Create(userID int, url string, interval time.Duration) (*monitor.Target, error) {
	userSite := model.UserTarget{
		UserID: userID,
		Target: &monitor.Target{
			URL:      url,
			Interval: interval,
			Enabled:  true,
			Status:   "pending",
		},
	}

	userSite.OnStatusUpdate = s.handleStatusUpdate

	newUserSite, err := s.repo.Create(userSite)
	if err != nil {
		return nil, fmt.Errorf("failed to create site: %w", err)
	}

	// Create a new site monitor
	if err := s.manager.RegisterSite(newUserSite.Target); err != nil {
		return nil, fmt.Errorf("failed to register site monitor: %w", err)
	}

	return newUserSite.Target, nil
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

func (s *TargetService) Update(site *monitor.Target) (*monitor.Target, error) {
	site.OnStatusUpdate = s.handleStatusUpdate

	// First update the site in the database
	updatedSite, err := s.repo.Update(site)
	if err != nil {
		return nil, fmt.Errorf("failed to update site: %w", err)
	}

	if existingSite, exists := s.manager.Targets[site.ID]; exists {
		existingSite.Update(updatedSite)
	} else {
		// Register new monitor if it doesn't exist
		if err := s.manager.RegisterSite(updatedSite); err != nil {
			return nil, fmt.Errorf("failed to register site monitor: %w", err)
		}
	}

	return updatedSite, nil
}

func (s *TargetService) Delete(id int) error {
	s.manager.RevokeSite(id)
	return s.repo.Delete(id)
}

func (s *TargetService) InitializeMonitoring() error {
	sites, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to fetch sites: %w", err)
	}

	for _, site := range sites {
		site.OnStatusUpdate = s.handleStatusUpdate

		if err := s.manager.RegisterSite(site); err != nil {
			return fmt.Errorf("failed to register site %s: %w", site.URL, err)
		}
	}

	return nil
}
