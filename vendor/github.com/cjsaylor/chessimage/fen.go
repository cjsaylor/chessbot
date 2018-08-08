package chessimage

import (
	"fmt"
	"strconv"
	"strings"
)

// For now we are only concerned about position of the pieces
func decodeFEN(sequence string) (board, error) {
	fen := strings.TrimSpace(sequence)
	parts := strings.Split(fen, " ")
	if len(parts) != 6 {
		return nil, fmt.Errorf("FEN invalid notiation %s must have 6 sections", fen)
	}
	board := board{}
	fenPositions := strings.Split(parts[0], "/")
	for rank, row := range fenPositions {
		for file, piece := range normalizeFENRank(row) {
			if ok := pieceNames[string(piece)]; ok != "" {

				board = append(board, position{tileFromRankFile(rank, file), string(piece)})
			}
		}
	}
	return board, nil
}

func normalizeFENRank(fenRank string) string {
	normalized := ""
	for _, symbol := range fenRank {
		skip, err := strconv.Atoi(string(symbol))
		if err == nil {
			normalized += strings.Repeat(" ", skip)
		} else {
			normalized += string(symbol)
		}
	}
	return normalized
}
