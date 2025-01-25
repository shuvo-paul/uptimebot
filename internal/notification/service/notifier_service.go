package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	notifCoer "github.com/shuvo-paul/uptimebot/internal/notification/core"
	"github.com/shuvo-paul/uptimebot/internal/notification/model"
	"github.com/shuvo-paul/uptimebot/internal/notification/provider"
	"github.com/shuvo-paul/uptimebot/internal/notification/repository"
)

type NotifierServiceInterface interface {
	Create(notifier *model.Notifier) error
	Get(id int64) (*model.Notifier, error)
	Update(id int, config json.RawMessage) (*model.Notifier, error)
	Delete(id int64) error
	ConfigureObservers(siteID int) error
	HandleSlackCallback(code string, siteId int) (*model.Notifier, error)
	ParseOAuthState(state string) (int, error)
	GetSubject() *notifCoer.Subject
}

type NotifierService struct {
	notifierRepo repository.NotifierRepositoryInterface
	subject      *notifCoer.Subject
}

var (
	// SlackTokenURL is the Slack OAuth token URL, can be overridden in tests
	SlackTokenURL = "https://slack.com/api/oauth.v2.access"
)

func NewNotifierService(
	notifierRepo repository.NotifierRepositoryInterface,
	subject *notifCoer.Subject,
) *NotifierService {
	if subject == nil {
		subject = notifCoer.NewSubject()
	}
	return &NotifierService{
		notifierRepo: notifierRepo,
		subject:      subject,
	}
}

// Create adds a new notifier
func (s *NotifierService) Create(notifier *model.Notifier) error {
	if _, err := s.notifierRepo.Create(notifier); err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}
	return nil
}

// Get retrieves a notifier by ID
func (s *NotifierService) Get(id int64) (*model.Notifier, error) {
	notifier, err := s.notifierRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifier: %w", err)
	}
	return notifier, nil
}

// Update modifies an existing notifier's configuration
func (s *NotifierService) Update(id int, config json.RawMessage) (*model.Notifier, error) {
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
	s.subject = notifCoer.NewSubject()

	notifiers, err := s.notifierRepo.GetBySiteID(siteID)
	if err != nil {
		return fmt.Errorf("failed to get notifiers: %w", err)
	}

	for _, notifier := range notifiers {
		switch notifier.Type {
		case model.NotifierTypeSlack:
			config, err := notifier.GetSlackConfig()
			if err != nil {
				return fmt.Errorf("failed to get slack config: %w", err)
			}
			observer := provider.NewSlackObserver(config.WebhookURL, http.DefaultClient)
			s.subject.Attach(observer)
		default:
			return fmt.Errorf("unsupported notifier type: %s", notifier.Type)
		}
	}

	return nil
}

func (s *NotifierService) HandleSlackCallback(code string, siteId int) (*model.Notifier, error) {
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

	notifier := &model.Notifier{
		SiteId: siteId,
		Type:   model.NotifierTypeSlack,
		Config: json.RawMessage(`{"webhook_url": "` + webhookUrl + `"}`),
	}

	return notifier, nil
}

func (s *NotifierService) ParseOAuthState(state string) (int, error) {
	parsedState, err := url.ParseQuery(state)

	if err != nil {
		return 0, fmt.Errorf("invalid state format: %w", err)
	}

	siteId, ok := parsedState["target_id"]

	if !ok || len(siteId) <= 0 || siteId[0] == "" {
		return 0, fmt.Errorf("missing site id in state")
	}

	siteIdInt, err := strconv.Atoi(siteId[0])
	if err != nil {
		return 0, fmt.Errorf("invalid site id format: %w", err)
	}

	return siteIdInt, nil
}

func (s *NotifierService) GetSubject() *notifCoer.Subject {
	return s.subject
}
