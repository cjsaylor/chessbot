// Package game contains all the logic for creating and manipulating a game
package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/notnil/chess"
)

// Challenge represents a challenge between two players
type Challenge struct {
	ChallengerID string
	ChallengedID string
	GameID       string
	ChannelID    string
}

type Color string

// White represents the color of the white set.
// Black represents the color of the black set.
// TakebackThreshold represents number of minutes to allow a takeback
const (
	White             Color         = "White"
	Black             Color         = "Black"
	TakebackThreshold time.Duration = time.Minute
)

var colorMap = map[Color]chess.Color{
	White: chess.White,
	Black: chess.Black,
}

// TimeProvider is a closure that returns the current time as determined by the provider
type TimeProvider func() time.Time

var defaultTimeProvider TimeProvider = func() time.Time {
	return time.Now()
}

// Player represents a human Chess player
type Player struct {
	ID    string
	color Color
}

// Game is the state of a game (active or not)
type Game struct {
	ID           string
	game         *chess.Game
	Players      map[Color]Player
	started      bool
	lastMoved    time.Time
	checkedTile  *chess.Square
	timeProvider TimeProvider
}

// NewGame will create a new game with typical starting positions
func NewGame(ID string, players ...Player) *Game {
	gm := &Game{
		ID:           ID,
		game:         chess.NewGame(chess.UseNotation(chess.LongAlgebraicNotation{})),
		lastMoved:    time.Time{},
		timeProvider: defaultTimeProvider,
	}
	attachPlayers(gm, players...)
	return gm
}

func attachPlayers(game *Game, players ...Player) {
	playerList := []Player{}
	playerList = append(playerList, players...)
	rand.Shuffle(2, func(i, j int) {
		playerList[i], playerList[j] = playerList[j], playerList[i]
	})
	playerList[0].color = White
	playerList[1].color = Black
	game.Players = map[Color]Player{
		White: playerList[0],
		Black: playerList[1],
	}
}

// NewGameFromFEN will create a new game with a given FEN starting position
func NewGameFromFEN(ID string, fen string, players ...Player) (*Game, error) {
	gameState, err := chess.FEN(fen)
	if err != nil {
		return &Game{}, err
	}
	game := &Game{
		ID:           ID,
		game:         chess.NewGame(gameState, chess.UseNotation(chess.LongAlgebraicNotation{})),
		timeProvider: defaultTimeProvider,
		lastMoved:    time.Time{},
		started:      true,
	}
	attachPlayers(game, players...)
	return game, nil
}

func NewGameFromPGN(ID string, pgn string, white Player, black Player) (*Game, error) {
	reader := strings.NewReader(pgn)
	gameState, err := chess.PGN(reader)
	if err != nil {
		return &Game{}, err
	}
	game := &Game{
		ID:           ID,
		game:         chess.NewGame(gameState, chess.UseNotation(chess.LongAlgebraicNotation{})),
		lastMoved:    time.Time{},
		timeProvider: defaultTimeProvider,
	}
	game.Players = make(map[Color]Player)
	white.color = White
	game.Players[White] = white
	black.color = Black
	game.Players[Black] = black
	return game, nil
}

// PlayerByID returns a reference to a player given their ID
func (g *Game) PlayerByID(ID string) (*Player, error) {
	for _, player := range g.Players {
		if player.ID == ID {
			return &player, nil
		}
	}
	return nil, errors.New("player not found with given ID")
}

// Resign will resign a player from the game
func (g *Game) Resign(resigner Player) {
	g.game.Resign(colorMap[resigner.color])
}

// TurnPlayer returns which player should move next
func (g *Game) TurnPlayer() Player {
	return g.Players[g.Turn()]
}

// Turn returns which color should move next
func (g *Game) Turn() Color {
	switch g.game.Position().Turn() {
	case chess.White:
		return White
	case chess.Black:
		return Black
	default:
		return White
	}
}

// FEN serializer
func (g *Game) FEN() string {
	return g.game.FEN()
}

// PGN serializer
func (g *Game) PGN() string {
	return g.game.String()
}

// Export a game in PGN format
func (g *Game) Export() string {
	regularNotation := chess.UseNotation(chess.AlgebraicNotation{})
	longNotation := chess.UseNotation(chess.LongAlgebraicNotation{})
	defer longNotation(g.game)
	g.game.AddTagPair("Site", "Slack ChessBot match")
	g.game.AddTagPair("White", g.Players[White].ID)
	g.game.AddTagPair("Black", g.Players[Black].ID)
	regularNotation(g.game)
	return g.game.String()
}

// Outcome determines the outcome of the game (or no outcome)
func (g *Game) Outcome() chess.Outcome {
	return g.game.Outcome()
}

// ResultText will show the outcome of the game in textual format
func (g *Game) ResultText() string {
	outcome := g.Outcome()
	if outcome == chess.Draw {
		return fmt.Sprintf("Game completed. %s by %s.", g.Outcome(), g.game.Method())
	}
	var winningPlayer Player
	if outcome == chess.WhiteWon {
		winningPlayer = g.Players[White]
	} else {
		winningPlayer = g.Players[Black]
	}
	return fmt.Sprintf("Congratulations, <@%v>! %s by %s", winningPlayer.ID, g.Outcome(), g.game.Method())
}

// LastMove returns the last move done of the game
func (g *Game) LastMove() *chess.Move {
	moves := g.game.Moves()
	if len(moves) == 0 {
		return nil
	}
	return moves[len(moves)-1]
}

// LastMoved is the last time a move was made
func (g *Game) LastMoved() time.Time {
	return g.lastMoved
}

// Move a Chess piece based on standard algabreic notation (d2d4, etc)
func (g *Game) Move(san string) (*chess.Move, error) {
	err := g.game.MoveStr(san)
	if err != nil {
		return nil, err
	}
	g.started = true
	g.lastMoved = g.timeProvider()
	return g.LastMove(), nil
}

// Start indicates the game has been started
func (g *Game) Start() {
	g.started = true
}

// Started determines if the game has been started
func (g *Game) Started() bool {
	return g.started
}

// ValidMoves returns a list of all moves available to the current player's turn
func (g *Game) ValidMoves() []*chess.Move {
	return g.game.ValidMoves()
}

// CheckedKing returns the square of a checked king if there is indeed a king in check.
func (g *Game) CheckedKing() chess.Square {
	squareMap := g.game.Position().Board().SquareMap()
	lastMovePiece := squareMap[g.LastMove().S2()]
	for square, piece := range squareMap {
		if piece.Type() == chess.King && piece.Color() == lastMovePiece.Color().Other() {
			return square
		}
	}
	return chess.NoSquare
}

// ErrGameHasNoMoves is an error representing an action that failed due to the game having no moves yet.
var ErrGameHasNoMoves = errors.New("game has no moves yet")

// ErrGameCompleted is an error representing an action that failed due to the game being completed.
var ErrGameCompleted = errors.New("game is already completed")

// ErrPlayerAlreadyMoved is an error representing an action that failed due to a another player move.
var ErrPlayerAlreadyMoved = errors.New("other player has already made a move")

// ErrPastTimeThreshold is an error representing an action that failed due to taking place after a specific threshold
var ErrPastTimeThreshold = fmt.Errorf("exceeded threshold of %v", TakebackThreshold)

// IsPastTakebackThreshold determines if the last move is within the allowable takeback time period.
func (g *Game) IsPastTakebackThreshold() bool {
	return g.timeProvider().Sub(g.LastMoved()) > TakebackThreshold
}

// Takeback reverts the game to the previous move prior to the last move.
// Note: If the first move of the game is taken back, the resulting move will be nil
func (g *Game) Takeback(requestingPlayer *Player) (*chess.Move, error) {
	if g.LastMove() == nil {
		return nil, ErrGameHasNoMoves
	}
	if g.Outcome() != chess.NoOutcome {
		return nil, ErrGameCompleted
	}
	turnPlayer := g.TurnPlayer()
	if requestingPlayer.ID == turnPlayer.ID {
		return nil, ErrPlayerAlreadyMoved
	}
	newGame := chess.NewGame(chess.UseNotation(chess.LongAlgebraicNotation{}))
	moves := g.game.Moves()
	withoutLast := moves[:len(moves)-1]
	for _, move := range withoutLast {
		newGame.Move(move)
	}
	g.game = newGame
	// Prevent cascading takebacks
	g.lastMoved = time.Time{}
	return g.LastMove(), nil
}

// SetTimeProvider allows the time provider to be overwritten (exclusively for testing)
func (g *Game) SetTimeProvider(provider TimeProvider) {
	g.timeProvider = provider
}

// String representation of the current game state (draws an ascii board)
func (g *Game) String() string {
	return g.game.Position().Board().Draw()
}
