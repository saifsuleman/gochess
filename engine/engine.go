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
- Implement killer moves [DONE]

- Implement opening book
- Implement endgame tablebases
- Implement passed pawn evaluation
- Implement king safety evaluation
- Implement SEE
- Implement dynamic time budget allocation

- Implement multi-threading
*/

const maxDepth = 24

type Engine struct {
	Board *core.Board
	TT *TranspositionalTable
	Deadline time.Time
	NodesSearched uint64
	Aborted bool
	KillerMoves [maxDepth + 1][2]core.Move
	HistoryTable [64][64]int
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

	e.OrderMoves(moves, 0)

	var bestMove core.Move = moves[0]
	var bestValue int
	depthReached := 0

	for depth := 1; depth <= maxDepth; depth++ {
		if e.TimeUp() {
			break
		}

		currentBestMove := moves[0]
		currentBestValue := math.MinInt
		completedSearch := true

		for _, move := range moves {
			e.Board.Push(&move)
			e.Aborted = false
			moveValue := -e.negamax(depth-1, math.MinInt, math.MaxInt, depth)
			e.Board.Pop()

			if e.Aborted {
				completedSearch = false
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

		if !completedSearch {
			break
		}

		bestMove = currentBestMove
		bestValue = currentBestValue
		depthReached = depth

		for i, m := range moves {
			if m == bestMove {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Best move: %v, depth: %d, nodes searched: %d at depth %d, Time taken: %s\n", bestMove, bestValue, e.NodesSearched, depthReached, elapsed)

	return &bestMove
}

func (e *Engine) addKillerMove(move core.Move, depth int) {
	if e.KillerMoves[depth][0] != move {
		e.KillerMoves[depth][1] = e.KillerMoves[depth][0]
		e.KillerMoves[depth][0] = move
	}
}
