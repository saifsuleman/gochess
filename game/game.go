package game

import (
	"gochess/core"
	"gochess/engine"
	"gochess/fen"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	BOARD_SIZE = 8
	TILE_SIZE  = 45
)

var (
	LIGHT_SQUARE = color.RGBA{R: 255, G: 206, B: 158, A: 255}
	DARK_SQUARE  = color.RGBA{R: 209, G: 139, B: 71, A: 255}
	HIGHLIGHT    = color.RGBA{R: 255, G: 255, B: 0, A: 128} // Previous move
	SELECT_COLOR = color.RGBA{R: 0, G: 255, B: 0, A: 128}   // Selected piece
)

var img *ebiten.Image

func Init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("spritesheet.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	Board        *core.Board
	engine       *engine.Engine
	selected     int
	dragging     bool
	dragStart    int
	prevMoveFrom int
	prevMoveTo   int
	mouseX       float64
	mouseY       float64
}

func NewGame(board *core.Board) *Game {
	engine := engine.NewEngine(board)
	game := &Game{Board: board, selected: -1, engine: engine}
	return game
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	g.mouseX = float64(mouseX)
	g.mouseY = float64(mouseY)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		square := g.xyToSquare(int(g.mouseX)/TILE_SIZE, int(g.mouseY)/TILE_SIZE)

		// Pick up piece
		if !g.dragging && square != -1 && g.Board.Pieces[square] != core.PieceNone {
			piece := g.Board.Pieces[square]
			if g.Board.WhiteToMove && piece.Color() == core.PieceColorWhite ||
				(!g.Board.WhiteToMove && piece.Color() == core.PieceColorBlack) {
				g.selected = square
				g.dragStart = square
				g.dragging = true
			}
		}

	} else if g.dragging {
		toSquare := g.xyToSquare(int(g.mouseX)/TILE_SIZE, int(g.mouseY)/TILE_SIZE)
		if toSquare != -1 && toSquare != g.dragStart {
			move := core.Move{
				From:      core.Position(g.dragStart),
				To:        core.Position(toSquare),
				Promotion: core.PieceNone, // Could add promotion UI later
			}

			if g.Board.IsMoveLegal(move) {
				g.Board.Push(&move)

				g.prevMoveFrom = g.dragStart
				g.prevMoveTo = toSquare

				go (func() {
					g.engine.Board = g.Board.Clone()
					bestMove := g.engine.FindBestMove(time.Millisecond * 1000)
					if bestMove != nil {
						g.Board.Push(bestMove)

						g.prevMoveFrom = int(bestMove.From)
						g.prevMoveTo = int(bestMove.To)
					}
				})()
			}
		}

		g.dragging = false
		g.selected = -1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.Board.Pop()

		lastMove := g.Board.LastMove()
		if lastMove != nil {
			g.prevMoveFrom = int(lastMove.From)
			g.prevMoveTo = int(lastMove.To)
		} else {
			g.prevMoveFrom = -1
			g.prevMoveTo = -1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.Board, _ = fen.LoadFromFEN(fen.DefaultFEN())
		g.engine = engine.NewEngine(g.Board)
		g.selected = -1
		g.dragging = false
		g.dragStart = -1
		g.prevMoveFrom = -1
		g.prevMoveTo = -1
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y := 0; y < BOARD_SIZE; y++ {
		for x := 0; x < BOARD_SIZE; x++ {
			square := g.xyToSquare(x, y)
			tileColor := LIGHT_SQUARE
			if (x+y)%2 == 1 {
				tileColor = DARK_SQUARE
			}
			ebitenutil.DrawRect(screen, float64(x*TILE_SIZE), float64(y*TILE_SIZE), TILE_SIZE, TILE_SIZE, tileColor)

			// Previous move highlight
			if square == g.prevMoveFrom || square == g.prevMoveTo {
				ebitenutil.DrawRect(screen, float64(x*TILE_SIZE), float64(y*TILE_SIZE), TILE_SIZE, TILE_SIZE, HIGHLIGHT)
			}

			// Selected highlight
			if square == g.selected {
				ebitenutil.DrawRect(screen, float64(x*TILE_SIZE), float64(y*TILE_SIZE), TILE_SIZE, TILE_SIZE, SELECT_COLOR)
			}

			piece := g.Board.Pieces[square]
			if piece != core.PieceNone {
				// Skip drawing dragged piece at original square
				if g.dragging && square == g.dragStart {
					continue
				}
				DrawPiece(screen, pieceID(piece), x, y)
			}
		}
	}

	// Draw dragged piece smoothly under the cursor
	if g.dragging && g.selected != -1 {
		DrawPieceAtPixel(screen, pieceID(g.Board.Pieces[g.selected]), g.mouseX, g.mouseY)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (width, height int) {
	return TILE_SIZE * BOARD_SIZE, TILE_SIZE * BOARD_SIZE
}

func DrawPiece(screen *ebiten.Image, pieceID, x, y int) {
	DrawPieceAtPixel(screen, pieceID, float64(x*TILE_SIZE)+TILE_SIZE/2, float64(y*TILE_SIZE)+TILE_SIZE/2)
}

func DrawPieceAtPixel(screen *ebiten.Image, pieceID int, px, py float64) {
	if img == nil {
		log.Println("Image not initialized")
		return
	}
	sx := (pieceID % 6) * TILE_SIZE
	sy := (pieceID / 6) * TILE_SIZE
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(px-TILE_SIZE/2, py-TILE_SIZE/2) // Center piece on cursor
	screen.DrawImage(img.SubImage(image.Rect(sx, sy, sx+TILE_SIZE, sy+TILE_SIZE)).(*ebiten.Image), op)
}

func (g *Game) xyToSquare(x, y int) int {
	if x < 0 || x >= BOARD_SIZE || y < 0 || y >= BOARD_SIZE {
		return -1
	}
	return (BOARD_SIZE-1-y)*BOARD_SIZE + x
}

func pieceID(piece core.Piece) int {
	switch piece {
	case core.PieceWhiteKing:
		return 0
	case core.PieceWhiteQueen:
		return 1
	case core.PieceWhiteBishop:
		return 2
	case core.PieceWhiteKnight:
		return 3
	case core.PieceWhiteRook:
		return 4
	case core.PieceWhitePawn:
		return 5
	case core.PieceBlackKing:
		return 6
	case core.PieceBlackQueen:
		return 7
	case core.PieceBlackBishop:
		return 8
	case core.PieceBlackKnight:
		return 9
	case core.PieceBlackRook:
		return 10
	case core.PieceBlackPawn:
		return 11
	default:
		return -1
	}
}
