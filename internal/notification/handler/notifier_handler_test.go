package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	notification "github.com/shuvo-paul/uptimebot/internal/notification/core"
	"github.com/shuvo-paul/uptimebot/internal/notification/model"
	"github.com/stretchr/testify/assert"
)

type MockNotifierService struct {
	createFunc              func(notifier *model.Notifier) error
	getFunc                 func(id int) (*model.Notifier, error)
	updateFunc              func(id int, config json.RawMessage) (*model.Notifier, error)
	deleteFunc              func(id int) error
	configureObserversFunc  func(targetID int) error
	handleSlackCallbackFunc func(code string, targetId int) (*model.Notifier, error)
	parseOAuthStateFunc     func(state string) (int, error)
}

func (m *MockNotifierService) Create(notifier *model.Notifier) error {
	return m.createFunc(notifier)
}

func (m *MockNotifierService) Get(id int) (*model.Notifier, error) {
	return m.getFunc(id)
}

func (m *MockNotifierService) Update(id int, config json.RawMessage) (*model.Notifier, error) {
	return m.updateFunc(id, config)
}

func (m *MockNotifierService) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *MockNotifierService) ConfigureObservers(targetID int) error {
	return m.configureObserversFunc(targetID)
}

func (m *MockNotifierService) HandleSlackCallback(code string, targetId int) (*model.Notifier, error) {
	return m.handleSlackCallbackFunc(code, targetId)
}

func (m *MockNotifierService) ParseOAuthState(state string) (int, error) {
	return m.parseOAuthStateFunc(state)
}

func (m *MockNotifierService) GetSubject() *notification.Subject {
	return nil
}

func TestNotifierHandler_AuthSlack(t *testing.T) {
	mockService := new(MockNotifierService)
	handler := NewNotifierHandler(mockService)

	t.Run("successful redirect", func(t *testing.T) {
		os.Setenv("SLACK_REDIRECT_URI", "http://example.com/callback")
		os.Setenv("SLACK_CLIENT_ID", "test_client_id")
		defer func() {
			os.Unsetenv("SLACK_REDIRECT_URI")
			os.Unsetenv("SLACK_CLIENT_ID")
		}()

		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/", nil)
		req.SetPathValue("targetId", "1")
		w := httptest.NewRecorder()

		handler.AuthSlack(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)

		expectedLocation := fmt.Sprintf(
			"https://slack.com/oauth/v2/authorize?" +
				"scope=incoming-webhook&" +
				"user_scope=&" +
				"redirect_uri=http://example.com/callback&" +
				"client_id=test_client_id&" +
				"state=target_id=1",
		)

		assert.Equal(t, expectedLocation, w.Header().Get("Location"))
	})

	t.Run("missing environment variables", func(t *testing.T) {
		// Setup - ensure env vars are not set
		os.Unsetenv("SLACK_REDIRECT_URI")
		os.Unsetenv("SLACK_CLIENT_ID")

		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/", nil)
		req.SetPathValue("targetId", "1")
		w := httptest.NewRecorder()

		handler.AuthSlack(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing environment variables")
	})
}

func TestNotifierController_AuthSlackCallback(t *testing.T) {
	mockService := new(MockNotifierService)
	mockService.handleSlackCallbackFunc = func(code string, targetId int) (*model.Notifier, error) {
		return &model.Notifier{ID: 1, TargetId: targetId}, nil
	}
	mockService.parseOAuthStateFunc = func(state string) (int, error) {
		return 1, nil
	}
	mockService.createFunc = func(notifier *model.Notifier) error {
		return nil
	}
	controller := NewNotifierHandler(mockService)

	t.Run("successful callback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/callback?code=test_code&state=target_id=1", nil)
		w := httptest.NewRecorder()

		controller.AuthSlackCallback(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets/1", w.Header().Get("Location"))
	})

	t.Run("invalid code", func(t *testing.T) {
		mockService.handleSlackCallbackFunc = func(code string, targetId int) (*model.Notifier, error) {
			return nil, fmt.Errorf("invalid code")
		}
		req := httptest.NewRequest(http.MethodGet, "/oauth/slack/callback?code=&state=target_id=1", nil)
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
