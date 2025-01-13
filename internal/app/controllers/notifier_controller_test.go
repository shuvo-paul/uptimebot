package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/stretchr/testify/assert"
)

type MockNotifierService struct {
	createFunc              func(notifier *models.Notifier) error
	getFunc                 func(id int64) (*models.Notifier, error)
	updateFunc              func(id int, config *models.NotifierConfig) (*models.Notifier, error)
	deleteFunc              func(id int64) error
	configureObserversFunc  func(siteID int) error
	handleSlackCallbackFunc func(code string, siteId int) (*models.Notifier, error)
	parseOAuthStateFunc     func(state string) (int, error)
}

func (m *MockNotifierService) Create(notifier *models.Notifier) error {
	return m.createFunc(notifier)
}

func (m *MockNotifierService) Get(id int64) (*models.Notifier, error) {
	return m.getFunc(id)
}

func (m *MockNotifierService) Update(id int, config *models.NotifierConfig) (*models.Notifier, error) {
	return m.updateFunc(id, config)
}

func (m *MockNotifierService) Delete(id int64) error {
	return m.deleteFunc(id)
}

func (m *MockNotifierService) ConfigureObservers(siteID int) error {
	return m.configureObserversFunc(siteID)
}

func (m *MockNotifierService) HandleSlackCallback(code string, siteId int) (*models.Notifier, error) {
	return m.handleSlackCallbackFunc(code, siteId)
}

func (m *MockNotifierService) ParseOAuthState(state string) (int, error) {
	return m.parseOAuthStateFunc(state)
}

func TestNotifierController_AuthSlack(t *testing.T) {
	mockService := new(MockNotifierService)
	controller := NewNotifierController(mockService)

	t.Run("successful redirect", func(t *testing.T) {
		os.Setenv("SLACK_REDIRECT_URI", "http://example.com/callback")
		os.Setenv("SLACK_CLIENT_ID", "test_client_id")
		defer func() {
			os.Unsetenv("SLACK_REDIRECT_URI")
			os.Unsetenv("SLACK_CLIENT_ID")
		}()

		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/", nil)
		w := httptest.NewRecorder()

		controller.AuthSlack(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)

		expectedLocation := fmt.Sprintf(
			"https://slack.com/oauth/v2/authorize?" +
				"scope=incoming-webhook&" +
				"user_scope=&" +
				"redirect_uri=http://example.com/callback&" +
				"client_id=test_client_id",
		)

		assert.Equal(t, expectedLocation, w.Header().Get("Location"))
	})

	t.Run("missing environment variables", func(t *testing.T) {
		// Setup - ensure env vars are not set
		os.Unsetenv("SLACK_REDIRECT_URI")
		os.Unsetenv("SLACK_CLIENT_ID")

		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/", nil)
		w := httptest.NewRecorder()

		controller.AuthSlack(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing environment variables")
	})
}

func TestNotifierController_AuthSlackCallback(t *testing.T) {
	mockService := new(MockNotifierService)
	mockService.handleSlackCallbackFunc = func(code string, siteId int) (*models.Notifier, error) {
		return &models.Notifier{ID: 1, SiteId: siteId}, nil
	}
	mockService.parseOAuthStateFunc = func(state string) (int, error) {
		return 1, nil
	}
	mockService.createFunc = func(notifier *models.Notifier) error {
		return nil
	}
	controller := NewNotifierController(mockService)

	t.Run("successful callback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/callback?code=test_code&state=site_id=1", nil)
		w := httptest.NewRecorder()

		controller.AuthSlackCallback(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites/1", w.Header().Get("Location"))
	})

	t.Run("invalid code", func(t *testing.T) {
		mockService.handleSlackCallbackFunc = func(code string, siteId int) (*models.Notifier, error) {
			return nil, fmt.Errorf("invalid code")
		}
		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/callback?code=&state=site_id=1", nil)
		w := httptest.NewRecorder()

		controller.AuthSlackCallback(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "invalid code")
	})

	t.Run("invalid state", func(t *testing.T) {
		mockService.parseOAuthStateFunc = func(state string) (int, error) {
			return 0, fmt.Errorf("invalid state")
		}
		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/callback?code=test_code&state=invalid_state", nil)
		w := httptest.NewRecorder()

		controller.AuthSlackCallback(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid state")
	})
}
