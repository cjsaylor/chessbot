package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessbot/integration"
)

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(97, 122))
	}
	return string(bytes)
}

const (
	export integration.CommandType = iota + integration.Help
	fen
	exit
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Game REPL")
	fmt.Println("Note: piece colors may appear reversed on dark background terminals.")
	gameID := randomString(20)
	store, _ := game.NewSqliteStore("./chessbot.db")
	fmt.Println("Game ID: " + gameID)
	initialState := ""
	if len(os.Args) > 1 {
		initialState = os.Args[1]
	}
	var gm *game.Game
	players := []game.Player{
		{ID: "player1"},
		{ID: "player2"},
	}
	if string(initialState) != "" {
		var err error
		gm, err = game.NewGameFromFEN(gameID, string(initialState), players...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		gm = game.NewGame(gameID, players...)

	}
	inputParser := integration.NewCommandParser([]integration.CommandPattern{
		{
			Type:    integration.Move,
			Pattern: regexp.MustCompile("^.*([a-h][1-8][a-h][1-8][qnrb]?).*$"),
		},
		{
			Type:    integration.Resign,
			Pattern: regexp.MustCompile("^.*resign.*"),
		},
		{
			Type:    integration.Takeback,
			Pattern: regexp.MustCompile("^.*takeback.*"),
		},
		{
			Type:    export,
			Pattern: regexp.MustCompile("^.*export.*$"),
		},
		{
			Type:    fen,
			Pattern: regexp.MustCompile("^.*fen.*$"),
		},
		{
			Type:    exit,
			Pattern: regexp.MustCompile("^.*exit.*$"),
		},
	})
	store.StoreGame(gameID, gm)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(gm)
	fmt.Printf("%v's turn (%v)\n", gm.TurnPlayer().ID, gm.Turn())
	fmt.Print("\n> ")
	for scanner.Scan() {
		input := scanner.Text()
		matchedCommand := inputParser.ParseInput(input)
		gm, err := store.RetrieveGame(gameID)
		if err != nil {
			fmt.Println("Error reading in game: ", err)
		}
		switch matchedCommand.Type {
		case fen:
			fmt.Println(gm.FEN())
			fmt.Print("\n> ")
			continue
		case export:
			fmt.Println(gm.Export())
			fmt.Print("\n> ")
			continue
		case exit:
			os.Exit(0)
		case integration.Move:
			moveCommand, err := matchedCommand.ToMove()
			if err != nil {
				fmt.Println(err)
				fmt.Print("\n> ")
				continue
			}
			_, err = gm.Move(moveCommand.LAN)
			if err != nil {
				fmt.Println(err)
				fmt.Print("\n> ")
				continue
			}
		case integration.Resign:
			gm.Resign(gm.TurnPlayer())
		case integration.Takeback:
			currentTurnPlayer := gm.TurnPlayer()
			var takebackPlayer game.Player
			for _, player := range players {
				if player.ID != currentTurnPlayer.ID {
					takebackPlayer = player
				}
			}
			if _, err := gm.Takeback(&takebackPlayer); err != nil {
				fmt.Println(err)
				fmt.Print("\n> ")
				continue
			}
		}

		store.StoreGame(gameID, gm)
		fmt.Println(gm)
		if outcome := gm.Outcome(); outcome != "*" {
			fmt.Println(gm.ResultText())
			os.Exit(0)
		}
		fmt.Printf("%v's turn (%v)\n", gm.TurnPlayer().ID, gm.Turn())
		fmt.Print("\n> ")
	}

}
