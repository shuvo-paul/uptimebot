package notification

import (
	"fmt"
	"time"
)

// Event represents a notification event
type Event struct {
	SiteURL    string
	Status     string
	Message    string
	OccurredAt time.Time
}

// Message represents a notification message
type Message struct {
	NotifierID string
	Event      Event
}

// Provider is the interface that wraps the basic Send method
type Provider interface {
	Send(message Message) error
}

// Notifier represents a notification channel
type Notifier struct {
	id       string
	provider Provider
}

// NewNotifier creates a new notifier with a specific sender
func NewNotifier(id string, provider Provider) *Notifier {
	return &Notifier{
		id:       id,
		provider: provider,
	}
}

// Send sends a notification through this notifier
func (n *Notifier) Send(event Event) error {
	msg := Message{
		Event:      event,
		NotifierID: n.id,
	}
	if err := n.provider.Send(msg); err != nil {
		return fmt.Errorf("notifier %s failed: %w", n.id, err)
	}
	return nil
}

// NotificationHub manages multiple notifiers
type NotificationHub struct {
	notifiers []*Notifier
}

// NewNotificationHub creates a new notification hub
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		notifiers: make([]*Notifier, 0),
	}
}

// RegisterNotifier adds a new notifier to the hub
func (h *NotificationHub) RegisterNotifier(notifier *Notifier) {
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
