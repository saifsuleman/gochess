package core

type Move struct {
	From Position
	To Position
	Piece Piece
	Promotion Piece
}

func (m Move) IsEnPassant() bool {
	if m.Piece != PieceWhitePawn && m.Piece != PieceBlackPawn {
		return false
	}

	return true // TODO
}

func (m Move) IsCastleKingside() bool {
	if m.Piece != PieceWhiteKing && m.Piece != PieceBlackKing {
		return false
	}

	return (m.From == 4 && m.To == 6) || (m.From == 60 && m.To == 62) // White or Black kingside castling
}

func (m Move) IsCastleQueenside() bool {
	if m.Piece != PieceWhiteKing && m.Piece != PieceBlackKing {
		return false
	}

	return (m.From == 4 && m.To == 2) || (m.From == 60 && m.To == 58) // White or Black queenside castling
}
