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
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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
	TakebackStorage  game.TakebackStorage
	LinkRenderer     rendering.RenderLink
}

// HandleChallenge does the necessary operations for action responses to player challenges.
func (s SlackActionHandler) HandleChallenge(w http.ResponseWriter, event slackevents.MessageAction) {
	results := challengerPattern.FindStringSubmatch(event.OriginalMessage.Text)
	challenge, err := s.ChallengeStorage.RetrieveChallenge(results[1], event.User.ID)
	if err != nil {
		log.Println(err)
		s.sendResponse(w, event.OriginalMessage, "Challenge automatically declined. We couldn't find it in our system.")
		return
	}
	if event.Actions[0].Value != "accept" {
		s.SlackClient.PostMessage(
			challenge.ChannelID,
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

// HandleTakeback performs necessary operations for action responses to player takeback requests.
func (s SlackActionHandler) HandleTakeback(w http.ResponseWriter, event slackevents.MessageAction) {
	// always remove the ephemeral message
	defer func() {
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ResponseType    string `json:"response_type"`
			Text            string `json:"text"`
			ReplaceOriginal bool   `json:"replace_original"`
			DeleteOriginal  bool   `json:"delete_original"`
		}{
			"ephemeral",
			"",
			true,
			true,
		})
	}()
	gameID := event.Actions[0].Name
	takeback, err := s.TakebackStorage.RetrieveTakeback(gameID)
	if err != nil && event.Actions[0].Value != "decline" {
		s.sendError(gameID, event.Channel.ID, "Could not verify the takeback request.")
		log.Printf("Takeback request failed: %v", err)
		return
	}
	if event.Actions[0].Value == "decline" {
		if err := s.TakebackStorage.RemoveTakeback(takeback); err != nil {
			log.Printf("Failed to remove takeback %v: %v\n", gameID, err)
		}
		s.sendError(gameID, event.Channel.ID, "Takeback request declined by player.")
		return
	}
	if !takeback.IsValidTakeback() {
		s.sendError(gameID, event.Channel.ID, "Takeback request is no longer valid.")
		return
	}
	approvingPlayer, err := takeback.CurrentGame.PlayerByID(event.User.ID)
	if err != nil {
		s.sendError(gameID, event.Channel.ID, fmt.Sprintf("Take back request failed: %v", err))
		return
	}
	player := takeback.CurrentGame.OtherPlayer(approvingPlayer)
	chessMove, err := takeback.CurrentGame.Takeback(player)
	if err != nil {
		s.sendError(gameID, event.Channel.ID, fmt.Sprintf("Take back request failed: %v", err))
		return
	}
	link, _ := s.LinkRenderer.CreateLink(takeback.CurrentGame)
	boardAttachment := slack.Attachment{
		ImageURL: link.String(),
		Color:    colorToHex[takeback.CurrentGame.Turn()],
	}
	if chessMove != nil {
		boardAttachment.Text = chessMove.String()
	}
	if err := s.GameStorage.StoreGame(gameID, takeback.CurrentGame); err != nil {
		s.sendError(gameID, event.Channel.ID, err.Error())
		return
	}
	s.SlackClient.PostMessage(
		event.Channel.ID,
		slack.MsgOptionText(fmt.Sprintf(
			"<@%v> requested a take back, it is now <@%v>'s turn again.",
			player.ID,
			takeback.CurrentGame.TurnPlayer().ID,
		), false),
		slack.MsgOptionAttachments(boardAttachment),
		slack.MsgOptionTS(gameID))
	if err := s.TakebackStorage.RemoveTakeback(takeback); err != nil {
		log.Printf("Failed to remove takeback %v: %v\n", gameID, err)
	}
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
	if event.Type != "interactive_message" {
		s.sendResponse(w, event.OriginalMessage, "Invalid action.")
		return
	}
	switch event.CallbackID {
	case "challenge_response":
		s.HandleChallenge(w, event)
	case "takeback_response":
		s.HandleTakeback(w, event)
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

func (s SlackActionHandler) sendError(gameID string, channel string, text string) {
	_, _, err := s.SlackClient.PostMessage(
		channel,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(gameID))
	if err != nil {
		log.Println(err)
	}
}
