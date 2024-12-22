package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackNotifier implements the Notifier interface for Slack
type SlackNotifier struct {
	webhookURL string
	client     HTTPClient
}

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(webhookURL string, client HTTPClient) *SlackNotifier {
	if client == nil {
		client = http.DefaultClient
	}
	return &SlackNotifier{
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

// Send implements the Notifier interface for Slack
func (s *SlackNotifier) Send(event Event) error {
	color := "warning"
	if event.Status == "up" {
		color = "good"
	} else if event.Status == "down" {
		color = "danger"
	}

	msg := slackMessage{
		Text: fmt.Sprintf("Site Status Alert for %s", event.SiteURL),
		Attachments: []attachment{
			{
				Color: color,
				Fields: []field{
					{Title: "Site URL", Value: event.SiteURL, Short: true},
					{Title: "Status", Value: event.Status, Short: true},
					{Title: "Time", Value: event.OccurredAt.String(), Short: true},
					{Title: "Message", Value: event.Message, Short: false},
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
