package services

import (
	"fmt"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/internal/app/repository"
	"github.com/shuvo-paul/sitemonitor/pkg/notification"
)

type NotifierService struct {
	notifierRepo    repository.NotifierRepositoryInterface
	notificationHub notification.NotificationHubInterface
}

func NewNotifierService(
	notifierRepo repository.NotifierRepositoryInterface,
) NotifierService {
	return NotifierService{
		notifierRepo:    notifierRepo,
		notificationHub: notification.NewNotificationHub(),
	}
}

func (s *NotifierService) SetupNotifier(siteID int) error {
	notifiers, err := s.notifierRepo.GetBySiteID(siteID)
	if err != nil {
		return err
	}

	for _, notifier := range notifiers {
		switch notifier.Config.Type {
		case models.NotifierTypeSlack:
			config, err := notifier.Config.GetSlackConfig()
			if err != nil {
				return err
			}
			s.notificationHub.Register(notification.NewSlackNotifier(config.WebhookURL, http.DefaultClient))
		default:
			return fmt.Errorf("unsupported notifier type: %s", notifier.Config.Type)
		}
	}

	return nil
}
