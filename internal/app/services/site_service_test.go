package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/pkg/monitor"
	"github.com/shuvo-paul/sitemonitor/pkg/notification"
	"github.com/stretchr/testify/assert"
)

// mockSiteRepository is a mock implementation of SiteRepositoryInterface
type mockSiteRepository struct {
	createFunc         func(userSite models.UserSite) (models.UserSite, error)
	getByIDFunc        func(id int) (*monitor.Site, error)
	getAllFunc         func() ([]*monitor.Site, error)
	updateFunc         func(site *monitor.Site) (*monitor.Site, error)
	deleteFunc         func(id int) error
	updateStatusFunc   func(site *monitor.Site, status string) error
	getAllByUserIDFunc func(userID int) ([]*monitor.Site, error)
}

func (m *mockSiteRepository) Create(userSite models.UserSite) (models.UserSite, error) {
	return m.createFunc(userSite)
}

func (m *mockSiteRepository) GetByID(id int) (*monitor.Site, error) {
	return m.getByIDFunc(id)
}

func (m *mockSiteRepository) GetAll() ([]*monitor.Site, error) {
	return m.getAllFunc()
}

func (m *mockSiteRepository) Update(site *monitor.Site) (*monitor.Site, error) {
	return m.updateFunc(site)
}

func (m *mockSiteRepository) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *mockSiteRepository) UpdateStatus(site *monitor.Site, status string) error {
	return m.updateStatusFunc(site, status)
}

func (m *mockSiteRepository) GetAllByUserID(userID int) ([]*monitor.Site, error) {
	return m.getAllByUserIDFunc(userID)
}

type mockNotifierService struct {
	configureObserversFunc func(siteID int) error
}

func (m *mockNotifierService) ConfigureObservers(siteID int) error {
	return m.configureObserversFunc(siteID)
}

func (m *mockNotifierService) Create(notifier *models.Notifier) error {
	return nil
}

func (m *mockNotifierService) Get(id int64) (*models.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) Update(id int, config *models.NotifierConfig) (*models.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) Delete(id int64) error {
	return nil
}

func (m *mockNotifierService) GetSubject() *notification.Subject {
	return nil
}

func (m *mockNotifierService) HandleSlackCallback(code string, siteId int) (*models.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) ParseOAuthState(state string) (int, error) {
	return 0, nil
}

func TestSiteService_Create(t *testing.T) {
	mockRepo := &mockSiteRepository{
		createFunc: func(userSite models.UserSite) (models.UserSite, error) {
			userSite.ID = 1
			return userSite, nil
		},
	}
	mockNotifierService := &mockNotifierService{}

	service := NewSiteService(mockRepo, mockNotifierService)

	t.Run("Site created successfully", func(t *testing.T) {
		url := "https://example.com"
		interval := time.Second * 30

		site, err := service.Create(1, url, interval)
		assert.NoError(t, err)
		assert.Equal(t, 1, site.ID)
		assert.Equal(t, url, site.URL)
		assert.Equal(t, interval, site.Interval)
		assert.True(t, site.Enabled)
		assert.Equal(t, "pending", site.Status)

		// Verify the site was registered with the monitor manager
		assert.Contains(t, service.manager.Sites, site.ID)
	})

	t.Run("Create fails", func(t *testing.T) {
		mockRepo.createFunc = func(userSite models.UserSite) (models.UserSite, error) {
			return models.UserSite{}, fmt.Errorf("database error")
		}

		_, err := service.Create(1, "https://example.com", time.Second*30)
		assert.Error(t, err)
	})
}

func TestSiteService_Update(t *testing.T) {
	mockRepo := &mockSiteRepository{
		updateFunc: func(site *monitor.Site) (*monitor.Site, error) {
			return site, nil
		},
	}
	service := NewSiteService(mockRepo, &mockNotifierService{})

	t.Run("Update existing site", func(t *testing.T) {
		// Create and register initial site
		site := &monitor.Site{
			ID:       1,
			URL:      "https://example.com",
			Interval: time.Second * 30,
			Enabled:  true,
			Status:   "up",
		}
		err := service.manager.RegisterSite(site)
		assert.NoError(t, err)

		// Update the site
		site.URL = "https://updated-example.com"
		site.Interval = time.Second * 60
		site.Enabled = false

		result, err := service.Update(site)
		assert.NoError(t, err)
		assert.Equal(t, site.URL, result.URL)
		assert.Equal(t, site.Interval, result.Interval)
		assert.Equal(t, site.Enabled, result.Enabled)

		// Verify the monitor was updated
		existingSite := service.manager.Sites[site.ID]
		assert.Equal(t, site.URL, existingSite.URL)
		assert.Equal(t, site.Interval, existingSite.Interval)
		assert.Equal(t, site.Enabled, existingSite.Enabled)
	})

	t.Run("Update fails", func(t *testing.T) {
		mockRepo.updateFunc = func(site *monitor.Site) (*monitor.Site, error) {
			return nil, fmt.Errorf("database error")
		}

		site := &monitor.Site{ID: 1}
		_, err := service.Update(site)
		assert.Error(t, err)
	})
}

func TestSiteService_Delete(t *testing.T) {
	mockRepo := &mockSiteRepository{
		deleteFunc: func(id int) error {
			return nil
		},
	}
	service := NewSiteService(mockRepo, &mockNotifierService{})

	t.Run("Delete existing site", func(t *testing.T) {
		// Register a site first
		site := &monitor.Site{
			ID:       1,
			URL:      "https://example.com",
			Interval: time.Second * 30,
			Enabled:  true,
		}
		err := service.manager.RegisterSite(site)
		assert.NoError(t, err)

		err = service.Delete(site.ID)
		assert.NoError(t, err)
		assert.NotContains(t, service.manager.Sites, site.ID)
	})

	t.Run("Delete fails", func(t *testing.T) {
		mockRepo.deleteFunc = func(id int) error {
			return fmt.Errorf("database error")
		}

		// Try to delete a site that doesn't exist in the manager
		err := service.Delete(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestSiteService_GetAllByUserID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		expectedSites := []*monitor.Site{
			{ID: 1, URL: "site1.com", Status: "up"},
			{ID: 2, URL: "site2.com", Status: "down"},
		}

		mockRepo := &mockSiteRepository{
			getAllByUserIDFunc: func(userID int) ([]*monitor.Site, error) {
				assert.Equal(t, 1, userID)
				return expectedSites, nil
			},
		}

		service := NewSiteService(mockRepo, &mockNotifierService{})
		sites, err := service.GetAllByUserID(1)

		assert.NoError(t, err)
		assert.Equal(t, expectedSites, sites)
	})

	t.Run("no sites found", func(t *testing.T) {
		mockRepo := &mockSiteRepository{
			getAllByUserIDFunc: func(userID int) ([]*monitor.Site, error) {
				return []*monitor.Site{}, nil
			},
		}

		service := NewSiteService(mockRepo, &mockNotifierService{})
		sites, err := service.GetAllByUserID(999)

		assert.NoError(t, err)
		assert.Empty(t, sites)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &mockSiteRepository{
			getAllByUserIDFunc: func(userID int) ([]*monitor.Site, error) {
				return nil, fmt.Errorf("database error")
			},
		}

		service := NewSiteService(mockRepo, &mockNotifierService{})
		sites, err := service.GetAllByUserID(1)

		assert.Error(t, err)
		assert.Nil(t, sites)
	})
}
