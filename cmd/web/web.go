package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cjsaylor/chessbot/config"
	"github.com/cjsaylor/chessbot/rendering"
)

func main() {
	config, err := config.ParseConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/board", rendering.BoardRenderHandler{})
	log.Printf("Listening on port %v\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil))
}
