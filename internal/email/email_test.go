package email

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	mail "github.com/xhit/go-simple-mail/v2"
)

func TestNewEmailService(t *testing.T) {
	// Set up test environment variables
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "test@example.com")
	os.Setenv("SMTP_PASSWORD", "password")
	defer func() {
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USERNAME")
		os.Unsetenv("SMTP_PASSWORD")
	}()

	service, err := NewEmailService()

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.server)
	assert.NotNil(t, service.mail)
	assert.Equal(t, "smtp.example.com", service.server.Host)
	assert.Equal(t, 587, service.server.Port)
	assert.Equal(t, "test@example.com", service.server.Username)
	assert.Equal(t, "password", service.server.Password)
	assert.Equal(t, mail.EncryptionSTARTTLS, service.server.Encryption)

	// Test missing configuration
	os.Unsetenv("SMTP_HOST")
	service, err = NewEmailService()
	assert.Error(t, err)
	assert.Nil(t, service)

	// Test invalid port
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_PORT", "invalid")
	service, err = NewEmailService()
	assert.Error(t, err)
	assert.Nil(t, service)
}

func TestNewEmail(t *testing.T) {
	// Test with SMTP_EMAIL_FROM environment variable set
	os.Setenv("SMTP_EMAIL_FROM", "sender@example.com")
	defer os.Unsetenv("SMTP_EMAIL_FROM")

	email := NewEmail()
	assert.NotNil(t, email)

	// Test with no SMTP_EMAIL_FROM environment variable
	os.Unsetenv("SMTP_EMAIL_FROM")
	email = NewEmail()
	assert.NotNil(t, email)
}

func TestEmailService_SetTo(t *testing.T) {
	service := &MailService{
		mail: NewEmail(),
	}

	recipient := "recipient@example.com"
	service.SetTo(recipient)

	// Verify that the recipient was added
	assert.Contains(t, service.mail.GetRecipients(), recipient)
}

func TestEmailService_SetSubject(t *testing.T) {
	service := &MailService{
		mail: NewEmail(),
	}

	subject := "Test Subject"
	service.SetSubject(subject)

	// Since go-simple-mail doesn't expose a way to get the subject directly,
	// we can only verify that the method doesn't panic
	assert.NotPanics(t, func() {
		service.SetSubject(subject)
	})
}

func TestEmailService_SetBody(t *testing.T) {
	service := &MailService{
		mail: NewEmail(),
	}

	body := "<h1>Test Body</h1>"
	service.SetBody(body)

	// Since go-simple-mail doesn't expose a way to get the body directly,
	// we can only verify that the method doesn't panic
	assert.NotPanics(t, func() {
		service.SetBody(body)
	})
}

func TestEmailService_SendEmail(t *testing.T) {
	// Set up test environment variables
	os.Setenv("SMTP_HOST", "localhost")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "test@example.com")
	os.Setenv("SMTP_PASSWORD", "password")
	defer func() {
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USERNAME")
		os.Unsetenv("SMTP_PASSWORD")
	}()

	// Create email service
	service, err := NewEmailService()
	assert.NoError(t, err)

	// Set up email content
	service.SetTo("recipient@example.com")
	service.SetSubject("Test Subject")
	service.SetBody("<h1>Test Body</h1>")

	// Test sending email (this will fail since we're not actually connecting to an SMTP server)
	err = service.SendEmail()
	assert.Error(t, err) // Expect an error since we're not actually connecting to a server
}

func TestEmailService_SetToMultiple(t *testing.T) {
	service := &MailService{
		mail: NewEmail(),
	}

	recipients := []string{"recipient1@example.com", "recipient2@example.com"}
	for _, recipient := range recipients {
		service.SetTo(recipient)
	}

	// Verify that all recipients were added
	for _, recipient := range recipients {
		assert.Contains(t, service.mail.GetRecipients(), recipient)
	}
}

func TestEmailService_SendEmailWithConnectionError(t *testing.T) {
	// Set up test environment variables with invalid host to force connection error
	os.Setenv("SMTP_HOST", "invalid.host")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "test@example.com")
	os.Setenv("SMTP_PASSWORD", "password")
	defer func() {
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USERNAME")
		os.Unsetenv("SMTP_PASSWORD")
	}()

	service, err := NewEmailService()
	assert.NoError(t, err)

	service.SetTo("recipient@example.com")
	service.SetSubject("Test Subject")
	service.SetBody("<h1>Test Body</h1>")

	err = service.SendEmail()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dial tcp: lookup invalid.host")
}
