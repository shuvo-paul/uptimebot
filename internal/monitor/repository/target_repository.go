package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
)

var (
	ErrTargetNotFound = errors.New("target not found")
)

type TargetRepositoryInterface interface {
	Create(model.UserTarget) (model.UserTarget, error)
	GetByID(int) (*monitor.Target, error)
	GetAll() ([]*monitor.Target, error)
	GetAllByUserID(userID int) ([]*monitor.Target, error)
	Update(*monitor.Target) (*monitor.Target, error)
	Delete(int) error
	UpdateStatus(*monitor.Target, string) error
}

var _ TargetRepositoryInterface = (*TargetRepository)(nil)

type TargetRepository struct {
	db *sql.DB
}

func NewTargetRepository(db *sql.DB) *TargetRepository {
	return &TargetRepository{db: db}
}

func (r *TargetRepository) formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func (r *TargetRepository) parseTime(s string) (time.Time, error) {
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

func (r *TargetRepository) Create(userTarget model.UserTarget) (model.UserTarget, error) {

	if userTarget.URL == "" {
		return model.UserTarget{}, fmt.Errorf("URL cannot be empty")
	}
	if _, err := url.Parse(userTarget.URL); err != nil {
		return model.UserTarget{}, fmt.Errorf("invalid URL: %w", err)
	}
	if userTarget.UserID <= 0 {
		return model.UserTarget{}, fmt.Errorf("invalid UserID: %d", userTarget.UserID)
	}

	query := `
		INSERT INTO target (url, user_id, status, enabled, interval, status_changed_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(
		query,
		userTarget.URL,
		userTarget.UserID,
		userTarget.Status,
		userTarget.Enabled,
		userTarget.Interval.Seconds(),
		r.formatTime(userTarget.StatusChangedAt),
	)
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to create target: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	userTarget.ID = int(id)
	userTarget.StatusChangedAt = userTarget.StatusChangedAt.UTC()
	return userTarget, nil
}

func (r *TargetRepository) GetByID(id int) (*monitor.Target, error) {
	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM target
		WHERE id = ?`

	target := &monitor.Target{}
	var intervalSeconds float64
	var statusChangedAtStr string

	err := r.db.QueryRow(query, id).Scan(
		&target.ID,
		&target.URL,
		&target.Status,
		&target.Enabled,
		&intervalSeconds,
		&statusChangedAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	target.Interval = time.Duration(intervalSeconds) * time.Second
	target.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status_changed_at: %w", err)
	}

	return target, nil
}

func (r *TargetRepository) GetAll() ([]*monitor.Target, error) {

	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM target`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets: %w", err)
	}
	defer rows.Close()

	var targets []*monitor.Target
	for rows.Next() {
		target := &monitor.Target{}
		var intervalSeconds float64
		var statusChangedAtStr string

		err := rows.Scan(
			&target.ID,
			&target.URL,
			&target.Status,
			&target.Enabled,
			&intervalSeconds,
			&statusChangedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target: %w", err)
		}

		target.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse status_changed_at: %w", err)
		}

		target.Interval = time.Duration(intervalSeconds) * time.Second
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating targets: %w", err)
	}

	return targets, nil
}

func (r *TargetRepository) GetAllByUserID(userID int) ([]*monitor.Target, error) {
	query := `
		SELECT id, url, status, enabled, interval, status_changed_at
		FROM target 
		WHERE user_id = ?`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets: %w", err)
	}
	defer rows.Close()

	var targets []*monitor.Target
	for rows.Next() {
		target := &monitor.Target{}
		var intervalSeconds float64
		var statusChangedAtStr string

		err := rows.Scan(
			&target.ID,
			&target.URL,
			&target.Status,
			&target.Enabled,
			&intervalSeconds,
			&statusChangedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target: %w", err)
		}

		target.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse status_changed_at: %w", err)
		}

		target.Interval = time.Duration(intervalSeconds) * time.Second
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating targets: %w", err)
	}

	return targets, nil
}

func (r *TargetRepository) Update(target *monitor.Target) (*monitor.Target, error) {
	query := `
		UPDATE target
		SET url = ?, status = ?, enabled = ?, interval = ?, status_changed_at = ?
		WHERE id = ?`

	result, err := r.db.Exec(
		query,
		target.URL,
		target.Status,
		target.Enabled,
		target.Interval.Seconds(),
		r.formatTime(target.StatusChangedAt),
		target.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update target: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrTargetNotFound
	}

	target.StatusChangedAt = target.StatusChangedAt.UTC()
	return target, nil
}

func (r *TargetRepository) UpdateStatus(target *monitor.Target, status string) error {
	query := `
		UPDATE target
		SET status = ?, status_changed_at = ?
		WHERE id = ?`

	result, err := r.db.Exec(query, status, target.StatusChangedAt, target.ID)
	if err != nil {
		return fmt.Errorf("failed to update target status: %w", err)
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

func (r *TargetRepository) Delete(targetId int) error {
	query := `DELETE FROM target WHERE id = ?`

	result, err := r.db.Exec(query, targetId)
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
