package service

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/auth/repository"
	"github.com/shuvo-paul/uptimebot/internal/email"
)

type TokenService struct {
	tokenRepo    repository.TokenRepositoryInterface
	emailService email.Mailer
	baseURL      string
	templates    map[model.TokenType]*template.Template
}

func NewTokenService(
	tokenRepo repository.TokenRepositoryInterface,
	emailService email.Mailer,
	baseURL string,
	templates map[model.TokenType]*template.Template,
) *TokenService {
	return &TokenService{
		tokenRepo:    tokenRepo,
		emailService: emailService,
		baseURL:      baseURL,
		templates:    templates,
	}
}

type TokenServiceInterface interface {
	createToken(userID int, tokenType model.TokenType, expiresIn time.Duration) (*model.Token, error)
	ValidateToken(token string, tokenType model.TokenType) (*model.Token, error)
	invalidateAndCreateNewToken(userID int, tokenType model.TokenType, expiresIn time.Duration) (*model.Token, error)
	SendToken(userID int, email string, tokenType model.TokenType, subject string, path string, expiresIn time.Duration) error
	MarkTokenAsUsed(tokenID int) error
}

// Ensure TokenService implements TokenServiceInterface
var _ TokenServiceInterface = (*TokenService)(nil)

func (s *TokenService) createToken(userID int, tokenType model.TokenType, expiresIn time.Duration) (*model.Token, error) {
	token := &model.Token{
		UserID:    userID,
		Token:     uuid.New().String(),
		Type:      tokenType,
		ExpiresAt: time.Now().Add(expiresIn),
		Used:      false,
	}

	return s.tokenRepo.SaveToken(token)
}

func (s *TokenService) ValidateToken(token string, tokenType model.TokenType) (*model.Token, error) {
	vToken, err := s.tokenRepo.GetTokenByValue(token)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	if vToken == nil {
		return nil, fmt.Errorf("token not found")
	}

	if err := vToken.ValidateType(tokenType); err != nil {
		return nil, err
	}

	if !vToken.IsValid() {
		return nil, fmt.Errorf("token is no longer valid")
	}

	return vToken, nil
}

// MarkTokenAsUsed marks a token as used in the repository
func (s *TokenService) MarkTokenAsUsed(tokenID int) error {
	if err := s.tokenRepo.MarkTokenUsed(tokenID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	return nil
}

func (s *TokenService) invalidateAndCreateNewToken(userID int, tokenType model.TokenType, expiresIn time.Duration) (*model.Token, error) {
	if err := s.tokenRepo.InvalidateExistingTokens(userID, tokenType); err != nil {
		return nil, fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	return s.createToken(userID, tokenType, expiresIn)
}

// emailParams contains all necessary parameters for sending token-based emails
type emailParams struct {
	// UserID is the unique identifier of the user
	UserID int `validate:"required"`
	// Email is the recipient's email address
	Email string `validate:"required,email"`
	// TokenType specifies the type of token (e.g., email verification, password reset)
	TokenType model.TokenType `validate:"required"`
	// Subject is the email subject line
	Subject string `validate:"required"`
	// Path is the URL path component for the token link
	Path string `validate:"required"`
	// Template is the HTML template used for rendering emails
	Template *template.Template `validate:"required"`
	// ExpiresIn is the duration until the token expires
	ExpiresIn time.Duration `validate:"required"`
}

// Validate checks if all required fields are properly set
func (p *emailParams) Validate() error {
	if p.UserID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	if p.Email == "" {
		return fmt.Errorf("email is required")
	}
	if p.TokenType == "" {
		return fmt.Errorf("token type is required")
	}

	if p.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if p.Path == "" {
		return fmt.Errorf("path is required")
	}

	if p.ExpiresIn <= 0 {
		return fmt.Errorf("expiration duration must be positive")
	}
	return nil
}

func (s *TokenService) sendTokenEmail(params emailParams) error {
	// Create a new token
	token, err := s.invalidateAndCreateNewToken(params.UserID, params.TokenType, params.ExpiresIn)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	// Generate token link
	tokenLink := fmt.Sprintf("%s/%s?token=%s", s.baseURL, params.Path, token.Token)

	// Send email using the email service
	if err := s.emailService.SetTo(params.Email); err != nil {
		return fmt.Errorf("failed to set email recipient: %w", err)
	}

	if err := s.emailService.SetSubject(params.Subject); err != nil {
		return fmt.Errorf("failed to set email subject: %w", err)
	}

	// Execute the email template with the token link
	var buf bytes.Buffer
	data := struct {
		TokenLink string
	}{
		TokenLink: tokenLink,
	}

	if err := params.Template.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	if err := s.emailService.SetBody(buf.String()); err != nil {
		return fmt.Errorf("failed to set email body: %w", err)
	}

	if err := s.emailService.SendEmail(); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *TokenService) SendToken(
	userID int,
	email string,
	tokenType model.TokenType,
	subject string,
	path string,
	expiresIn time.Duration,
) error {
	template, ok := s.templates[tokenType]
	if !ok {
		return fmt.Errorf("template not found for token type: %s", tokenType)
	}

	return s.sendTokenEmail(emailParams{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		Subject:   subject,
		Path:      path,
		Template:  template,
		ExpiresIn: expiresIn,
	})
}
