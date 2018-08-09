// Package rendering is responsible for rendering game boards by ID in their current state
package rendering

import (
	"image/png"
	"log"
	"net/http"
	"time"

	"github.com/cjsaylor/chessbot/game"
	"github.com/cjsaylor/chessimage"
	"github.com/notnil/chess"
)

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

// BoardRenderHandler handles all image requests from Slack
type BoardRenderHandler struct {
	GameStorage game.GameStorage
}

// ServeHTTP is a request handler
func (b BoardRenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	gameID, ok := r.URL.Query()["game_id"]
	if !ok || len(gameID[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	game, err := b.GameStorage.RetrieveGame(gameID[0])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fen := game.FEN()
	board, err := chessimage.NewRendererFromFEN(string(fen))
	if lastMove := game.LastMove(); lastMove != nil {
		from, _ := chessimage.TileFromAN(lastMove.S1().String())
		to, _ := chessimage.TileFromAN(lastMove.S2().String())
		board.SetLastMove(chessimage.LastMove{
			From: from,
			To:   to,
		})
		if game.LastMove().HasTag(chess.Check) {
			square := game.CheckedKing()
			if square != chess.NoSquare {
				tile, _ := chessimage.TileFromAN(square.String())
				board.SetCheckTile(tile)
			}
		}
	}
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}
	image, err := board.Render(chessimage.Options{AssetPath: "./assets/"})
	if err != nil {
		log.Println(err)
	}
	png.Encode(w, image)
}
