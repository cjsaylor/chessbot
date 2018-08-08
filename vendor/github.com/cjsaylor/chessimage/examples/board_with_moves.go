// +build ignore

package main

import (
	"image/png"
	"log"
	"os"

	"github.com/cjsaylor/chessimage"
)

func main() {
	board, err := chessimage.NewRendererFromFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1")
	if err != nil {
		log.Fatalln(err)
	}
	board.SetLastMove(chessimage.LastMove{
		From: chessimage.E2,
		To:   chessimage.E4,
	})
	image, err := board.Render(chessimage.Options{
		AssetPath: "./assets/",
	})
	if err != nil {
		log.Fatalln(err)
	}
	png.Encode(os.Stdout, image)
}
