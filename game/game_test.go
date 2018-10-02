package game_test

import (
	"testing"

	"github.com/cjsaylor/chessbot/game"
)

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
