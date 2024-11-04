package controllers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"

	"github.com/shuvo-paul/sitemonitor/models"
)

// Mock UserService
type mockUserService struct {
	createUserFunc   func(*models.User) (*models.User, error)
	authenticateFunc func(string, string) (*models.User, error)
}

func (m *mockUserService) CreateUser(user *models.User) (*models.User, error) {
	return m.createUserFunc(user)
}

func (m *mockUserService) Authenticate(email, password string) (*models.User, error) {
	return m.authenticateFunc(email, password)
}

// Mock SessionService
type mockSessionService struct {
	createSessionFunc   func(int) (*models.Session, string, error)
	deleteSessionFunc   func(int) error
	validateSessionFunc func(string) (*models.Session, error)
}

func (m *mockSessionService) CreateSession(userID int) (*models.Session, string, error) {
	return m.createSessionFunc(userID)
}

func (m *mockSessionService) DeleteSession(sessionID int) error {
	return m.deleteSessionFunc(sessionID)
}

func (m *mockSessionService) ValidateSession(token string) (*models.Session, error) {
	return m.validateSessionFunc(token)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		formData       url.Values
		mockUserFunc   func(*models.User) (*models.User, error)
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
			mockUserFunc: func(u *models.User) (*models.User, error) {
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

			controller := NewUserController(mockUser, mockSession)
			controller.Template.Register = template.Must(template.New("register").Parse("dummy"))

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
	tests := []struct {
		name           string
		formData       url.Values
		mockAuthFunc   func(string, string) (*models.User, error)
		mockSessFunc   func(int) (*models.Session, string, error)
		expectedStatus int
		expectedPath   string
	}{
		{
			name: "successful login",
			formData: url.Values{
				"email":    {"test@example.com"},
				"password": {"password123"},
			},
			mockAuthFunc: func(email, password string) (*models.User, error) {
				return &models.User{ID: 1, Email: email}, nil
			},
			mockSessFunc: func(userID int) (*models.Session, string, error) {
				return &models.Session{}, "session-token", nil
			},
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/",
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

			controller := NewUserController(mockUser, mockSession)
			controller.Template.Login = template.Must(template.New("login").Parse("dummy"))

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
