package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Email EmailConfig
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func Load() (*Config, error) {
	emailConfig, err := loadEmailConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load email config: %v", err)
	}

	return &Config{
		Email: emailConfig,
	}, nil
}

func loadEmailConfig() (EmailConfig, error) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_EMAIL_FROM")

	if host == "" || port == "" || username == "" || password == "" || from == "" {
		return EmailConfig{}, fmt.Errorf("missing required email configuration")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return EmailConfig{}, fmt.Errorf("invalid port number: %v", err)
	}

	return EmailConfig{
		Host:     host,
		Port:     portNum,
		Username: username,
		Password: password,
		From:     from,
	}, nil
}
