package engine

import (
	"gochess/core"
	"slices"
)

func (e *Engine) MoveScore(move core.Move) (score int) {
	score = 0

	switch move.Promotion {
	case core.PieceTypeQueen:
		score += 900
	case core.PieceTypeRook:
		score += 500
	case core.PieceTypeBishop:
		score += 300
	case core.PieceTypeKnight:
		score += 350 // knight promotions are badass
	}

	attacker := e.Board.Pieces[move.From]
	victim := e.Board.Pieces[move.To]

	if (attacker&core.PieceTypeMask) == core.PieceTypePawn && move.To == e.Board.EnPassantTarget {
		victim = core.PieceTypePawn | ^(attacker & core.PieceColorMask)
	}

	if victim != core.PieceNone {
		score += 100 + 10*e.PieceValue(victim) - e.PieceValue(attacker)
	}

	return score
}

func (e *Engine) OrderMoves(moves []core.Move) {
	slices.SortStableFunc(moves, func(a, b core.Move) int {
		return e.MoveScore(b) - e.MoveScore(a)
	})
}
