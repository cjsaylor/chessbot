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
