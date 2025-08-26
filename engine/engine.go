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
	randomIndex := rand.Intn(len(moves))
	return &moves[randomIndex]
}

func (e *Engine) Evaluate() int {
	score := 0
	mask := e.board.AllPieces
	for mask != 0 {
		lsb := mask.PopLSB()
		piece := e.board.Pieces[lsb]
		value := e.PieceValue(piece)
		if piece.Color() == core.PieceColorWhite {
			score += value
		} else {
			score -= value
		}
	}
	return score
}

func (e *Engine) alphaBeta(board *core.Board, depth int, alpha int, beta int, maximizing bool) int {
	if depth == 0 {
		return e.Evaluate()
	}

	moves := board.GenerateLegalMoves()
	if len(moves) == 0 {
		return e.Evaluate()
	}

	if maximizing {
		maxEval := math.MaxInt
		for _, move := range moves {
			e.board.Push(&move)
			eval := e.alphaBeta(board, depth - 1, alpha, beta, false)
			e.board.Pop()

			if eval > maxEval {
				maxEval = eval
			}

			if eval > alpha {
				alpha = eval
			}

			if alpha > beta {
				break
			}
		}
		return maxEval
	} else {
		minEval := math.MinInt
		for _, move := range moves {
			e.board.Push(&move)
			eval := e.alphaBeta(board, depth - 1, alpha, beta, true)
			e.board.Pop()
			if eval < minEval {
				minEval = eval
			}
			if eval < beta {
				beta = eval
			}
			if alpha > beta {
				break
			}
		}
		return minEval
	}
}

func (e *Engine) PieceValue(piece core.Piece) int {
	type_ := piece.Type()
	switch type_ {
	case core.PieceTypeNone:
		return 0
	case core.PieceTypePawn:
		return 100
	case core.PieceTypeKnight:
		return 320
	case core.PieceTypeBishop:
		return 330
	case core.PieceTypeRook:
		return 500
	case core.PieceTypeQueen:
		return 900
	case core.PieceTypeKing:
		return 20000
	default:
		return 0
	}
}
