package core

import (
	"fmt"
	"log"
)

const (
	CastlingRightsNone = 0b0000
	CastlingWhiteKingside = 0b0001
	CastlingWhiteQueenside = 0b0010
	CastlingBlackKingside = 0b0100
	CastlingBlackQueenside = 0b1000
)

type Position uint8

type MoveHistoryEntry struct {
	From Position
	To Position
	Promotion Piece
	Captured Piece

	EnPassantTarget Position
	CastlingRights 	uint8
	IsEnPassant 		bool
}

type Board struct {
	Pieces [64]Piece
	PieceBitboards [2][6]Bitboard
	AllPieces Bitboard
	WhitePieces Bitboard
	BlackPieces Bitboard
	WhiteToMove bool
	EnPassantTarget Position
	CastlingRights uint8
	MoveHistory []MoveHistoryEntry
}

func NewBoard() *Board {
    return &Board{
				Pieces:                 [64]Piece{},
        PieceBitboards:           [2][6]Bitboard{},
				WhitePieces:             0,
				BlackPieces:             0,
        WhiteToMove:      true,
        EnPassantTarget:  64,
        CastlingRights:   CastlingWhiteKingside | CastlingWhiteQueenside |
                          CastlingBlackKingside | CastlingBlackQueenside,
				MoveHistory:      []MoveHistoryEntry{},
    }
}

func (b *Board) LastMove() *MoveHistoryEntry {
	if len(b.MoveHistory) == 0 {
		return nil
	}

	return &b.MoveHistory[len(b.MoveHistory)-1]
}

// Why are we pass
func (b *Board) Push(move *Move) {
	from := move.From
	to := move.To
	piece := b.Pieces[from]
	captured := b.Pieces[to]

	castlingRights := b.CastlingRights

	if captured != PieceNone {
		b.RemovePiece(to, captured)
	}

	b.RemovePiece(from, piece)

	if move.Promotion == PieceNone {
		b.AddPiece(to, piece)
	} else {
		b.AddPiece(to, move.Promotion)
	}

	type_ := piece & PieceTypeMask
	color := piece & PieceColorMask
	var enPassantTarget Position
	isEnPassant := false

	switch type_ {
	case PieceTypePawn:
		if color == PieceColorWhite && to - from == 7 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget - 8]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget - 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorWhite && to - from == 9 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget - 8]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget - 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorBlack && from - to == 7 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget + 8]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget + 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorBlack && from - to == 9 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget + 8]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget + 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorWhite && to - from == 16 {
			enPassantTarget = to - 8
		} else if color == PieceColorBlack && from - to == 16 {
			enPassantTarget = to + 8
		} else {
			enPassantTarget = 64 // Reset En Passant target
		}
	case PieceTypeKing:
		// TODO: replace these weird if/else with something like a switch for a faster dispatch
		if color == PieceColorWhite && from == 4 && to == 6 {
			b.RemovePiece(7, PieceWhiteRook)
			b.AddPiece(5, PieceWhiteRook)
			b.CastlingRights &^= CastlingWhiteKingside | CastlingWhiteQueenside
		} else if color == PieceColorWhite && from == 4 && to == 2 {
			b.RemovePiece(0, PieceWhiteRook)
			b.AddPiece(3, PieceWhiteRook)
			b.CastlingRights &^= CastlingWhiteQueenside | CastlingWhiteKingside
		} else if color == PieceColorBlack && from == 60 && to == 62 {
			b.RemovePiece(63, PieceBlackRook)
			b.AddPiece(61, PieceBlackRook)
			b.CastlingRights &^= CastlingBlackKingside | CastlingBlackQueenside
		} else if color == PieceColorBlack && from == 60 && to == 58 {
			b.RemovePiece(56, PieceBlackRook)
			b.AddPiece(59, PieceBlackRook)
			b.CastlingRights &^= CastlingBlackQueenside | CastlingBlackKingside
		} else {
			if color == PieceColorWhite {
				b.CastlingRights &^= (CastlingWhiteKingside | CastlingWhiteQueenside)
			} else {
				b.CastlingRights &^= (CastlingBlackKingside | CastlingBlackQueenside)
			}
		}
	case PieceTypeRook:
		if color == PieceColorWhite {
			switch from {
			case 0:
				b.CastlingRights &^= CastlingWhiteQueenside
			case 7:
				b.CastlingRights &^= CastlingWhiteKingside
			}
		} else {
			switch from {
			case 56:
				b.CastlingRights &^= CastlingBlackQueenside
			case 63:
				b.CastlingRights &^= CastlingBlackKingside
			}
		}
	}

	b.MoveHistory = append(b.MoveHistory, MoveHistoryEntry{
		From:      from,
		To:        to,
		Promotion: move.Promotion,
		Captured:  captured,
		IsEnPassant: isEnPassant,
		EnPassantTarget: b.EnPassantTarget,
		CastlingRights: castlingRights,
	})

	b.EnPassantTarget = enPassantTarget
	b.WhiteToMove = !b.WhiteToMove
}

func (b *Board) Pop() {
	if len(b.MoveHistory) == 0 {
		return
	}

	lastMove := b.MoveHistory[len(b.MoveHistory)-1]
	b.MoveHistory = b.MoveHistory[:len(b.MoveHistory)-1]

	from := lastMove.From
	to := lastMove.To
	piece := b.Pieces[to]
	captured := lastMove.Captured

	b.RemovePiece(to, piece)
	if captured != PieceNone && !lastMove.IsEnPassant {
		b.AddPiece(to, captured)
	}

	if lastMove.Promotion == PieceNone {
		b.AddPiece(from, piece)
	} else {
		b.AddPiece(from, piece & PieceColorMask | PieceTypePawn) // Revert to pawn
	}

	type_ := piece & PieceTypeMask
	color := piece & PieceColorMask

	switch type_ {
	case PieceTypePawn:
		if lastMove.IsEnPassant {
			if color == PieceColorWhite {
				b.AddPiece(lastMove.EnPassantTarget - 8, PieceBlackPawn)
			} else {
				b.AddPiece(lastMove.EnPassantTarget + 8, PieceWhitePawn)
			}
		}
	case PieceTypeKing:
		if color == PieceColorWhite && from == 4 && to == 6 {
			b.RemovePiece(5, PieceWhiteRook)
			b.AddPiece(7, PieceWhiteRook)
		}
		if color == PieceColorWhite && from == 4 && to == 2 {
			b.RemovePiece(3, PieceWhiteRook)
			b.AddPiece(0, PieceWhiteRook)
		}
		if color == PieceColorBlack && from == 60 && to == 62 {
			b.RemovePiece(61, PieceBlackRook)
			b.AddPiece(63, PieceBlackRook)
		}
		if color == PieceColorBlack && from == 60 && to == 58 {
			b.RemovePiece(59, PieceBlackRook)
			b.AddPiece(56, PieceBlackRook)
		}
	}

	b.WhiteToMove = !b.WhiteToMove
	b.EnPassantTarget = lastMove.EnPassantTarget
	b.CastlingRights = lastMove.CastlingRights
}

func (b *Board) RemovePiece(pos Position, piece Piece) {
	if b.Pieces[pos] != piece {
		log.Printf("Trying to remove piece %d at position %d, but found piece %d instead", piece, pos, b.Pieces[pos])
		piece = b.Pieces[pos]
	}

	if pos >= 64 || piece == PieceNone {
		return
	}
	b.Pieces[pos] = PieceNone
	b.AllPieces &^= (1 << pos)
	color := (piece & PieceColorMask) >> 3
	type_ := int(piece & PieceTypeMask) - 1

	if type_ >= 0 {
		b.PieceBitboards[color][type_] &^= (1 << pos)
	}

	if color == 0 {
		b.WhitePieces &^= (1 << pos)
	} else {
		b.BlackPieces &^= (1 << pos)
	}
}

func (b *Board) AddPiece(pos Position, piece Piece) {
	if pos == 7 {
		fmt.Printf("Adding piece %d at position %d\n", piece, pos)
	}

	if pos >= 64 || piece == PieceNone {
		return
	}

	b.Pieces[pos] = piece
	b.AllPieces |= (1 << pos)
	color := (piece & PieceColorMask) >> 3
	type_ := (piece & PieceTypeMask) - 1
	b.PieceBitboards[color][type_] |= (1 << pos)
	if color == 0 {
		b.WhitePieces |= (1 << pos)
	} else {
		b.BlackPieces |= (1 << pos)
	}
}


func (b *Board) WhiteCanCastleKingside() bool {
	return (b.CastlingRights & CastlingWhiteKingside) != 0
}

func (b *Board) WhiteCanCastleQueenside() bool {
	return (b.CastlingRights & CastlingWhiteQueenside) != 0
}

func (b *Board) BlackCanCastleKingside() bool {
	return (b.CastlingRights & CastlingBlackKingside) != 0
}

func (b *Board) BlackCanCastleQueenside() bool {
	return (b.CastlingRights & CastlingBlackQueenside) != 0
}

func (b *Board) ToAlgebraNotation(move Move) string {
	from := move.From
	to := move.To
	promotion := move.Promotion

	s := make([]byte, 0, 5)
	s = append(s, byte('a' + (from & 7)))
	s = append(s, byte('1' + (from >> 3)))
	s = append(s, byte('a' + (to & 7)))
	s = append(s, byte('1' + (to >> 3)))
	if promotion != PieceNone {
		switch promotion {
		case PieceWhiteQueen, PieceBlackQueen:
			s = append(s, 'q')
		case PieceWhiteRook, PieceBlackRook:
			s = append(s, 'r')
		case PieceWhiteBishop, PieceBlackBishop:
			s = append(s, 'b')
		case PieceWhiteKnight, PieceBlackKnight:
			s = append(s, 'n')
		}
	}
	return string(s)
}
