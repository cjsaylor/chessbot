// Package analysis allows for games to be analyzed by a third party analysis service
package analysis

import (
	"log"
	"net/http"
	"net/url"

	"github.com/cjsaylor/chessbot/game"
)

// Analyzer should interact with an external service and return a link to view the analysis
type Analyzer interface {
	Analyze(*game.Game) (*url.URL, error)
}

// Handler is an http handler that will redirect the user to an analysis of their game
type Handler struct {
	gameStorage game.GameStorage
	analyzer    Analyzer
}

// NewHTTPHandler returns an instance of an analysis endpoint handler
func NewHTTPHandler(store game.GameStorage, analyzer Analyzer) *Handler {
	return &Handler{
		gameStorage: store,
		analyzer:    analyzer,
	}
}

func (a Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	gameID := r.URL.Query().Get("game_id")
	gm, err := a.gameStorage.RetrieveGame(gameID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	analysisURL, err := a.analyzer.Analyze(gm)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, analysisURL.String(), http.StatusTemporaryRedirect)
}
