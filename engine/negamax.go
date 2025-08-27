package engine

import "gochess/core"

const (
	MateScore = 30000
	MateThreshold = MateScore - 1000
)

func (e *Engine) negamax(depth int, alpha, beta, rootDepth int) int {
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
		return e.quiscence(alpha, beta, rootDepth)
	}

	// TODO: we should definitely have some protections against zugzwang somehow
	// note: this appears to cause a significant amount of blunders in the opening... this should be double checked
	// for now let's not use it, because the engine seems really strong without it

	// if depth >= 3 && !e.Board.InCheck(e.Board.WhiteToMove) {
	// 	// TODO: also reset en passant counter, and half-clock full-clock counter when we implement them
	// 	e.Board.WhiteToMove = !e.Board.WhiteToMove
	// 	nullScore := -e.negamax(depth-3, -beta, -beta+1, rootDepth) // reduction R=2
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
			e.TT.Store(key, depth, bestScore, FlagLower, bestMove)
			return bestScore
		}
	}

	e.OrderMoves(moves, rootDepth)

	firstMove := true
	for i, move := range moves {
		if move == ttMove {
			continue
		}

		isCapture := (move.To == board.EnPassantTarget) || ((1 << move.To) & board.AllPieces) != 0
		board.Push(&move)
		enemyInCheck := board.InCheck(board.WhiteToMove)
		searchDepth := depth - 1
		if depth >= 3 && !isCapture && !enemyInCheck && i > 3 {
			searchDepth = depth - 2
		}

		var score int
		if firstMove {
			score = -e.negamax(searchDepth, -beta, -alpha, rootDepth)
			firstMove = false
		} else {
			score = -e.negamax(searchDepth, -alpha - 1, -alpha, rootDepth)
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
			isCapture := move.To == board.EnPassantTarget || ((1 << move.To) & board.AllPieces) != 0
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

	e.TT.Store(key, depth, bestScore, bound, bestMove)

	return bestScore
}
