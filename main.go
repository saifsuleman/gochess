package main

import (
	"fmt"
	"gochess/core"
	"gochess/fen"
	"gochess/game"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	board, err := fen.LoadFromFEN(fen.DefaultFEN())
	if err != nil {
		log.Fatal(err)
	}

	for depth := range 5 {
		moveCount := countMoves(board, depth, false)
		fmt.Printf("ILLLegal Move count at ply %d: %d\n", depth, moveCount)
	}

	for depth := range 5 {
		moveCount := countMoves(board, depth, true)
		fmt.Printf("Legal Move count at ply %d: %d\n", depth, moveCount)
	}
}

func countMoves(board *core.Board, depth int, legal bool) int {
	if depth == 0 {
		return 1
	}

	var moves []core.Move
	if legal {
		moves = board.GenerateLegalMoves()
	} else {
		moves = board.GeneratePseudoLegalMoves()
	}
	if len(moves) == 0 {
		return 0
	}

	totalMoves := 0
	for _, move := range moves {
		board.Push(&move)
		totalMoves += countMoves(board, depth - 1, legal)
		board.Pop()
	}

	return totalMoves
}

func play() {
	core.MoveTest()

	game.Init()
	game := game.NewGame()
	ebiten.SetWindowTitle("hi")
	ebiten.SetWindowSize(45*8,45*8)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
