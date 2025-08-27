package core

import (
	"math/rand"
	"time"
)

const (
	numPieceTypes = 12
	numSquares    = 64
)

var zobristTable [numPieceTypes][numSquares]uint64
var zobristBlackToMove uint64
var zobristCastlingRights [16]uint64
var zobristEnPassantFile [8]uint64

func zobristInit() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for pt := range numPieceTypes {
		for sq := range numSquares {
			zobristTable[pt][sq] = rng.Uint64()
		}
	}

	zobristBlackToMove = rng.Uint64()
	for i := range 16 {
		zobristCastlingRights[i] = rng.Uint64()
	}

	for i := range 8 {
		zobristEnPassantFile[i] = rng.Uint64()
	}
}

func pieceToZobristIndex(piece Piece) int {
	if piece.Color() == PieceColorWhite {
		switch piece.Type() {
		case PieceTypePawn:
			return 0
		case PieceTypeKnight:
			return 1
		case PieceTypeBishop:
			return 2
		case PieceTypeRook:
			return 3
		case PieceTypeQueen:
			return 4
		case PieceTypeKing:
			return 5
		}
	} else {
		switch piece.Type() {
		case PieceTypePawn:
			return 6
		case PieceTypeKnight:
			return 7
		case PieceTypeBishop:
			return 8
		case PieceTypeRook:
			return 9
		case PieceTypeQueen:
			return 10
		case PieceTypeKing:
			return 11
		}
	}

	return -1
}

func (b *Board) ComputeZobristHash() uint64 {
	var hash uint64
	mask := b.AllPieces
	for mask != 0 {
		sq := mask.PopLSB()
		piece := b.Pieces[sq]
		idx := pieceToZobristIndex(piece)
		if idx != -1 {
			hash ^= zobristTable[idx][sq]
		}
	}

	if !b.WhiteToMove {
		hash ^= zobristBlackToMove
	}

	hash ^= zobristCastlingRights[b.CastlingRights]
	if b.EnPassantTarget != 64 {
		file := b.EnPassantTarget & 7
		hash ^= zobristEnPassantFile[file]
	}

	return hash
}
