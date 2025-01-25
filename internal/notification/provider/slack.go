package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	notification "github.com/shuvo-paul/uptimebot/internal/notification/core"
)

// SlackObserver implements the Observer interface for Slack notifications
type SlackObserver struct {
	webhookURL string
	client     HTTPClient
}

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewSlackObserver creates a new Slack observer
func NewSlackObserver(webhookURL string, client HTTPClient) *SlackObserver {
	if client == nil {
		client = http.DefaultClient
	}
	return &SlackObserver{
		webhookURL: webhookURL,
		client:     client,
	}
}

type slackMessage struct {
	Text        string       `json:"text"`
	Attachments []attachment `json:"attachments"`
}

type attachment struct {
	Color  string  `json:"color"`
	Fields []field `json:"fields"`
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Notify implements the Observer interface
func (s *SlackObserver) Notify(state notification.State) error {
	color := "warning"
	if state.Status == "up" {
		color = "good"
	} else if state.Status == "down" {
		color = "danger"
	}

	msg := slackMessage{
		Text: fmt.Sprintf("Status Update for %s", state.Name),
		Attachments: []attachment{
			{
				Color: color,
				Fields: []field{
					{Title: "Name", Value: state.Name, Short: true},
					{Title: "Status", Value: state.Status, Short: true},
					{Title: "Time", Value: state.UpdatedAt.String(), Short: true},
					{Title: "Message", Value: state.Message, Short: false},
				},
			},
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
