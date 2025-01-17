package models

import (
	"encoding/json"
	"fmt"
)

// NotifierType represents the type of notifier
type NotifierType string

const (
	NotifierTypeSlack NotifierType = "slack"
	NotifierTypeEmail NotifierType = "email"
)

// NotifierConfig represents the configuration for a notifier
type NotifierConfig struct {
	Type   NotifierType    `db:"type"`
	Config json.RawMessage `db:"config"`
}

// FromString converts a JSON string to NotifierConfig
func (n *NotifierConfig) FromString(configString string) error {
	err := json.Unmarshal([]byte(configString), n)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// ToString converts NotifierConfig to a JSON string
func (n *NotifierConfig) ToString() (string, error) {
	configBytes, err := json.Marshal(n)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(configBytes), nil
}

// Scan implements sql.Scanner interface
func (n *NotifierConfig) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, n)
}

// Value implements driver.Valuer interface
func (n *NotifierConfig) Value() (interface{}, error) {
	return json.Marshal(n)
}

// Notifier represents a notification channel configuration
type Notifier struct {
	ID     int64           `db:"id"`
	SiteId int             `db:"site_id"`
	Config *NotifierConfig `db:"config"`
}

// SlackConfig represents Slack notifier configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// EmailConfig represents email notifier configuration
type EmailConfig struct {
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	From       string   `json:"from"`
	Recipients []string `json:"recipients"`
}

// GetSlackConfig parses and returns Slack configuration
func (n *NotifierConfig) GetSlackConfig() (*SlackConfig, error) {
	if n.Type != NotifierTypeSlack {
		return nil, nil
	}
	var config SlackConfig
	if err := json.Unmarshal(n.Config, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetEmailConfig parses and returns email configuration
func (n *NotifierConfig) GetEmailConfig() (*EmailConfig, error) {
	if n.Type != NotifierTypeEmail {
		return nil, nil
	}
	var config EmailConfig
	if err := json.Unmarshal(n.Config, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
