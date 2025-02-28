package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/renderer"
	"github.com/shuvo-paul/uptimebot/internal/templates"
	"github.com/shuvo-paul/uptimebot/pkg/flash"
)

// Mock AuthService
type mockAuthService struct {
	createUserFunc     func(*model.User) (*model.User, error)
	authenticateFunc   func(string, string) (*model.User, error)
	getUserByIdFunc    func(id int) (*model.User, error)
	getUserByEmailFunc func(email string) (*model.User, error)
	sendTokenFunc      func(int, string, model.TokenType) error
	updatePasswordFunc func(int, string) error
	resetPasswordFunc  func(string, string) error
	validateTokenFunc  func(string, model.TokenType) (*model.Token, error)
}

func (m *mockAuthService) CreateUser(user *model.User) (*model.User, error) {
	return m.createUserFunc(user)
}

func (m *mockAuthService) Authenticate(email, password string) (*model.User, error) {
	return m.authenticateFunc(email, password)
}

func (m *mockAuthService) GetUserByID(id int) (*model.User, error) {
	return m.getUserByIdFunc(id)
}

func (m *mockAuthService) GetUserByEmail(email string) (*model.User, error) {
	return m.getUserByEmailFunc(email)
}

func (m *mockAuthService) ValidateToken(token string, tokenType model.TokenType) (*model.Token, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token, tokenType)
	}
	return nil, nil
}

func (m *mockAuthService) VerifyEmail(token string) error {
	return nil
}

func (m *mockAuthService) SendToken(id int, email string, tokenType model.TokenType) error {
	return m.sendTokenFunc(id, email, tokenType)
}

func (m *mockAuthService) UpdatePassword(userID int, newPassword string) error {
	if m.updatePasswordFunc != nil {
		return m.updatePasswordFunc(userID, newPassword)
	}
	return nil
}

func (m *mockAuthService) ResetPassword(token string, newPassword string) error {
	if m.resetPasswordFunc != nil {
		return m.resetPasswordFunc(token, newPassword)
	}
	return nil
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
	mockFlashStore := flash.NewFlashStore()
	templateRenderer := renderer.New(templates.TemplateFS, mockFlashStore)

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
			mockUser := &mockAuthService{
				createUserFunc: tt.mockUserFunc,
			}
			mockSession := &mockSessionService{}

			mockFlash := flash.NewFlashStore()

			controller := NewAuthHandler(mockUser, mockSession, mockFlash)
			controller.Template.Register = templateRenderer.GetTemplate("pages:register")

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
	mockFlashStore := flash.NewFlashStore()
	templateRenderer := renderer.New(templates.TemplateFS, mockFlashStore)

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
			expectedPath:   "/targets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUser := &mockAuthService{
				authenticateFunc: tt.mockAuthFunc,
			}
			mockSession := &mockSessionService{
				createSessionFunc: tt.mockSessFunc,
			}

			mockFlash := &flash.MockFlashStore{}

			controller := NewAuthHandler(mockUser, mockSession, mockFlash)
			controller.Template.Login = templateRenderer.GetTemplate("pages:login")

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

func TestSendVerificationMail(t *testing.T) {
	tests := []struct {
		name              string
		queryParams       string
		mockGetUserFunc   func(int) (*model.User, error)
		mockSendEmailFunc func(int, string) error
	}{
		{
			name:        "successful verification email send",
			queryParams: "user_id=1",
			mockGetUserFunc: func(id int) (*model.User, error) {
				return &model.User{ID: 1, Email: "test@example.com"}, nil
			},
			mockSendEmailFunc: func(id int, email string) error {
				return nil
			},
		},
		{
			name:        "missing user ID",
			queryParams: "",
		},
		{
			name:        "invalid user ID",
			queryParams: "user_id=invalid",
		},
		{
			name:        "user not found",
			queryParams: "user_id=999",
			mockGetUserFunc: func(id int) (*model.User, error) {
				return nil, fmt.Errorf("user not found")
			},
		},
		{
			name:        "email sending fails",
			queryParams: "user_id=1",
			mockGetUserFunc: func(id int) (*model.User, error) {
				return &model.User{ID: 1, Email: "test@example.com"}, nil
			},
			mockSendEmailFunc: func(id int, email string) error {
				return fmt.Errorf("failed to send email")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUser := &mockAuthService{
				getUserByIdFunc: tt.mockGetUserFunc,
				sendTokenFunc: func(id int, email string, tokenType model.TokenType) error {
					if tt.mockSendEmailFunc != nil {
						return tt.mockSendEmailFunc(id, email)
					}
					return nil
				},
			}
			mockSession := &mockSessionService{}
			mockFlash := &flash.MockFlashStore{}

			handler := NewAuthHandler(mockUser, mockSession, mockFlash)

			req := httptest.NewRequest(http.MethodGet, "/verify-email?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.SendVerificationEmail(w, req)

		})
	}
}
