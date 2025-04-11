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

// mockTargetRepository is a mock implementation of TargetRepositoryInterface
type mockTargetRepository struct {
	createFunc         func(model.UserTarget) (model.UserTarget, error)
	getByIDFunc        func(id int) (model.UserTarget, error)
	getAllFunc         func() ([]model.UserTarget, error)
	updateFunc         func(target model.UserTarget) (model.UserTarget, error)
	deleteFunc         func(id int) error
	updateStatusFunc   func(target *monitor.Target, status string) error
	getAllByUserIDFunc func(userID int) ([]model.UserTarget, error)
}

func (m *mockTargetRepository) Create(userTarget model.UserTarget) (model.UserTarget, error) {
	return m.createFunc(userTarget)
}

func (m *mockTargetRepository) GetByID(id int) (model.UserTarget, error) {
	return m.getByIDFunc(id)
}

func (m *mockTargetRepository) GetAll() ([]model.UserTarget, error) {
	return m.getAllFunc()
}

func (m *mockTargetRepository) Update(target model.UserTarget) (model.UserTarget, error) {
	return m.updateFunc(target)
}

func (m *mockTargetRepository) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *mockTargetRepository) UpdateStatus(target *monitor.Target, status string) error {
	return m.updateStatusFunc(target, status)
}

func (m *mockTargetRepository) GetAllByUserID(userID int) ([]model.UserTarget, error) {
	return m.getAllByUserIDFunc(userID)
}

type mockNotifierService struct {
	configureObserversFunc func(targetID int) error
}

func (m *mockNotifierService) ConfigureObservers(targetID int) error {
	return m.configureObserversFunc(targetID)
}

func (m *mockNotifierService) Create(notifier *alertModel.Notifier) error {
	return nil
}

func (m *mockNotifierService) Get(id int) (*alertModel.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) Update(id int, config json.RawMessage) (*alertModel.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) Delete(id int) error {
	return nil
}

func (m *mockNotifierService) GetSubject() *notifCore.Subject {
	return nil
}

func (m *mockNotifierService) HandleSlackCallback(code string, targetID int) (*alertModel.Notifier, error) {
	return nil, nil
}

func (m *mockNotifierService) ParseOAuthState(state string) (int, error) {
	return 0, nil
}

func TestTargetService_Create(t *testing.T) {
	mockRepo := &mockTargetRepository{
		createFunc: func(userTarget model.UserTarget) (model.UserTarget, error) {
			userTarget.ID = 1
			return userTarget, nil
		},
		getAllByUserIDFunc: func(userID int) ([]model.UserTarget, error) {
			return []model.UserTarget{}, nil
		},
	}
	mockNotifierService := &mockNotifierService{}

	service := NewTargetService(mockRepo, mockNotifierService)

	t.Run("Target created successfully", func(t *testing.T) {
		url := "https://example.com"
		interval := time.Second * 30

		target, err := service.Create(1, url, interval)
		assert.NoError(t, err)
		assert.Equal(t, 1, target.ID)
		assert.Equal(t, url, target.URL)
		assert.Equal(t, interval, target.Interval)
		assert.True(t, target.Enabled)
		assert.Equal(t, "pending", target.Status)

		// Verify the target was registered with the monitor manager
		assert.Contains(t, service.manager.Targets, target.ID)
	})

	t.Run("Create fails", func(t *testing.T) {
		mockRepo.createFunc = func(userTarget model.UserTarget) (model.UserTarget, error) {
			return model.UserTarget{}, fmt.Errorf("database error")
		}

		_, err := service.Create(1, "https://example.com", time.Second*30)
		assert.Error(t, err)
	})
}

func TestTargetService_Update(t *testing.T) {
	mockRepo := &mockTargetRepository{
		updateFunc: func(target model.UserTarget) (model.UserTarget, error) {
			return target, nil
		},
		getByIDFunc: func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: 1,
				Target: &monitor.Target{
					ID:       id,
					URL:      "https://example.com",
					Interval: time.Second * 30,
					Enabled:  true,
					Status:   "up",
				},
			}, nil
		},
	}
	service := NewTargetService(mockRepo, &mockNotifierService{})

	t.Run("Update existing target", func(t *testing.T) {
		// Create and register initial target
		userTarget := model.UserTarget{
			UserID: 1,
			Target: &monitor.Target{
				ID:       1,
				URL:      "https://example.com",
				Interval: time.Second * 30,
				Enabled:  true,
				Status:   "up",
			},
		}
		err := service.manager.RegisterTarget(userTarget.Target)
		assert.NoError(t, err)

		// Update the target
		userTarget.URL = "https://updated-example.com"
		userTarget.Interval = time.Second * 60
		userTarget.Enabled = false

		result, err := service.Update(userTarget, userTarget.UserID)
		assert.NoError(t, err)
		assert.Equal(t, userTarget.URL, result.URL)
		assert.Equal(t, userTarget.Interval, result.Interval)
		assert.Equal(t, userTarget.Enabled, result.Enabled)

		// Verify the monitor was updated
		existingTarget := service.manager.Targets[userTarget.ID]
		assert.Equal(t, userTarget.URL, existingTarget.URL)
		assert.Equal(t, userTarget.Interval, existingTarget.Interval)
		assert.Equal(t, userTarget.Enabled, existingTarget.Enabled)
	})

	t.Run("Update fails", func(t *testing.T) {
		mockRepo.updateFunc = func(target model.UserTarget) (model.UserTarget, error) {
			return model.UserTarget{}, fmt.Errorf("database error")
		}

		userTarget := model.UserTarget{UserID: 1, Target: &monitor.Target{ID: 1}}
		_, err := service.Update(userTarget, userTarget.UserID)
		assert.Error(t, err)
	})

	t.Run("Update unauthorized target", func(t *testing.T) {
		userTarget := model.UserTarget{UserID: 1, Target: &monitor.Target{
			ID:       1,
			URL:      "https://example.com",
			Interval: time.Second * 30,
			Enabled:  true,
			Status:   "up",
		}}
		_, err := service.Update(userTarget, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestTargetService_Delete(t *testing.T) {
	mockRepo := &mockTargetRepository{
		deleteFunc: func(id int) error {
			return nil
		},
		getByIDFunc: func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: 1,
				Target: &monitor.Target{ID: id},
			}, nil
		},
	}
	service := NewTargetService(mockRepo, &mockNotifierService{})

	t.Run("Delete existing target", func(t *testing.T) {
		// Register a target first
		target := &monitor.Target{
			ID:       1,
			URL:      "https://example.com",
			Interval: time.Second * 30,
			Enabled:  true,
		}
		err := service.manager.RegisterTarget(target)
		assert.NoError(t, err)

		err = service.Delete(1, 1) // UserID 1 deleting target 1
		assert.NoError(t, err)
		assert.NotContains(t, service.manager.Targets, target.ID)
	})

	t.Run("Delete unauthorized target", func(t *testing.T) {
		mockRepo.getByIDFunc = func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: 2, // Different user ID
				Target: &monitor.Target{ID: id},
			}, nil
		}

		err := service.Delete(1, 1) // UserID 1 trying to delete target owned by UserID 2
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("Delete fails", func(t *testing.T) {
		mockRepo.deleteFunc = func(id int) error {
			return fmt.Errorf("database error")
		}

		mockRepo.getByIDFunc = func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: id, // Different user ID
				Target: &monitor.Target{ID: id},
			}, nil
		}

		err := service.Delete(1, 1)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "database error")
	})
}

func TestTargetService_ToggleEnabled(t *testing.T) {
	mockRepo := &mockTargetRepository{
		getByIDFunc: func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: 1,
				Target: &monitor.Target{
					ID:       id,
					Enabled:  true,
					URL:      "https://example.com",
					Interval: time.Second * 30,
				},
			}, nil
		},
		updateFunc: func(target model.UserTarget) (model.UserTarget, error) {
			return target, nil
		},
	}
	service := NewTargetService(mockRepo, &mockNotifierService{})

	t.Run("Toggle target successfully", func(t *testing.T) {
		// Register initial target
		target := &monitor.Target{
			ID:       1,
			Enabled:  true,
			URL:      "https://example.com",
			Interval: time.Second * 30,
		}
		err := service.manager.RegisterTarget(target)
		assert.NoError(t, err)

		// Toggle the target
		result, err := service.ToggleEnabled(1, 1)
		assert.NoError(t, err)
		assert.False(t, result.Enabled)

		// Verify the monitor was updated
		existingTarget := service.manager.Targets[target.ID]
		assert.False(t, existingTarget.Enabled)
	})

	t.Run("Toggle unauthorized target", func(t *testing.T) {
		mockRepo.getByIDFunc = func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: 2, // Different user ID
				Target: &monitor.Target{ID: id},
			}, nil
		}

		_, err := service.ToggleEnabled(1, 1) // UserID 1 trying to toggle target owned by UserID 2
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unauthorized")
	})

	t.Run("Toggle fails - database error", func(t *testing.T) {
		mockRepo.getByIDFunc = func(id int) (model.UserTarget, error) {
			return model.UserTarget{
				UserID: 1,
				Target: &monitor.Target{ID: id},
			}, nil
		}
		mockRepo.updateFunc = func(target model.UserTarget) (model.UserTarget, error) {
			return model.UserTarget{}, fmt.Errorf("database error")
		}

		_, err := service.ToggleEnabled(1, 1)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "database error")
	})
}

func TestTargetService_GetAllByUserID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		expectedTargets := []model.UserTarget{
			{
				UserID: 1,
				Target: &monitor.Target{ID: 1, URL: "target1.com", Status: "up"},
			},
			{
				UserID: 1,
				Target: &monitor.Target{ID: 2, URL: "target2.com", Status: "down"},
			},
		}

		mockRepo := &mockTargetRepository{
			getAllByUserIDFunc: func(userID int) ([]model.UserTarget, error) {
				assert.Equal(t, 1, userID)
				return expectedTargets, nil
			},
		}

		service := NewTargetService(mockRepo, &mockNotifierService{})
		targets, err := service.GetAllByUserID(1)

		assert.NoError(t, err)
		assert.Equal(t, expectedTargets, targets)
	})

	t.Run("no targets found", func(t *testing.T) {
		mockRepo := &mockTargetRepository{
			getAllByUserIDFunc: func(userID int) ([]model.UserTarget, error) {
				return []model.UserTarget{}, nil
			},
		}

		service := NewTargetService(mockRepo, &mockNotifierService{})
		targets, err := service.GetAllByUserID(999)

		assert.NoError(t, err)
		assert.Empty(t, targets)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &mockTargetRepository{
			getAllByUserIDFunc: func(userID int) ([]model.UserTarget, error) {
				return nil, fmt.Errorf("database error")
			},
		}

		service := NewTargetService(mockRepo, &mockNotifierService{})
		targets, err := service.GetAllByUserID(1)

		assert.Error(t, err)
		assert.Nil(t, targets)
	})
}
