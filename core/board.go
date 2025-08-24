package core

const (
	CastlingRightsNone = 0b0000
	CastlingWhiteKingside = 0b0001
	CastlingWhiteQueenside = 0b0010
	CastlingBlackKingside = 0b0100
	CastlingBlackQueenside = 0b1000
)

type Bitboard uint64
type Position uint8

type UndoInfo struct {
	MoveFrom 				Position
	MoveTo 					Position
	CapturedPiece 	Piece
	CastlingRights 	uint8
	EnPassantTarget Position
}

type Board struct {
	Pieces [64]Piece
	PieceBitboards [2][6]Bitboard
	WhitePieces Bitboard
	BlackPieces Bitboard
	WhiteToMove bool
	EnPassantTarget Position
	CastlingRights uint8
	UndoInfo []UndoInfo
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
    }
}

func (b *Board) removePiece(pos Position, piece Piece) {
	if pos >= 64 || piece == PieceNone {
		return
	}
	b.Pieces[pos] = PieceNone
	color := (piece & PieceColorMask) >> 3
	type_ := (piece & PieceTypeMask) - 1
	b.PieceBitboards[color][type_] &^= (1 << pos)
	if color == 0 {
		b.WhitePieces &^= (1 << pos)
	} else {
		b.BlackPieces &^= (1 << pos)
	}
}

func (b *Board) addPiece(pos Position, piece Piece) {
	if pos >= 64 || piece == PieceNone {
		return
	}
	b.Pieces[pos] = piece
	color := (piece & PieceColorMask) >> 3
	type_ := (piece & PieceTypeMask) - 1
	b.PieceBitboards[color][type_] |= (1 << pos)
	if color == 0 {
		b.WhitePieces |= (1 << pos)
	} else {
		b.BlackPieces |= (1 << pos)
	}
}

func (b *Board) Pop() {
	if len(b.UndoInfo) == 0 {
		return
	}
	undo := b.UndoInfo[len(b.UndoInfo)-1]
	b.UndoInfo = b.UndoInfo[:len(b.UndoInfo)-1]

	from := undo.MoveFrom
	to := undo.MoveTo
	captured := undo.CapturedPiece

	b.addPiece(from, b.Pieces[to])
	b.removePiece(to, b.Pieces[to])

	if captured != PieceNone {
		b.addPiece(to, captured)
	}

	b.EnPassantTarget = undo.EnPassantTarget
	b.CastlingRights = undo.CastlingRights
	b.WhiteToMove = !b.WhiteToMove
}

func (b *Board) Push(move Move) {
	from := move.From
	to := move.To
	piece := b.Pieces[from]
	captured := b.Pieces[to]

	undo := UndoInfo{
		MoveFrom: from,
		MoveTo: to,
		CapturedPiece: captured,
		CastlingRights: b.CastlingRights,
		EnPassantTarget: b.EnPassantTarget,
	}
	b.UndoInfo = append(b.UndoInfo, undo)

	b.EnPassantTarget = 64

	if captured != PieceNone {
		b.removePiece(to, captured)
	}

	b.removePiece(from, piece)
	b.addPiece(to, piece)

	pieceType := piece & PieceTypeMask

	switch pieceType {
	case PieceTypePawn:
		if abs(int(to) - int(from)) == 16 {
			if b.WhiteToMove {
				b.EnPassantTarget = from + 8
			} else {
				b.EnPassantTarget = from - 8
			}
		}

		if move.IsEnPassant() {
			var capSq Position
			if b.WhiteToMove {
				capSq = to - 8
			} else {
				capSq = to + 8
			}
			captured = b.Pieces[capSq]
			b.removePiece(capSq, captured)
		}

		if move.Promotion != PieceNone {
			b.removePiece(to, piece)
			b.addPiece(to, move.Promotion)
		}
	case PieceTypeKing:
		if b.WhiteToMove {
			b.CastlingRights &^= CastlingWhiteKingside | CastlingWhiteQueenside
		} else {
			b.CastlingRights &^= CastlingBlackKingside | CastlingBlackQueenside
		}

		if move.IsCastleKingside() {
			if b.WhiteToMove {
				b.removePiece(7, PieceWhiteRook)
				b.addPiece(5, PieceWhiteRook)
			} else {
				b.removePiece(63, PieceBlackRook)
				b.addPiece(61, PieceBlackRook)
			}
		} else if move.IsCastleQueenside() {
			if b.WhiteToMove {
				b.removePiece(0, PieceWhiteRook)
				b.addPiece(3, PieceWhiteRook)
			} else {
				b.removePiece(56, PieceBlackRook)
				b.addPiece(59, PieceBlackRook)
			}
		}
	case PieceTypeRook:
		switch from {
		case 0:
			b.CastlingRights &^= CastlingWhiteQueenside
		case 7:
			b.CastlingRights &^= CastlingWhiteKingside
		case 56:
			b.CastlingRights &^= CastlingBlackQueenside
		case 63:
			b.CastlingRights &^= CastlingBlackKingside
		}
	}

	// TODO: check if captured type is a rook and is the right color
	switch to {
	case 0:
		b.CastlingRights &^= CastlingWhiteQueenside
	case 7:
		b.CastlingRights &^= CastlingWhiteKingside
	case 56:
		b.CastlingRights &^= CastlingBlackQueenside
	case 63:
		b.CastlingRights &^= CastlingBlackKingside
	}

	b.WhiteToMove = !b.WhiteToMove
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
