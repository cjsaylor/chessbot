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

type BoardRenderHandler struct {
	GameStorage game.GameStorage
}

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
			// If in checkmate, there won't be a valid move. Figure out a different way to highlight in that case
			moves := game.ValidMoves()
			if len(moves) > 0 {
				tile, _ := chessimage.TileFromAN(moves[0].S1().String())
				board.SetCheckTile(tile)
			}
		}
	}
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}
	image, err := board.Render(chessimage.Options{AssetPath: "../chessimage/assets/"})
	if err != nil {
		log.Println(err)
	}
	png.Encode(w, image)
}
