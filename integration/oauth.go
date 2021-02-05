package integration

import (
	"fmt"
	"log"
	"net/http"

	"github.com/slack-go/slack"
)

// SlackOauthHandler will respond to all Slack authorization callbacks
type SlackOauthHandler struct {
	SlackClientID     string
	SlackClientSecret string
	SlackAppID        string
	AuthStore         AuthStorage
}

// SlackV2OauthHandler will respond to all Slack V2 authorization callbacks
type SlackV2OauthHandler struct {
	SlackClientID     string
	SlackClientSecret string
	SlackAppID        string
	AuthStore         AuthStorage
}

// ServeHTTP handles oauth token exchange requests for Slack's V1 oauth2 implementation.
// This is currently deprecated
func (s SlackOauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("invalid oauth code requested")
		return
	}
	resp, err := slack.GetOAuthResponse(http.DefaultClient, s.SlackClientID, s.SlackClientSecret, code, "")
	if err != nil {
		log.Println(err)
		return
	}
	err = s.AuthStore.StoreAuthToken(resp.TeamID, resp.Bot.BotAccessToken)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	http.Redirect(w, r, fmt.Sprintf("https://slack.com/app_redirect?app=%v", s.SlackAppID), http.StatusSeeOther)
}

// ServiceHTTP handles oauth token exchange requests for Slack's V2 oauth2 implementation
func (s SlackV2OauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("invalid oauth code requested")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request: code is required to exchange tokens."))
		return
	}
	resp, err := slack.GetOAuthV2Response(http.DefaultClient, s.SlackClientID, s.SlackClientSecret, code, "")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("There was a problem exchanging code for token. Please try again."))
		return
	}
	err = s.AuthStore.StoreAuthToken(resp.Team.ID, resp.AccessToken)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was a problem storing your authentication token. Please try again."))
	}
	http.Redirect(w, r, fmt.Sprintf("https://slack.com/app_redirect?app=%v", s.SlackAppID), http.StatusSeeOther)
}
