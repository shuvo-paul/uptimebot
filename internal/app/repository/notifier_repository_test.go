package repository

import (
	"encoding/json"
	"testing"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/internal/app/testutil"
	"github.com/stretchr/testify/assert"
)

func createNotifier() (*models.Notifier, error) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

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
	db := testutil.NewInMemoryDB()
	defer db.Close()

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
	db := testutil.NewInMemoryDB()
	defer db.Close()

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
	db := testutil.NewInMemoryDB()
	defer db.Close()

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

func TestNotifierRepository_GetBySiteID(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	repo := NewNotifierRepository(db)

	t.Run("NoNotifiers", func(t *testing.T) {
		notifiers, err := repo.GetBySiteID(999)
		assert.NoError(t, err)
		assert.Empty(t, notifiers)
	})

	t.Run("MultipleNotifiers", func(t *testing.T) {
		// Create multiple notifiers for the same site
		siteID := 1
		notifier1 := &models.Notifier{
			SiteId: siteID,
			Config: &models.NotifierConfig{
				Type:   models.NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test1"}`),
			},
		}
		notifier2 := &models.Notifier{
			SiteId: siteID,
			Config: &models.NotifierConfig{
				Type:   models.NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`),
			},
		}

		// Create notifier for different site
		otherNotifier := &models.Notifier{
			SiteId: siteID + 1,
			Config: &models.NotifierConfig{
				Type:   models.NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/other"}`),
			},
		}

		// Save all notifiers
		err := repo.Create(notifier1)
		assert.NoError(t, err)
		err = repo.Create(notifier2)
		assert.NoError(t, err)
		err = repo.Create(otherNotifier)
		assert.NoError(t, err)

		// Fetch notifiers for the site
		notifiers, err := repo.GetBySiteID(siteID)
		assert.NoError(t, err)
		assert.Len(t, notifiers, 2)

		// Verify notifier details
		for _, n := range notifiers {
			assert.Equal(t, siteID, n.SiteId)
			assert.Equal(t, models.NotifierTypeSlack, n.Config.Type)

			config, err := n.Config.GetSlackConfig()
			assert.NoError(t, err)
			assert.Contains(t, config.WebhookURL, "https://hooks.slack.com/test")
		}

		// Verify other site's notifier is not included
		notifiers, err = repo.GetBySiteID(siteID + 1)
		assert.NoError(t, err)
		assert.Len(t, notifiers, 1)
		assert.Equal(t, siteID+1, notifiers[0].SiteId)
	})
}
