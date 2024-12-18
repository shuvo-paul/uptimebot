package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	statusUp     = "up"
	statusError  = "error"
	statusDown   = "down"
	statusPaused = "paused"
)

// ClientConfig holds HTTP client configuration
type ClientConfig struct {
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// DefaultClientConfig provides sensible defaults
var DefaultClientConfig = ClientConfig{
	Timeout:         10 * time.Second,
	MaxIdleConns:    100,
	IdleConnTimeout: 90 * time.Second,
}

// DefaultClient provides a default HTTP client using DefaultClientConfig
var DefaultClient = &http.Client{
	Timeout: DefaultClientConfig.Timeout,
	Transport: &http.Transport{
		MaxIdleConns:    DefaultClientConfig.MaxIdleConns,
		IdleConnTimeout: DefaultClientConfig.IdleConnTimeout,
	},
}

type StatusUpdateCallback func(*Site, string) error

type Site struct {
	ID              int
	URL             string
	Status          string
	Enabled         bool
	Interval        time.Duration
	StatusChangedAt time.Time
	mu              sync.RWMutex
	cancelFunc      context.CancelFunc
	Client          *http.Client
	OnStatusUpdate  StatusUpdateCallback
}

func (s *Site) Check() error {
	defer func(startStatus string) {
		slog.Info("Site check completed", "URL", s.URL, "fromStatus", startStatus, "toStatus", s.Status)
	}(s.Status)

	r, err := s.Client.Get(s.URL)

	if err != nil {
		s.updateStatus(statusError)
		return fmt.Errorf("connection error: %w", err)
	}

	defer r.Body.Close()

	if r.StatusCode >= 400 {
		s.updateStatus(statusDown)
		return fmt.Errorf("HTTP error: %d", r.StatusCode)
	}

	s.updateStatus(statusUp)

	return nil
}

func (s *Site) updateStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Status != status {
		s.Status = status
		s.StatusChangedAt = time.Now()

		if s.OnStatusUpdate != nil {
			if err := s.OnStatusUpdate(s, status); err != nil {
				slog.Error("Failed to persist status update", "site", s.URL, "error", err)
			}
		}
	}
}

func (s *Site) Update(updatedSite *Site) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.URL = updatedSite.URL
	s.Interval = updatedSite.Interval
	s.Enabled = updatedSite.Enabled
}

type Manager struct {
	mu    sync.Mutex
	Sites map[int]*Site
}

func NewManager() *Manager {
	return &Manager{
		Sites: make(map[int]*Site),
	}
}

func (m *Manager) RegisterSite(site *Site) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Sites[site.ID]; ok {
		return fmt.Errorf("site %s already being monitored", site.URL)
	}

	if site.Client == nil {
		site.Client = DefaultClient
	}

	ctx, cancel := context.WithCancel(context.Background())
	site.cancelFunc = cancel

	m.Sites[site.ID] = site

	go func() {
		ticker := time.NewTicker(site.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("Monitoring stopperd", "site", site.URL)
				m.mu.Lock()
				delete(m.Sites, site.ID)
				m.mu.Unlock()
				return
			case <-ticker.C:
				if !site.Enabled {
					continue
				}
				if err := site.Check(); err != nil {
					slog.Error("Site check failed", "site", site.URL, "error", err)
				}
			}
		}

	}()

	slog.Info("Monitoring started", "site", site.URL)
	return nil
}

func (m *Manager) RevokeSite(siteID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if site, exist := m.Sites[siteID]; exist {
		site.cancelFunc()
		delete(m.Sites, siteID)
		slog.Info("Monitoring Stopped", "Site", site.URL)
	} else {
		slog.Info("Site removed, but no monitoring was active", "siteID", siteID)
	}
}
