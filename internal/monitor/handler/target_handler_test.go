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
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
	"github.com/shuvo-paul/uptimebot/internal/monitor/service"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/internal/templates"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
	"github.com/stretchr/testify/assert"
)

// Mock TargetService
type mockTargetService struct {
	getAllFunc               func() ([]model.UserTarget, error)
	getByIDFunc              func(id, userID int) (model.UserTarget, error)
	createFunc               func(userID int, url string, interval time.Duration) (model.UserTarget, error)
	updateFunc               func(target model.UserTarget, userID int) (model.UserTarget, error)
	deleteFunc               func(id, userID int) error
	getAllByUserIDFunc       func(userID int) ([]model.UserTarget, error)
	initializeMonitoringFunc func() error
	toggleEnabledFunc        func(id, userID int) (model.UserTarget, error)
}

func (m *mockTargetService) GetAll() ([]model.UserTarget, error) {
	return m.getAllFunc()
}

func (m *mockTargetService) GetByID(id, userID int) (model.UserTarget, error) {
	return m.getByIDFunc(id, userID)
}

func (m *mockTargetService) Create(userID int, url string, interval time.Duration) (model.UserTarget, error) {
	return m.createFunc(userID, url, interval)
}

func (m *mockTargetService) Update(target model.UserTarget, userID int) (model.UserTarget, error) {
	return m.updateFunc(target, userID)
}

func (m *mockTargetService) Delete(id, userID int) error {
	return m.deleteFunc(id, userID)
}

func (m *mockTargetService) GetAllByUserID(userID int) ([]model.UserTarget, error) {
	return m.getAllByUserIDFunc(userID)
}

func (m *mockTargetService) InitializeMonitoring() error {
	if m.initializeMonitoringFunc != nil {
		return m.initializeMonitoringFunc()
	}
	return nil
}

func (m *mockTargetService) ToggleEnabled(id, userID int) (model.UserTarget, error) {
	return m.toggleEnabledFunc(id, userID)
}

func TestTargetHandler_List(t *testing.T) {
	mockFlashStore := flash.NewMockFlashStore()
	mockService := &mockTargetService{
		getAllByUserIDFunc: func(userID int) ([]model.UserTarget, error) {
			return []model.UserTarget{{
				UserID: userID,
				Target: &monitor.Target{ID: 1, URL: "http://example.com", Interval: 60 * time.Second},
			}}, nil
		},
		initializeMonitoringFunc: func() error { return nil },
	}

	handler := NewTargetHandler(mockService, &flash.MockFlashStore{})
	templateRenderer := renderer.New(templates.TemplateFS, mockFlashStore)
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
	mockFlashStore := flash.NewMockFlashStore()
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockTargetService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS, mockFlashStore)
		handler.Template.Create = templateRenderer.GetTemplate("pages:targets/create")

		req := httptest.NewRequest(http.MethodGet, "/targets/create", nil)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		mockService := &mockTargetService{
			createFunc: func(userID int, url string, interval time.Duration) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: userID,
					Target: &monitor.Target{ID: 1, URL: url, Interval: interval},
				}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

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
			createFunc: func(userID int, url string, interval time.Duration) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: userID,
					Target: &monitor.Target{ID: 1, URL: url, Interval: interval},
				}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

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
	mockFlashStore := flash.NewMockFlashStore()
	t.Run("GET request", func(t *testing.T) {
		mockService := &mockTargetService{
			getByIDFunc: func(id, userID int) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: userID,
					Target: &monitor.Target{ID: id, URL: "http://example.com", Interval: 60 * time.Second},
				}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})
		templateRenderer := renderer.New(templates.TemplateFS, mockFlashStore)
		handler.Template.Edit = templateRenderer.GetTemplate("pages:targets/edit")

		req := httptest.NewRequest(http.MethodGet, "/targets/1/edit", nil)
		req.SetPathValue("id", "1")

		// Add user to context
		user := &authModel.User{ID: 1}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Edit(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request - success", func(t *testing.T) {
		target := &monitor.Target{ID: 1, URL: "http://example.com", Interval: 60 * time.Second}
		mockService := &mockTargetService{
			getByIDFunc: func(id, userID int) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: userID,
					Target: target,
				}, nil
			},
			updateFunc: func(ut model.UserTarget, userID int) (model.UserTarget, error) {
				return ut, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

		form := url.Values{}
		form.Add("url", "http://example.com")
		form.Add("interval", "60")

		req := httptest.NewRequest(http.MethodPost, "/targets/1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")

		// Add user to context
		user := &authModel.User{ID: 1}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Edit(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})
}

func TestTargetHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := &mockTargetService{
			deleteFunc: func(id, userID int) error {
				return nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/1/delete", nil)
		req.SetPathValue("id", "1")

		// Add user to context
		user := &authModel.User{ID: 1}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})

	t.Run("invalid ID", func(t *testing.T) {
		mockService := &mockTargetService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/invalid/delete", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTargetHandler_ToggleEnabled(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		target := &monitor.Target{
			ID:       1,
			URL:      "http://example.com",
			Interval: 60 * time.Second,
			Enabled:  false,
		}
		mockService := &mockTargetService{
			getByIDFunc: func(id, userID int) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: userID,
					Target: target,
				}, nil
			},
			toggleEnabledFunc: func(id, userID int) (model.UserTarget, error) {
				target.Enabled = !target.Enabled
				return model.UserTarget{
					UserID: userID,
					Target: target,
				}, nil
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/1/toggle", nil)
		req.SetPathValue("id", "1")

		// Add user to context
		user := &authModel.User{ID: 1}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.ToggleEnabled(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})

	t.Run("unauthorized access", func(t *testing.T) {
		target := &monitor.Target{
			ID:       1,
			URL:      "http://example.com",
			Interval: 60 * time.Second,
			Enabled:  false,
		}
		mockService := &mockTargetService{
			getByIDFunc: func(id, userID int) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: 1, // Original owner's ID
					Target: target,
				}, nil
			},
			toggleEnabledFunc: func(id, userID int) (model.UserTarget, error) {
				return model.UserTarget{
					UserID: 1,
					Target: target,
				}, service.ErrUnauthorized
			},
			initializeMonitoringFunc: func() error { return nil },
		}

		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/1/toggle", nil)
		req.SetPathValue("id", "1")

		// Add different user to context
		user := &authModel.User{ID: 2}
		ctx := authService.WithUser(req.Context(), user)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.ToggleEnabled(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/targets", w.Header().Get("Location"))
	})

	t.Run("invalid ID", func(t *testing.T) {
		mockService := &mockTargetService{
			initializeMonitoringFunc: func() error { return nil },
		}
		handler := NewTargetHandler(mockService, &flash.MockFlashStore{})

		req := httptest.NewRequest(http.MethodPost, "/targets/invalid/toggle", nil)
		req.SetPathValue("id", "invalid")
		w := httptest.NewRecorder()

		handler.ToggleEnabled(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
