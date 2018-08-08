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
	initialState := os.Args[1]
	var gm game.Game
	if string(initialState) != "" {
		var err error
		gm, err = game.NewGameFromFEN(string(initialState))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		gm = game.NewGame()
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(gm)
	fmt.Println(gm.Turn())
	fmt.Print("\n> ")
	for scanner.Scan() {
		input := scanner.Text()
		if input == "fen" {
			fmt.Println(gm.FEN())
			fmt.Print("\n> ")
			continue
		}
		err := gm.Move(input)
		if err != nil {
			fmt.Println(err)
			fmt.Print("\n> ")
			continue
		}
		fmt.Print(gm)
		if outcome := gm.Outcome(); outcome != "*" {
			fmt.Println(gm.ResultText())
			os.Exit(0)
		}
		fmt.Println(gm.Turn())
		fmt.Print("\n> ")
	}

}
