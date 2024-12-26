package services

import (
	"encoding/json"
	"testing"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/stretchr/testify/assert"
)

type mockNotifierRepository struct {
	GetBySiteIDFunc func(int) ([]*models.Notifier, error)
}

func (m mockNotifierRepository) Create(notifier *models.Notifier) error {
	return nil
}

func (m mockNotifierRepository) Get(id int64) (*models.Notifier, error) {
	return nil, nil
}

func (m mockNotifierRepository) Update(id int, config *models.NotifierConfig) (*models.Notifier, error) {
	return nil, nil
}

func (m mockNotifierRepository) Delete(id int64) error {
	return nil
}

func (m mockNotifierRepository) GetBySiteID(id int) ([]*models.Notifier, error) {
	return m.GetBySiteIDFunc(id)
}

func TestSetupNotifier(t *testing.T) {
	mockRepo := mockNotifierRepository{}
	mockRepo.GetBySiteIDFunc = func(id int) ([]*models.Notifier, error) {
		notifier := []*models.Notifier{
			{
				ID:     1,
				SiteId: id,
				Config: &models.NotifierConfig{
					Type:   models.NotifierTypeSlack,
					Config: json.RawMessage(`{"webhook_url": "https://example.com/slack/webhook"}`),
				},
			},
		}
		return notifier, nil
	}

	notifierService := NewNotifierService(mockRepo)

	err := notifierService.SetupNotifier(1)
	assert.NoError(t, err)
}
