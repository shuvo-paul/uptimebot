package notification

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockHTTPClient is a manual mock implementation of HTTPClient
type MockHTTPClient struct {
	requests []*http.Request
	response *http.Response
	err      error
}

func NewMockHTTPClient(statusCode int, err error) *MockHTTPClient {
	return &MockHTTPClient{
		requests: make([]*http.Request, 0),
		response: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(nil),
		},
		err: err,
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)
	return m.response, m.err
}

func TestSlackNotifier_Send(t *testing.T) {
	tests := []struct {
		name        string
		event       Event
		statusCode  int
		err         error
		wantErr     bool
		checkFields bool
	}{
		{
			name: "successful send",
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "Site is up",
				OccurredAt: time.Now(),
			},
			statusCode:  http.StatusOK,
			err:         nil,
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "down status",
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "down",
				Message:    "Site is down",
				OccurredAt: time.Now(),
			},
			statusCode:  http.StatusOK,
			err:         nil,
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "api error",
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "Site is up",
				OccurredAt: time.Now(),
			},
			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("api error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockHTTPClient(tt.statusCode, tt.err)
			notifier := NewSlackNotifier("https://hooks.slack.com/test", mockClient)

			err := notifier.Send(tt.event)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, mockClient.requests, 1)

			req := mockClient.requests[0]
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

			if tt.checkFields {
				var msg slackMessage
				err := json.NewDecoder(req.Body).Decode(&msg)
				assert.NoError(t, err)

				assert.Contains(t, msg.Text, tt.event.SiteURL)
				assert.Len(t, msg.Attachments, 1)

				attachment := msg.Attachments[0]
				expectedColor := "warning"
				if tt.event.Status == "up" {
					expectedColor = "good"
				} else if tt.event.Status == "down" {
					expectedColor = "danger"
				}
				assert.Equal(t, expectedColor, attachment.Color)

				assert.Len(t, attachment.Fields, 4)
				fields := make(map[string]string)
				for _, f := range attachment.Fields {
					fields[f.Title] = f.Value
				}

				assert.Equal(t, tt.event.SiteURL, fields["Site URL"])
				assert.Equal(t, tt.event.Status, fields["Status"])
				assert.Equal(t, tt.event.Message, fields["Message"])
				assert.Contains(t, fields["Time"], tt.event.OccurredAt.String())
			}
		})
	}
}
