package engine

import "gochess/core"

const (
	MateScore = 30000
	MateThreshold = MateScore - 1000
)

func (e *Engine) negamax(depth int, alpha, beta int) int {
	e.NodesSearched++

	if e.NodesSearched % 2048 == 0 && e.TimeUp() {
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
		return e.quiscence(alpha, beta)
	}

	// Null move pruning
	// if depth >= 3 && !e.Board.InCheck(e.Board.WhiteToMove) {
	// 	e.Board.WhiteToMove = !e.Board.WhiteToMove // TODO: need to restore state better
	// 	nullScore := -e.negamax(depth-3, -beta, -beta+1) // reduction R=2
	// 	e.Board.WhiteToMove = !e.Board.WhiteToMove
	// 	if nullScore >= beta {
	// 		return nullScore
	// 	}
	// }

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
	var enemyBB core.Bitboard

	if board.WhiteToMove {
		enemyBB = board.BlackPieces
	} else {
		enemyBB = board.WhitePieces
	}

	if ttMove != (core.Move{}) {
		board.Push(&ttMove)
		score := -e.negamax(depth-1, -beta, -alpha)
		board.Pop()

		bestScore = score
		bestMove = ttMove
		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			e.TT.Store(key, depth, bestScore, FlagLower, bestMove)
			return bestScore
		}
	}

	e.OrderMoves(moves)

	firstMove := true
	for i, move := range moves {
		if move == ttMove {
			continue
		}

		isCapture := (move.To == board.EnPassantTarget) || ((1 << move.To) & enemyBB != 0)
		board.Push(&move)
		enemyInCheck := board.InCheck(board.WhiteToMove)
		searchDepth := depth - 1
		if depth >= 3 && !isCapture && !enemyInCheck && i > 3 {
			searchDepth = depth - 2
		}

		var score int
		if firstMove {
			score = -e.negamax(searchDepth, -beta, -alpha)
			firstMove = false
		} else {
			score = -e.negamax(searchDepth, -alpha - 1, -alpha)
			if score > alpha && score < beta {
				score = -e.negamax(searchDepth, -beta, -alpha)
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

	e.TT.Store(key, depth, bestScore, bound, bestMove)

	return bestScore
}
