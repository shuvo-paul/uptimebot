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

// Mock SiteService
type mockSiteService struct {
	getAllFunc               func() ([]*monitor.Target, error)
	getByIDFunc              func(id int) (*monitor.Target, error)
	createFunc               func(userID int, url string, interval time.Duration) (*monitor.Target, error)
	updateFunc               func(site *monitor.Target) (*monitor.Target, error)
	deleteFunc               func(id int) error
	getAllByUserIDFunc       func(userID int) ([]*monitor.Target, error)
	initializeMonitoringFunc func() error
}

func (m *mockSiteService) GetAll() ([]*monitor.Target, error) {
	return m.getAllFunc()
}

func (m *mockSiteService) GetByID(id int) (*monitor.Target, error) {
	return m.getByIDFunc(id)
}

func (m *mockSiteService) Create(userID int, url string, interval time.Duration) (*monitor.Target, error) {
	return m.createFunc(userID, url, interval)
}

func (m *mockSiteService) Update(site *monitor.Target) (*monitor.Target, error) {
	return m.updateFunc(site)
}

func (m *mockSiteService) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *mockSiteService) GetAllByUserID(userID int) ([]*monitor.Target, error) {
	return m.getAllByUserIDFunc(userID)
}

func (m *mockSiteService) InitializeMonitoring() error {
	if m.initializeMonitoringFunc != nil {
		return m.initializeMonitoringFunc()
	}
	return nil
}

func TestSiteHandler_List(t *testing.T) {
	mockService := &mockSiteService{
		getAllByUserIDFunc: func(userID int) ([]*monitor.Target, error) {
			return []*monitor.Target{{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}}, nil
		},
		initializeMonitoringFunc: func() error { return nil },
	}

	handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})
	templateRenderer := renderer.New(templates.TemplateFS)
	handler.Template.List = templateRenderer.Parse("sites/list.html")

	req := httptest.NewRequest(http.MethodGet, "/sites", nil)
	user := &authModel.User{ID: 1}
	ctx := authService.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSiteHandler_Create(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockSiteService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS)
		handler.Template.Create = templateRenderer.Parse("sites/create.html")

		req := httptest.NewRequest(http.MethodGet, "/sites/create", nil)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		mockService := &mockSiteService{
			createFunc: func(userID int, url string, interval time.Duration) (*monitor.Target, error) {
				return &monitor.Target{ID: 1, URL: url, Interval: interval}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/sites/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Add user to context
		user := &authModel.User{ID: 1, Name: "Test User"}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites", w.Header().Get("Location"))
	})

	t.Run("POST request - no user in context", func(t *testing.T) {
		mockService := &mockSiteService{
			createFunc: func(userID int, url string, interval time.Duration) (*monitor.Target, error) {
				return &monitor.Target{ID: 1, URL: url, Interval: interval}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/sites/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSiteHandler_Edit(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockSiteService{
			getByIDFunc: func(id int) (*monitor.Target, error) {
				return &monitor.Target{ID: id, URL: "http://example.com", Interval: 60 * time.Second}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS)
		handler.Template.Edit = templateRenderer.Parse("sites/edit.html")

		req := httptest.NewRequest(http.MethodGet, "/sites/1/edit", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler.Edit(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		site := &monitor.Target{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}
		mockService := &mockSiteService{
			getByIDFunc: func(id int) (*monitor.Target, error) {
				return site, nil
			},
			updateFunc: func(site *monitor.Target) (*monitor.Target, error) {
				return site, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/sites/1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler.Edit(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites", w.Header().Get("Location"))
	})
}

func TestSitehandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := &mockSiteService{
			deleteFunc: func(id int) error {
				return nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/sites/1/delete", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites", w.Header().Get("Location"))
	})

	t.Run("invalid ID", func(t *testing.T) {
		mockService := &mockSiteService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &testutil.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/sites/invalid/delete", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
