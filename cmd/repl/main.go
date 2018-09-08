package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/cjsaylor/chessbot/game"
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
		game.Player{ID: "player1"},
		game.Player{ID: "player2"},
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
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(gm)
	fmt.Printf("%v's turn (%v)\n", gm.TurnPlayer().ID, gm.Turn())
	fmt.Print("\n> ")
	for scanner.Scan() {
		input := scanner.Text()
		switch scanner.Text() {
		case "fen":
			fmt.Println(gm.FEN())
			fmt.Print("\n> ")
			continue
		case "export":
			fmt.Println(gm.Export())
			fmt.Print("\n> ")
			continue
		case "resign":
			gm.Resign(gm.TurnPlayer())
		case "exit":
			os.Exit(0)
		default:
			_, err := gm.Move(input)
			if err != nil {
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
