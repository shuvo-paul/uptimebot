package services

import (
	"fmt"
	"net/http"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/internal/app/repository"
	"github.com/shuvo-paul/sitemonitor/pkg/notification"
)

type NotifierServiceInterface interface {
	ConfigureObservers(siteID int) error
}

type NotifierService struct {
	notifierRepo repository.NotifierRepositoryInterface
	Subject      *notification.Subject
}

func NewNotifierService(
	notifierRepo repository.NotifierRepositoryInterface,
	subject *notification.Subject,
) *NotifierService {
	if subject == nil {
		subject = notification.NewSubject()
	}
	return &NotifierService{
		notifierRepo: notifierRepo,
		Subject:      subject,
	}
}

// ConfigureObservers configures observers for a specific site
func (s *NotifierService) ConfigureObservers(siteID int) error {
	// First detach any existing observers
	// This ensures we don't have duplicate observers if called multiple times
	s.Subject = notification.NewSubject()

	notifiers, err := s.notifierRepo.GetBySiteID(siteID)
	if err != nil {
		return fmt.Errorf("failed to get notifiers: %w", err)
	}

	for _, notifier := range notifiers {
		switch notifier.Config.Type {
		case models.NotifierTypeSlack:
			config, err := notifier.Config.GetSlackConfig()
			if err != nil {
				return fmt.Errorf("failed to get slack config: %w", err)
			}
			observer := notification.NewSlackObserver(config.WebhookURL, http.DefaultClient)
			s.Subject.Attach(observer)
		default:
			return fmt.Errorf("unsupported notifier type: %s", notifier.Config.Type)
		}
	}

	return nil
}
