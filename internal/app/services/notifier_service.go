package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
	"github.com/shuvo-paul/sitemonitor/internal/app/repository"
	"github.com/shuvo-paul/sitemonitor/pkg/notification"
)

type NotifierServiceInterface interface {
	Create(notifier *models.Notifier) error
	Get(id int64) (*models.Notifier, error)
	Update(id int, config *models.NotifierConfig) (*models.Notifier, error)
	Delete(id int64) error
	ConfigureObservers(siteID int) error
	HandleSlackCallback(code string, siteId int) (*models.Notifier, error)
	ParseOAuthState(state string) (int, error)
}

type NotifierService struct {
	notifierRepo repository.NotifierRepositoryInterface
	Subject      *notification.Subject
}

var (
	// SlackTokenURL is the Slack OAuth token URL, can be overridden in tests
	SlackTokenURL = "https://slack.com/api/oauth.v2.access"
)

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

// Create adds a new notifier
func (s *NotifierService) Create(notifier *models.Notifier) error {
	if err := s.notifierRepo.Create(notifier); err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}
	return nil
}

// Get retrieves a notifier by ID
func (s *NotifierService) Get(id int64) (*models.Notifier, error) {
	notifier, err := s.notifierRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifier: %w", err)
	}
	return notifier, nil
}

// Update modifies an existing notifier's configuration
func (s *NotifierService) Update(id int, config *models.NotifierConfig) (*models.Notifier, error) {
	notifier, err := s.notifierRepo.Update(id, config)
	if err != nil {
		return nil, fmt.Errorf("failed to update notifier: %w", err)
	}
	return notifier, nil
}

// Delete removes a notifier
func (s *NotifierService) Delete(id int64) error {
	if err := s.notifierRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete notifier: %w", err)
	}
	return nil
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

func (s *NotifierService) HandleSlackCallback(code string, siteId int) (*models.Notifier, error) {
	clientId := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")

	if code == "" || clientId == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing code or client credentials")
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
	}

	resp, err := http.PostForm(SlackTokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	incomingWebhook, ok := result["incoming_webhook"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to get incoming webhook")
	}

	webhookUrl, ok := incomingWebhook["url"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get incoming webhook url")
	}

	notifier := &models.Notifier{
		SiteId: siteId,
		Config: &models.NotifierConfig{
			Type:   models.NotifierTypeSlack,
			Config: json.RawMessage(webhookUrl),
		},
	}

	return notifier, nil
}

func (s *NotifierService) ParseOAuthState(state string) (int, error) {
	parsedState, err := url.ParseQuery(state)

	if err != nil {
		return 0, fmt.Errorf("invalid state format: %w", err)
	}

	siteId, ok := parsedState["site_id"]

	if !ok || len(siteId) <= 0 || siteId[0] == "" {
		return 0, fmt.Errorf("missing site id in state")
	}

	siteIdInt, err := strconv.Atoi(siteId[0])
	if err != nil {
		return 0, fmt.Errorf("invalid site id format: %w", err)
	}

	return siteIdInt, nil
}
