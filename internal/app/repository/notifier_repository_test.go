package repository

import (
	"encoding/json"
	"testing"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func createNotifier() (*models.Notifier, error) {
	notifierRepo := NewNotifierRepository(db)
	notifier := &models.Notifier{
		SiteId: 1,
		Config: &models.NotifierConfig{
			Type:   models.NotifierTypeSlack,
			Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
		},
	}

	err := notifierRepo.Create(notifier)
	return notifier, err
}

func TestNotifierRepository_Create(t *testing.T) {
	notifierRepo := NewNotifierRepository(db)

	notifier := &models.Notifier{
		SiteId: 1,
		Config: &models.NotifierConfig{
			Type:   models.NotifierTypeSlack,
			Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
		},
	}

	err := notifierRepo.Create(notifier)

	assert.NoError(t, err)
	assert.NotEmpty(t, notifier.ID)

}

func TestNotifierRepository_Get(t *testing.T) {
	notifierRepo := NewNotifierRepository(db)

	t.Run("NotFound", func(t *testing.T) {
		notifier, err := notifierRepo.Get(1)
		assert.NoError(t, err)
		assert.Nil(t, notifier)
	})

	t.Run("Found", func(t *testing.T) {
		notifier := &models.Notifier{
			SiteId: 2,
			Config: &models.NotifierConfig{
				Type:   models.NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
		}
		err := notifierRepo.Create(notifier)

		assert.NoError(t, err)
		assert.NotEmpty(t, notifier.ID)

		savedNotifier, err := notifierRepo.Get(notifier.ID)
		assert.NoError(t, err)
		assert.NotNil(t, savedNotifier)
		assert.Equal(t, notifier.ID, savedNotifier.ID)
		assert.Equal(t, notifier.Config.Type, savedNotifier.Config.Type)
		assert.JSONEq(t, string(notifier.Config.Config), string(savedNotifier.Config.Config))
	})
}

func TestNotifierRepository_Update(t *testing.T) {
	notifierRepo := NewNotifierRepository(db)

	t.Run("NotFound", func(t *testing.T) {
		config := &models.NotifierConfig{
			Type:   models.NotifierTypeSlack,
			Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`),
		}

		notifier, err := notifierRepo.Update(99, config)
		assert.Error(t, err)
		assert.Nil(t, notifier)
	})

	t.Run("Success", func(t *testing.T) {
		// First create a notifier
		notifier := &models.Notifier{
			SiteId: 1,
			Config: &models.NotifierConfig{
				Type:   models.NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
		}
		err := notifierRepo.Create(notifier)
		assert.NoError(t, err)
		assert.NotEmpty(t, notifier.ID)

		// Update the config
		newConfig := &models.NotifierConfig{
			Type:   models.NotifierTypeSlack,
			Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`),
		}

		updatedNotifier, err := notifierRepo.Update(int(notifier.ID), newConfig)
		assert.NoError(t, err)
		assert.NotNil(t, updatedNotifier)
		assert.Equal(t, notifier.ID, updatedNotifier.ID)
		assert.Equal(t, notifier.Config.Type, updatedNotifier.Config.Type)
		assert.JSONEq(t, string(newConfig.Config), string(updatedNotifier.Config.Config))
	})
}

func TestNotifierRepository_Delete(t *testing.T) {
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
