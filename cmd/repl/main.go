package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cjsaylor/chessbot/game"
)

const foolsMate = "rnbqkbnr/pppp1ppp/8/8/5P2/8/PPPPP2P/RNBQKBNR b KQkq - 03"

func main() {
	fmt.Println("Game REPL")
	fmt.Println("Note: piece colors may appear reversed on dark background terminals.")
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
		gm, err = game.NewGameFromFEN("1", string(initialState), players...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		gm = game.NewGame("1", players...)

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
		fmt.Println(gm)
		if outcome := gm.Outcome(); outcome != "*" {
			fmt.Println(gm.ResultText())
			os.Exit(0)
		}
		fmt.Printf("%v's turn (%v)\n", gm.TurnPlayer().ID, gm.Turn())
		fmt.Print("\n> ")
	}

}
