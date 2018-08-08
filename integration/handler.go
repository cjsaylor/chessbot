package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

type SlackHandler struct {
	BotToken          string
	VerificationToken string
	SigningKey        string
	Hostname          string
	SlackClient       *slack.Client
}

const requestVersion = "v0"

func (s SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{s.VerificationToken}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}
	if event.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	} else if event.Type == slackevents.CallbackEvent {
		if s.SlackClient == nil {
			s.SlackClient = slack.New(s.BotToken)
		}
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			log.Print(ev.Text)
			postParams := slack.PostMessageParameters{
				ThreadTimestamp: ev.TimeStamp,
				Attachments: []slack.Attachment{
					slack.Attachment{
						Text:     "Game in progress",
						ImageURL: s.Hostname + "/board?fen=r1b1k2r/pp1n1ppp/2n1p3/q1bpP3/N2N1P2/P3B3/1PP3PP/R2QKB1R%20w%20KQkq%20-%203%2013&last_move=b6%20a5&checked_tile=e1",
					},
				},
			}
			s.SlackClient.PostMessage(ev.Channel, "Yes, hello.", postParams)
		}
	}

}

// Not using this for now since the challenge request doesn't appear to send it
// Also we'd need to implement this in a form that slackevents.ParseEvent() can use for verification
func (s SlackHandler) validateSignature(r *http.Request, body string) bool {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	requestSignature := r.Header.Get("X-Slack-Signature")
	compiled := fmt.Sprintf("%v:%v:%v", requestVersion, timestamp, body)
	mac := hmac.New(sha256.New, []byte(s.SigningKey))
	mac.Write([]byte(compiled))
	expectedSignature := mac.Sum(nil)
	return hmac.Equal(expectedSignature, []byte(requestSignature))
}
