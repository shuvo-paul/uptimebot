package notification

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockObserver is a mock implementation of the Observer interface
type MockObserver struct {
	states []State
	err    error
}

func NewMockObserver(err error) *MockObserver {
	return &MockObserver{
		states: make([]State, 0),
		err:    err,
	}
}

func (m *MockObserver) Notify(state State) error {
	if m.err != nil {
		return m.err
	}
	m.states = append(m.states, state)
	return nil
}

func TestSubject_Notify(t *testing.T) {
	tests := []struct {
		name           string
		observers      []*MockObserver
		state          State
		expectedErrors int
	}{
		{
			name: "all observers succeed",
			observers: []*MockObserver{
				NewMockObserver(nil),
				NewMockObserver(nil),
			},
			state: State{
				Name:      "test-system",
				Status:    "up",
				Message:   "test message",
				UpdatedAt: time.Now(),
			},
			expectedErrors: 0,
		},
		{
			name: "some observers fail",
			observers: []*MockObserver{
				NewMockObserver(nil),
				NewMockObserver(fmt.Errorf("update error")),
				NewMockObserver(nil),
			},
			state: State{
				Name:      "test-system",
				Status:    "up",
				Message:   "test message",
				UpdatedAt: time.Now(),
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewSubject()
			for _, observer := range tt.observers {
				subject.Attach(observer)
			}

			errors := subject.Notify(tt.state)
			assert.Len(t, errors, tt.expectedErrors)

			for _, observer := range tt.observers {
				if observer.err == nil {
					assert.Len(t, observer.states, 1)
					assert.Equal(t, tt.state, observer.states[0])
				} else {
					assert.Len(t, observer.states, 0)
				}
			}
		})
	}
}

func TestSubject_Detach(t *testing.T) {
	subject := NewSubject()
	observer1 := NewMockObserver(nil)
	observer2 := NewMockObserver(nil)

	subject.Attach(observer1)
	subject.Attach(observer2)

	state := State{
		Name:      "test-system",
		Status:    "up",
		Message:   "test message",
		UpdatedAt: time.Now(),
	}

	// Both observers should receive the state update
	subject.Notify(state)
	assert.Len(t, observer1.states, 1)
	assert.Len(t, observer2.states, 1)

	// Detach observer1
	subject.Detach(observer1)

	// Only observer2 should receive the state update
	subject.Notify(state)
	assert.Len(t, observer1.states, 1) // Still 1 from before
	assert.Len(t, observer2.states, 2) // Got both updates
}
