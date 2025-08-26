package engine

import (
	"fmt"
	"gochess/core"
	"math"
	"time"
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
	Board *core.Board
}

func NewEngine(board *core.Board) *Engine {
	return &Engine{Board: board}
}

func (e *Engine) FindBestMove() *core.Move {
	start := time.Now()

	moves := e.Board.GenerateLegalMoves()
	if len(moves) == 0 {
		return nil
	}

	e.OrderMoves(moves)

	bestMove := moves[0]
	bestValue := math.MinInt
	depth := 5

	for _, move := range moves {
		e.Board.Push(&move)
		moveValue := -e.negamax(depth-1, math.MinInt, math.MaxInt)
		e.Board.Pop()

		if moveValue > bestValue {
			bestValue = moveValue
			bestMove = move
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Best move: %v, Value: %d, Time taken: %s\n", bestMove, bestValue, elapsed)

	return &bestMove
}
