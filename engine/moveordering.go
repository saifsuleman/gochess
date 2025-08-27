package engine

import (
	"gochess/core"
	"slices"
)

func (e *Engine) MVVLVA(move core.Move) (score int) {
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

func (e *Engine) OrderMovesQ(moves []core.Move, depth int) {
	slices.SortStableFunc(moves, func(a, b core.Move) int {
		scoreA := e.SEE(a) + e.killerHistoryScore(a, depth)
		scoreB := e.SEE(b) + e.killerHistoryScore(b, depth)
		return scoreB - scoreA // Descending order
	})
}

func (e *Engine) OrderMoves(moves []core.Move, depth int) {
	slices.SortStableFunc(moves, func(a, b core.Move) int {
		scoreA := e.MVVLVA(a) + e.killerHistoryScore(a, depth)
		scoreB := e.MVVLVA(b) + e.killerHistoryScore(b, depth)
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

func (e *Engine) SEE(move core.Move) int {
	from := move.From
	to := move.To

	board := e.Board
	attackerIsWhite := (board.Pieces[from] & core.PieceColorMask) == core.PieceColorWhite
	attackers := make([]int, 0)
	defenders := make([]int, 0)
	attackersBB := board.GetAttackingBitboard(to, attackerIsWhite)
	defendersBB := board.GetAttackingBitboard(to, !attackerIsWhite)
	for sq := range 64 {
		if (attackersBB>>sq)&1 != 0 {
			attackers = append(attackers, sq)
		}
		if (defendersBB>>sq)&1 != 0 {
			defenders = append(defenders, sq)
		}
	}

	victim := board.Pieces[to]
	gains := []int{e.PieceValue(victim)}
	turn := attackerIsWhite

	for {
		var next int
		var nextVal int
		if turn == attackerIsWhite {
			if len(attackers) == 0 {
				break
			}
			next = e.leastValuablePiece(board, attackers)
			nextVal = e.PieceValue(board.Pieces[next])
			attackers = removeSquare(attackers, next)
		} else {
			if len(defenders) == 0 {
				break
			}
			next = e.leastValuablePiece(board, defenders)
			nextVal = e.PieceValue(board.Pieces[next])
			defenders = removeSquare(defenders, next)
		}

		lastGain := gains[len(gains)-1]
		gains = append(gains, nextVal-lastGain)
		turn = !turn
	}

	for i := len(gains) - 2; i >= 0; i-- {
		gains[i] = max(-gains[i+1], gains[i])
	}

	return gains[0]
}

func (e *Engine) leastValuablePiece(board *core.Board, squares []int) int {
	least := squares[0]
	val := e.PieceValue(board.Pieces[least])
	for _, sq := range squares[1:] {
		v := e.PieceValue(board.Pieces[sq])
		if v < val {
			val = v
			least = sq
		}
	}
	return least
}

func removeSquare(slice []int, sq int) []int {
	for i, s := range slice {
		if s == sq {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

