package model

import (
	"encoding/json"
)

// NotifierType represents the type of notifier
type NotifierType string

const (
	NotifierTypeSlack NotifierType = "slack"
	NotifierTypeEmail NotifierType = "email"
)

// Notifier represents a notification channel configuration
type Notifier struct {
	ID       int             `db:"id"`
	TargetId int             `db:"target_id"`
	Type     NotifierType    `db:"type"`
	Config   json.RawMessage `db:"config"`
}

// SlackConfig represents Slack notifier configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// EmailConfig represents email notifier configuration
type EmailConfig struct {
	Recipients []string `json:"recipients"`
}

// GetSlackConfig parses and returns Slack configuration
func (n *Notifier) GetSlackConfig() (*SlackConfig, error) {
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
func (n *Notifier) GetEmailConfig() (*EmailConfig, error) {
	if n.Type != NotifierTypeEmail {
		return nil, nil
	}
	var config EmailConfig
	if err := json.Unmarshal(n.Config, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
