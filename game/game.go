package game

import (
	"fmt"

	"github.com/notnil/chess"
)

type Color string

const (
	White Color = "White"
	Black Color = "Black"
)

type Game struct {
	game *chess.Game
}

func NewGame() Game {
	return Game{
		game: chess.NewGame(chess.UseNotation(chess.LongAlgebraicNotation{})),
	}
}

func NewGameFromFEN(fen string) (Game, error) {
	gameState, err := chess.FEN(fen)
	if err != nil {
		return Game{}, err
	}
	return Game{
		game: chess.NewGame(gameState, chess.UseNotation(chess.LongAlgebraicNotation{})),
	}, nil
}

func (g Game) Turn() Color {
	switch g.game.Position().Turn() {
	case chess.White:
		return White
	case chess.Black:
		return Black
	default:
		return White
	}
}

func (g Game) FEN() string {
	return g.game.FEN()
}

func (g Game) PGN() string {
	return g.game.String()
}

func (g Game) Outcome() chess.Outcome {
	return g.game.Outcome()
}

func (g Game) ResultText() string {
	return fmt.Sprintf("Game completed. %s by %s.", g.Outcome(), g.game.Method())
}

func (g Game) Move(gridMove string) error {
	return g.game.MoveStr(gridMove)
}

func (g Game) String() string {
	return g.game.Position().Board().Draw()
}
