package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/database"
	monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	"github.com/shuvo-paul/uptimebot/internal/monitor/model"
)

var (
	ErrTargetNotFound = errors.New("target not found")
)

type TargetRepositoryInterface interface {
	Create(model.UserTarget) (model.UserTarget, error)
	GetByID(int) (model.UserTarget, error)
	GetAll() ([]model.UserTarget, error)
	GetAllByUserID(userID int) ([]model.UserTarget, error)
	Update(model.UserTarget) (model.UserTarget, error)
	Delete(int) error
	UpdateStatus(*monitor.Target, string) error
}

var _ TargetRepositoryInterface = (*TargetRepository)(nil)

type TargetRepository struct {
	db database.Querier
}

func NewTargetRepository(db database.Querier) *TargetRepository {
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
		INSERT INTO target (url, user_id, status, enabled, interval, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRow(
		query,
		userTarget.URL,
		userTarget.UserID,
		userTarget.Status,
		userTarget.Enabled,
		userTarget.Interval.Seconds(),
		r.formatTime(userTarget.StatusChangedAt),
	).Scan(&userTarget.ID)

	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to create target: %w", err)
	}

	userTarget.StatusChangedAt = userTarget.StatusChangedAt.UTC()
	return userTarget, nil
}

func (r *TargetRepository) GetByID(id int) (model.UserTarget, error) {
	query := `
		SELECT id, url, status, enabled, interval, changed_at, user_id
		FROM target
		WHERE id = $1`

	userTarget := model.UserTarget{Target: &monitor.Target{}}
	var intervalSeconds float64
	var statusChangedAtStr string

	err := r.db.QueryRow(query, id).Scan(
		&userTarget.ID,
		&userTarget.URL,
		&userTarget.Status,
		&userTarget.Enabled,
		&intervalSeconds,
		&statusChangedAtStr,
		&userTarget.UserID,
	)

	if err == sql.ErrNoRows {
		return model.UserTarget{}, ErrTargetNotFound
	}
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to get target: %w", err)
	}

	userTarget.Interval = time.Duration(intervalSeconds) * time.Second
	userTarget.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to parse changed_at: %w", err)
	}

	return userTarget, nil
}

func (r *TargetRepository) GetAll() ([]model.UserTarget, error) {
	query := `
		SELECT id, url, status, enabled, interval, changed_at, user_id
		FROM target`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets: %w", err)
	}
	defer rows.Close()

	var targets []model.UserTarget
	for rows.Next() {
		userTarget := model.UserTarget{Target: &monitor.Target{}}
		var intervalSeconds float64
		var statusChangedAtStr string

		err = rows.Scan(
			&userTarget.ID,
			&userTarget.URL,
			&userTarget.Status,
			&userTarget.Enabled,
			&intervalSeconds,
			&statusChangedAtStr,
			&userTarget.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target: %w", err)
		}

		userTarget.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse changed_at: %w", err)
		}

		userTarget.Interval = time.Duration(intervalSeconds) * time.Second
		targets = append(targets, userTarget)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating targets: %w", err)
	}

	return targets, nil
}

func (r *TargetRepository) GetAllByUserID(userID int) ([]model.UserTarget, error) {
	query := `
		SELECT id, url, status, enabled, interval, changed_at, user_id
		FROM target 
		WHERE user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets: %w", err)
	}
	defer rows.Close()

	var targets []model.UserTarget
	for rows.Next() {
		userTarget := model.UserTarget{Target: &monitor.Target{}}
		var intervalSeconds float64
		var statusChangedAtStr string

		err = rows.Scan(
			&userTarget.ID,
			&userTarget.URL,
			&userTarget.Status,
			&userTarget.Enabled,
			&intervalSeconds,
			&statusChangedAtStr,
			&userTarget.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target: %w", err)
		}

		userTarget.StatusChangedAt, err = r.parseTime(statusChangedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse changed_at: %w", err)
		}

		userTarget.Interval = time.Duration(intervalSeconds) * time.Second
		targets = append(targets, userTarget)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating targets: %w", err)
	}

	return targets, nil
}

func (r *TargetRepository) Update(userTarget model.UserTarget) (model.UserTarget, error) {
	query := `
		UPDATE target
		SET url = $1, status = $2, enabled = $3, interval = $4, changed_at = $5
		WHERE id = $6`

	result, err := r.db.Exec(
		query,
		userTarget.URL,
		userTarget.Status,
		userTarget.Enabled,
		userTarget.Interval.Seconds(),
		r.formatTime(userTarget.StatusChangedAt),
		userTarget.ID,
	)
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to update target: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.UserTarget{}, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.UserTarget{}, ErrTargetNotFound
	}

	userTarget.StatusChangedAt = userTarget.StatusChangedAt.UTC()
	return userTarget, nil
}

func (r *TargetRepository) UpdateStatus(target *monitor.Target, status string) error {
	query := `
		UPDATE target
		SET status = $1, changed_at = $2
		WHERE id = $3`

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
	query := `DELETE FROM target WHERE id = $1`

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
