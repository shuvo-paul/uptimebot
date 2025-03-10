package service

import (
	"fmt"
	"html/template"
	"testing"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	mockRepo "github.com/shuvo-paul/uptimebot/internal/auth/repository"
	mockEmail "github.com/shuvo-paul/uptimebot/internal/email"
	"github.com/stretchr/testify/assert"
)

func TestTokenService_CreateToken(t *testing.T) {
	tokenRepo := &mockRepo.TokenRepositoryMock{}
	emailService := &mockEmail.EmailServiceMock{}
	baseURL := "http://localhost:8080"

	templates := map[model.TokenType]*template.Template{
		model.TokenTypeEmailVerification: template.Must(template.New("email_verification").Parse("{{.TokenLink}}")),
		model.TokenTypePasswordReset:     template.Must(template.New("password_reset").Parse("{{.TokenLink}}")),
	}

	service := NewTokenService(tokenRepo, emailService, baseURL, templates)

	tests := []struct {
		name      string
		userID    int
		tokenType model.TokenType
		expiresIn time.Duration
		wantErr   bool
	}{
		{
			name:      "successful token creation",
			userID:    1,
			tokenType: model.TokenTypeEmailVerification,
			expiresIn: 24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "repository error",
			userID:    2,
			tokenType: model.TokenTypePasswordReset,
			expiresIn: time.Hour,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				tokenRepo.SaveTokenFunc = func(token *model.Token) (*model.Token, error) {
					return nil, fmt.Errorf("mock error")
				}
			} else {
				tokenRepo.SaveTokenFunc = func(token *model.Token) (*model.Token, error) {
					token.ID = 1
					return token, nil
				}
			}

			token, err := service.createToken(tt.userID, tt.tokenType, tt.expiresIn)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, tt.userID, token.UserID)
				assert.Equal(t, tt.tokenType, token.Type)
				assert.False(t, token.Used)
				assert.WithinDuration(t, time.Now().Add(tt.expiresIn), token.ExpiresAt, time.Second)
			}
		})
	}
}

func TestTokenService_ValidateToken(t *testing.T) {
	tokenRepo := &mockRepo.TokenRepositoryMock{}
	emailService := &mockEmail.EmailServiceMock{}
	baseURL := "http://localhost:8080"

	templates := map[model.TokenType]*template.Template{
		model.TokenTypeEmailVerification: template.Must(template.New("email_verification").Parse("{{.TokenLink}}")),
		model.TokenTypePasswordReset:     template.Must(template.New("password_reset").Parse("{{.TokenLink}}")),
	}
	service := NewTokenService(tokenRepo, emailService, baseURL, templates)

	tests := []struct {
		name      string
		token     string
		tokenType model.TokenType
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "valid token",
			token:     "valid-token",
			tokenType: model.TokenTypeEmailVerification,
			setupMock: func() {
				tokenRepo.GetTokenByValueFunc = func(token string) (*model.Token, error) {
					return &model.Token{
						ID:        1,
						Token:     "valid-token",
						Type:      model.TokenTypeEmailVerification,
						ExpiresAt: time.Now().Add(time.Hour),
						Used:      false,
					}, nil
				}
				tokenRepo.MarkTokenUsedFunc = func(tokenID int) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:      "token not found",
			token:     "invalid-token",
			tokenType: model.TokenTypeEmailVerification,
			setupMock: func() {
				tokenRepo.GetTokenByValueFunc = func(token string) (*model.Token, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
		{
			name:      "expired token",
			token:     "expired-token",
			tokenType: model.TokenTypeEmailVerification,
			setupMock: func() {
				tokenRepo.GetTokenByValueFunc = func(token string) (*model.Token, error) {
					return &model.Token{
						ID:        1,
						Token:     "expired-token",
						Type:      model.TokenTypeEmailVerification,
						ExpiresAt: time.Now().Add(-time.Hour),
						Used:      false,
					}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			token, err := service.ValidateToken(tt.token, tt.tokenType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, tt.token, token.Token)
				assert.Equal(t, tt.tokenType, token.Type)
			}
		})
	}
}

func TestTokenService_SendToken(t *testing.T) {
	tokenRepo := &mockRepo.TokenRepositoryMock{}
	emailService := &mockEmail.EmailServiceMock{}
	baseURL := "http://localhost:8080"

	templates := map[model.TokenType]*template.Template{
		model.TokenTypeEmailVerification: template.Must(template.New("email_verification").Parse("{{.TokenLink}}")),
		model.TokenTypePasswordReset:     template.Must(template.New("password_reset").Parse("{{.TokenLink}}")),
	}
	s := NewTokenService(tokenRepo, emailService, baseURL, templates)

	tests := []struct {
		name      string
		userID    int
		email     string
		tokenType model.TokenType
		subject   string
		path      string
		expiresIn time.Duration
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "successful email verification token sending",
			userID:    1,
			email:     "test@example.com",
			tokenType: model.TokenTypeEmailVerification,
			subject:   "Verify Your Email Address",
			path:      "verify-email",
			expiresIn: 24 * time.Hour,
			setupMock: func() {
				tokenRepo.InvalidateExistingTokensFunc = func(userID int, tokenType model.TokenType) error {
					return nil
				}
				tokenRepo.SaveTokenFunc = func(token *model.Token) (*model.Token, error) {
					token.ID = 1
					return token, nil
				}
				emailService.SetToFunc = func(to string) error { return nil }
				emailService.SetSubjectFunc = func(subject string) error { return nil }
				emailService.SetBodyFunc = func(body string) error { return nil }
				emailService.SendEmailFunc = func() error { return nil }
			},
			wantErr: false,
		},
		{
			name:      "successful password reset token sending",
			userID:    1,
			email:     "test@example.com",
			tokenType: model.TokenTypePasswordReset,
			subject:   "Reset Your Password",
			path:      "reset-password",
			expiresIn: time.Hour,
			setupMock: func() {
				tokenRepo.InvalidateExistingTokensFunc = func(userID int, tokenType model.TokenType) error {
					return nil
				}
				tokenRepo.SaveTokenFunc = func(token *model.Token) (*model.Token, error) {
					token.ID = 1
					return token, nil
				}
				emailService.SetToFunc = func(to string) error { return nil }
				emailService.SetSubjectFunc = func(subject string) error { return nil }
				emailService.SetBodyFunc = func(body string) error { return nil }
				emailService.SendEmailFunc = func() error { return nil }
			},
			wantErr: false,
		},
		{
			name:      "email service error",
			userID:    1,
			email:     "test@example.com",
			tokenType: model.TokenTypeEmailVerification,
			subject:   "Verify Your Email Address",
			path:      "verify-email",
			expiresIn: 24 * time.Hour,
			setupMock: func() {
				tokenRepo.InvalidateExistingTokensFunc = func(userID int, tokenType model.TokenType) error {
					return nil
				}
				tokenRepo.SaveTokenFunc = func(token *model.Token) (*model.Token, error) {
					token.ID = 1
					return token, nil
				}
				emailService.SetToFunc = func(to string) error { return nil }
				emailService.SetSubjectFunc = func(subject string) error { return nil }
				emailService.SetBodyFunc = func(body string) error { return nil }
				emailService.SendEmailFunc = func() error { return fmt.Errorf("failed to send email") }
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			params := emailParams{
				UserID:    tt.userID,
				Email:     tt.email,
				TokenType: tt.tokenType,
				Subject:   tt.subject,
				Path:      tt.path,
				Template:  template.Must(template.New("test").Parse(`{{.TokenLink}}`)),
				ExpiresIn: tt.expiresIn,
			}

			err := s.sendTokenEmail(params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenService_MarkTokenAsUsed(t *testing.T) {
	tokenRepo := &mockRepo.TokenRepositoryMock{}
	emailService := &mockEmail.EmailServiceMock{}
	baseURL := "http://localhost:8080"
	tmpl := template.Must(template.New("test").Parse("{{.TokenLink}}"))

	service := NewTokenService(tokenRepo, emailService, baseURL, map[model.TokenType]*template.Template{
		model.TokenTypeEmailVerification: tmpl,
		model.TokenTypePasswordReset:     tmpl,
	})

	tests := []struct {
		name      string
		tokenID   int
		setupMock func()
		wantErr   bool
	}{
		{
			name:    "successful token marking",
			tokenID: 1,
			setupMock: func() {
				tokenRepo.MarkTokenUsedFunc = func(tokenID int) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "repository error",
			tokenID: 2,
			setupMock: func() {
				tokenRepo.MarkTokenUsedFunc = func(tokenID int) error {
					return fmt.Errorf("failed to mark token as used")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := service.MarkTokenAsUsed(tt.tokenID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
