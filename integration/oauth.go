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

	var teamID string
	var token string
	resp, err := slack.GetOAuthV2Response(http.DefaultClient, s.SlackClientID, s.SlackClientSecret, code, "")
	// During transition to slack's oauth v2.
	if err.Error() == "oauth_authorization_url_mismatch" {
		log.Println(err)
		log.Println("Attempting fallback to v1 of Slack's oauth")
		respV1, err := slack.GetOAuthResponse(http.DefaultClient, s.SlackClientID, s.SlackClientSecret, code, "")
		if err == nil {
			teamID = respV1.TeamID
			token = respV1.Bot.BotAccessToken
		}
	} else if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("There was an issue exchanging tokens with Slack. Please try again."))
		return
	} else {
		teamID = resp.Team.ID
		token = resp.AccessToken
	}

	err = s.AuthStore.StoreAuthToken(teamID, token)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was an issue exchanging tokens with Slack. Please try again."))
		return
	}
	http.Redirect(w, r, fmt.Sprintf("https://slack.com/app_redirect?app=%v", s.SlackAppID), http.StatusSeeOther)
}
