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

	if ((attacker & core.PieceTypeMask) == core.PieceTypePawn) && (move.To == move.From+16 || move.To == move.From-16) {
		score += 50
	}

	if victim != core.PieceNone {
		score += 100 + 10*e.PieceValue(victim) - e.PieceValue(attacker)
	}

	return score
}

func (e *Engine) OrderMoves(moves []core.Move, depth int) {
	slices.SortStableFunc(moves, func(a, b core.Move) int {
		scoreA := e.MoveScore(a) + e.killerHistoryScore(a, depth)
		scoreB := e.MoveScore(b) + e.killerHistoryScore(b, depth)
		return scoreB - scoreA // Descending order
	})
}

func (e *Engine) killerHistoryScore(move core.Move, depth int) int {
	if move == e.KillerMoves[depth][0] {
		return 100_000
	}

	if move == e.KillerMoves[depth][1] {
		return 80_000
	}

	isCapture := (move.To == e.Board.EnPassantTarget) || ((1<<move.To)&e.Board.AllPieces) != 0
	if !isCapture {
		return e.HistoryTable[move.From][move.To]
	}

	return 0
}
