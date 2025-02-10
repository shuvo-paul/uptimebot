package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	authModel "github.com/shuvo-paul/uptimebot/internal/auth/model"
	authService "github.com/shuvo-paul/uptimebot/internal/auth/service"
	monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/internal/templates"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// Mock TargetService
type mockTargetService struct {
	getAllFunc               func() ([]*monitor.Target, error)
	getByIDFunc              func(id int) (*monitor.Target, error)
	createFunc               func(userID int, url string, interval time.Duration) (*monitor.Target, error)
	updateFunc               func(target *monitor.Target) (*monitor.Target, error)
	deleteFunc               func(id int) error
	getAllByUserIDFunc       func(userID int) ([]*monitor.Target, error)
	initializeMonitoringFunc func() error
}

func (m *mockTargetService) GetAll() ([]*monitor.Target, error) {
	return m.getAllFunc()
}

func (m *mockTargetService) GetByID(id int) (*monitor.Target, error) {
	return m.getByIDFunc(id)
}

func (m *mockTargetService) Create(userID int, url string, interval time.Duration) (*monitor.Target, error) {
	return m.createFunc(userID, url, interval)
}

func (m *mockTargetService) Update(target *monitor.Target) (*monitor.Target, error) {
	return m.updateFunc(target)
}

func (m *mockTargetService) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *mockTargetService) GetAllByUserID(userID int) ([]*monitor.Target, error) {
	return m.getAllByUserIDFunc(userID)
}

func (m *mockTargetService) InitializeMonitoring() error {
	if m.initializeMonitoringFunc != nil {
		return m.initializeMonitoringFunc()
	}
	return nil
}

func TestTargetHandler_List(t *testing.T) {
	mockService := &mockTargetService{
		getAllByUserIDFunc: func(userID int) ([]*monitor.Target, error) {
			return []*monitor.Target{{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}}, nil
		},
		initializeMonitoringFunc: func() error { return nil },
	}

	handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})
	templateRenderer := renderer.New(templates.TemplateFS)
	handler.Template.List = templateRenderer.GetTemplate("pages:targets/list")

	req := httptest.NewRequest(http.MethodGet, "/targets", nil)
	user := &authModel.User{ID: 1}
	ctx := authService.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTargetHandler_Create(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockTargetService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS)
		handler.Template.Create = templateRenderer.GetTemplate("targets/create")

		req := httptest.NewRequest(http.MethodGet, "/targets/create", nil)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		mockService := &mockTargetService{
			createFunc: func(userID int, url string, interval time.Duration) (*monitor.Target, error) {
				return &monitor.Target{ID: 1, URL: url, Interval: interval}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/targets/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Add user to context
		user := &authModel.User{ID: 1, Name: "Test User"}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})

	t.Run("POST request - no user in context", func(t *testing.T) {
		mockService := &mockTargetService{
			createFunc: func(userID int, url string, interval time.Duration) (*monitor.Target, error) {
				return &monitor.Target{ID: 1, URL: url, Interval: interval}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/targets/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTargetHandler_Edit(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockTargetService{
			getByIDFunc: func(id int) (*monitor.Target, error) {
				return &monitor.Target{ID: id, URL: "http://example.com", Interval: 60 * time.Second}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS)
		handler.Template.Edit = templateRenderer.GetTemplate("pages:targets/edit")

		req := httptest.NewRequest(http.MethodGet, "/targets/1/edit", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler.Edit(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		target := &monitor.Target{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}
		mockService := &mockTargetService{
			getByIDFunc: func(id int) (*monitor.Target, error) {
				return target, nil
			},
			updateFunc: func(t *monitor.Target) (*monitor.Target, error) {
				return t, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/targets/1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler.Edit(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})
}

func TestTargetHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := &mockTargetService{
			deleteFunc: func(id int) error {
				return nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/1/delete", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})

	t.Run("invalid ID", func(t *testing.T) {
		mockService := &mockTargetService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/invalid/delete", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
