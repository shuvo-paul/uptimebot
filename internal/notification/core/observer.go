package notification

import "time"

// State represents the current state that observers are interested in
type State struct {
	Name      string    // Name of what is being observed
	Status    string    // Current status
	Message   string    // Additional details
	UpdatedAt time.Time // When the state was last updated
}

// Observer defines the interface for objects that should be notified of state changes
type Observer interface {
	Notify(State) error
}

// Subject maintains a list of observers and notifies them of state changes
type Subject struct {
	observers []Observer
}

// NewSubject creates a new subject
func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

// Attach adds an observer to the subject
func (s *Subject) Attach(observer Observer) {
	s.observers = append(s.observers, observer)
}

// Detach removes an observer from the subject
func (s *Subject) Detach(observer Observer) {
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

// Notify sends the state to all observers
func (s *Subject) Notify(state State) []error {
	var errors []error
	for _, observer := range s.observers {
		if err := observer.Notify(state); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
