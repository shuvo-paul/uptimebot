package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/internal/templates"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
)

// Mock UserService
type mockUserService struct {
	createUserFunc   func(*model.User) (*model.User, error)
	authenticateFunc func(string, string) (*model.User, error)
	getUserByIdFunc  func(id int) (*model.User, error)
}

func (m *mockUserService) CreateUser(user *model.User) (*model.User, error) {
	return m.createUserFunc(user)
}

func (m *mockUserService) Authenticate(email, password string) (*model.User, error) {
	return m.authenticateFunc(email, password)
}

func (m *mockUserService) GetUserByID(id int) (*model.User, error) {
	return m.getUserByIdFunc(id)
}

// Mock SessionService
type mockSessionService struct {
	createSessionFunc   func(int) (*model.Session, string, error)
	deleteSessionFunc   func(string) error
	validateSessionFunc func(string) (*model.Session, error)
}

func (m *mockSessionService) CreateSession(userID int) (*model.Session, string, error) {
	return m.createSessionFunc(userID)
}

func (m *mockSessionService) DeleteSession(token string) error {
	return m.deleteSessionFunc(token)
}

func (m *mockSessionService) ValidateSession(token string) (*model.Session, error) {
	return m.validateSessionFunc(token)
}

func TestRegister(t *testing.T) {
	templateRenderer := renderer.New(templates.TemplateFS)

	tests := []struct {
		name           string
		formData       url.Values
		mockUserFunc   func(*model.User) (*model.User, error)
		expectedStatus int
		expectedPath   string
	}{
		{
			name: "successful registration",
			formData: url.Values{
				"username": {"testuser"},
				"email":    {"test@example.com"},
				"password": {"password123"},
			},
			mockUserFunc: func(u *model.User) (*model.User, error) {
				u.ID = 1
				return u, nil
			},
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/login",
		},
		// Add more test cases for validation errors, service errors, etc.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUser := &mockUserService{
				createUserFunc: tt.mockUserFunc,
			}
			mockSession := &mockSessionService{}

			mockFlash := &testutil.MockFlashStore{}

			controller := NewUserHandler(mockUser, mockSession, mockFlash)
			controller.Template.Register = templateRenderer.Parse("register.html")

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			controller.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, w.Code)
			}

			if location := w.Header().Get("Location"); location != tt.expectedPath {
				t.Errorf("expected redirect to %s; got %s", tt.expectedPath, location)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	templateRenderer := renderer.New(templates.TemplateFS)

	tests := []struct {
		name           string
		formData       url.Values
		mockAuthFunc   func(string, string) (*model.User, error)
		mockSessFunc   func(int) (*model.Session, string, error)
		expectedStatus int
		expectedPath   string
	}{
		{
			name: "successful login",
			formData: url.Values{
				"email":    {"test@example.com"},
				"password": {"password123"},
			},
			mockAuthFunc: func(email, password string) (*model.User, error) {
				return &model.User{ID: 1, Email: email}, nil
			},
			mockSessFunc: func(userID int) (*model.Session, string, error) {
				return &model.Session{}, "session-token", nil
			},
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/sites",
		},
		// Add more test cases for invalid credentials, service errors, etc.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUser := &mockUserService{
				authenticateFunc: tt.mockAuthFunc,
			}
			mockSession := &mockSessionService{
				createSessionFunc: tt.mockSessFunc,
			}

			mockFlash := &testutil.MockFlashStore{}

			controller := NewUserHandler(mockUser, mockSession, mockFlash)
			controller.Template.Login = templateRenderer.Parse("login.html")

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			controller.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, w.Code)
			}

			if location := w.Header().Get("Location"); location != tt.expectedPath {
				t.Errorf("expected redirect to %s; got %s", tt.expectedPath, location)
			}
		})
	}
}
