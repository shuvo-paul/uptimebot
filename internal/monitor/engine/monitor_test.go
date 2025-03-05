package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTargetCheck(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create target using DefaultClient
	target := &Target{
		ID:       1,
		URL:      ts.URL,
		Interval: time.Minute,
		Enabled:  true,
		Client:   DefaultClient,
	}

	err := target.Check()
	if err != nil {
		t.Errorf("Expected successful check, got error: %v", err)
	}

	if target.Status != statusUp {
		t.Errorf("Expected status %s, got %s", statusUp, target.Status)
	}
}

func TestTargetCheckTimeout(t *testing.T) {
	// Create a test server that delays response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create target with a client that has a short timeout
	client := &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	target := &Target{
		ID:       1,
		URL:      ts.URL,
		Interval: time.Minute,
		Enabled:  true,
		Client:   client,
		Status:   statusUp, // Initial status
	}

	// Check should return timeout error but not change status
	err := target.Check()
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Verify status hasn't changed due to timeout
	if target.Status != statusUp {
		t.Errorf("Status should not change on timeout, expected %s, got %s", statusUp, target.Status)
	}
}

func TestTarget_updateStatus(t *testing.T) {
	target := &Target{
		Status:          statusUp,
		StatusChangedAt: time.Now().Add(-10 * time.Minute),
	}

	newStatus := statusDown
	target.updateStatus(newStatus)

	if target.Status != newStatus {
		t.Errorf("expected status %q, got %q", newStatus, target.Status)
	}

	if time.Since(target.StatusChangedAt) > time.Second {
		t.Errorf("StatusChangedAt was not updated correctly")
	}
}
