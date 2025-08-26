package engine

func (e *Engine) negamax(depth int, alpha, beta int) int {
    if depth == 0 {
        return e.quiscence(alpha, beta)
    }

		board := e.Board
    moves := board.GenerateLegalMoves()
    if len(moves) == 0 {
        if board.InCheck(board.WhiteToMove) {
            return 30000 - depth
        }
        return 0 // stalemate
    }

    best := -100000
    for _, move := range moves {
        board.Push(&move)
        score := -e.negamax(depth - 1, -beta, -alpha)
        board.Pop()

        if score > best {
            best = score
        }
        if best > alpha {
            alpha = best
        }
        if alpha >= beta {
            break
        }
    }
    return best
}
