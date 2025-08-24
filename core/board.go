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
	IsEnPassant			bool
	IsCastleKingside bool
	IsCastleQueenside bool
	PromotedFrom		Piece  // Store original piece type for promotion undo
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

func (b *Board) DeepCopy() *Board {
    copyBoard := *b // shallow copy first

    // Deep copy the UndoInfo slice
    if b.UndoInfo != nil {
        copyBoard.UndoInfo = make([]UndoInfo, len(b.UndoInfo))
        copy(copyBoard.UndoInfo, b.UndoInfo)
    }

    return &copyBoard
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
	movedPiece := b.Pieces[to] // Piece that was moved (after the move)

	// Undo the main move
	b.RemovePiece(to, movedPiece)
	b.AddPiece(from, movedPiece)

	// Restore captured piece if any
	if undo.CapturedPiece != PieceNone {
		if undo.IsEnPassant {
			// En passant: captured piece is not at `to`, but adjacent rank
			var capPos Position
			if (movedPiece&PieceColorMask)>>3 == 0 { // White pawn
				capPos = to - 8
			} else {
				capPos = to + 8
			}
			b.AddPiece(capPos, undo.CapturedPiece)
		} else {
			// Normal capture
			b.AddPiece(to, undo.CapturedPiece)
		}
	}

	// Restore special state
	b.EnPassantTarget = undo.EnPassantTarget
	b.CastlingRights = undo.CastlingRights

	// Undo castling rook moves
	pieceType := movedPiece & PieceTypeMask
	if pieceType == PieceTypeKing {
		if undo.IsCastleKingside {
			rookFrom, rookTo := Position(0), Position(0)
			rookPiece := PieceNone
			if (movedPiece & PieceColorMask) == 0 { // White
				rookFrom = 7
				rookTo = 5
				rookPiece = PieceWhiteRook
			} else { // Black
				rookFrom = 63
				rookTo = 61
				rookPiece = PieceBlackRook
			}
			b.RemovePiece(rookTo, rookPiece)
			b.AddPiece(rookFrom, rookPiece)
		} else if undo.IsCastleQueenside {
			rookFrom, rookTo := Position(0), Position(0)
			rookPiece := PieceNone
			if (movedPiece & PieceColorMask) == 0 { // White
				rookFrom = 0
				rookTo = 3
				rookPiece = PieceWhiteRook
			} else { // Black
				rookFrom = 56
				rookTo = 59
				rookPiece = PieceBlackRook
			}
			b.RemovePiece(rookTo, rookPiece)
			b.AddPiece(rookFrom, rookPiece)
		}
	}

	// Undo promotion: restore original piece if this was a promotion
	if undo.PromotedFrom != PieceNone {
		b.RemovePiece(from, movedPiece)
		b.AddPiece(from, undo.PromotedFrom)
	}

	// Toggle turn back
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
		IsEnPassant: move.IsEnPassant(),
		IsCastleKingside: move.IsCastleKingside(),
		IsCastleQueenside: move.IsCastleQueenside(),
		PromotedFrom: PieceNone,
	}

	b.EnPassantTarget = 64

	// Handle captured piece
	if captured != PieceNone {
		b.RemovePiece(to, captured)
	}

	// Make the basic move
	b.RemovePiece(from, piece)
	b.AddPiece(to, piece)

	pieceType := piece & PieceTypeMask
	switch pieceType {
	case PieceTypePawn:
		// Double pawn move - set en passant target
		if abs(int(to) - int(from)) == 16 {
			if b.WhiteToMove {
				b.EnPassantTarget = from + 8
			} else {
				b.EnPassantTarget = from - 8
			}
		}

		// En passant capture
		if b.EnPassantTarget == move.To && move.IsEnPassant() {
			var capSq Position
			if b.WhiteToMove {
				capSq = to - 8
			} else {
				capSq = to + 8
			}
			capturedPawn := b.Pieces[capSq]
			undo.CapturedPiece = capturedPawn // Store the actually captured piece
			b.RemovePiece(capSq, capturedPawn)
		}

		// Promotion
		if move.Promotion != PieceNone {
			undo.PromotedFrom = piece // Store original piece for undo
			b.RemovePiece(to, piece)
			b.AddPiece(to, move.Promotion)
		}

	case PieceTypeKing:
		// King move removes castling rights
		if b.WhiteToMove {
			b.CastlingRights &^= CastlingWhiteKingside | CastlingWhiteQueenside
		} else {
			b.CastlingRights &^= CastlingBlackKingside | CastlingBlackQueenside
		}

		// Handle castling rook moves
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
		// Rook move removes castling rights for that rook
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

	// Handle captured rook affecting castling rights
	if captured != PieceNone && (captured&PieceTypeMask) == PieceTypeRook {
		switch to {
		case 0:
			if captured == PieceWhiteRook {
				b.CastlingRights &^= CastlingWhiteQueenside
			}
		case 7:
			if captured == PieceWhiteRook {
				b.CastlingRights &^= CastlingWhiteKingside
			}
		case 56:
			if captured == PieceBlackRook {
				b.CastlingRights &^= CastlingBlackQueenside
			}
		case 63:
			if captured == PieceBlackRook {
				b.CastlingRights &^= CastlingBlackKingside
			}
		}
	}

	// Store the updated undo info
	b.UndoInfo = append(b.UndoInfo, undo)

	// Toggle turn
	b.WhiteToMove = !b.WhiteToMove
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
