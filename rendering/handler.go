package rendering

import (
	"image/png"
	"log"
	"net/http"
	"strings"

	"github.com/cjsaylor/chessimage"
)

type BoardRenderHandler struct{}

func (b BoardRenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	fen, ok := r.URL.Query()["fen"]
	if !ok || len(fen[0]) < 1 {
		log.Println("Missing fen parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	board, err := chessimage.NewRendererFromFEN(string(fen[0]))
	if lastMove, ok := r.URL.Query()["last_move"]; ok {
		parts := strings.Split(lastMove[0], " ")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		from, err := chessimage.TileFromAN(parts[0])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		to, err := chessimage.TileFromAN(parts[1])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		board.SetLastMove(chessimage.LastMove{
			From: from,
			To:   to,
		})
	}
	if checkedTile, ok := r.URL.Query()["checked_tile"]; ok {
		tile, err := chessimage.TileFromAN(checkedTile[0])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		board.SetCheckTile(tile)
	}
	image, err := board.Render(chessimage.Options{AssetPath: "../chessimage/assets/"})
	if err != nil {
		log.Println(err)
	}
	png.Encode(w, image)
}
