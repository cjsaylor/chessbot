package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/slack"
	"github.com/cjsaylor/slack/slackevents"
)

type SlackHandler struct {
	BotToken          string
	VerificationToken string
	SigningKey        string
	Hostname          string
	SlackClient       *slack.Client
}

const requestVersion = "v0"

var games map[string]*game.Game

func init() {
	games = make(map[string]*game.Game)
}

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
			gm := games[ev.ThreadTimeStamp]
			if gm == nil {
				gm := game.NewGame(game.Player{
					ID: ev.User,
				}, game.Player{
					ID: ev.User,
				})
				games[ev.TimeStamp] = gm
				s.SlackClient.PostMessage(ev.Channel, fmt.Sprintf("<@%v>'s (%v) turn.", gm.Players[gm.Turn()].ID, gm.Turn()), slack.PostMessageParameters{
					ThreadTimestamp: ev.TimeStamp,
					Attachments: []slack.Attachment{
						slack.Attachment{
							Text:     "Opening",
							ImageURL: fmt.Sprintf("%v/board?fen=%v", s.Hostname, url.QueryEscape(gm.FEN())),
						},
					},
				})
				break
			}
			input := strings.Split(ev.Text, " ")
			if ev.User != gm.Players[gm.Turn()].ID {
				log.Println("ignoreing player input as it is not their turn")
			}
			err := gm.Move(input[1])
			if err != nil {
				s.SlackClient.PostMessage(ev.Channel, err.Error(), slack.PostMessageParameters{
					ThreadTimestamp: ev.TimeStamp,
				})
				break
			}
			s.SlackClient.PostMessage(ev.Channel, fmt.Sprintf("<@%v>'s (%v) turn.", gm.Players[gm.Turn()].ID, gm.Turn()), slack.PostMessageParameters{
				ThreadTimestamp: ev.TimeStamp,
				Attachments: []slack.Attachment{
					slack.Attachment{
						Text:     "TODO - Put last move here",
						ImageURL: fmt.Sprintf("%v/board?fen=%v", s.Hostname, url.QueryEscape(gm.FEN())),
					},
				},
				LinkNames: 1,
			})
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
