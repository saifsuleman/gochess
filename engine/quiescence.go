package engine

func (e *Engine) quiscence(alpha, beta int) (score int, nodes int) {
	standPat := e.Evaluate()

	if standPat >= beta {
		return beta, 1
	}

	if standPat > alpha {
		alpha = standPat
	}

	moves := e.Board.GenerateLegalCaptures()
	e.OrderMoves(moves)
	for _, move := range moves {
		e.Board.Push(&move)
		score, nodes := e.quiscence(-beta, -alpha)
		score = -score
		e.Board.Pop()

		if score >= beta {
			return score, nodes
		}

		if score > alpha {
			alpha = score
		}
	}

	return alpha, nodes
}
