package controllers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shuvo-paul/sitemonitor/internal/app/renderer"
	"github.com/shuvo-paul/sitemonitor/internal/app/testutil"
	"github.com/shuvo-paul/sitemonitor/pkg/monitor"
	"github.com/shuvo-paul/sitemonitor/web/templates"
	"github.com/stretchr/testify/assert"
)

// Mock SiteService
type mockSiteService struct {
	getAllFunc               func() ([]*monitor.Site, error)
	getByIDFunc              func(id int) (*monitor.Site, error)
	createFunc               func(url string, interval time.Duration) (*monitor.Site, error)
	updateFunc               func(site *monitor.Site) (*monitor.Site, error)
	deleteFunc               func(id int) error
	initializeMonitoringFunc func() error
}

func (m *mockSiteService) GetAll() ([]*monitor.Site, error) {
	return m.getAllFunc()
}

func (m *mockSiteService) GetByID(id int) (*monitor.Site, error) {
	return m.getByIDFunc(id)
}

func (m *mockSiteService) Create(url string, interval time.Duration) (*monitor.Site, error) {
	return m.createFunc(url, interval)
}

func (m *mockSiteService) Update(site *monitor.Site) (*monitor.Site, error) {
	return m.updateFunc(site)
}

func (m *mockSiteService) Delete(id int) error {
	return m.deleteFunc(id)
}

func (m *mockSiteService) InitializeMonitoring() error {
	if m.initializeMonitoringFunc != nil {
		return m.initializeMonitoringFunc()
	}
	return nil
}

func TestSiteController_List(t *testing.T) {
	mockService := &mockSiteService{
		getAllFunc: func() ([]*monitor.Site, error) {
			return []*monitor.Site{{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}}, nil
		},
		initializeMonitoringFunc: func() error { return nil },
	}

	controller := NewSiteController(mockService, &testutil.MockFlashStore{})
	templateRenderer := renderer.New(templates.TemplateFS)
	controller.Template.List = templateRenderer.Parse("sites/list.html")

	req := httptest.NewRequest(http.MethodGet, "/sites", nil)
	w := httptest.NewRecorder()

	controller.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSiteController_Create(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockSiteService{
			initializeMonitoringFunc: func() error { return nil },
		}
		controller := NewSiteController(mockService, &testutil.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS)
		controller.Template.Create = templateRenderer.Parse("sites/create.html")

		req := httptest.NewRequest(http.MethodGet, "/sites/create", nil)
		w := httptest.NewRecorder()

		controller.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		mockService := &mockSiteService{
			createFunc: func(url string, interval time.Duration) (*monitor.Site, error) {
				return &monitor.Site{ID: 1, URL: url, Interval: interval}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		controller := NewSiteController(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/sites/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		controller.Create(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites", w.Header().Get("Location"))
	})
}

func TestSiteController_Edit(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockSiteService{
			getByIDFunc: func(id int) (*monitor.Site, error) {
				return &monitor.Site{ID: id, URL: "http://example.com", Interval: 60 * time.Second}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		controller := NewSiteController(mockService, &testutil.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS)
		controller.Template.Edit = templateRenderer.Parse("sites/edit.html")

		req := httptest.NewRequest(http.MethodGet, "/sites/1/edit", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		controller.Edit(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		site := &monitor.Site{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}
		mockService := &mockSiteService{
			getByIDFunc: func(id int) (*monitor.Site, error) {
				return site, nil
			},
			updateFunc: func(site *monitor.Site) (*monitor.Site, error) {
				return site, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		controller := NewSiteController(mockService, &testutil.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/sites/1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		controller.Edit(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites", w.Header().Get("Location"))
	})
}

func TestSiteController_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := &mockSiteService{
			deleteFunc: func(id int) error {
				return nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		controller := NewSiteController(mockService, &testutil.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/sites/1/delete", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		controller.Delete(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/sites", w.Header().Get("Location"))
	})

	t.Run("invalid ID", func(t *testing.T) {
		mockService := &mockSiteService{
			initializeMonitoringFunc: func() error { return nil },
		}
		controller := NewSiteController(mockService, &testutil.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/sites/invalid/delete", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()

		controller.Delete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
