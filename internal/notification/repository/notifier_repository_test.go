package repository

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	authModel "github.com/shuvo-paul/uptimebot/internal/auth/model"
	authRepo "github.com/shuvo-paul/uptimebot/internal/auth/repository"
	core "github.com/shuvo-paul/uptimebot/internal/monitor/engine"
	monitorModel "github.com/shuvo-paul/uptimebot/internal/monitor/model"
	monitorRepo "github.com/shuvo-paul/uptimebot/internal/monitor/repository"
	"github.com/shuvo-paul/uptimebot/internal/notification/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func createTestUser(t *testing.T, tx *sql.Tx) *authModel.User {
	userRepo := authRepo.NewUserRepository(tx)
	user, err := userRepo.SaveUser(&authModel.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
	})

	if err != nil {
		return nil
	}

	return user
}

func createTestTarget(t *testing.T, tx *sql.Tx, user *authModel.User) int {
	// Create target
	targetRepo := monitorRepo.NewTargetRepository(tx)
	target, err := targetRepo.Create(monitorModel.UserTarget{
		UserID: user.ID,
		Target: &core.Target{
			URL:             "example.org",
			Status:          "up",
			Enabled:         true,
			Interval:        30 * time.Second,
			StatusChangedAt: time.Now(),
		},
	})
	assert.NoError(t, err)

	return target.ID
}

func TestNotifierRepository_Create(t *testing.T) {
	tx := testutil.GetTestTx(t)
	notifierRepo := NewNotifierRepository(tx)

	user := createTestUser(t, tx)
	targetID := createTestTarget(t, tx, user)

	slackNotifier := &model.Notifier{
		TargetId: targetID,
		Type:     model.NotifierTypeSlack,
		Config:   json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
	}

	emailNotifier := &model.Notifier{
		TargetId: targetID,
		Type:     model.NotifierTypeEmail,
		Config:   json.RawMessage(`{"recipients": ["test@example.com"]}`),
	}

	tests := []struct {
		name     string
		notifier *model.Notifier
		want     *model.Notifier
		wantErr  bool
	}{
		{
			name:     "valid slack config",
			notifier: slackNotifier,
			want:     slackNotifier,
			wantErr:  false,
		},
		{
			name: "invalid json",
			notifier: &model.Notifier{
				TargetId: targetID,
				Type:     model.NotifierTypeSlack,
				Config:   json.RawMessage(`invalid json`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "valid email config",
			notifier: emailNotifier,
			want:     emailNotifier,
			wantErr:  false,
		},
		{
			name: "invalid target id",
			notifier: &model.Notifier{
				TargetId: 999,
				Type:     model.NotifierTypeSlack,
				Config:   json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newNotifier, err := notifierRepo.Create(tt.notifier)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, newNotifier.ID)
			assert.Equal(t, tt.want.TargetId, newNotifier.TargetId)
			assert.Equal(t, tt.want.Type, newNotifier.Type)
			assert.JSONEq(t, string(tt.want.Config), string(newNotifier.Config))
		})
	}

}

func TestNotifierRepository_Get(t *testing.T) {
	tx := testutil.GetTestTx(t)
	notifierRepo := NewNotifierRepository(tx)
	user := createTestUser(t, tx)
	targetID := createTestTarget(t, tx, user)

	t.Run("NotFound", func(t *testing.T) {
		notifier, err := notifierRepo.Get(1)
		assert.NoError(t, err)
		assert.Nil(t, notifier)
	})

	t.Run("Found", func(t *testing.T) {
		config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`)
		notifier := &model.Notifier{
			TargetId: targetID,
			Type:     model.NotifierTypeSlack,
			Config:   config,
		}
		created, err := notifierRepo.Create(notifier)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		savedNotifier, err := notifierRepo.Get(created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, savedNotifier)
		assert.Equal(t, created.ID, savedNotifier.ID)
		assert.Equal(t, created.Type, savedNotifier.Type)
		assert.JSONEq(t, string(config), string(savedNotifier.Config))
	})
}

func TestNotifierRepository_Update(t *testing.T) {
	tx := testutil.GetTestTx(t)
	notifierRepo := NewNotifierRepository(tx)
	user := createTestUser(t, tx)
	targetID := createTestTarget(t, tx, user)

	t.Run("NotFound", func(t *testing.T) {
		config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`)
		notifier, err := notifierRepo.Update(99, config)
		assert.Error(t, err)
		assert.Nil(t, notifier)
	})

	t.Run("Success", func(t *testing.T) {
		initialConfig := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`)
		notifier := &model.Notifier{
			TargetId: targetID,
			Type:     model.NotifierTypeSlack,
			Config:   initialConfig,
		}
		created, err := notifierRepo.Create(notifier)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		newConfig := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`)
		updated, err := notifierRepo.Update(created.ID, newConfig)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.Type, updated.Type)
		assert.JSONEq(t, string(newConfig), string(updated.Config))
	})
}

func TestNotifierRepository_GetByTargetID(t *testing.T) {
	tx := testutil.GetTestTx(t)
	repo := NewNotifierRepository(tx)
	user := createTestUser(t, tx)
	targetID := createTestTarget(t, tx, user)
	targetID2 := createTestTarget(t, tx, user)

	t.Run("NoNotifiers", func(t *testing.T) {
		notifiers, err := repo.GetByTargetID(999)
		assert.NoError(t, err)
		assert.Empty(t, notifiers)
	})

	t.Run("MultipleNotifiers", func(t *testing.T) {
		config1 := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test1"}`)
		config2 := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`)
		otherConfig := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/other"}`)

		notifier1 := &model.Notifier{
			TargetId: targetID,
			Type:     model.NotifierTypeSlack,
			Config:   config1,
		}
		notifier2 := &model.Notifier{
			TargetId: targetID,
			Type:     model.NotifierTypeSlack,
			Config:   config2,
		}
		otherNotifier := &model.Notifier{
			TargetId: targetID2,
			Type:     model.NotifierTypeSlack,
			Config:   otherConfig,
		}

		// Save all notifiers
		_, err := repo.Create(notifier1)
		assert.NoError(t, err)
		_, err = repo.Create(notifier2)
		assert.NoError(t, err)
		_, err = repo.Create(otherNotifier)
		assert.NoError(t, err)

		// Fetch notifiers for the target
		notifiers, err := repo.GetByTargetID(targetID)
		assert.NoError(t, err)
		assert.Len(t, notifiers, 2)

		// Verify notifier details
		for _, n := range notifiers {
			assert.Equal(t, targetID, n.TargetId)
			assert.Equal(t, model.NotifierTypeSlack, n.Type)
			config, err := n.GetSlackConfig()
			assert.NoError(t, err)
			assert.Contains(t, config.WebhookURL, "https://hooks.slack.com/test")
		}

		// Verify other target's notifier is not included
		otherNotifiers, err := repo.GetByTargetID(targetID2)
		assert.NoError(t, err)
		assert.Len(t, otherNotifiers, 1)
		assert.Equal(t, targetID2, otherNotifiers[0].TargetId)
	})
}
