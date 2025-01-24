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
		SiteId: 1,
		Type:   model.NotifierTypeSlack,
		Config: config,
	}

	return notifierRepo.Create(notifier)
}

func TestNotifierRepository_Create(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	notifierRepo := NewNotifierRepository(db)

	config := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`)
	notifier := &model.Notifier{
		SiteId: 1,
		Type:   model.NotifierTypeSlack,
		Config: config,
	}

	newNotifier, err := notifierRepo.Create(notifier)
	assert.NoError(t, err)
	assert.NotEmpty(t, newNotifier.ID)
	assert.Equal(t, notifier.SiteId, newNotifier.SiteId)
	assert.Equal(t, notifier.Type, newNotifier.Type)
	assert.JSONEq(t, string(config), string(newNotifier.Config))
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
			SiteId: 2,
			Type:   model.NotifierTypeSlack,
			Config: config,
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
			SiteId: 1,
			Type:   model.NotifierTypeSlack,
			Config: initialConfig,
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
		config1 := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test1"}`)
		config2 := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test2"}`)
		otherConfig := json.RawMessage(`{"webhook_url": "https://hooks.slack.com/other"}`)

		notifier1 := &model.Notifier{
			SiteId: siteID,
			Type:   model.NotifierTypeSlack,
			Config: config1,
		}
		notifier2 := &model.Notifier{
			SiteId: siteID,
			Type:   model.NotifierTypeSlack,
			Config: config2,
		}
		otherNotifier := &model.Notifier{
			SiteId: siteID + 1,
			Type:   model.NotifierTypeSlack,
			Config: otherConfig,
		}

		// Save all notifiers
		_, err := repo.Create(notifier1)
		assert.NoError(t, err)
		_, err = repo.Create(notifier2)
		assert.NoError(t, err)
		_, err = repo.Create(otherNotifier)
		assert.NoError(t, err)

		// Fetch notifiers for the site
		notifiers, err := repo.GetBySiteID(siteID)
		assert.NoError(t, err)
		assert.Len(t, notifiers, 2)

		// Verify notifier details
		for _, n := range notifiers {
			assert.Equal(t, siteID, n.SiteId)
			assert.Equal(t, model.NotifierTypeSlack, n.Type)

			config, err := n.GetSlackConfig()
			assert.NoError(t, err)
			assert.Contains(t, config.WebhookURL, "https://hooks.slack.com/test")
		}

		// Verify other site's notifier is not included
		otherNotifiers, err := repo.GetBySiteID(siteID + 1)
		assert.NoError(t, err)
		assert.Len(t, otherNotifiers, 1)
		assert.Equal(t, siteID+1, otherNotifiers[0].SiteId)
	})
}
