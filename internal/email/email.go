package email

import (
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/config"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mailer interface {
	SetTo(string) error
	SetSubject(string) error
	SetBody(string) error
	SendEmail() error
}

var _ Mailer = (*MailService)(nil)

func NewEmailService(config *config.EmailConfig) (*MailService, error) {
	if config == nil {
		return nil, fmt.Errorf("email configuration cannot be nil")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port number: port must be between 1 and 65535")
	}

	server := mail.NewSMTPClient()

	server.Host = config.Host
	server.Port = config.Port
	server.Username = config.Username
	server.Password = config.Password
	server.Encryption = mail.EncryptionSTARTTLS

	return &MailService{
		server: server,
		mail:   NewEmail(config.From),
	}, nil
}

func NewEmail(from string) *mail.Email {
	mail := mail.NewMSG()

	if from != "" {
		mail.SetFrom(from)
	}

	return mail
}

type MailService struct {
	server *mail.SMTPServer
	mail   *mail.Email
}

func (e *MailService) SetTo(to string) error {
	if to == "" {
		return fmt.Errorf("recipient email cannot be empty")
	}
	e.mail.AddTo(to)
	return nil
}

func (e *MailService) SetSubject(subject string) error {
	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	e.mail.SetSubject(subject)
	return nil
}

func (e *MailService) SetBody(body string) error {
	if body == "" {
		return fmt.Errorf("email body cannot be empty")
	}
	e.mail.SetBody(mail.TextHTML, body)
	return nil
}

func (e *MailService) SendEmail() error {
	server, err := e.server.Connect()
	if err != nil {
		return err
	}
	defer server.Close()

	if err := e.mail.Send(server); err != nil {
		return err
	}
	return nil
}
