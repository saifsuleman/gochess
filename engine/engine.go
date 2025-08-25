package engine

import (
	"fmt"
	"gochess/core"
	"math/rand"
)

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
	fmt.Println("legal moves", moves)
	fmt.Println("random index", randomIndex)
	return &moves[randomIndex]
}
