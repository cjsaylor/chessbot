package analysis

import (
	"net/url"

	"github.com/cjsaylor/chessbot/game"
)

// ChesscomAnalyzer provides a way to setup an analysis of a game
type ChesscomAnalyzer struct {
	AffiliateCode string
}

// NewChesscomAnalyzer returns an analyzer for chess.com
func NewChesscomAnalyzer(affiliateCode string) *ChesscomAnalyzer {
	return &ChesscomAnalyzer{
		AffiliateCode: affiliateCode,
	}
}

// Analyze a game and return a URL to that analysis
func (c ChesscomAnalyzer) Analyze(gm *game.Game) (*url.URL, error) {
	data := url.Values{}
	data.Add("pgn", gm.Export())
	data.Add("ref_id", c.AffiliateCode)

	return url.Parse("https://www.chess.com/analysis?" + data.Encode())
}
