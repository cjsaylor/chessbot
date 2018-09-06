// Package integration is for integrating the chess game engine into slack callbacks
package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessbot/rendering"
	"github.com/cjsaylor/slack"
	"github.com/cjsaylor/slack/slackevents"
	"github.com/notnil/chess"
)

// SlackHandler will respond to all Slack event callback subscriptions
type SlackHandler struct {
	VerificationToken string
	SigningKey        string
	Hostname          string
	SlackClient       *slack.Client
	AuthStorage       AuthStorage
	GameStorage       game.GameStorage
	ChallengeStorage  game.ChallengeStorage
	LinkRenderer      rendering.RenderLink
}

const requestVersion = "v0"

type command uint8

const (
	unknownCommand command = iota
	challengeCommand
	moveCommand
	resignCommand
	helpCommand
)

var commandPatterns = map[command]*regexp.Regexp{
	challengeCommand: regexp.MustCompile("^<@[\\w|\\d]+>.*challenge <@([\\w\\d]+)>.*$"),
	moveCommand:      regexp.MustCompile("^<@[\\w|\\d]+> .*([a-h][1-8][a-h][1-8]).*$"),
	resignCommand:    regexp.MustCompile("^<@[\\w|\\d]+>.*resign.*$"),
	helpCommand:      regexp.MustCompile(".*help.*"),
}

var colorToHex = map[game.Color]string{
	game.Black: "#000000",
	game.White: "#eeeeee",
}

func (s SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{
		VerificationToken: s.VerificationToken,
	}))
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
		case *slackevents.AppMentionEvent:
			var gameID string
			if ev.ThreadTimeStamp == "" {
				gameID = ev.TimeStamp
			} else {
				gameID = ev.ThreadTimeStamp
			}
			handler := unknownCommand
			var captures []string
			for command, regex := range commandPatterns {
				results := regex.FindStringSubmatch(ev.Text)
				if len(results) > 0 {
					handler = command
					captures = results[1:]
				}
			}
			switch handler {
			case unknownCommand:
				s.sendError(gameID, ev.Channel, "Does not compute. :(")
			case moveCommand:
				s.handleMoveCommand(gameID, captures[0], ev)
			case challengeCommand:
				s.handleChallengeCommand(gameID, captures[0], ev)
			case resignCommand:
				s.handleResignCommand(gameID, ev)
			case helpCommand:
				s.handleHelpCommand(gameID, ev)
			}
		}
	}
}

func (s SlackHandler) handleMoveCommand(gameID string, move string, ev *slackevents.AppMentionEvent) {
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
	chessMove, err := gm.Move(move)
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
		s.SlackClient.PostMessage(ev.Channel, fmt.Sprintf("<@%v>'s (%v) turn.", gm.TurnPlayer().ID, gm.Turn()), slack.PostMessageParameters{
			ThreadTimestamp: ev.TimeStamp,
			Attachments:     []slack.Attachment{boardAttachment},
		})
	}
}

func (s SlackHandler) displayEndGame(gm *game.Game, ev *slackevents.AppMentionEvent) {
	fenAttachment := slack.Attachment{
		Title: "FEN",
		Text:  gm.FEN(),
	}
	pgnAttachment := slack.Attachment{
		Title: "PGN",
		Text:  gm.PGN(),
	}
	link, _ := s.LinkRenderer.CreateLink(gm)
	boardAttachment := slack.Attachment{
		Text:     gm.LastMove().String(),
		ImageURL: link.String(),
	}
	s.SlackClient.PostMessage(ev.Channel, gm.ResultText(), slack.PostMessageParameters{
		ThreadTimestamp: ev.TimeStamp,
		Attachments:     []slack.Attachment{boardAttachment, fenAttachment, pgnAttachment},
	})
	// @todo persist record to some incremental storage (redis, etc)
}

func (s SlackHandler) handleChallengeCommand(gameID string, challengedUser string, ev *slackevents.AppMentionEvent) {
	if _, err := s.GameStorage.RetrieveGame(gameID); err == nil {
		s.SlackClient.PostMessage(ev.Channel, "A game already exists in this thread. Try making a new thread.", slack.PostMessageParameters{
			ThreadTimestamp: gameID,
		})
		return
	}
	_, _, channelID, err := s.SlackClient.OpenIMChannel(challengedUser)
	if err != nil {
		log.Printf("unable to challenge %v: %v", challengedUser, err)
		s.sendError(gameID, ev.Channel, "Unable to challenge that player. :(")
		return
	}
	challenge := &game.Challenge{
		ChallengerID: ev.User,
		ChallengedID: challengedUser,
		GameID:       gameID,
		ChannelID:    ev.Channel,
	}
	s.ChallengeStorage.StoreChallenge(challenge)
	s.SlackClient.PostMessage(channelID, fmt.Sprintf("<@%v> has challenged you to a game of chess!", ev.User), slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			slack.Attachment{
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
		},
	})
}

func (s SlackHandler) handleResignCommand(gameID string, ev *slackevents.AppMentionEvent) {
	gm, err := s.GameStorage.RetrieveGame(gameID)
	if err != nil {
		log.Println(err)
		return
	}
	player, err := gm.PlayerByID(ev.User)
	gm.Resign(*player)
	s.displayEndGame(gm, ev)
}

func (s SlackHandler) handleHelpCommand(gameID string, ev *slackevents.AppMentionEvent) {
	text := "You can use ChessBot to play Chess with other teammates."
	s.SlackClient.PostMessage(ev.Channel, text, slack.PostMessageParameters{
		ThreadTimestamp: gameID,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Pretext:    "To start a new game, challenge another teammate by mentioning them.",
				Title:      "Start a new game",
				Text:       "Example: `@ChessBot challenge @<target_player>`",
				MarkdownIn: []string{"text"},
			},
			slack.Attachment{
				Pretext:    "To move a piece during a game, put in the square of the piece you want to move and the destination.",
				Title:      "Move a piece",
				Text:       "Example: `@ChessBot e2e4`",
				MarkdownIn: []string{"text"},
			},
			slack.Attachment{
				Pretext:    "To resign the game, mention the chess bot and say \"resign\"",
				Title:      "Resignation",
				Text:       "Example: `@ChessBot resign",
				MarkdownIn: []string{"text"},
			},
		},
	})
}

func (s SlackHandler) sendError(gameID string, channel string, text string) {
	_, _, err := s.SlackClient.PostMessage(channel, text, slack.PostMessageParameters{
		ThreadTimestamp: gameID,
	})
	if err != nil {
		log.Println(err)
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
