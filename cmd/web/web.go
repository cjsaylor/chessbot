package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cjsaylor/chessbot/config"
	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessbot/integration"
	"github.com/cjsaylor/chessbot/rendering"
)

func main() {
	config, err := config.ParseConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	var gameStorage game.GameStorage
	var challengeStorage game.ChallengeStorage
	if config.SqlitePath != "" {
		sqlStore, err := game.NewSqliteStore(config.SqlitePath)
		if err != nil {
			log.Fatal(err)
		}
		gameStorage = sqlStore
		challengeStorage = sqlStore
	} else {
		memoryStore := game.NewMemoryStore()
		gameStorage = memoryStore
		challengeStorage = memoryStore
	}
	renderLink := rendering.NewRenderLink(config.Hostname, config.SigningKey)
	http.Handle("/board", rendering.BoardRenderHandler{
		LinkRenderer: renderLink,
	})
	http.Handle("/slack", integration.SlackHandler{
		BotToken:          config.SlackBotToken,
		VerificationToken: config.SlackVerificationToken,
		Hostname:          config.Hostname,
		GameStorage:       gameStorage,
		ChallengeStorage:  challengeStorage,
		LinkRenderer:      renderLink,
	})
	http.Handle("/slack/action", integration.SlackActionHandler{
		BotToken:          config.SlackBotToken,
		VerificationToken: config.SlackVerificationToken,
		Hostname:          config.Hostname,
		GameStorage:       gameStorage,
		ChallengeStorage:  challengeStorage,
		LinkRenderer:      renderLink,
	})
	log.Printf("Listening on port %v\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil))
}
