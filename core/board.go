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

func BoardEquals(a, b *Board) bool {
	if a.WhiteToMove != b.WhiteToMove ||
		a.EnPassantTarget != b.EnPassantTarget ||
		a.CastlingRights != b.CastlingRights ||
		len(a.UndoInfo) != len(b.UndoInfo) {
		return false
	}

	for i := range 64 {
		if a.Pieces[i] != b.Pieces[i] {
			return false
		}
	}

	for color := range 2 {
		for type_ := range 6 {
			if a.PieceBitboards[color][type_] != b.PieceBitboards[color][type_] {
				return false
			}
		}
	}

	return true
}

func (b *Board) RemovePiece(pos Position, piece Piece) {
	if pos >= 64 || piece == PieceNone {
		return
	}
	b.Pieces[pos] = PieceNone
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
	movedPiece := b.Pieces[to]  // The piece that moved

	// Undo move: put moving piece back to 'from'
	b.RemovePiece(to, movedPiece)
	b.AddPiece(from, movedPiece)

	// Restore captured piece
	if undo.CapturedPiece != PieceNone {
		b.AddPiece(to, undo.CapturedPiece)
	}

	// Restore en passant target and castling rights
	b.EnPassantTarget = undo.EnPassantTarget
	b.CastlingRights = undo.CastlingRights

	// Handle undoing castling rook move
	pieceType := movedPiece & PieceTypeMask
	if pieceType == PieceTypeKing {
		switch from {
		case 4: // White king starting square
			// Kingside castle
			switch to {
			case 6:
				b.RemovePiece(5, PieceWhiteRook)
				b.AddPiece(7, PieceWhiteRook)
			case 2: // Queenside castle
				b.RemovePiece(3, PieceWhiteRook)
				b.AddPiece(0, PieceWhiteRook)
			}
		case 60: // Black king starting square
			// Kingside castle
			if to == 62 {
				b.RemovePiece(61, PieceBlackRook)
				b.AddPiece(63, PieceBlackRook)
			} else if to == 58 { // Queenside castle
				b.RemovePiece(59, PieceBlackRook)
				b.AddPiece(56, PieceBlackRook)
			}
		}
	}

	// Handle undoing en passant capture
	if pieceType == PieceTypePawn && abs(int(to)-int(from)) == 7 || abs(int(to)-int(from)) == 9 {
		if undo.CapturedPiece != PieceNone && (to == b.EnPassantTarget) {
			// Captured pawn was removed via en passant, restore it
			capPos := Position(0)
			if (movedPiece&PieceColorMask)>>3 == 0 { // White pawn
				capPos = to - 8
			} else {
				capPos = to + 8
			}
			b.AddPiece(capPos, undo.CapturedPiece)
		}
	}

	// Toggle side to move
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
		b.RemovePiece(to, captured)
	}

	b.RemovePiece(from, piece)
	b.AddPiece(to, piece)

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
			b.RemovePiece(capSq, captured)
		}

		if move.Promotion != PieceNone {
			b.RemovePiece(to, piece)
			b.AddPiece(to, move.Promotion)
		}
	case PieceTypeKing:
		if b.WhiteToMove {
			b.CastlingRights &^= CastlingWhiteKingside | CastlingWhiteQueenside
		} else {
			b.CastlingRights &^= CastlingBlackKingside | CastlingBlackQueenside
		}

		if move.IsCastleKingside() {
			if b.WhiteToMove {
				b.RemovePiece(7, PieceWhiteRook)
				b.AddPiece(5, PieceWhiteRook)
			} else {
				b.RemovePiece(63, PieceBlackRook)
				b.AddPiece(61, PieceBlackRook)
			}
		} else if move.IsCastleQueenside() {
			if b.WhiteToMove {
				b.RemovePiece(0, PieceWhiteRook)
				b.AddPiece(3, PieceWhiteRook)
			} else {
				b.RemovePiece(56, PieceBlackRook)
				b.AddPiece(59, PieceBlackRook)
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
