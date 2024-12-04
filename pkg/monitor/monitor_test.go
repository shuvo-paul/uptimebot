package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSite(t *testing.T) {
	// Test with default config
	site := NewSite(1, "https://example.com", time.Minute, DefaultClientConfig)

	if site.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	// Test with custom config
	customConfig := ClientConfig{
		Timeout:         5 * time.Second,
		MaxIdleConns:    50,
		IdleConnTimeout: 30 * time.Second,
	}

	site = NewSite(2, "https://example.com", time.Minute, customConfig)

	if site.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestSiteCheck(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Test with custom shorter timeout
	site := NewSite(1, ts.URL, time.Minute, ClientConfig{
		Timeout:         2 * time.Second,
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	})

	err := site.Check()
	if err != nil {
		t.Errorf("Expected successful check, got error: %v", err)
	}

	if site.Status != statusUp {
		t.Errorf("Expected status %s, got %s", statusUp, site.Status)
	}
}

func TestSite_updateStatus(t *testing.T) {
	site := &Site{
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
