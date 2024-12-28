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

func TestSlackObserver_Notify(t *testing.T) {
	tests := []struct {
		name        string
		state       State
		statusCode  int
		err         error
		wantErr     bool
		checkFields bool
	}{
		{
			name: "successful notification",
			state: State{
				Name:      "test-system",
				Status:    "up",
				Message:   "System is up",
				UpdatedAt: time.Now(),
			},
			statusCode:  http.StatusOK,
			err:         nil,
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "down status",
			state: State{
				Name:      "test-system",
				Status:    "down",
				Message:   "System is down",
				UpdatedAt: time.Now(),
			},
			statusCode:  http.StatusOK,
			err:         nil,
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "api error",
			state: State{
				Name:      "test-system",
				Status:    "up",
				Message:   "System is up",
				UpdatedAt: time.Now(),
			},
			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("api error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockHTTPClient(tt.statusCode, tt.err)
			observer := NewSlackObserver("https://hooks.slack.com/test", mockClient)

			err := observer.Notify(tt.state)

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

				assert.Contains(t, msg.Text, tt.state.Name)
				assert.Len(t, msg.Attachments, 1)

				attachment := msg.Attachments[0]
				expectedColor := "warning"
				if tt.state.Status == "up" {
					expectedColor = "good"
				} else if tt.state.Status == "down" {
					expectedColor = "danger"
				}
				assert.Equal(t, expectedColor, attachment.Color)

				assert.Len(t, attachment.Fields, 4)
				fields := make(map[string]string)
				for _, f := range attachment.Fields {
					fields[f.Title] = f.Value
				}

				assert.Equal(t, tt.state.Name, fields["Name"])
				assert.Equal(t, tt.state.Status, fields["Status"])
				assert.Equal(t, tt.state.Message, fields["Message"])
				assert.Contains(t, fields["Time"], tt.state.UpdatedAt.String())
			}
		})
	}
}
