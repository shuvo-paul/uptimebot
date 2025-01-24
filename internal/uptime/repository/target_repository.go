package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/uptime/model"
	"github.com/shuvo-paul/uptimebot/internal/uptime/monitor"
)

var (
	ErrTargetNotFound = errors.New("target not found")
)

type SiteRepositoryInterface interface {
	Create(model.UserTarget) (model.UserTarget, error)
	GetByID(int) (*monitor.Target, error)
	GetAll() ([]*monitor.Target, error)
	GetAllByUserID(userID int) ([]*monitor.Target, error)
	Update(*monitor.Target) (*monitor.Target, error)
	Delete(int) error
	UpdateStatus(*monitor.Target, string) error
}

var _ SiteRepositoryInterface = (*SiteRepository)(nil)

type SiteRepository struct {
	db *sql.DB
}

func NewSiteRepository(db *sql.DB) *SiteRepository {
	return &SiteRepository{db: db}
}

func (r *SiteRepository) formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func (r *SiteRepository) parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try parsing RFC3339Nano format first
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}

	// If that fails, try parsing the alternative format
	return time.Parse("2006-01-02 15:04:05.999999999-07:00", s)
}

func (r *SiteRepository) Create(userSite model.UserTarget) (model.UserTarget, error) {

	if userSite.URL == "" {
		return model.UserTarget{}, fmt.Errorf("URL cannot be empty")
	}
	if _, err := url.Parse(userSite.URL); err != nil {
		return model.UserTarget{}, fmt.Errorf("invalid URL: %w", err)
	}
	if userSite.UserID <= 0 {
		return model.UserTarget{}, fmt.Errorf("invalid UserID: %d", userSite.UserID)
	}

	query := `
		INSERT INTO target (url, user_id, status, enabled, interval, status_changed_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(
		query,
		userSite.URL,
		userSite.UserID,
		userSite.Status,
		userSite.Enabled,
		userSite.Interval.Seconds(),
		r.formatTime(userSite.StatusChangedAt),
	)
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to create site: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	userSite.ID = int(id)
	userSite.StatusChangedAt = userSite.StatusChangedAt.UTC()
	return userSite, nil
}

func (r *SiteRepository) GetByID(id int) (*monitor.Target, error) {
	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM target
		WHERE id = ?`

	site := &monitor.Target{}
	var intervalSeconds float64
	var statusChangedAtStr string

	err := r.db.QueryRow(query, id).Scan(
		&site.ID,
		&site.URL,
		&site.Status,
		&site.Enabled,
		&intervalSeconds,
		&statusChangedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get site: %w", err)
	}

	site.Interval = time.Duration(intervalSeconds) * time.Second
	site.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status_changed_at: %w", err)
	}

	return site, nil
}

func (r *SiteRepository) GetAll() ([]*monitor.Target, error) {

	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM target`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sites: %w", err)
	}
	defer rows.Close()

	var sites []*monitor.Target
	for rows.Next() {
		site := &monitor.Target{}
		var intervalSeconds float64
		var statusChangedAtStr string

		err := rows.Scan(
			&site.ID,
			&site.URL,
			&site.Status,
			&site.Enabled,
			&intervalSeconds,
			&statusChangedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}

		site.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse status_changed_at: %w", err)
		}

		site.Interval = time.Duration(intervalSeconds) * time.Second
		sites = append(sites, site)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sites: %w", err)
	}

	return sites, nil
}

func (r *SiteRepository) GetAllByUserID(userID int) ([]*monitor.Target, error) {
	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM target 
		WHERE user_id = ?`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sites: %w", err)
	}
	defer rows.Close()

	var sites []*monitor.Target
	for rows.Next() {
		site := &monitor.Target{}
		var intervalSeconds float64
		var statusChangedAtStr string

		err := rows.Scan(
			&site.ID,
			&site.URL,
			&site.Status,
			&site.Enabled,
			&intervalSeconds,
			&statusChangedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}

		site.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse status_changed_at: %w", err)
		}

		site.Interval = time.Duration(intervalSeconds) * time.Second
		sites = append(sites, site)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sites: %w", err)
	}

	return sites, nil
}

func (r *SiteRepository) Update(site *monitor.Target) (*monitor.Target, error) {
	query := `
		UPDATE target
		SET url = ?, status = ?, enabled = ?, interval = ?, status_changed_at = ?
		WHERE id = ?`

	result, err := r.db.Exec(
		query,
		site.URL,
		site.Status,
		site.Enabled,
		site.Interval.Seconds(),
		r.formatTime(site.StatusChangedAt),
		site.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update site: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrTargetNotFound
	}

	site.StatusChangedAt = site.StatusChangedAt.UTC()
	return site, nil
}

func (r *SiteRepository) UpdateStatus(site *monitor.Target, status string) error {
	query := `
		UPDATE target
		SET status = ?, status_changed_at = ?
		WHERE id = ?`

	result, err := r.db.Exec(query, status, site.StatusChangedAt, site.ID)
	if err != nil {
		return fmt.Errorf("failed to update site status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTargetNotFound
	}

	return nil
}

func (r *SiteRepository) Delete(siteId int) error {
	query := `DELETE FROM target WHERE id = ?`

	result, err := r.db.Exec(query, siteId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTargetNotFound
	}

	return nil
}
