package repository

import (
	"encoding/json"
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/notification/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func createNotifier() (*model.Notifier, error) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	notifierRepo := NewNotifierRepository(db)
	config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`)
	notifier := &model.Notifier{
		TargetId: 1,
		Type:     model.NotifierTypeSlack,
		Config:   config,
	}

	return notifierRepo.Create(notifier)
}

func TestNotifierRepository_Create(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	notifierRepo := NewNotifierRepository(db)

	slackNotifier := &model.Notifier{
		TargetId: 1,
		Type:     model.NotifierTypeSlack,
		Config:   json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
	}

	emailNotifier := &model.Notifier{
		TargetId: 1,
		Type:     model.NotifierTypeEmail,
		Config:   json.RawMessage(`{"recipients": ["EMAIL", "EMAIL", "EMAIL"]}`),
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
				Type:   model.NotifierTypeSlack,
				Config: json.RawMessage(`invalid json`),
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
			name: "invalid json",
			notifier: &model.Notifier{
				Type:   model.NotifierTypeEmail,
				Config: json.RawMessage(`invalid json`),
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
	db := testutil.NewInMemoryDB()
	defer db.Close()

	notifierRepo := NewNotifierRepository(db)

	t.Run("NotFound", func(t *testing.T) {
		notifier, err := notifierRepo.Get(1)
		assert.NoError(t, err)
		assert.Nil(t, notifier)
	})

	t.Run("Found", func(t *testing.T) {
		config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`)
		notifier := &model.Notifier{
			TargetId: 2,
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
	db := testutil.NewInMemoryDB()
	defer db.Close()

	notifierRepo := NewNotifierRepository(db)

	t.Run("NotFound", func(t *testing.T) {
		config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`)
		notifier, err := notifierRepo.Update(99, config)
		assert.Error(t, err)
		assert.Nil(t, notifier)
	})

	t.Run("Success", func(t *testing.T) {
		// First create a notifier
		initialConfig := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`)
		notifier := &model.Notifier{
			TargetId: 1,
			Type:     model.NotifierTypeSlack,
			Config:   initialConfig,
		}
		created, err := notifierRepo.Create(notifier)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		// Update the config
		newConfig := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`)
		updated, err := notifierRepo.Update(int(created.ID), newConfig)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.Type, updated.Type)
		assert.JSONEq(t, string(newConfig), string(updated.Config))
	})
}

func TestNotifierRepository_Delete(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	repo := NewNotifierRepository(db)

	t.Run("NotFound", func(t *testing.T) {
		err := repo.Delete(1)
		assert.NoError(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		notifier, err := createNotifier()
		assert.NoError(t, err)
		assert.NotEmpty(t, notifier.ID)

		err = repo.Delete(notifier.ID)
		assert.NoError(t, err)
	})
}

func TestNotifierRepository_GetByTargetID(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	repo := NewNotifierRepository(db)

	t.Run("NoNotifiers", func(t *testing.T) {
		notifiers, err := repo.GetByTargetID(999)
		assert.NoError(t, err)
		assert.Empty(t, notifiers)
	})

	t.Run("MultipleNotifiers", func(t *testing.T) {
		// Create multiple notifiers for the same target
		targetID := 1
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
			TargetId: targetID + 1,
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
		otherNotifiers, err := repo.GetByTargetID(targetID + 1)
		assert.NoError(t, err)
		assert.Len(t, otherNotifiers, 1)
		assert.Equal(t, targetID+1, otherNotifiers[0].TargetId)
	})
}
