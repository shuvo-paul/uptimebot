package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/shuvo-paul/sitemonitor/pkg/monitor"
)

var (
	ErrSiteNotFound = errors.New("site not found")
)

type SiteRepositoryInterface interface {
	Create(*monitor.Site) (*monitor.Site, error)
	GetByID(int) (*monitor.Site, error)
	GetAll() ([]*monitor.Site, error)
	Update(*monitor.Site) (*monitor.Site, error)
	Delete(int) error
	UpdateStatus(*monitor.Site, string) error
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

func (r *SiteRepository) Create(site *monitor.Site) (*monitor.Site, error) {

	if site.URL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}
	if _, err := url.Parse(site.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	query := `
		INSERT INTO sites (url, status, enabled, interval, status_changed_at)
		VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(
		query,
		site.URL,
		site.Status,
		site.Enabled,
		site.Interval.Seconds(),
		r.formatTime(site.StatusChangedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create site: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	site.ID = int(id)
	site.StatusChangedAt = site.StatusChangedAt.UTC()
	return site, nil
}

func (r *SiteRepository) GetByID(id int) (*monitor.Site, error) {
	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM sites
		WHERE id = ?`

	site := &monitor.Site{}
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
		return nil, ErrSiteNotFound
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

func (r *SiteRepository) GetAll() ([]*monitor.Site, error) {
	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM sites`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sites: %w", err)
	}
	defer rows.Close()

	var sites []*monitor.Site
	for rows.Next() {
		site := &monitor.Site{}
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

func (r *SiteRepository) Update(site *monitor.Site) (*monitor.Site, error) {
	query := `
		UPDATE sites
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
		return nil, ErrSiteNotFound
	}

	site.StatusChangedAt = site.StatusChangedAt.UTC()
	return site, nil
}

func (r *SiteRepository) UpdateStatus(site *monitor.Site, status string) error {
	query := `
		UPDATE sites
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
		return ErrSiteNotFound
	}

	return nil
}

func (r *SiteRepository) Delete(siteId int) error {
	query := `DELETE FROM sites WHERE id = ?`

	result, err := r.db.Exec(query, siteId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSiteNotFound
	}

	return nil
}
