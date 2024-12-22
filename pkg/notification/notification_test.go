package notification

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockNotifier is a mock implementation of the Notifier interface
type MockNotifier struct {
	events []Event
	err    error
}

func NewMockNotifier(err error) *MockNotifier {
	return &MockNotifier{
		events: make([]Event, 0),
		err:    err,
	}
}

func (m *MockNotifier) Send(event Event) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}

func TestNotificationHub_Notify(t *testing.T) {
	tests := []struct {
		name           string
		notifiers      []*MockNotifier
		event          Event
		expectedErrors int
	}{
		{
			name: "all notifiers succeed",
			notifiers: []*MockNotifier{
				NewMockNotifier(nil),
				NewMockNotifier(nil),
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
			name: "some notifiers fail",
			notifiers: []*MockNotifier{
				NewMockNotifier(nil),
				NewMockNotifier(fmt.Errorf("send error")),
				NewMockNotifier(nil),
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
			for _, notifier := range tt.notifiers {
				hub.Register(notifier)
			}

			errors := hub.Notify(tt.event)
			assert.Len(t, errors, tt.expectedErrors)

			for _, notifier := range tt.notifiers {
				if notifier.err == nil {
					assert.Len(t, notifier.events, 1)
					assert.Equal(t, tt.event, notifier.events[0])
				} else {
					assert.Len(t, notifier.events, 0)
				}
			}
		})
	}
}
