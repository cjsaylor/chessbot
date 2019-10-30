// Package integration is for integrating the chess game engine into slack callbacks
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessbot/rendering"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/notnil/chess"
)

// SlackHandler will respond to all Slack event callback subscriptions
type SlackHandler struct {
	SigningKey       string
	Hostname         string
	SlackClient      *slack.Client
	AuthStorage      AuthStorage
	GameStorage      game.GameStorage
	ChallengeStorage game.ChallengeStorage
	LinkRenderer     rendering.RenderLink
}

const requestVersion = "v0"

type command uint8

// SlackCommandPatterns is a list of patterns specific to how text is transmitted in the Slack platform.
var slackCommandPatterns = []CommandPattern{
	{
		Type:    Challenge,
		Pattern: regexp.MustCompile("^<@[\\w|\\d]+>.*challenge.*?<@([\\w\\d]+)>.*$"),
	},
	{
		Type:    Move,
		Pattern: regexp.MustCompile("^<@[\\w|\\d]+> .*([a-h][1-8][a-h][1-8][qnrb]?).*$"),
	},
	{
		Type:    Resign,
		Pattern: regexp.MustCompile("^<@[\\w|\\d]+>.*resign.*$"),
	},
	{
		Type:    Takeback,
		Pattern: regexp.MustCompile("^<@[\\w|\\d]+>.*take\\s?back.*$"),
	},
	{
		Type:    Help,
		Pattern: regexp.MustCompile(".*help.*"),
	},
}

// SlackCommandParser is an instance of the command parse specific to Slack platform formatting.
var slackCommandParser = NewCommandParser(slackCommandPatterns)

var colorToHex = map[game.Color]string{
	game.Black: "#000000",
	game.White: "#eeeeee",
}

func (s SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

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

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
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
			botToken, err := s.AuthStorage.GetAuthToken(event.TeamID)
			if err != nil {
				log.Panicln(err)
			}
			s.SlackClient = slack.New(botToken)
		}
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			matched := slackCommandParser.ParseInput(ev.Text)
			if ev.ChannelType == "im" && ev.BotID == "" && matched.Type == Help {
				s.SlackClient.PostMessage(
					ev.Channel,
					slack.MsgOptionText("You can use ChessBot to play Chess with other teammates.", false),
					slack.MsgOptionAttachments(getHelpAttachments()...))
			}
		case *slackevents.AppMentionEvent:
			var gameID string
			if ev.ThreadTimeStamp == "" {
				gameID = ev.TimeStamp
			} else {
				gameID = ev.ThreadTimeStamp
			}
			matched := slackCommandParser.ParseInput(ev.Text)
			switch matched.Type {
			case Unknown:
				s.sendErrorWithHelp(gameID, ev.Channel, "Sorry, I don't understand what you said.")
			case Challenge:
				challengeCommand, _ := matched.ToChallenge()
				s.handleChallengeCommand(gameID, challengeCommand, ev)
			case Move:
				moveCommand, _ := matched.ToMove()
				s.handleMoveCommand(gameID, moveCommand, ev)
			case Resign:
				s.handleResignCommand(gameID, ev)
			case Takeback:
				s.handleTakebackCommand(gameID, ev)
			case Help:
				s.handleHelpCommand(gameID, ev)
			}
		}
	}
}

func (s SlackHandler) handleMoveCommand(gameID string, moveCommand *MoveCommand, ev *slackevents.AppMentionEvent) {
	gm, err := s.GameStorage.RetrieveGame(gameID)
	if err != nil {
		log.Println(err)
		return
	}
	player := gm.TurnPlayer()
	if ev.User != player.ID {
		s.sendError(gameID, ev.Channel, "Please wait for your turn.")
		return
	}
	chessMove, err := gm.Move(moveCommand.LAN)
	if err != nil {
		s.sendError(gameID, ev.Channel, err.Error())
		return
	}
	if err := s.GameStorage.StoreGame(gameID, gm); err != nil {
		s.sendError(gameID, ev.Channel, err.Error())
		return
	}
	link, _ := s.LinkRenderer.CreateLink(gm)
	boardAttachment := slack.Attachment{
		Text:     chessMove.String(),
		ImageURL: link.String(),
		Color:    colorToHex[gm.Turn()],
	}
	if outcome := gm.Outcome(); outcome != chess.NoOutcome {
		s.displayEndGame(gm, ev)
	} else {
		s.SlackClient.PostMessage(
			ev.Channel,
			slack.MsgOptionText(fmt.Sprintf("<@%v>'s (%v) turn.", gm.TurnPlayer().ID, gm.Turn()), false),
			slack.MsgOptionAttachments(boardAttachment),
			slack.MsgOptionTS(ev.TimeStamp))
	}
}

func (s SlackHandler) displayEndGame(gm *game.Game, ev *slackevents.AppMentionEvent) {
	pgnAttachment := slack.Attachment{
		Title:     "Analysis",
		TitleLink: s.Hostname + "/analyze?game_id=" + gm.ID,
		Text:      gm.Export(),
	}
	link, _ := s.LinkRenderer.CreateLink(gm)
	boardAttachment := slack.Attachment{
		Text:     gm.LastMove().String(),
		ImageURL: link.String(),
	}
	s.SlackClient.PostMessage(
		ev.Channel,
		slack.MsgOptionText(gm.ResultText(), false),
		slack.MsgOptionTS(ev.TimeStamp),
		slack.MsgOptionAttachments(boardAttachment, pgnAttachment))
	// @todo persist record to some incremental storage (redis, etc)
}

func (s SlackHandler) handleChallengeCommand(gameID string, command *ChallengeCommand, ev *slackevents.AppMentionEvent) {
	if _, err := s.GameStorage.RetrieveGame(gameID); err == nil {
		s.sendErrorWithHelp(gameID, ev.Channel, "A game already exists in this thread. Try making a new thread.")
		return
	}
	_, _, channelID, err := s.SlackClient.OpenIMChannel(command.ChallengedID)
	if err != nil {
		log.Printf("unable to challenge %v: %v", command.ChallengedID, err)
		s.sendError(gameID, ev.Channel, "Unable to challenge that player.")
		return
	}
	challenge := &game.Challenge{
		ChallengerID: ev.User,
		ChallengedID: command.ChallengedID,
		GameID:       gameID,
		ChannelID:    ev.Channel,
	}
	s.ChallengeStorage.StoreChallenge(challenge)
	s.SlackClient.PostMessage(
		channelID,
		slack.MsgOptionText(fmt.Sprintf("<@%v> has challenged you to a game of chess!", ev.User), false),
		slack.MsgOptionAttachments(slack.Attachment{
			Text:       "Do you accept?",
			Fallback:   "Unable to accept the challenge.",
			CallbackID: "challenge_response",
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Name:  "challenge",
					Text:  "Accept Challenge",
					Type:  "button",
					Value: "accept",
				},
				slack.AttachmentAction{
					Name:  "challenge",
					Text:  "Decline",
					Type:  "button",
					Style: "danger",
					Value: "reject",
				},
			},
		},
		))
	s.SlackClient.PostEphemeral(ev.Channel, ev.User, slack.MsgOptionText("Challenge has been sent.", false))
}

func (s SlackHandler) handleResignCommand(gameID string, ev *slackevents.AppMentionEvent) {
	gm, err := s.GameStorage.RetrieveGame(gameID)
	if err != nil {
		log.Println(err)
		return
	}
	player, err := gm.PlayerByID(ev.User)
	if err != nil {
		s.sendError(gameID, ev.Channel, "I couldn't find you as part of this game.")
		return
	}
	gm.Resign(*player)
	s.displayEndGame(gm, ev)
}

func (s SlackHandler) handleTakebackCommand(gameID string, ev *slackevents.AppMentionEvent) {
	gm, err := s.GameStorage.RetrieveGame(gameID)
	if err != nil {
		log.Println(err)
		return
	}
	player, err := gm.PlayerByID(ev.User)
	if err != nil {
		s.sendError(gameID, ev.Channel, "I couldn't find you as part of this game.")
		return
	}
	chessMove, err := gm.Takeback(player)
	if err != nil {
		s.sendError(gameID, ev.Channel, fmt.Sprintf("Take back request failed: %v", err))
		return
	}
	link, _ := s.LinkRenderer.CreateLink(gm)
	boardAttachment := slack.Attachment{
		ImageURL: link.String(),
		Color:    colorToHex[gm.Turn()],
	}
	if chessMove != nil {
		boardAttachment.Text = chessMove.String()
	}
	if err := s.GameStorage.StoreGame(gameID, gm); err != nil {
		s.sendError(gameID, ev.Channel, err.Error())
		return
	}
	s.SlackClient.PostMessage(
		ev.Channel,
		slack.MsgOptionText(fmt.Sprintf("<@%v> requested a take back, it is now <@%v>'s turn again.", player.ID, gm.TurnPlayer().ID), false),
		slack.MsgOptionAttachments(boardAttachment),
		slack.MsgOptionTS(ev.TimeStamp))
}

func getHelpAttachments() []slack.Attachment {
	return []slack.Attachment{
		slack.Attachment{
			Title: "Challenge Player",
			Text:  "To challenge a player, mention @chessbot and say \"challenge @player_to_challenge\".",
		},
		slack.Attachment{
			Title: "Making a move",
			Text:  "To make a move playing, mention @chessbot and say \"d2d4\" which are the grid position of the piece you wish to move and the destination.",
		},
		slack.Attachment{
			Pretext:   "For additional help visit our website.",
			Title:     "ChessBot Help",
			TitleLink: "https://www.chris-saylor.com/chessbot",
		},
	}
}

func (s SlackHandler) handleHelpCommand(gameID string, ev *slackevents.AppMentionEvent) {
	text := slack.MsgOptionText("You can use ChessBot to play Chess with other teammates.", false)
	if ev.ThreadTimeStamp == "" || ev.TimeStamp == ev.ThreadTimeStamp {
		s.SlackClient.PostMessage(
			ev.Channel,
			text,
			slack.MsgOptionAttachments(getHelpAttachments()...))
		return
	}
	s.SlackClient.PostMessage(
		ev.Channel,
		text,
		slack.MsgOptionTS(gameID),
		slack.MsgOptionAttachments(getHelpAttachments()...))
}

func (s SlackHandler) sendError(gameID string, channel string, text string) {
	_, _, err := s.SlackClient.PostMessage(
		channel,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(gameID))
	if err != nil {
		log.Println(err)
	}
}

func (s SlackHandler) sendErrorWithHelp(gameID string, channel string, text string) {
	_, _, err := s.SlackClient.PostMessage(
		channel,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(gameID),
		slack.MsgOptionAttachments(getHelpAttachments()...))
	if err != nil {
		log.Println(err)
	}
}
