// +build ignore

package main

import (
	"image/png"
	"log"
	"os"

	"github.com/cjsaylor/chessimage"
)

func main() {
	board, err := chessimage.NewRendererFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		log.Fatalln(err)
	}
	image, err := board.Render(chessimage.Options{
		AssetPath: "./assets/",
	})
	if err != nil {
		log.Fatalln(err)
	}
	png.Encode(os.Stdout, image)
}
