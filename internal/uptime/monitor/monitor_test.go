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

	// Create site using DefaultClient
	site := &Target{
		ID:       1,
		URL:      ts.URL,
		Interval: time.Minute,
		Enabled:  true,
		Client:   DefaultClient,
	}

	err := site.Check()
	if err != nil {
		t.Errorf("Expected successful check, got error: %v", err)
	}

	if site.Status != statusUp {
		t.Errorf("Expected status %s, got %s", statusUp, site.Status)
	}
}

func TestTarget_updateStatus(t *testing.T) {
	site := &Target{
		Status:          statusUp,
		StatusChangedAt: time.Now().Add(-10 * time.Minute),
	}

	newStatus := statusDown
	site.updateStatus(newStatus)

	if site.Status != newStatus {
		t.Errorf("expected status %q, got %q", newStatus, site.Status)
	}

	if time.Since(site.StatusChangedAt) > time.Second {
		t.Errorf("StatusChangedAt was not updated correctly")
	}
}
