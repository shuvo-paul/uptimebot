package services

import (
	"fmt"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/repository"
	"github.com/shuvo-paul/sitemonitor/pkg/monitor"
)

type SiteServiceInterface interface {
	Create(url string, interval time.Duration) (*monitor.Site, error)
	GetByID(id int) (*monitor.Site, error)
	GetAll() ([]*monitor.Site, error)
	Update(site *monitor.Site) (*monitor.Site, error)
	Delete(id int) error
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

func (s *SiteService) Create(url string, interval time.Duration) (*monitor.Site, error) {
	site := &monitor.Site{
		URL:      url,
		Interval: interval,
		Enabled:  true,
		Status:   "pending",
	}

	site, err := s.repo.Create(site)
	if err != nil {
		return nil, fmt.Errorf("failed to create site: %w", err)
	}

	// Create a new site monitor
	if err := s.manager.RegisterSite(site); err != nil {
		return nil, fmt.Errorf("failed to register site monitor: %w", err)
	}

	return site, nil
}

func (s *SiteService) GetByID(id int) (*monitor.Site, error) {
	return s.repo.GetByID(id)
}

func (s *SiteService) GetAll() ([]*monitor.Site, error) {
	return s.repo.GetAll()
}

func (s *SiteService) Update(site *monitor.Site) (*monitor.Site, error) {
	// First update the site in the database
	updatedSite, err := s.repo.Update(site)
	if err != nil {
		return nil, fmt.Errorf("failed to update site: %w", err)
	}

	// Check if the site monitor exists
	if existingSite, exists := s.manager.Sites[site.ID]; exists {
		// Update all fields of the existing site
		existingSite.URL = updatedSite.URL
		existingSite.Interval = updatedSite.Interval
		existingSite.Enabled = updatedSite.Enabled
		existingSite.Status = updatedSite.Status
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
