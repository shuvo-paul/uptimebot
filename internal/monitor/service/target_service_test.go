package service

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
	notifCore "github.com/shuvo-paul/uptimebot/internal/notification/core"
	alertModel "github.com/shuvo-paul/uptimebot/internal/notification/model"
	"github.com/stretchr/testify/assert"
)

// mockSiteRepository is a mock implementation of SiteRepositoryInterface
type mockSiteRepository struct {
	createFunc         func(model.UserTarget) (model.UserTarget, error)
	getByIDFunc        func(id int) (*monitor.Target, error)
	getAllFunc         func() ([]*monitor.Target, error)
	updateFunc         func(site *monitor.Target) (*monitor.Target, error)
	deleteFunc         func(id int) error
	updateStatusFunc   func(site *monitor.Target, status string) error
	getAllByUserIDFunc func(userID int) ([]*monitor.Target, error)
}

func (m *mockSiteRepository) Create(userTarget model.UserTarget) (model.UserTarget, error) {
	return m.createFunc(userTarget)
}

func (m *mockSiteRepository) GetByID(id int) (*monitor.Target, error) {
	return m.getByIDFunc(id)
}

func (m *mockSiteRepository) GetAll() ([]*monitor.Target, error) {
	return m.getAllFunc()
}

func (m *mockSiteRepository) Update(site *monitor.Target) (*monitor.Target, error) {
	return m.updateFunc(site)
}

func (m *mockSiteRepository) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *mockSiteRepository) UpdateStatus(site *monitor.Target, status string) error {
	return m.updateStatusFunc(site, status)
}

func (m *mockSiteRepository) GetAllByUserID(userID int) ([]*monitor.Target, error) {
	return m.getAllByUserIDFunc(userID)
}

type mockNotifierService struct {
	configureObserversFunc func(siteID int) error
}

func (m *mockNotifierService) ConfigureObservers(siteID int) error {
	return m.configureObserversFunc(siteID)
}

func (m *mockNotifierService) Create(notifier *alertModel.Notifier) error {
	return nil
}

func (m *mockNotifierService) Get(id int64) (*alertModel.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) Update(id int, config json.RawMessage) (*alertModel.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) Delete(id int64) error {
	return nil
}

func (m *mockNotifierService) GetSubject() *notifCore.Subject {
	return nil
}

func (m *mockNotifierService) HandleSlackCallback(code string, siteId int) (*alertModel.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) ParseOAuthState(state string) (int, error) {
	return 0, nil
}

func TestSiteService_Create(t *testing.T) {
	mockRepo := &mockSiteRepository{
		createFunc: func(userTarget model.UserTarget) (model.UserTarget, error) {
			userTarget.ID = 1
			return userTarget, nil
		},
	}
	mockNotifierService := &mockNotifierService{}

	service := NewTargetService(mockRepo, mockNotifierService)

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
		assert.Contains(t, service.manager.Targets, site.ID)
	})

	t.Run("Create fails", func(t *testing.T) {
		mockRepo.createFunc = func(userTarget model.UserTarget) (model.UserTarget, error) {
			return model.UserTarget{}, fmt.Errorf("database error")
		}

		_, err := service.Create(1, "https://example.com", time.Second*30)
		assert.Error(t, err)
	})
}

func TestSiteService_Update(t *testing.T) {
	mockRepo := &mockSiteRepository{
		updateFunc: func(site *monitor.Target) (*monitor.Target, error) {
			return site, nil
		},
	}
	service := NewTargetService(mockRepo, &mockNotifierService{})

	t.Run("Update existing site", func(t *testing.T) {
		// Create and register initial site
		site := &monitor.Target{
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
		existingSite := service.manager.Targets[site.ID]
		assert.Equal(t, site.URL, existingSite.URL)
		assert.Equal(t, site.Interval, existingSite.Interval)
		assert.Equal(t, site.Enabled, existingSite.Enabled)
	})

	t.Run("Update fails", func(t *testing.T) {
		mockRepo.updateFunc = func(site *monitor.Target) (*monitor.Target, error) {
			return nil, fmt.Errorf("database error")
		}

		site := &monitor.Target{ID: 1}
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
	service := NewTargetService(mockRepo, &mockNotifierService{})

	t.Run("Delete existing site", func(t *testing.T) {
		// Register a site first
		site := &monitor.Target{
			ID:       1,
			URL:      "https://example.com",
			Interval: time.Second * 30,
			Enabled:  true,
		}
		err := service.manager.RegisterSite(site)
		assert.NoError(t, err)

		err = service.Delete(site.ID)
		assert.NoError(t, err)
		assert.NotContains(t, service.manager.Targets, site.ID)
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
		expectedSites := []*monitor.Target{
			{ID: 1, URL: "site1.com", Status: "up"},
			{ID: 2, URL: "site2.com", Status: "down"},
		}

		mockRepo := &mockSiteRepository{
			getAllByUserIDFunc: func(userID int) ([]*monitor.Target, error) {
				assert.Equal(t, 1, userID)
				return expectedSites, nil
			},
		}

		service := NewTargetService(mockRepo, &mockNotifierService{})
		sites, err := service.GetAllByUserID(1)

		assert.NoError(t, err)
		assert.Equal(t, expectedSites, sites)
	})

	t.Run("no sites found", func(t *testing.T) {
		mockRepo := &mockSiteRepository{
			getAllByUserIDFunc: func(userID int) ([]*monitor.Target, error) {
				return []*monitor.Target{}, nil
			},
		}

		service := NewTargetService(mockRepo, &mockNotifierService{})
		sites, err := service.GetAllByUserID(999)

		assert.NoError(t, err)
		assert.Empty(t, sites)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &mockSiteRepository{
			getAllByUserIDFunc: func(userID int) ([]*monitor.Target, error) {
				return nil, fmt.Errorf("database error")
			},
		}

		service := NewTargetService(mockRepo, &mockNotifierService{})
		sites, err := service.GetAllByUserID(1)

		assert.Error(t, err)
		assert.Nil(t, sites)
	})
}
