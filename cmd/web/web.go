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

var memoryStore *game.MemoryStore

func init() {
	memoryStore = game.NewMemoryStore()
	// @todo implement a store that is persistent
}

func main() {
	config, err := config.ParseConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/board", rendering.BoardRenderHandler{})
	http.Handle("/slack", integration.SlackHandler{
		BotToken:          config.SlackBotToken,
		VerificationToken: config.SlackVerificationToken,
		Hostname:          config.Hostname,
		Storage:           memoryStore,
	})
	log.Printf("Listening on port %v\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil))
}
