package engine

import (
	"gochess/core"
)

func (e *Engine) Evaluate() int {
	score := 0
	mask := e.board.AllPieces
	for mask != 0 {
		lsb := mask.PopLSB()
		piece := e.board.Pieces[lsb]
		value := e.PieceValue(piece)
		if piece.Color() == core.PieceColorWhite {
			score += value
		} else {
			score -= value
		}
	}
	return score
}

func (e *Engine) PieceValue(piece core.Piece) int {
	type_ := piece.Type()
	switch type_ {
	case core.PieceTypeNone:
		return 0
	case core.PieceTypePawn:
		return 100
	case core.PieceTypeKnight:
		return 320
	case core.PieceTypeBishop:
		return 330
	case core.PieceTypeRook:
		return 500
	case core.PieceTypeQueen:
		return 900
	case core.PieceTypeKing:
		return 20000
	default:
		return 0
	}
}
