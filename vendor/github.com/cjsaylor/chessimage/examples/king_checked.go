// --+build ignore

package main

import (
	"image/png"
	"log"
	"os"

	"github.com/cjsaylor/chessimage"
)

func main() {
	board, err := chessimage.NewRendererFromFEN("r1r3k1/4bp2/ppn5/3pP1N1/1P1P1Pq1/P1NQ3b/3B1P2/R1R3K1 w - - 1 24")
	if err != nil {
		log.Fatalln(err)
	}
	board.SetLastMove(chessimage.LastMove{
		From: chessimage.D7,
		To:   chessimage.G4,
	})
	board.SetCheckTile(chessimage.G1)
	image, err := board.Render(chessimage.Options{AssetPath: "./assets/"})
	if err != nil {
		log.Fatalln(err)
	}
	png.Encode(os.Stdout, image)
}
