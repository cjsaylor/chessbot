package analysis

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/cjsaylor/chessbot/game"
)

const analysisHost = "https://lichess.org"

// LichessAnalyzer provides a way to setup an analysis of a game
type LichessAnalyzer struct{}

// Analyze a game and return a URL to that analysis
func (l LichessAnalyzer) Analyze(gm *game.Game) (*url.URL, error) {
	data := url.Values{}
	data.Add("pgn", gm.Export())
	data.Add("analysis", "on")

	req, _ := http.NewRequest(
		http.MethodPost,
		analysisHost+"/import",
		bytes.NewBufferString(data.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Slack/Chessbot")
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	if resp.StatusCode != http.StatusSeeOther {
		return nil, fmt.Errorf("server responded with an unexpected status code to our request: %v", resp.StatusCode)
	}
	return url.Parse(analysisHost + resp.Header.Get("Location"))
}
