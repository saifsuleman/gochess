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
	Deadline time.Time
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

	moves := e.Board.GenerateLegalMoves()
	if len(moves) == 0 {
		return nil
	}

	e.OrderMoves(moves)

	var maxDepth int = 20 // time budget will take over before we ever finish thsi search
	var bestMove core.Move
	var bestValue int
	var totalNodes int

	for depth := 1; depth <= maxDepth; depth++ {
		currentBestMove := moves[0]
		currentBestValue := math.MinInt
		completedAllMoves := true

		for _, move := range moves {
			e.Board.Push(&move)
			moveValue, nodes := e.negamax(depth-1, math.MinInt, math.MaxInt)
			moveValue = -moveValue
			totalNodes += nodes
			e.Board.Pop()

			if e.TimeUp() {
				completedAllMoves = false
				break
			}

			if moveValue > currentBestValue {
				currentBestValue = moveValue
				currentBestMove = move
			}

			if currentBestValue >= MateThreshold {
				break
			}
		}

		// We don't use this result if its not a full completion, because that's unreliable
		if !completedAllMoves {
			break
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
	fmt.Printf("Best move: %v, Value: %d, Nodes: %d, Time taken: %s\n", bestMove, bestValue, totalNodes, elapsed)

	return &bestMove
}
