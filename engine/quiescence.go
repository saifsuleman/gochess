package engine

func (e *Engine) quiscence(alpha, beta, rootDepth int) (score int) {
	e.NodesSearched++
	standPat := e.Evaluate()

	if standPat >= beta {
		return beta
	}

	if standPat > alpha {
		alpha = standPat
	}

	moves := e.Board.GenerateLegalCaptures()
	e.OrderMoves(moves, rootDepth)
	for _, move := range moves {
		e.Board.Push(&move)
		score := -e.quiscence(-beta, -alpha, rootDepth)
		e.Board.Pop()

		if score >= beta {
			return score
		}

		if score > alpha {
			alpha = score
		}
	}

	return alpha
}
