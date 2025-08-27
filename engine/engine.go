package engine

import (
	"fmt"
	"gochess/core"
	"math"
	"time"
)

/*
TODO:
- Implement iterative deepening [DONE]
- Implement move ordering [DONE]
- Implement transposition tables [DONE]
- Implement quiescence search [DONE]
- Implement null move pruning [DONE]
- Implement late move reductions (LMR) [DONE]
- Implement time management [DONE]
- Basic material exchange evaluation [DONE]
- Implement piece-square-tables [DONE]

- Implement killer moves
- Implement opening book
- Implement endgame tablebases
- Implement passed pawn evaluation
- Implement king safety evaluation
- Implement SEE

- Implement multi-threading
*/

type Engine struct {
	Board *core.Board
	TT *TranspositionalTable
	Deadline time.Time
	NodesSearched uint64
	Aborted bool
}

func (e *Engine) TimeUp() bool {
	return time.Now().After(e.Deadline)
}

func NewEngine(board *core.Board) *Engine {
	return &Engine{Board: board, TT: NewTranspositionalTable(256)}
}

func (e *Engine) FindBestMove(timeBudget time.Duration) *core.Move {
	start := time.Now()

	e.Deadline = start.Add(timeBudget)
	e.NodesSearched = 0

	moves := e.Board.GenerateLegalMoves()
	if len(moves) == 0 {
		return nil
	}

	e.OrderMoves(moves)

	var maxDepth int = 30 // time budget will take over before we ever finish thsi search
	var bestMove core.Move
	var bestValue int
	depthReached := 0

	for depth := 1; depth <= maxDepth; depth++ {
		if e.TimeUp() {
			break
		}

		currentBestMove := moves[0]
		currentBestValue := math.MinInt

		for _, move := range moves {
			e.Board.Push(&move)
			e.Aborted = false
			moveValue := -e.negamax(depth-1, math.MinInt, math.MaxInt)
			e.Board.Pop()

			if e.Aborted {
				break // Do not use a partially searched path
			}

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

		depthReached = depth
	}

	elapsed := time.Since(start)
	fmt.Printf("Best move: %v, depth: %d, nodes searched: %d at depth %d, Time taken: %s\n", bestMove, bestValue, e.NodesSearched, depthReached, elapsed)

	return &bestMove
}
