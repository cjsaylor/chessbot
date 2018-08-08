package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cjsaylor/chessbot/config"
	"github.com/cjsaylor/chessbot/integration"
	"github.com/cjsaylor/chessbot/rendering"
)

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
	})
	log.Printf("Listening on port %v\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil))
}
