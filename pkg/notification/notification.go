package notification

import (
	"time"
)

// Event represents a notification event
type Event struct {
	SiteURL    string
	Status     string
	Message    string
	OccurredAt time.Time
}

// Notifier is the interface that wraps the basic Send method
type Notifier interface {
	Send(event Event) error
}

// NotificationHub manages multiple notifiers
type NotificationHub struct {
	notifiers []Notifier
}

// NewNotificationHub creates a new notification hub
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		notifiers: make([]Notifier, 0),
	}
}

// Register adds a new notifier to the hub
func (h *NotificationHub) Register(notifier Notifier) {
	h.notifiers = append(h.notifiers, notifier)
}

// Notify sends an event to all registered notifiers
func (h *NotificationHub) Notify(event Event) []error {
	var errors []error
	for _, notifier := range h.notifiers {
		if err := notifier.Send(event); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
