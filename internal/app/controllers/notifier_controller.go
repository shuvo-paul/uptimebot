package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/shuvo-paul/sitemonitor/internal/app/services"
)

type NotifierController struct {
	notifierService services.NotifierServiceInterface
}

func NewNotifierController(notifierService services.NotifierServiceInterface) *NotifierController {
	return &NotifierController{
		notifierService: notifierService,
	}
}

func (c *NotifierController) AuthSlack(w http.ResponseWriter, r *http.Request) {
	siteId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid site ID", http.StatusBadRequest)
		return
	}

	redirectUri := os.Getenv("SLACK_REDIRECT_URI")
	clientId := os.Getenv("SLACK_CLIENT_ID")

	if redirectUri == "" || clientId == "" {
		http.Error(w, "Missing environment variables", http.StatusBadRequest)
		return
	}

	oauthLink := fmt.Sprintf("https://slack.com/oauth/v2/authorize?scope=incoming-webhook&user_scope=&redirect_uri=%s&client_id=%s&state=site_id=%d", redirectUri, clientId, siteId)
	http.Redirect(w, r, oauthLink, http.StatusSeeOther)
}

func (c *NotifierController) AuthSlackCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	siteId, err := c.notifierService.ParseOAuthState(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notifier, err := c.notifierService.HandleSlackCallback(code, siteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.notifierService.Create(notifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/sites/%d", siteId), http.StatusSeeOther)
}
