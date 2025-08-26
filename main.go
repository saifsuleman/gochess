package main

import (
	"gochess/core"
	"gochess/game"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	core.MoveTest()

	game.Init()
	game := game.NewGame()
	ebiten.SetWindowTitle("hi")
	ebiten.SetWindowSize(45*8,45*8)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
