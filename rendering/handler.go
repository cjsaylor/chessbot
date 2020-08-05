// Package rendering is responsible for rendering game boards by ID in their current state
package rendering

import (
	"image/png"
	"log"
	"net/http"

	"github.com/cjsaylor/chessimage"
)

// BoardRenderHandler handles all image requests from Slack
type BoardRenderHandler struct {
	LinkRenderer RenderLink
}

// ServeHTTP is a request handler
func (b BoardRenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	fen := query.Get("fen")
	if fen == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !b.LinkRenderer.ValidateLink(*r.URL) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	board, err := chessimage.NewRendererFromFEN(fen)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if from := query.Get("from"); from != "" {
		to := query.Get("to")
		tFrom, _ := chessimage.TileFromAN(from)
		tTo, _ := chessimage.TileFromAN(to)
		board.SetLastMove(chessimage.LastMove{
			From: tFrom,
			To:   tTo,
		})
	}
	if check := query.Get("check"); check != "" {
		tCheck, _ := chessimage.TileFromAN(check)
		board.SetCheckTile(tCheck)
	}

	inverted := query.Get("inverted") == "true"
	image, err := board.Render(chessimage.Options{AssetPath: "./assets/", Inverted: inverted})
	if err != nil {
		log.Println(err)
	}
	w.Header().Add("Cache-Control", "max-age=7776000")
	png.Encode(w, image)
}
