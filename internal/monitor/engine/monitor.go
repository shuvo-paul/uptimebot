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

type StatusUpdateCallback func(*Target, string) error

type Target struct {
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

func (s *Target) Check() error {
	defer func(startStatus string) {
		slog.Info("Target check completed", "URL", s.URL, "fromStatus", startStatus, "toStatus", s.Status)
	}(s.Status)

	r, err := s.Client.Get(s.URL)

	if err != nil {
		// Check if the error is a timeout error
		if timeoutErr, ok := err.(interface{ Timeout() bool }); ok && timeoutErr.Timeout() {
			// Log the timeout but don't update status or trigger notification
			slog.Info("Target check timeout", "URL", s.URL, "error", err)
			return fmt.Errorf("timeout error: %v", err)
		}
		// For non-timeout errors, update status and trigger notification
		s.updateStatus(statusError)
		return fmt.Errorf("connection error: %v", err)
	}

	defer r.Body.Close()

	if r.StatusCode >= 400 {
		s.updateStatus(statusDown)
		return fmt.Errorf("HTTP error: %d", r.StatusCode)
	}

	s.updateStatus(statusUp)

	return nil
}

func (s *Target) updateStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Status != status {
		s.Status = status
		s.StatusChangedAt = time.Now()

		if s.OnStatusUpdate != nil {
			if err := s.OnStatusUpdate(s, status); err != nil {
				slog.Error("Failed to persist status update", "Target", s.URL, "error", err)
			}
		}
	}
}

func (s *Target) Update(updatedTarget *Target) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.URL = updatedTarget.URL
	s.Interval = updatedTarget.Interval
	s.Enabled = updatedTarget.Enabled
}

type Manager struct {
	mu      sync.Mutex
	Targets map[int]*Target
}

func NewManager() *Manager {
	return &Manager{
		Targets: make(map[int]*Target),
	}
}

func (m *Manager) RegisterTarget(target *Target) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Targets[target.ID]; ok {
		return fmt.Errorf("Target %s already being monitored", target.URL)
	}

	if target.Client == nil {
		target.Client = DefaultClient
	}

	ctx, cancel := context.WithCancel(context.Background())
	target.cancelFunc = cancel

	m.Targets[target.ID] = target

	go func() {
		ticker := time.NewTicker(target.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("Monitoring stopped", "Target", target.URL)
				m.mu.Lock()
				delete(m.Targets, target.ID)
				m.mu.Unlock()
				return
			case <-ticker.C:
				if !target.Enabled {
					continue
				}
				if err := target.Check(); err != nil {
					slog.Error("Target check failed", "Target", target.URL, "error", err)
				}
			}
		}

	}()

	slog.Info("Monitoring started", "Target", target.URL)
	return nil
}

func (m *Manager) RevokeTarget(targetID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if target, exist := m.Targets[targetID]; exist {
		target.cancelFunc()
		delete(m.Targets, targetID)
		slog.Info("Monitoring Stopped", "Target", target.URL)
	} else {
		slog.Info("Target removed, but no monitoring was active", "targetID", targetID)
	}
}
