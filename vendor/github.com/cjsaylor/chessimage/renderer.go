package chessimage

import (
	"image"
	"log"

	findfont "github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

var pieceNames = map[string]string{
	"b": "bd.png",
	"B": "bl.png",
	"k": "kd.png",
	"K": "kl.png",
	"n": "nd.png",
	"N": "nl.png",
	"p": "pd.png",
	"P": "pl.png",
	"q": "qd.png",
	"Q": "ql.png",
	"r": "rd.png",
	"R": "rl.png",
}

const (
	defaultBoardSize   = 512
	defaultPieceRatio  = 0.8
	fileSymbols        = "abcdefgh"
	fileSymbolsReverse = "hgfedcba"
	rankSymbols        = "12345678"
	rankSymbolsReverse = "87654321"
)

var (
	colorLight        = []int{239, 218, 183}
	colorDark         = []int{180, 135, 102}
	colorHighlight    = []int{205, 210, 122}
	colorHighlightDim = []int{170, 160, 75}
	colorCheck        = []int{227, 30, 32}
)

type drawSize struct {
	gridSize               int
	pieceSize, pieceOffset int
}

// Options holds all possible rendering options for customization
type Options struct {
	AssetPath  string
	Resizer    draw.Scaler
	BoardSize  int
	PieceRatio float64
	Inverted   bool
}

// Renderer is responsible for rendering the board, pieces, rank/file, and tile highlights
type Renderer struct {
	context   *gg.Context
	board     board
	drawSize  drawSize
	checkTile Tile
	lastMove  *LastMove
}

// NewRendererFromFEN prepares a renderer for use with given FEN string
func NewRendererFromFEN(fen string) (*Renderer, error) {
	board, err := decodeFEN(fen)
	if err != nil {
		return nil, err
	}
	return &Renderer{
		board:     board,
		checkTile: NoTile,
	}, nil
}

// SetCheckTile - Sets the check tile
func (r *Renderer) SetCheckTile(tile Tile) {
	// @todo validate it is within the range of proper tiles
	r.checkTile = tile
}

// SetLastMove - Sets the last move
func (r *Renderer) SetLastMove(lastMove LastMove) {
	r.lastMove = &lastMove
}

// Render the chess board with given items
func (r *Renderer) Render(options Options) (image.Image, error) {
	if options.BoardSize <= 0 {
		options.BoardSize = defaultBoardSize
	}
	if options.PieceRatio <= 0.0 {
		options.PieceRatio = defaultPieceRatio
	}
	if options.Resizer == nil {
		options.Resizer = draw.CatmullRom
	}
	r.drawSize = calcDrawSize(options)
	r.context = gg.NewContext(options.BoardSize, options.BoardSize)
	r.drawBackground()
	r.highlightCells(options)
	r.drawCheckTile(options)
	r.drawRankFile(options)
	if err := r.drawBoard(options); err != nil {
		return nil, err
	}
	return r.context.Image(), nil
}

func (r *Renderer) drawBackground() {
	gridSize := r.drawSize.gridSize
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			r.context.DrawRectangle(float64(row*gridSize), float64(col*gridSize), float64(gridSize), float64(gridSize))
			if (col+row)%2 == 0 {
				r.context.SetRGB255(colorLight[0], colorLight[1], colorLight[2])
			} else {
				r.context.SetRGB255(colorDark[0], colorDark[1], colorDark[2])
			}
			r.context.Fill()
		}
	}
}

func (r *Renderer) highlightCells(o Options) {
	if r.lastMove == nil {
		return
	}

	var lastMoveFromRank, lastMoveToRank, lastMoveFromFile, lastMoveToFile int
	if o.Inverted {
		lastMoveFromRank = r.lastMove.From.rankInverted()
		lastMoveFromFile = r.lastMove.From.fileInverted()
		lastMoveToRank = r.lastMove.To.rankInverted()
		lastMoveToFile = r.lastMove.To.fileInverted()
	} else {
		lastMoveFromRank = r.lastMove.From.rank()
		lastMoveFromFile = r.lastMove.From.file()
		lastMoveToRank = r.lastMove.To.rank()
		lastMoveToFile = r.lastMove.To.file()
	}

	gridSize := r.drawSize.gridSize
	r.context.DrawRectangle(
		float64(lastMoveFromFile*gridSize),
		float64(lastMoveFromRank*gridSize),
		float64(gridSize),
		float64(gridSize))
	r.context.SetRGB255(colorHighlight[0], colorHighlight[1], colorHighlight[2])
	r.context.Fill()
	r.context.DrawRectangle(
		float64(lastMoveToFile*gridSize),
		float64(lastMoveToRank*gridSize),
		float64(gridSize), float64(gridSize))
	r.context.SetRGB255(colorHighlightDim[0], colorHighlightDim[1], colorHighlightDim[2])
	r.context.Fill()
}

func (r *Renderer) drawCheckTile(o Options) {
	if r.checkTile == NoTile {
		return
	}
	var checkTileFile, checkTileRank int
	if o.Inverted {
		checkTileFile = r.checkTile.fileInverted()
		checkTileRank = r.checkTile.rankInverted()
	} else {
		checkTileFile = r.checkTile.file()
		checkTileRank = r.checkTile.rank()
	}
	gridSize := float64(r.drawSize.gridSize)
	r.context.DrawRectangle(
		float64(checkTileFile)*gridSize,
		float64(checkTileRank)*gridSize,
		gridSize,
		gridSize,
	)
	r.context.SetRGB255(colorCheck[0], colorCheck[1], colorCheck[2])
	r.context.Fill()
}

func (r *Renderer) drawBoard(o Options) error {
	for _, position := range r.board {
		if err := r.drawPiece(position, o.AssetPath, o.Resizer, o.Inverted); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) drawRankFile(o Options) error {
	var symbols string
	fontPath, err := findfont.Find("arial.ttf")
	if err != nil {
		return err
	}
	if err := r.context.LoadFontFace(fontPath, 14); err != nil {
		return err
	}

	if o.Inverted {
		symbols = fileSymbolsReverse
	} else {
		symbols = fileSymbols
	}
	for i, symbol := range symbols {
		var color []int
		if i%2 == 0 {
			color = colorLight
		} else {
			color = colorDark
		}
		r.context.SetRGB255(color[0], color[1], color[2])
		r.context.DrawString(string(symbol), float64(r.drawSize.gridSize*i+2), float64(o.BoardSize-3))
	}

	if o.Inverted {
		symbols = rankSymbols
	} else {
		symbols = rankSymbolsReverse
	}
	for i, symbol := range symbols {
		var color []int
		if i%2 == 0 {
			color = colorLight
		} else {
			color = colorDark
		}
		r.context.SetRGB255(color[0], color[1], color[2])
		r.context.DrawString(string(symbol), float64(o.BoardSize-10), float64(r.drawSize.gridSize*i+12))
	}

	return nil
}

func (r *Renderer) drawPiece(piece position, assetPath string, resizer draw.Scaler, inverted bool) error {
	// Todo move this to runtime cache function
	png, err := gg.LoadPNG(assetPath + pieceNames[string(piece.pieceSymbol)])
	if err != nil {
		return err
	}
	resized := resizeImage(png, r.drawSize, resizer)
	if err != nil {
		log.Fatal(err)
	}
	gridSize := r.drawSize.gridSize
	pieceOffset := r.drawSize.pieceOffset

	var pieceRank, pieceFile int
	if inverted {
		pieceRank = piece.tile.rankInverted()
		pieceFile = piece.tile.fileInverted()
	} else {
		pieceRank = piece.tile.rank()
		pieceFile = piece.tile.file()
	}

	r.context.DrawImage(resized, gridSize*(pieceRank)+pieceOffset, gridSize*(pieceFile)+pieceOffset)
	return nil
}

func resizeImage(piece image.Image, drawSize drawSize, resizer draw.Scaler) *image.RGBA {
	rect := image.Rect(0, 0, drawSize.pieceSize, drawSize.pieceSize)
	dst := image.NewRGBA(rect)
	draw.BiLinear.Scale(dst, rect, piece, piece.Bounds(), draw.Over, nil)
	return dst
}

func calcDrawSize(o Options) drawSize {
	gridSize := o.BoardSize / 8
	pieceSize := int(float64(gridSize) * o.PieceRatio)
	return drawSize{
		gridSize:    gridSize,
		pieceSize:   int(pieceSize),
		pieceOffset: int((gridSize - pieceSize) / 2),
	}
}
