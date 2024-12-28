package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/pkg/notification"
	"github.com/stretchr/testify/assert"
)

// mockNotifierRepository is a mock implementation of NotifierRepositoryInterface
type mockNotifierRepository struct {
	getBySiteIDFunc func(siteID int) ([]*models.Notifier, error)
	createFunc      func(notifier *models.Notifier) error
	getFunc         func(id int64) (*models.Notifier, error)
	updateFunc      func(id int, config *models.NotifierConfig) (*models.Notifier, error)
	deleteFunc      func(id int64) error
}

func (m *mockNotifierRepository) GetBySiteID(siteID int) ([]*models.Notifier, error) {
	return m.getBySiteIDFunc(siteID)
}

func (m *mockNotifierRepository) Create(notifier *models.Notifier) error {
	return m.createFunc(notifier)
}

func (m *mockNotifierRepository) Get(id int64) (*models.Notifier, error) {
	return m.getFunc(id)
}

func (m *mockNotifierRepository) Update(id int, config *models.NotifierConfig) (*models.Notifier, error) {
	return m.updateFunc(id, config)
}

func (m *mockNotifierRepository) Delete(id int64) error {
	return m.deleteFunc(id)
}

// mockObserver is a mock implementation of the Observer interface
type mockObserver struct {
	state notification.State
	err   error
}

func newMockObserver(err error) *mockObserver {
	return &mockObserver{err: err}
}

func (m *mockObserver) Notify(state notification.State) error {
	if m.err != nil {
		return m.err
	}
	m.state = state
	return nil
}

func TestNotifierService_Create(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	service := NewNotifierService(mockRepo, nil)

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.createFunc = func(notifier *models.Notifier) error {
			return nil
		}

		notifier := &models.Notifier{
			SiteId: 1,
			Config: &models.NotifierConfig{
				Type: models.NotifierTypeSlack,
				Config: []byte(`{
					"webhook_url": "https://hooks.slack.com/test"
				}`),
			},
		}

		err := service.Create(notifier)
		assert.NoError(t, err)
	})

	t.Run("creation fails", func(t *testing.T) {
		mockRepo.createFunc = func(notifier *models.Notifier) error {
			return fmt.Errorf("db error")
		}

		err := service.Create(&models.Notifier{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create notifier")
	})
}

func TestNotifierService_Get(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	service := NewNotifierService(mockRepo, nil)

	t.Run("successful retrieval", func(t *testing.T) {
		expected := &models.Notifier{
			ID:     1,
			SiteId: 1,
			Config: &models.NotifierConfig{
				Type: models.NotifierTypeSlack,
			},
		}

		mockRepo.getFunc = func(id int64) (*models.Notifier, error) {
			return expected, nil
		}

		notifier, err := service.Get(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, notifier)
	})

	t.Run("retrieval fails", func(t *testing.T) {
		mockRepo.getFunc = func(id int64) (*models.Notifier, error) {
			return nil, fmt.Errorf("db error")
		}

		notifier, err := service.Get(1)
		assert.Error(t, err)
		assert.Nil(t, notifier)
		assert.Contains(t, err.Error(), "failed to get notifier")
	})
}

func TestNotifierService_Update(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	service := NewNotifierService(mockRepo, nil)

	t.Run("successful update", func(t *testing.T) {
		config := &models.NotifierConfig{
			Type: models.NotifierTypeSlack,
			Config: []byte(`{
				"webhook_url": "https://hooks.slack.com/new"
			}`),
		}

		expected := &models.Notifier{
			ID:     1,
			SiteId: 1,
			Config: config,
		}

		mockRepo.updateFunc = func(id int, cfg *models.NotifierConfig) (*models.Notifier, error) {
			return expected, nil
		}

		notifier, err := service.Update(1, config)
		assert.NoError(t, err)
		assert.Equal(t, expected, notifier)
	})

	t.Run("update fails", func(t *testing.T) {
		mockRepo.updateFunc = func(id int, cfg *models.NotifierConfig) (*models.Notifier, error) {
			return nil, fmt.Errorf("db error")
		}

		notifier, err := service.Update(1, &models.NotifierConfig{})
		assert.Error(t, err)
		assert.Nil(t, notifier)
		assert.Contains(t, err.Error(), "failed to update notifier")
	})
}

func TestNotifierService_Delete(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	service := NewNotifierService(mockRepo, nil)

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.deleteFunc = func(id int64) error {
			return nil
		}

		err := service.Delete(1)
		assert.NoError(t, err)
	})

	t.Run("deletion fails", func(t *testing.T) {
		mockRepo.deleteFunc = func(id int64) error {
			return fmt.Errorf("db error")
		}

		err := service.Delete(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete notifier")
	})
}

func TestNotifierService_ConfigureObservers(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	subject := notification.NewSubject()
	service := NewNotifierService(mockRepo, subject)

	t.Run("successful configuration with slack observer", func(t *testing.T) {
		mockRepo.getBySiteIDFunc = func(siteID int) ([]*models.Notifier, error) {
			return []*models.Notifier{
				{
					Config: &models.NotifierConfig{
						Type: models.NotifierTypeSlack,
						Config: []byte(`{
							"webhook_url": "https://hooks.slack.com/test"
						}`),
					},
				},
			}, nil
		}

		err := service.ConfigureObservers(1)
		assert.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.getBySiteIDFunc = func(siteID int) ([]*models.Notifier, error) {
			return nil, fmt.Errorf("db error")
		}

		err := service.ConfigureObservers(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get notifiers")
	})
}

func TestNotifierService_Subject(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	subject := notification.NewSubject()
	service := NewNotifierService(mockRepo, subject)

	// Create and attach mock observers
	observer1 := newMockObserver(nil)
	observer2 := newMockObserver(fmt.Errorf("notification error"))
	service.Subject.Attach(observer1)
	service.Subject.Attach(observer2)

	// Test notification using Subject directly
	state := notification.State{
		Name:      "test-system",
		Status:    "up",
		Message:   "System is up",
		UpdatedAt: time.Now(),
	}
	errors := service.Subject.Notify(state)

	// Verify results
	assert.Len(t, errors, 1) // One observer should fail
	assert.Equal(t, state, observer1.state)
	assert.Empty(t, observer2.state) // Failed observer shouldn't have state
}
