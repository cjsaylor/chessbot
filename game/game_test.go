package game_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cjsaylor/chessbot/game"
)

func init() {
	rand.Seed(0)
}

func TestExport(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	gm.Move("d2d4")
	gm.Move("d7d5")
	result := gm.Export()
	expected := "[Site \"Slack ChessBot match\"]\n[White \"a\"]\n[Black \"b\"]\n\n1.d4 d5  *"
	if result != expected {
		t.Errorf("expected %v got %v", expected, result)
	}
}

func TestExportAndMoveMore(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	gm.Move("d2d4")
	// The implementation modifies the game in place.
	// This test verifies the notation returns to LAN.
	gm.Export()
	if _, err := gm.Move("d7d5"); err != nil {
		t.Error(err)
	}
}

func TestLastMoveTimeRecorded(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	if !gm.LastMoved().IsZero() {
		t.Error("expected last move to be a zero value")
	}
	gm.Move("d2d4")
	if gm.LastMoved().IsZero() {
		t.Error("expected the last moved date to be set")
	}
}

func TestTakebackRequestWithinThreshold(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	now := time.Now()
	gm.SetTimeProvider(func() time.Time {
		return now
	})
	gm.Move("d2d4")
	if pastThreshold := gm.IsPastTakebackThreshold(); pastThreshold {
		t.Error("expected response to be false, got true")
	}
	gm.SetTimeProvider(func() time.Time {
		return now.Add(game.TakebackThreshold + time.Second)
	})
	if pastThreshold := gm.IsPastTakebackThreshold(); !pastThreshold {
		t.Error("expected response to be true, got false")
	}
}

func TestTakebackRequestWithCorrectPlayer(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	takebackPlayer := gm.TurnPlayer()
	gm.Move("d2d4")
	otherTakebackPlayer := gm.TurnPlayer()
	if _, err := gm.Takeback(&takebackPlayer); err != nil {
		t.Error(err)
		t.Error("expected takeback to succeed for the player that initiated the move")
	}
	gm.Move("d2d4")
	if _, err := gm.Takeback(&otherTakebackPlayer); err == nil {
		t.Error("expected takeback to fail due to \"current\" player requesting a takeback")
	}
	gm.Move("d7d5")
	gm.Takeback(&otherTakebackPlayer)
	if isPastThreshold := gm.IsPastTakebackThreshold(); !isPastThreshold {
		t.Error("expected a time threshold error due to the previous takeback wiping out the last moved time period")
	}
}

func TestTakebackRequestWithCompletedGame(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	gm.Move("d2d4")
	resigner, _ := gm.PlayerByID("a")
	gm.Resign(*resigner)
	if _, err := gm.Takeback(resigner); err != game.ErrGameCompleted {
		t.Error(err)
		t.Error("expected the takeback to fail due to the game being over")
	}
}

func TestTakebackRequestWithNoMoves(t *testing.T) {
	gm := game.NewGame("1234", []game.Player{
		game.Player{
			ID: "a",
		},
		game.Player{
			ID: "b",
		},
	}...)
	requester := gm.TurnPlayer()
	if _, err := gm.Takeback(&requester); err != game.ErrGameHasNoMoves {
		t.Error(err)
		t.Error("expected the takeback to fail due to not having any moves in the game yet")
	}
}
