package email

import (
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/config"
	"github.com/stretchr/testify/assert"
	mail "github.com/xhit/go-simple-mail/v2"
)

func TestNewEmailService(t *testing.T) {
	// Create test config
	emailConfig := &config.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "sender@example.com",
	}

	service, err := NewEmailService(emailConfig)

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
	service, err = NewEmailService(nil)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Equal(t, "email configuration cannot be nil", err.Error())

	// Test invalid port
	invalidPortConfig := &config.EmailConfig{
		Host:     "smtp.example.com",
		Port:     0, // Invalid port number
		Username: "test@example.com",
		Password: "password",
		From:     "sender@example.com",
	}
	service, err = NewEmailService(invalidPortConfig)
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Equal(t, "invalid port number: port must be between 1 and 65535", err.Error())
}

func TestNewEmail(t *testing.T) {
	// Test with from address
	email := NewEmail("sender@example.com")
	assert.NotNil(t, email)

	// Test with empty from address
	email = NewEmail("")
	assert.NotNil(t, email)
}

func TestEmailService_SetTo(t *testing.T) {
	service := &MailService{
		mail: NewEmail("sender@example.com"),
	}

	recipient := "recipient@example.com"
	service.SetTo(recipient)

	// Verify that the recipient was added
	assert.Contains(t, service.mail.GetRecipients(), recipient)
}

func TestEmailService_SetSubject(t *testing.T) {
	service := &MailService{
		mail: NewEmail("sender@example.com"),
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
		mail: NewEmail("sender@example.com"),
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
	// Create test config
	emailConfig := &config.EmailConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "sender@example.com",
	}

	// Create email service
	service, err := NewEmailService(emailConfig)
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
		mail: NewEmail("sender@example.com"),
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
	// Create test config with invalid host
	emailConfig := &config.EmailConfig{
		Host:     "invalid.host",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "sender@example.com",
	}

	service, err := NewEmailService(emailConfig)
	assert.NoError(t, err)

	service.SetTo("recipient@example.com")
	service.SetSubject("Test Subject")
	service.SetBody("<h1>Test Body</h1>")

	err = service.SendEmail()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dial tcp: lookup invalid.host")
}
