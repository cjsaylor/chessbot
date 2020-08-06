package rendering

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/cjsaylor/chessbot/game"
	"github.com/notnil/chess"
)

// RenderLink is a simple struct for creating valid external board URLs
type RenderLink struct {
	hostName   string
	signingKey string
}

// CreateLink returns an externally accessible board URL at the current game state
func (r RenderLink) CreateLink(gm *game.Game) (*url.URL, error) {
	fen := gm.FEN()
	sig := sha256.New()
	sig.Write([]byte(fen + r.signingKey))
	from, to, check := "", "", ""
	if lastMove := gm.LastMove(); lastMove != nil {
		from = lastMove.S1().String()
		to = lastMove.S2().String()
		if lastMove.HasTag(chess.Check) {
			square := gm.CheckedKing()
			check = square.String()
		}
	}
	u, _ := url.Parse(fmt.Sprintf("%v/board.png", r.hostName))
	q := u.Query()
	q.Add("fen", fen)
	q.Add("signature", hex.EncodeToString(sig.Sum(nil)))
	q.Add("from", from)
	q.Add("to", to)
	q.Add("check", check)
	if gm.Turn() == game.Black {
		q.Add("inverted", "true")
	}
	u.RawQuery = q.Encode()
	return u, nil
}

// ValidateLink ensures that the link signatuer is signed properly with the app signing key
func (r RenderLink) ValidateLink(url url.URL) bool {
	sig := sha256.New()
	sig.Write([]byte(url.Query().Get("fen") + r.signingKey))
	return hex.EncodeToString(sig.Sum(nil)) == url.Query().Get("signature")
}

// NewRenderLink creates a new RenderLink struct instance
func NewRenderLink(hostname string, signingKey string) RenderLink {
	return RenderLink{
		hostName:   hostname,
		signingKey: signingKey,
	}
}
