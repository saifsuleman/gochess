package engine

import (
	"gochess/core"
	"math"
	"math/rand"
)

/*
TODO:
- Implement a more sophisticated evaluation function
- Implement iterative deepening
- Implement move ordering
- Implement transposition tables
- Implement quiescence search
- Implement opening book
- Implement endgame tablebases
- Implement passed pawn evaluation
- Implement king safety evaluation
- Implement piece-square tables
- Implement time management
- Implement multi-threading
*/

type Engine struct {
	board *core.Board
}

func NewEngine(board *core.Board) *Engine {
	return &Engine{board: board}
}

func (e *Engine) FindBestMove() *core.Move {
	moves := e.board.GenerateLegalMoves()
	if len(moves) == 0 {
		return nil
	}

	bestMove := moves[0]
	bestValue := math.MinInt
	depth := 3

	// Shuffle moves to introduce some randomness
	rand.Shuffle(len(moves), func(i, j int) {
		moves[i], moves[j] = moves[j], moves[i]
	})

	for _, move := range moves {
		e.board.Push(&move)
		moveValue := e.alphaBeta(e.board, depth-1, math.MinInt, math.MaxInt, false)
		e.board.Pop()

		if moveValue > bestValue {
			bestValue = moveValue
			bestMove = move
		}
	}

	return &bestMove
}
