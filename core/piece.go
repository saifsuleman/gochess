package core

const (
	PieceTypeMask  		= 0b0111
	PieceColorMask 		= 0b1000

	PieceTypeNone     = 0b0000
	PieceTypePawn     = 0b0001
	PieceTypeKnight   = 0b0010
	PieceTypeBishop   = 0b0011
	PieceTypeRook     = 0b0100
	PieceTypeQueen    = 0b0101
	PieceTypeKing     = 0b0110

	PieceColorWhite  	= 0b0000
	PieceColorBlack  	= 0b1000

	PieceNone					  = Piece(PieceTypeNone) // No piece, treated as white for simplicity
	PieceWhitePawn 			= Piece(PieceTypePawn | PieceColorWhite)
	PieceWhiteKnight 		= Piece(PieceTypeKnight | PieceColorWhite)
	PieceWhiteBishop 		= Piece(PieceTypeBishop | PieceColorWhite)
	PieceWhiteRook 			= Piece(PieceTypeRook | PieceColorWhite)
	PieceWhiteQueen 		= Piece(PieceTypeQueen | PieceColorWhite)
	PieceWhiteKing 			= Piece(PieceTypeKing | PieceColorWhite)
	PieceBlackPawn 			= Piece(PieceTypePawn | PieceColorBlack)
	PieceBlackKnight 		= Piece(PieceTypeKnight | PieceColorBlack)
	PieceBlackBishop 		= Piece(PieceTypeBishop | PieceColorBlack)
	PieceBlackRook	 		= Piece(PieceTypeRook | PieceColorBlack)
	PieceBlackQueen		 	= Piece(PieceTypeQueen | PieceColorBlack)
	PieceBlackKing			= Piece(PieceTypeKing | PieceColorBlack)
)

type Piece uint8
