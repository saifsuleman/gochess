package game

import "image/color"
import "gochess/core"

const (
	BOARD_SIZE = 8
	TILE_SIZE = 80
	PANEL_HEIGHT = 24
)

var (
	LIGHT_SQUARE = color.RGBA{ R: 255, G: 206, B: 158, A: 255 }
	DARK_SQUARE = color.RGBA{ R: 209, G: 139, B: 71, A: 255 }
	HIGHLIGHT = color.RGBA{ R: 255, G: 255, B: 0, A: 128 } // Semi-transparent yellow for highlights
	SELECT_COLOR = color.RGBA{ R: 0, G: 255, B: 0, A: 128 } // Semi-transparent green for selected piece
	HOVER_COLOR = color.RGBA{ R: 0, G: 0, B: 255, A: 128 } // Semi-transparent blue for hovered piece
	TEXT_COLOR = color.RGBA{ R: 0, G: 0, B: 0, A: 255 } // Black text color
)

type Color int
const (
	White Color = iota
	Black
)


type PieceType int
const (
	None PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

func (p Piece) Glyph() {
}

