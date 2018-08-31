package game_test

import (
	"testing"

	"github.com/cjsaylor/chessbot/game"
)

func TestGameSavesAndIsRetrievable(t *testing.T) {
	db, err := game.NewSqliteStore("../chessbot.db")
	if err != nil {
		t.Error(err)
	}
	gm := game.NewGame(game.Player{ID: "1"}, game.Player{ID: "2"})
	if err := db.StoreGame("1234", gm); err != nil {
		t.Error(err)
	}
	gm, err = db.RetrieveGame("1234")
	if err != nil {
		t.Error(err)
	}
}

func TestRetrieveGameThatDoesNotExist(t *testing.T) {
	db, err := game.NewSqliteStore("../chessbot.db")
	if err != nil {
		t.Error(err)
	}
	if _, err := db.RetrieveGame("NOEXISTID"); err == nil {
		t.Error("should throw and error because the game was not found")
	}
}
