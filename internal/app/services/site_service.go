package services

import (
	"fmt"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/internal/app/repository"
	"github.com/shuvo-paul/sitemonitor/pkg/monitor"
)

type SiteServiceInterface interface {
	Create(userID int, url string, interval time.Duration) (*monitor.Site, error)
	GetByID(id int) (*monitor.Site, error)
	GetAll() ([]*monitor.Site, error)
	Update(site *monitor.Site) (*monitor.Site, error)
	Delete(id int) error
	InitializeMonitoring() error
}

var _ SiteServiceInterface = (*SiteService)(nil)

type SiteService struct {
	repo    repository.SiteRepositoryInterface
	manager *monitor.Manager
}

func NewSiteService(repo repository.SiteRepositoryInterface) *SiteService {
	return &SiteService{
		repo:    repo,
		manager: monitor.NewManager(),
	}
}

func (s *SiteService) handleStatusUpdate(site *monitor.Site, status string) error {
	return s.repo.UpdateStatus(site, status)
}

func (s *SiteService) Create(userID int, url string, interval time.Duration) (*monitor.Site, error) {
	userSite := models.UserSite{
		UserID: userID,
		Site: &monitor.Site{
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
	if err := s.manager.RegisterSite(newUserSite.Site); err != nil {
		return nil, fmt.Errorf("failed to register site monitor: %w", err)
	}

	return newUserSite.Site, nil
}

func (s *SiteService) GetByID(id int) (*monitor.Site, error) {
	return s.repo.GetByID(id)
}

func (s *SiteService) GetAll() ([]*monitor.Site, error) {
	return s.repo.GetAll()
}

func (s *SiteService) Update(site *monitor.Site) (*monitor.Site, error) {
	site.OnStatusUpdate = s.handleStatusUpdate

	// First update the site in the database
	updatedSite, err := s.repo.Update(site)
	if err != nil {
		return nil, fmt.Errorf("failed to update site: %w", err)
	}

	if existingSite, exists := s.manager.Sites[site.ID]; exists {
		existingSite.Update(updatedSite)
	} else {
		// Register new monitor if it doesn't exist
		if err := s.manager.RegisterSite(updatedSite); err != nil {
			return nil, fmt.Errorf("failed to register site monitor: %w", err)
		}
	}

	return updatedSite, nil
}

func (s *SiteService) Delete(id int) error {
	s.manager.RevokeSite(id)
	return s.repo.Delete(id)
}

func (s *SiteService) InitializeMonitoring() error {
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
