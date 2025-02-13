package email

import (
	"fmt"
	"os"
	"strconv"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Mailer interface {
	SetTo(string) error
	SetSubject(string) error
	SetBody(string) error
	SendEmail() error
}

var _ Mailer = (*MailService)(nil)

func NewEmailService() (*MailService, error) {
	server := mail.NewSMTPClient()

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	if host == "" || port == "" || username == "" || password == "" {
		return nil, fmt.Errorf("missing email config")

	}
	server.Host = host
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}
	server.Port = portNum
	server.Username = username
	server.Password = password
	server.Encryption = mail.EncryptionSTARTTLS

	return &MailService{
		server: server,
		mail:   NewEmail(),
	}, nil
}

func NewEmail() *mail.Email {
	from := os.Getenv("SMTP_EMAIL_FROM")

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
