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
- Implement iterative deepening [DONE]
- Implement move ordering [DONE]
- Implement transposition tables [DONE]
- Implement quiescence search [DONE]
- Implement null move pruning [DONE]
- Implement late move reductions (LMR) [DONE]
- Implement killer moves
- Implement opening book
- Implement endgame tablebases
- Implement passed pawn evaluation
- Implement king safety evaluation
- Implement piece-square-tables
- Implement time management
- Implement multi-threading
*/

type Engine struct {
	Board *core.Board
	TT *TranspositionalTable
}

func NewEngine(board *core.Board) *Engine {
	return &Engine{Board: board, TT: NewTranspositionalTable(256)}
}

func (e *Engine) FindBestMove(timeBudget time.Duration) *core.Move {
	start := time.Now()

	moves := e.Board.GenerateLegalMoves()
	if len(moves) == 0 {
		return nil
	}

	e.OrderMoves(moves)

	var maxDepth int = 6 // time budget will take over before we ever finish thsi search
	var bestMove core.Move
	var bestValue int

	for depth := 1; depth <= maxDepth; depth++ {
		currentBestMove := moves[0]
		currentBestValue := math.MinInt

		for _, move := range moves {
			e.Board.Push(&move)
			moveValue := -e.negamax(depth-1, math.MinInt, math.MaxInt)
			e.Board.Pop()

			if moveValue > currentBestValue {
				currentBestValue = moveValue
				currentBestMove = move
			}

			if currentBestValue >= MateThreshold {
				break
			}
		}

		bestMove = currentBestMove
		bestValue = currentBestValue

		for i, m := range moves {
			if m == bestMove {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}

		// TODO: some form of check to see if we have enough time to continue
	}

	elapsed := time.Since(start)
	fmt.Printf("Best move: %v, Value: %d, Time taken: %s\n", bestMove, bestValue, elapsed)

	return &bestMove
}
