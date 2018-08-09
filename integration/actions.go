package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/slack"
	"github.com/cjsaylor/slack/slackevents"
)

type SlackActionHandler struct {
	BotToken          string
	VerificationToken string
	SigningKey        string
	Hostname          string
	SlackClient       *slack.Client
	GameStorage       game.GameStorage
	ChallengeStorage  game.ChallengeStorage
}

func (s SlackActionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	payload, _ := url.QueryUnescape(body[8:])
	event, err := slackevents.ParseActionEvent(payload, slackevents.OptionVerifyToken(&slackevents.TokenComparator{s.VerificationToken}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}
	if event.Actions[0].Value != "accept" {
		return
	}
	if s.SlackClient == nil {
		s.SlackClient = slack.New(s.BotToken)
	}
	results := regexp.MustCompile("^<@([\\w|\\d]+).*$").FindStringSubmatch(event.OriginalMessage.Text)
	challenge, err := s.ChallengeStorage.RetrieveChallenge(results[1], event.User.Id)
	if err != nil {
		log.Println(err)
		return
	}
	gameID := challenge.GameID
	gm := game.NewGame(game.Player{
		ID: event.User.Id,
	}, game.Player{
		ID: results[1],
	})
	s.GameStorage.StoreGame(gameID, gm)
	gm.Start()
	s.SlackClient.PostMessage(challenge.ChannelID, fmt.Sprintf("<@%v>'s (%v) turn.", gm.TurnPlayer().ID, gm.Turn()), slack.PostMessageParameters{
		ThreadTimestamp: gameID,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:     fmt.Sprintf("<@%v> has accepted. Here is the opening.", event.User.Id),
				ImageURL: fmt.Sprintf("%v/board?game_id=%v&ts=%v", s.Hostname, gameID, event.ActionTimestamp),
			},
		},
	})
}

// Not using this for now since the challenge request doesn't appear to send it
// Also we'd need to implement this in a form that slackevents.ParseEvent() can use for verification
func (s SlackActionHandler) validateSignature(r *http.Request, body string) bool {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	requestSignature := r.Header.Get("X-Slack-Signature")
	compiled := fmt.Sprintf("%v:%v:%v", requestVersion, timestamp, body)
	mac := hmac.New(sha256.New, []byte(s.SigningKey))
	mac.Write([]byte(compiled))
	expectedSignature := mac.Sum(nil)
	return hmac.Equal(expectedSignature, []byte(requestSignature))
}
