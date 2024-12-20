package notification

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockSender is a manual mock implementation of the Sender interface
type MockSender struct {
	messages []Message
	err      error
}

func NewMockSender(err error) *MockSender {
	return &MockSender{
		messages: make([]Message, 0),
		err:      err,
	}
}

func (m *MockSender) Send(message Message) error {
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, message)
	return nil
}

func TestNotifier_Send(t *testing.T) {
	tests := []struct {
		name      string
		sender    *MockSender
		event     Event
		wantError bool
	}{
		{
			name:   "successful send",
			sender: NewMockSender(nil),
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "Site is up",
				OccurredAt: time.Now(),
			},
			wantError: false,
		},
		{
			name:   "sender error",
			sender: NewMockSender(fmt.Errorf("send error")),
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "down",
				Message:    "Site is down",
				OccurredAt: time.Now(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier := NewNotifier("test", tt.sender)
			err := notifier.Send(tt.event)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tt.sender.messages, 1)
				assert.Equal(t, tt.event, tt.sender.messages[0].Event)
				assert.Equal(t, "test", tt.sender.messages[0].NotifierID)
			}
		})
	}
}

func TestNotificationHub_Send(t *testing.T) {
	tests := []struct {
		name           string
		senders        []*MockSender
		event          Event
		expectedErrors int
	}{
		{
			name: "all senders succeed",
			senders: []*MockSender{
				NewMockSender(nil),
				NewMockSender(nil),
			},
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "test message",
				OccurredAt: time.Now(),
			},
			expectedErrors: 0,
		},
		{
			name: "some senders fail",
			senders: []*MockSender{
				NewMockSender(nil),
				NewMockSender(fmt.Errorf("send error")),
				NewMockSender(nil),
			},
			event: Event{
				SiteURL:    "https://example.com",
				Status:     "up",
				Message:    "test message",
				OccurredAt: time.Now(),
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewNotificationHub()
			for i, sender := range tt.senders {
				hub.RegisterNotifier(NewNotifier(fmt.Sprintf("test%d", i), sender))
			}

			errors := hub.Notify(tt.event)
			assert.Len(t, errors, tt.expectedErrors)

			for i, sender := range tt.senders {
				if sender.err == nil {
					assert.Len(t, sender.messages, 1)
					assert.Equal(t, tt.event, sender.messages[0].Event)
					assert.Equal(t, fmt.Sprintf("test%d", i), sender.messages[0].NotifierID)
				} else {
					assert.Len(t, sender.messages, 0)
				}
			}
		})
	}
}
