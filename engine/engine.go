package engine

import (
	"gochess/core"
	"log"
	"math"
)

const (
	PawnValue	= 100
	KnightValue = 300
	BishopValue = 300
	RookValue   = 500
	QueenValue  = 900
	KingValue   = 10000
)

const (
	FlagUpper = iota
	FlagExact
	FlagLower
)

type Engine struct {
	board *core.Board
	table map[string]int
}

type TTEntry struct {
	depth int
	score int
	bestMove core.Move
	flag int
}

func NewEngine(board *core.Board) *Engine {
	return &Engine{
		board: board,
	}
}

func (e *Engine) FindBestMove() *core.Move {
	var bestMove *core.Move = nil
	bestScore := math.MinInt32
	moveCount := 0

	// Iterate through all possible moves
	for _, move := range e.board.GenerateLegalMoves() {
		boardCopy := *e.board

		e.board.Push(move)

		// Evaluate the position after the move
		score, count := e.AlphaBeta(3, math.MinInt32, math.MaxInt32, false)
		moveCount += count

		e.board.Pop()

		if !core.BoardEquals(e.board, &boardCopy) {
			e.board = &boardCopy // Restore the board state
		}

		if score > bestScore {
			bestScore = score
			bestMove = &move
		}
	}

	if bestMove == nil {
		return nil // No valid moves found
	}
	return bestMove
}

func (e *Engine) EvaluatePosition() int {
	var score int = 0
	for _, piece := range e.board.Pieces {
		if piece == core.PieceNone {
			continue
		}

		type_ := piece.Type()
		color := piece.Color()

		var value int
		switch type_ {
		case core.PieceTypePawn:
			value = PawnValue
		case core.PieceTypeKnight:
			value = KnightValue
		case core.PieceTypeBishop:
			value = BishopValue
		case core.PieceTypeRook:
			value = RookValue
		case core.PieceTypeQueen:
			value = QueenValue
		case core.PieceTypeKing:
			value = KingValue
		}

		perspective := 1
		if color == core.PieceColorBlack {
			perspective = -1
		}

		score += perspective * value
	}

	return score
}

func (e *Engine) AlphaBeta(depth int, alpha int, beta int, maximizing bool) (int, int) {
	if depth == 0 {
		return e.EvaluatePosition(), 1
	}

	moves := e.board.GenerateLegalMoves()

	totalMoves := 0
	if maximizing {
		maxEval := math.MinInt32
		for _, move := range moves {
			boardCopy := *e.board

			e.board.Push(move)
			eval, moveCount := e.AlphaBeta(depth-1, alpha, beta, false)
			e.board.Pop()

			if !core.BoardEquals(e.board, &boardCopy) {
				log.Println("Board state mismatch after move:", move, move.IsEnPassant())
				e.board = &boardCopy // Restore the board state
			}

			if eval > maxEval {
				maxEval = eval
			}

			totalMoves += moveCount
			alpha = int(math.Max(float64(alpha), float64(eval)))

			if beta <= alpha {
				break // Beta cut-off
			}
		}

		return maxEval, totalMoves
	} else {
		minEval := math.MaxInt32
		for _, move := range moves {
			boardCopy := *e.board

			e.board.Push(move)
			eval, moveCount := e.AlphaBeta(depth-1, alpha, beta, true)
			e.board.Pop()

			if !core.BoardEquals(e.board, &boardCopy) {
				log.Println("Board state mismatch after move:", move, move.IsEnPassant())
				e.board = &boardCopy // Restore the board state
			}

			if eval < minEval {
				minEval = eval
			}

			totalMoves += moveCount
			beta = int(math.Min(float64(beta), float64(eval)))

			if beta <= alpha {
				break // Alpha cut-off
			}
		}

		return minEval, totalMoves
	}
}
