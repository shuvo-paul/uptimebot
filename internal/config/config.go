package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Email    EmailConfig
	Database DatabaseConfig
	BaseURL  string
	Port     int
}

type DatabaseConfig struct {
	URL   string
	Token string
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

	dbConfig, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %v", err)
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("missing required base URL configuration")
	}

	// Get port from environment variable with default value of 8080
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		portNum, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %v", err)
		}
		port = portNum
	}

	return &Config{
		Email:    emailConfig,
		Database: dbConfig,
		BaseURL:  baseURL,
		Port:     port,
	}, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")

	if url == "" || token == "" {
		return DatabaseConfig{}, fmt.Errorf("missing required database configuration")
	}

	return DatabaseConfig{
		URL:   url,
		Token: token,
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
