package main

import (
	"fmt"
	"gochess/core"
	"gochess/fen"
	"gochess/game"
	"log"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/hajimehoshi/ebiten/v2"
)

// const FEN = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
var FEN = fen.DefaultFEN()

func sqToPen(sq core.Position) string {
	rank := sq >> 3
	file := sq & 7
	return string('a' + file) + string('1' + rank)
}

func main() {
	board, err := fen.LoadFromFEN(FEN)
	if err != nil {
		log.Fatal(err)
	}
	// play(board)

	chessGame := chess.NewGame()
	fn, err := chess.FEN(FEN)
	if err != nil {
		log.Fatal(err)
	}
	fn(chessGame)
	countMovesAndTrack(board, chessGame, 3)


	start := time.Now()
	for depth := range 10 {
		moves := countMoves(board, depth)
		fmt.Printf("Depth %d: %d moves... elapsed: %s\n", depth, moves, time.Since(start))
	}
}

func FindChessMove(cg *chess.Game, from, to core.Position, moves []chess.Move) *chess.Move {
	for _, m := range moves {
		if sqToPen(from) == m.S1().String() && sqToPen(to) == m.S2().String() {
			return &m
		}
	}
	return nil
}

func countMoves(board *core.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	moves := board.GenerateLegalMoves()
	totalN := 0
	for _, move := range moves {
		board.Push(&move)
		n := countMoves(board, depth - 1)
		totalN += n
		board.Pop()
	}
	return totalN
}

func countMovesAndTrack(board *core.Board, cb *chess.Game, depth int) int {
	if depth == 0 {
		return 1
	}

	moves := board.GenerateLegalMoves()
	otherMoves := cb.ValidMoves()

	if len(moves) != len(otherMoves) {
		for _, m := range moves {
			found := FindChessMove(cb, m.From, m.To, otherMoves) != nil
			if !found {
				log.Printf("Extra move: %s -> %s (%s)\n", sqToPen(m.From), sqToPen(m.To), cb.FEN())
			}
		}

		for _, m := range otherMoves {
			found := false
			for _, m2 := range moves {
				if m.S1().String() == sqToPen(m2.From) && m.S2().String() == sqToPen(m2.To) {
					found = true
					break
				}
			}
			if !found {
				log.Printf("Missing move: %s -> %s (%s)\n", m.S1().String(), m.S2().String(), cb.FEN())
			}
		}

		log.Printf("Move count mismatch: %d vs %d (%s)\n", len(moves), len(otherMoves), cb.FEN())
		play(board)
	}

	if len(moves) == 0 {
		return 0
	}

	totalMoves := 0
	for _, move := range moves {
		chessMove := FindChessMove(cb, move.From, move.To, otherMoves)
		if chessMove == nil {
			log.Fatalf("Could not find move: %s -> %s (%s) .... %s\n", sqToPen(move.From), sqToPen(move.To), cb.FEN(), fen.BoardToFEN(board))
		}
		cb.Move(chessMove, nil)
		board.Push(&move)
		n := countMovesAndTrack(board, cb, depth - 1)
		totalMoves += n
		board.Pop()
		cb.GoBack()
	}

	return totalMoves
}

func play(board *core.Board) {
	game.Init()
	game := game.NewGame(board)
	ebiten.SetWindowTitle("hi")
	ebiten.SetWindowSize(45*8,45*8)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
