package core

type Move struct {
	From Position
	To Position
	Promotion Piece
}

func (b *Board) IsMoveLegal(move Move) bool {
	return true
}

func (b *Board) generatePseudoLegalMoves() []Move {
	moves := []Move{}
	for _, piece := range b.Pieces {
		type_ :=  piece & PieceTypeMask
		color := piece & PieceColorMask

		switch type_ {
		case PieceTypePawn:
		}
	}
}
