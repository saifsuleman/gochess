package core

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

func (b *Board) Push(move *Move) {
	from := move.From
	to := move.To
	piece := b.Pieces[from]
	captured := b.Pieces[to]

	if captured != PieceNone {
		b.RemovePiece(to, captured)
	}

	b.RemovePiece(from, piece)
	b.AddPiece(to, piece)

	type_ := piece & PieceTypeMask
	color := piece & PieceColorMask
	var enPassantTarget Position
	isEnPassant := false

	switch type_ {
	case PieceTypePawn:
		if move.Promotion != PieceNone {
			b.RemovePiece(to, piece)
			b.AddPiece(to, move.Promotion)
		}

		if color == PieceColorWhite && to - from == 7 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget - 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorWhite && to - from == 9 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget - 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorBlack && from - to == 7 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget]
			if removedPiece != PieceNone {
				b.RemovePiece(b.EnPassantTarget + 8, removedPiece)
				captured = removedPiece
				isEnPassant = true
			}
		} else if color == PieceColorBlack && from - to == 9 && b.EnPassantTarget == to {
			removedPiece := b.Pieces[b.EnPassantTarget]
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
			b.CastlingRights &^= CastlingWhiteKingside
		} else if color == PieceColorWhite && from == 4 && to == 2 {
			b.RemovePiece(0, PieceWhiteRook)
			b.AddPiece(3, PieceWhiteRook)
			b.CastlingRights &^= CastlingWhiteQueenside
		} else if color == PieceColorBlack && from == 60 && to == 62 {
			b.RemovePiece(63, PieceBlackRook)
			b.AddPiece(61, PieceBlackRook)
			b.CastlingRights &^= CastlingBlackKingside
		} else if color == PieceColorBlack && from == 60 && to == 58 {
			b.RemovePiece(56, PieceBlackRook)
			b.AddPiece(59, PieceBlackRook)
			b.CastlingRights &^= CastlingBlackQueenside
		}
	}

	b.MoveHistory = append(b.MoveHistory, MoveHistoryEntry{
		From:      from,
		To:        to,
		Promotion: move.Promotion,
		Captured:  captured,
		IsEnPassant: isEnPassant,
		EnPassantTarget: b.EnPassantTarget,
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

	b.AddPiece(from, piece)

	if lastMove.Promotion != PieceNone {
		b.RemovePiece(to, lastMove.Promotion)
		b.AddPiece(from, piece) // Revert promotion
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
			b.CastlingRights |= CastlingWhiteKingside
		}
		if color == PieceColorWhite && from == 4 && to == 2 {
			b.RemovePiece(3, PieceWhiteRook)
			b.AddPiece(0, PieceWhiteRook)
			b.CastlingRights |= CastlingWhiteQueenside
		}
		if color == PieceColorBlack && from == 60 && to == 62 {
			b.RemovePiece(61, PieceBlackRook)
			b.AddPiece(63, PieceBlackRook)
			b.CastlingRights |= CastlingBlackKingside
		}
		if color == PieceColorBlack && from == 60 && to == 58 {
			b.RemovePiece(59, PieceBlackRook)
			b.AddPiece(56, PieceBlackRook)
			b.CastlingRights |= CastlingBlackQueenside
		}
	}

	b.WhiteToMove = !b.WhiteToMove
	b.EnPassantTarget = lastMove.EnPassantTarget
}

func (b *Board) RemovePiece(pos Position, piece Piece) {
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
