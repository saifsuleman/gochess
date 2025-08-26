package engine

import (
	"gochess/core"
	"math"
)

func (e *Engine) alphaBeta(board *core.Board, depth int, alpha int, beta int, maximizing bool) int {
	if depth == 0 {
		return e.Evaluate()
	}

	moves := board.GenerateLegalMoves()
	if len(moves) == 0 {
		return e.Evaluate()
	}

	if maximizing {
		maxEval := math.MaxInt
		for _, move := range moves {
			e.board.Push(&move)
			eval := e.alphaBeta(board, depth - 1, alpha, beta, false)
			e.board.Pop()

			if eval > maxEval {
				maxEval = eval
			}

			if eval > alpha {
				alpha = eval
			}

			if alpha > beta {
				break
			}
		}
		return maxEval
	} else {
		minEval := math.MinInt
		for _, move := range moves {
			e.board.Push(&move)
			eval := e.alphaBeta(board, depth - 1, alpha, beta, true)
			e.board.Pop()
			if eval < minEval {
				minEval = eval
			}
			if eval < beta {
				beta = eval
			}
			if alpha > beta {
				break
			}
		}
		return minEval
	}
}
