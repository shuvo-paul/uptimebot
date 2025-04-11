package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/database"
	"github.com/shuvo-paul/uptimebot/internal/notification/model"
)

type NotifierRepositoryInterface interface {
	Create(*model.Notifier) (*model.Notifier, error)
	Get(int) (*model.Notifier, error)
	Update(int, json.RawMessage) (*model.Notifier, error)
	Delete(int) error
	GetByTargetID(int) ([]*model.Notifier, error)
}

var _ NotifierRepositoryInterface = (*NotifierRepository)(nil)

// NotifierRepository handles database operations for notifiers
type NotifierRepository struct {
	db database.Querier
}

// NewNotifierRepository creates a new notifier repository
func NewNotifierRepository(db database.Querier) *NotifierRepository {
	return &NotifierRepository{db: db}
}

// Create inserts a new notifier into the database
func (r *NotifierRepository) Create(notifier *model.Notifier) (*model.Notifier, error) {
	// Validate config based on notifier type
	if notifier.Type == model.NotifierTypeSlack {
		config, err := notifier.GetSlackConfig()
		if err != nil {
			return nil, fmt.Errorf("invalid slack config: %w", err)
		}
		if config == nil || config.WebhookURL == "" {
			return nil, fmt.Errorf("webhook URL is required for slack notifier")
		}
	} else if notifier.Type == model.NotifierTypeEmail {
		config, err := notifier.GetEmailConfig()
		if err != nil {
			return nil, fmt.Errorf("invalid email config: %w", err)
		}
		if config == nil {
			return nil, fmt.Errorf("email configuration is required")
		}
	}

	query := `
		INSERT INTO notifier (target_id, type, config)
		VALUES ($1, $2, $3)
		RETURNING id, target_id, type, config
	`

	newNotifier := &model.Notifier{}
	err := r.db.QueryRow(query, notifier.TargetId, notifier.Type, notifier.Config).Scan(
		&newNotifier.ID,
		&newNotifier.TargetId,
		&newNotifier.Type,
		&newNotifier.Config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notifier: %w", err)
	}

	return newNotifier, nil
}

// Get retrieves a notifier by ID
func (r *NotifierRepository) Get(id int) (*model.Notifier, error) {
	query := `
		SELECT id, target_id, type, config
		FROM notifier
		WHERE id = $1
	`

	notifier := &model.Notifier{}
	err := r.db.QueryRow(query, id).Scan(
		&notifier.ID,
		&notifier.TargetId,
		&notifier.Type,
		&notifier.Config,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notifier: %w", err)
	}

	return notifier, nil
}

// Update updates a notifier's configuration
func (r *NotifierRepository) Update(id int, config json.RawMessage) (*model.Notifier, error) {
	query := `
		UPDATE notifier
		SET config = $1
		WHERE id = $2
		RETURNING id, target_id, type, config
	`

	notifier := &model.Notifier{}
	err := r.db.QueryRow(query, config, id).Scan(
		&notifier.ID,
		&notifier.TargetId,
		&notifier.Type,
		&notifier.Config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update: %w", err)
	}

	return notifier, nil
}

// Delete removes a notifier from the database
func (r *NotifierRepository) Delete(id int) error {
	query := `DELETE FROM notifier WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notifier: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return nil
	}

	return nil
}

// GetByTargetID retrieves all notifiers for a specific target
func (r *NotifierRepository) GetByTargetID(targetID int) ([]*model.Notifier, error) {
	query := `
		SELECT id, target_id, type, config
		FROM notifier
		WHERE target_id = $1
	`

	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifiers: %w", err)
	}
	defer rows.Close()

	var notifiers []*model.Notifier
	for rows.Next() {
		notifier := &model.Notifier{}
		err := rows.Scan(
			&notifier.ID,
			&notifier.TargetId,
			&notifier.Type,
			&notifier.Config,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notifier: %w", err)
		}

		notifiers = append(notifiers, notifier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifiers: %w", err)
	}

	return notifiers, nil
}
