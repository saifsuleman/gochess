package engine

import (
	"gochess/core"
)

const (
	MateScore     = 30000
	MateThreshold = MateScore - 1000
)

func (e *Engine) negamax(depth int, alpha, beta, rootDepth int) int {
	e.NodesSearched++

	if e.NodesSearched%2048 == 0 && e.TimeUp() {
		e.Aborted = true
		return 0
	}

	originalAlpha := alpha
	key := e.Board.ComputeZobristHash()

	var ttMove core.Move
	if ok, score, _, m := e.TT.ProbeCut(key, depth, alpha, beta, e.Board.Ply); ok {
		return score
	} else {
		ttMove = m
	}

	if depth <= 0 {
		return e.quiscence(alpha, beta, rootDepth)
	}

	// TODO: keep an eye on this, I've found sometimes it makes the engine worse
	// i don't trust you
	isNullWindow := beta-alpha == 1
	nmpMask := e.Board.AllPieces
	nmpMask ^= e.Board.PieceBitboards[0][core.PieceTypeKing-1] | e.Board.PieceBitboards[1][core.PieceTypeKing-1]
	nmpMask ^= e.Board.PieceBitboards[0][core.PieceTypePawn-1] | e.Board.PieceBitboards[1][core.PieceTypePawn-1]
	if isNullWindow && depth >= 3 && nmpMask != 0 && !e.Board.InCheck(e.Board.WhiteToMove) && e.Evaluate() >= beta {
		enPassant := e.Board.EnPassantTarget
		e.Board.EnPassantTarget = 64
		e.Board.WhiteToMove = !e.Board.WhiteToMove

		nullScore := -e.negamax(depth-3, -beta, -beta+1, rootDepth) // reduction R=2

		e.Board.WhiteToMove = !e.Board.WhiteToMove
		e.Board.EnPassantTarget = enPassant

		if nullScore >= beta {
			depth -= 4
			if depth <= 0 {
				return e.quiscence(alpha, beta, rootDepth)
			}
		}
	}

	board := e.Board
	moves := board.GenerateLegalMoves()
	if len(moves) == 0 {
		if board.InCheck(board.WhiteToMove) {
			return -MateScore + board.Ply // checkmate
		}
		return 0 // stalemate
	}

	var bestMove core.Move
	var bestScore int = -100000

	if ttMove != (core.Move{}) {
		board.Push(&ttMove)
		score := -e.negamax(depth-1, -beta, -alpha, rootDepth)
		board.Pop()

		bestScore = score
		bestMove = ttMove
		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			e.TT.Store(key, depth, bestScore, FlagLower, bestMove, board.Ply)
			return bestScore
		}
	}

	e.OrderMoves(moves, rootDepth)

	firstMove := true
	for i, move := range moves {
		if move == ttMove {
			continue
		}

		isCapture := (move.To == board.EnPassantTarget) || ((1<<move.To)&board.AllPieces) != 0
		board.Push(&move)
		searchDepth := depth - 1

		// LMR
		enemyInCheck := board.InCheck(board.WhiteToMove)
		if depth >= 3 && !isCapture && !enemyInCheck && i > 3 {
			searchDepth = depth - 2
		}

		var score int
		if firstMove {
			score = -e.negamax(searchDepth, -beta, -alpha, rootDepth)
			firstMove = false
		} else {
			score = -e.negamax(searchDepth, -alpha-1, -alpha, rootDepth)
			if score > alpha && score < beta {
				score = -e.negamax(searchDepth, -beta, -alpha, rootDepth)
			}
		}

		board.Pop()

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		if bestScore > alpha {
			alpha = bestScore
		}
		if alpha >= beta {
			isCapture := move.To == board.EnPassantTarget || ((1<<move.To)&board.AllPieces) != 0
			if !isCapture {
				e.addKillerMove(move, depth)
				e.HistoryTable[move.From][move.To] += depth * depth
			}
			break
		}
	}

	// Store result in TT
	var bound Bound
	switch {
	case bestScore <= originalAlpha:
		bound = FlagUpper
	case bestScore >= beta:
		bound = FlagLower
	default:
		bound = FlagExact
	}

	e.TT.Store(key, depth, bestScore, bound, bestMove, board.Ply)

	return bestScore
}
