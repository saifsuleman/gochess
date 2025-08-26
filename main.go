package main

import (
	"gochess/core"
	"gochess/fen"
	"gochess/game"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var FEN = fen.DefaultFEN()

func sqToPen(sq core.Position) string {
	rank := sq >> 3
	file := sq & 7
	return string('a'+file) + string('1'+rank)
}

func main() {
	board, err := fen.LoadFromFEN(FEN)
	if err != nil {
		log.Fatal(err)
	}
	play(board)
}

func play(board *core.Board) {
	game.Init()
	game := game.NewGame(board)
	ebiten.SetWindowTitle("hi")
	ebiten.SetWindowSize(45*8, 45*8)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
