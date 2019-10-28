package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessbot/rendering"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

var challengerPattern = regexp.MustCompile("^<@([\\w|\\d]+).*$")

// SlackActionHandler will respond to all Slack integration component requests
type SlackActionHandler struct {
	SigningKey       string
	Hostname         string
	SlackClient      *slack.Client
	AuthStorage      AuthStorage
	GameStorage      game.GameStorage
	ChallengeStorage game.ChallengeStorage
	LinkRenderer     rendering.RenderLink
}

func (s SlackActionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	if len(body) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	secretsVerifier, err := slack.NewSecretsVerifier(r.Header, s.SigningKey)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	secretsVerifier.Write([]byte(body))
	if err := secretsVerifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	payload, _ := url.QueryUnescape(body[8:])
	event, err := slackevents.ParseActionEvent(payload, slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}
	if s.SlackClient == nil {
		botToken, err := s.AuthStorage.GetAuthToken(event.Team.ID)
		if err != nil {
			log.Panicln(err)
		}
		s.SlackClient = slack.New(botToken)
	}
	if event.Type != "interactive_message" && event.CallbackID != "challenge_response" {
		s.sendResponse(w, event.OriginalMessage, "Invalid action.")
		return
	}
	results := challengerPattern.FindStringSubmatch(event.OriginalMessage.Text)
	challenge, err := s.ChallengeStorage.RetrieveChallenge(results[1], event.User.ID)
	if err != nil {
		log.Println(err)
		s.sendResponse(w, event.OriginalMessage, "Challenge automatically declined. We couldn't find it in our system.")
		return
	}
	if event.Actions[0].Value != "accept" {
		s.SlackClient.PostMessage(
			challenge.GameID,
			slack.MsgOptionText("Challenge declined by player.", false),
			slack.MsgOptionTS(challenge.GameID))
		s.sendResponse(w, event.OriginalMessage, "Declined.")
		if err := s.ChallengeStorage.RemoveChallenge(challenge.ChallengerID, challenge.ChallengedID); err != nil {
			log.Printf("Failed to remove challenge %v: %v\n", challenge, err)
		}
		return
	}
	gameID := challenge.GameID
	gm := game.NewGame(gameID, game.Player{
		ID: event.User.ID,
	}, game.Player{
		ID: challenge.ChallengerID,
	})
	s.GameStorage.StoreGame(gameID, gm)
	gm.Start()
	link, _ := s.LinkRenderer.CreateLink(gm)
	s.SlackClient.PostMessage(
		challenge.ChannelID,
		slack.MsgOptionText(fmt.Sprintf("<@%v>'s (%v) turn.", gm.TurnPlayer().ID, gm.Turn()), false),
		slack.MsgOptionTS(gameID),
		slack.MsgOptionAttachments(slack.Attachment{
			Text:     fmt.Sprintf("<@%v> has accepted. Here is the opening.", event.User.ID),
			ImageURL: link.String(),
		}))
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
