package notification

import (
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockSMTPClient is a manual mock implementation of SMTPClient
type MockSMTPClient struct {
	calls []struct {
		addr string
		auth smtp.Auth
		from string
		to   []string
		msg  []byte
	}
	err error
}

func NewMockSMTPClient(err error) *MockSMTPClient {
	return &MockSMTPClient{
		calls: make([]struct {
			addr string
			auth smtp.Auth
			from string
			to   []string
			msg  []byte
		}, 0),
		err: err,
	}
}

func (m *MockSMTPClient) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.calls = append(m.calls, struct {
		addr string
		auth smtp.Auth
		from string
		to   []string
		msg  []byte
	}{addr, a, from, to, msg})
	return m.err
}

func TestEmailNotifier_Send(t *testing.T) {
	tests := []struct {
		name       string
		config     SMTPConfig
		recipients []string
		event      Event
		err        error
		wantErr    bool
	}{
		{
			name: "successful send",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test",
				Password: "pass",
				From:     "from@example.com",
			},
			recipients: []string{"test@example.com"},
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "Site is up",
				OccurredAt: time.Now(),
			},
			err:     nil,
			wantErr: false,
		},
		{
			name: "no recipients",
			config: SMTPConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test",
				Password: "pass",
				From:     "from@example.com",
			},
			recipients: []string{},
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "Site is up",
				OccurredAt: time.Now(),
			},
			err:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockSMTPClient(tt.err)
			notifier := NewEmailNotifier(tt.config, tt.recipients, mockClient)

			err := notifier.Send(tt.event)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Len(t, mockClient.calls, 0)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, mockClient.calls, 1)

			call := mockClient.calls[0]
			expectedAddr := tt.config.Host + ":" + "587"
			assert.Equal(t, expectedAddr, call.addr)
			assert.Equal(t, tt.config.From, call.from)
			assert.Equal(t, tt.recipients, call.to)

			msgStr := string(call.msg)
			assert.Contains(t, msgStr, "From: "+tt.config.From)
			assert.Contains(t, msgStr, "To: "+strings.Join(tt.recipients, ","))
			assert.Contains(t, msgStr, tt.event.SiteURL)
			assert.Contains(t, msgStr, tt.event.Status)
			assert.Contains(t, msgStr, tt.event.Message)
			assert.Contains(t, msgStr, tt.event.OccurredAt.String())
		})
	}
}

func TestEmailNotifier_AddRecipients(t *testing.T) {
	notifier := NewEmailNotifier(SMTPConfig{}, []string{"initial@example.com"}, nil)

	// Test adding single recipient
	notifier.AddRecipients("test1@example.com")
	assert.Len(t, notifier.recipients, 2)
	assert.Contains(t, notifier.recipients, "test1@example.com")

	// Test adding multiple recipients
	notifier.AddRecipients("test2@example.com", "test3@example.com")
	assert.Len(t, notifier.recipients, 4)
	assert.Contains(t, notifier.recipients, "test2@example.com")
	assert.Contains(t, notifier.recipients, "test3@example.com")
}

func TestEmailNotifier_SetRecipients(t *testing.T) {
	notifier := NewEmailNotifier(SMTPConfig{}, []string{"old@example.com"}, nil)

	// Test setting new recipients
	newRecipients := []string{"new1@example.com", "new2@example.com"}
	notifier.SetRecipients(newRecipients)

	assert.Len(t, notifier.recipients, 2)
	assert.Equal(t, newRecipients, notifier.recipients)
	assert.NotContains(t, notifier.recipients, "old@example.com")
}
