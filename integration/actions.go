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
	"regexp"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessbot/rendering"
	"github.com/cjsaylor/slack"
	"github.com/cjsaylor/slack/slackevents"
)

// SlackActionHandler will respond to all Slack integration component requests
type SlackActionHandler struct {
	VerificationToken string
	SigningKey        string
	Hostname          string
	SlackClient       *slack.Client
	AuthStorage       AuthStorage
	GameStorage       game.GameStorage
	ChallengeStorage  game.ChallengeStorage
	LinkRenderer      rendering.RenderLink
}

func (s SlackActionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	if len(body) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, _ := url.QueryUnescape(body[8:])
	event, err := slackevents.ParseActionEvent(payload, slackevents.OptionVerifyToken(&slackevents.TokenComparator{
		VerificationToken: s.VerificationToken,
	}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}
	if s.SlackClient == nil {
		botToken, err := s.AuthStorage.GetAuthToken(event.Team.Id)
		if err != nil {
			log.Panicln(err)
		}
		s.SlackClient = slack.New(botToken)
	}
	results := regexp.MustCompile("^<@([\\w|\\d]+).*$").FindStringSubmatch(event.OriginalMessage.Text)
	challenge, err := s.ChallengeStorage.RetrieveChallenge(results[1], event.User.Id)
	if err != nil {
		log.Println(err)
		s.sendResponse(w, event.OriginalMessage, "Challenge automatically declined. We couldn't find it in our system.")
		return
	}
	if event.Actions[0].Value != "accept" {
		s.SlackClient.PostMessage(challenge.ChannelID, "Challenge declined by player.", slack.PostMessageParameters{
			ThreadTimestamp: challenge.GameID,
		})
		s.sendResponse(w, event.OriginalMessage, "Declined.")
		if err := s.ChallengeStorage.RemoveChallenge(challenge.ChallengerID, challenge.ChallengedID); err != nil {
			log.Printf("Failed to remove challenge %v: %v\n", challenge, err)
		}
		return
	}
	gameID := challenge.GameID
	gm := game.NewGame(gameID, game.Player{
		ID: event.User.Id,
	}, game.Player{
		ID: results[1],
	})
	s.GameStorage.StoreGame(gameID, gm)
	gm.Start()
	link, _ := s.LinkRenderer.CreateLink(gm)
	s.SlackClient.PostMessage(challenge.ChannelID, fmt.Sprintf("<@%v>'s (%v) turn.", gm.TurnPlayer().ID, gm.Turn()), slack.PostMessageParameters{
		ThreadTimestamp: gameID,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:     fmt.Sprintf("<@%v> has accepted. Here is the opening.", event.User.Id),
				ImageURL: link.String(),
			},
		},
	})
	s.sendResponse(w, event.OriginalMessage, ":ok: Game begun!")
	if err := s.ChallengeStorage.RemoveChallenge(challenge.ChallengerID, challenge.ChallengedID); err != nil {
		log.Printf("Failed to remove challenge %v: %v\n", challenge, err)
	}
}

func (s SlackActionHandler) sendResponse(w http.ResponseWriter, original slack.Message, text string) {
	original.ReplaceOriginal = true
	original.Attachments[0].Actions = []slack.AttachmentAction{}
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: text,
			Value: "",
			Short: false,
		},
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
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
