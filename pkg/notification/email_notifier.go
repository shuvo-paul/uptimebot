package notification

import (
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPClient interface for sending emails
type SMTPClient interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// defaultSMTPClient implements SMTPClient using smtp.SendMail
type defaultSMTPClient struct{}

func (c *defaultSMTPClient) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, msg)
}

// SMTPConfig holds the SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// EmailNotifier implements the Notifier interface for email
type EmailNotifier struct {
	config     SMTPConfig
	recipients []string
	client     SMTPClient
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config SMTPConfig, recipients []string, client SMTPClient) *EmailNotifier {
	if client == nil {
		client = &defaultSMTPClient{}
	}
	return &EmailNotifier{
		config:     config,
		recipients: recipients,
		client:     client,
	}
}

// AddRecipients adds new email recipients
func (e *EmailNotifier) AddRecipients(emails ...string) {
	e.recipients = append(e.recipients, emails...)
}

// SetRecipients replaces all existing recipients with new ones
func (e *EmailNotifier) SetRecipients(emails []string) {
	e.recipients = emails
}

// Send implements the Notifier interface for email
func (e *EmailNotifier) Send(event Event) error {
	if len(e.recipients) == 0 {
		return fmt.Errorf("no recipients configured for email notification")
	}

	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	subject := fmt.Sprintf("Site Status Alert: %s is %s", event.SiteURL, event.Status)
	body := fmt.Sprintf(`
Site Status Alert

Site URL: %s
Status: %s
Time: %s

Message: %s
`, event.SiteURL, event.Status, event.OccurredAt.String(), event.Message)

	emailMsg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", e.config.From, strings.Join(e.recipients, ","), subject, body)

	return e.client.SendMail(addr, auth, e.config.From, e.recipients, []byte(emailMsg))
}
