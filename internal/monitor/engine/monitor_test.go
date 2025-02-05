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
