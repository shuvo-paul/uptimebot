package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/shuvo-paul/uptimebot/internal/notification/model"
	"github.com/shuvo-paul/uptimebot/internal/notification/provider"
	"github.com/stretchr/testify/assert"
)

// mockNotifierRepository is a mock implementation of NotifierRepositoryInterface
type mockNotifierRepository struct {
	getBySiteIDFunc func(siteID int) ([]*model.Notifier, error)
	createFunc      func(notifier *model.Notifier) (*model.Notifier, error)
	getFunc         func(id int64) (*model.Notifier, error)
	updateFunc      func(id int, config json.RawMessage) (*model.Notifier, error)
	deleteFunc      func(id int64) error
}

func (m *mockNotifierRepository) GetBySiteID(siteID int) ([]*model.Notifier, error) {
	return m.getBySiteIDFunc(siteID)
}

func (m *mockNotifierRepository) Create(notifier *model.Notifier) (*model.Notifier, error) {
	return m.createFunc(notifier)
}

func (m *mockNotifierRepository) Get(id int64) (*model.Notifier, error) {
	return m.getFunc(id)
}

func (m *mockNotifierRepository) Update(id int, config json.RawMessage) (*model.Notifier, error) {
	return m.updateFunc(id, config)
}

func (m *mockNotifierRepository) Delete(id int64) error {
	return m.deleteFunc(id)
}

// mockObserver is a mock implementation of the Observer interface
type mockObserver struct {
	state provider.State
	err   error
}

func newMockObserver(err error) *mockObserver {
	return &mockObserver{err: err}
}

func (m *mockObserver) Notify(state provider.State) error {
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
		mockRepo.createFunc = func(notifier *model.Notifier) (*model.Notifier, error) {
			return &model.Notifier{
				ID:     1,
				SiteId: 1,
				Type:   model.NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			}, nil
		}

		notifier := &model.Notifier{
			SiteId: 1,
			Type:   model.NotifierTypeSlack,
			Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
		}

		err := service.Create(notifier)
		assert.NoError(t, err)
	})

	t.Run("creation fails", func(t *testing.T) {
		mockRepo.createFunc = func(notifier *model.Notifier) (*model.Notifier, error) {
			return nil, fmt.Errorf("db error")
		}

		err := service.Create(&model.Notifier{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create notifier")
	})
}

func TestNotifierService_Get(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	service := NewNotifierService(mockRepo, nil)

	t.Run("successful retrieval", func(t *testing.T) {
		expected := &model.Notifier{
			ID:     1,
			SiteId: 1,
			Type:   model.NotifierTypeSlack,
			Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
		}

		mockRepo.getFunc = func(id int64) (*model.Notifier, error) {
			return expected, nil
		}

		notifier, err := service.Get(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, notifier)
	})

	t.Run("retrieval fails", func(t *testing.T) {
		mockRepo.getFunc = func(id int64) (*model.Notifier, error) {
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
		config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/new"}`)

		expected := &model.Notifier{
			ID:     1,
			SiteId: 1,
			Type:   model.NotifierTypeSlack,
			Config: config,
		}

		mockRepo.updateFunc = func(id int, cfg json.RawMessage) (*model.Notifier, error) {
			return expected, nil
		}

		notifier, err := service.Update(1, config)
		assert.NoError(t, err)
		assert.Equal(t, expected, notifier)
	})

	t.Run("update fails", func(t *testing.T) {
		mockRepo.updateFunc = func(id int, cfg json.RawMessage) (*model.Notifier, error) {
			return nil, fmt.Errorf("db error")
		}

		notifier, err := service.Update(1, json.RawMessage{})
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
	subject := provider.NewSubject()
	service := NewNotifierService(mockRepo, subject)

	t.Run("successful configuration with slack observer", func(t *testing.T) {
		mockRepo.getBySiteIDFunc = func(siteID int) ([]*model.Notifier, error) {
			return []*model.Notifier{
				{
					ID:     1,
					SiteId: 1,
					Type:   model.NotifierTypeSlack,
					Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
				},
			}, nil
		}

		err := service.ConfigureObservers(1)
		assert.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.getBySiteIDFunc = func(siteID int) ([]*model.Notifier, error) {
			return nil, fmt.Errorf("db error")
		}

		err := service.ConfigureObservers(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get notifiers")
	})
}

func TestNotifierService_Subject(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	subject := provider.NewSubject()
	service := NewNotifierService(mockRepo, subject)

	// Create and attach mock observers
	observer1 := newMockObserver(nil)
	observer2 := newMockObserver(fmt.Errorf("notification error"))
	service.subject.Attach(observer1)
	service.subject.Attach(observer2)

	// Test notification using Subject directly
	state := provider.State{
		Name:      "test-system",
		Status:    "up",
		Message:   "System is up",
		UpdatedAt: time.Now(),
	}
	errors := service.subject.Notify(state)

	// Verify results
	assert.Len(t, errors, 1) // One observer should fail
	assert.Equal(t, state, observer1.state)
	assert.Empty(t, observer2.state) // Failed observer shouldn't have state
}

func TestNotifierService_ParseOAuthState(t *testing.T) {
	mockRepo := &mockNotifierRepository{}
	service := NewNotifierService(mockRepo, nil)

	t.Run("successful parsing", func(t *testing.T) {
		state := "target_id=1"
		siteId, err := service.ParseOAuthState(state)
		assert.NoError(t, err)
		assert.Equal(t, 1, siteId)
	})

	t.Run("invalid state", func(t *testing.T) {
		state := "%invalid_state"
		_, err := service.ParseOAuthState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid state format")
	})

	t.Run("missing site id", func(t *testing.T) {
		state := "target_id="
		_, err := service.ParseOAuthState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing site id in state")
	})

	t.Run("invalid site id", func(t *testing.T) {
		state := "target_id=invalid"
		_, err := service.ParseOAuthState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid site id format")
	})
}

func TestNotifierService_HandleSlackCallback(t *testing.T) {
	// Create a mock HTTP server to simulate Slack's OAuth API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/oauth.v2.access" {
			t.Errorf("Expected /api/oauth.v2.access path, got %s", r.URL.Path)
		}

		err := r.ParseForm()
		if err != nil {
			t.Fatal(err)
		}

		// Verify required OAuth parameters
		if code := r.Form.Get("code"); code != "test_code" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":    false,
				"error": "invalid_code",
			})
			return
		}

		// Return successful response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": true,
			"incoming_webhook": map[string]interface{}{
				"url": "https://hooks.slack.com/services/TEST/WEBHOOK/URL",
			},
		})
	}))
	defer mockServer.Close()

	// Set environment variables for testing
	os.Setenv("SLACK_CLIENT_ID", "test_client_id")
	os.Setenv("SLACK_CLIENT_SECRET", "test_client_secret")

	tests := []struct {
		name      string
		code      string
		siteID    int
		wantErr   bool
		errString string
	}{
		{
			name:    "successful callback",
			code:    "test_code",
			siteID:  123,
			wantErr: false,
		},
		{
			name:      "empty code",
			code:      "",
			siteID:    123,
			wantErr:   true,
			errString: "missing code or client credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with mock repository
			mockRepo := &mockNotifierRepository{}
			service := NewNotifierService(mockRepo, nil)

			// Override the Slack API URL to point to our mock server
			originalURL := SlackTokenURL
			SlackTokenURL = mockServer.URL + "/api/oauth.v2.access"
			defer func() { SlackTokenURL = originalURL }()

			notifier, err := service.HandleSlackCallback(tt.code, tt.siteID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				assert.Nil(t, notifier)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, notifier)
			assert.Equal(t, tt.siteID, notifier.SiteId)
			assert.Equal(t, model.NotifierTypeSlack, notifier.Type)
			assert.Contains(t, string(notifier.Config), "hooks.slack.com")
		})
	}
}
