package core

// ============================================================================
// PUBLIC API
// ============================================================================

// GenerateLegalMoves returns all legal moves for the current position
func (b *Board) GenerateLegalMoves() []Move {
	pseudoMoves := b.GeneratePseudoLegalMoves()
	legalMoves := make([]Move, 0, len(pseudoMoves))

	for _, move := range pseudoMoves {
		if b.IsMoveLegal(move) {
			legalMoves = append(legalMoves, move)
		}
	}

	return legalMoves
}

// GenerateLegalCaptures returns all legal capturing moves for the current position
func (b *Board) GenerateLegalCaptures() []Move {
	pseudoMoves := b.GeneratePseudoLegalMoves()
	legalCaptures := make([]Move, 0, len(pseudoMoves))

	for _, move := range pseudoMoves {
		// Check if move is a capture (target square occupied or en passant)
		isCapture := b.Pieces[move.To] != PieceNone ||
			(move.Piece&PieceTypeMask == PieceTypePawn && Position(move.To) == b.EnPassantTarget)

		if isCapture && b.IsMoveLegal(move) {
			legalCaptures = append(legalCaptures, move)
		}
	}

	return legalCaptures
}

// IsMoveLegal checks if a move is legal (doesn't leave king in check)
func (b *Board) IsMoveLegal(move Move) bool {
	piece := move.Piece

	// Basic validation
	if piece == PieceNone || b.Pieces[move.From] != piece {
		return false
	}
	if sameColor(b.Pieces[move.To], piece) {
		return false
	}

	// Ensure move is pseudo-legal before expensive operations
	if !b.isPseudoLegal(move) {
		return false
	}

	// Test legality by making the move and checking for check
	b.Push(move)
	kingInCheck := b.isKingInCheck(!b.WhiteToMove) // WhiteToMove flipped after Push
	b.Pop()

	return !kingInCheck
}

// ============================================================================
// PSEUDO-LEGAL MOVE GENERATION
// ============================================================================

// GeneratePseudoLegalMoves generates all pseudo-legal moves (ignoring check)
func (b *Board) GeneratePseudoLegalMoves() []Move {
	var moves []Move
	occupancy := b.WhitePieces | b.BlackPieces
	sideIdx := sideIndex(b.WhiteToMove)

	// Generate moves for each piece type
	moves = append(moves, b.generatePawnMoves(sideIdx, occupancy)...)
	moves = append(moves, b.generatePieceMoves(sideIdx, PieceTypeKnight)...)
	moves = append(moves, b.generatePieceMoves(sideIdx, PieceTypeBishop)...)
	moves = append(moves, b.generatePieceMoves(sideIdx, PieceTypeRook)...)
	moves = append(moves, b.generatePieceMoves(sideIdx, PieceTypeQueen)...)
	moves = append(moves, b.generateKingMoves(sideIdx)...)

	return moves
}

// isPseudoLegal validates if a move follows piece movement rules (ignoring check)
func (b *Board) isPseudoLegal(move Move) bool {
	piece := move.Piece

	// Basic validation
	if piece == PieceNone || b.Pieces[move.From] != piece || sameColor(b.Pieces[move.To], piece) {
		return false
	}

	// Check movement rules by piece type
	pieceType := piece & PieceTypeMask
	switch pieceType {
	case PieceTypePawn:
		return b.isPawnMoveLegal(move)
	case PieceTypeKnight:
		return knightMask(move.From)&bb(move.To) != 0
	case PieceTypeBishop:
		return b.isSliderMoveLegal(move.From, move.To, bishopDirections)
	case PieceTypeRook:
		return b.isSliderMoveLegal(move.From, move.To, rookDirections)
	case PieceTypeQueen:
		return b.isSliderMoveLegal(move.From, move.To, queenDirections)
	case PieceTypeKing:
		return b.isKingMoveLegal(move)
	}

	return false
}

// ============================================================================
// PIECE-SPECIFIC MOVE GENERATORS
// ============================================================================

// generatePieceMoves generates moves for knights, bishops, rooks, and queens
func (b *Board) generatePieceMoves(sideIdx int, pieceType Piece) []Move {
	var moves []Move
	pieceBitboard := b.PieceBitboards[sideIdx][pieceType-1]

	for pieceBitboard != 0 {
		fromSquare := popLSB(&pieceBitboard)
		piece := b.Pieces[fromSquare]

		switch pieceType {
		case PieceTypeKnight:
			moves = append(moves, b.generateKnightMoves(Position(fromSquare), piece)...)
		case PieceTypeBishop:
			moves = append(moves, b.generateSlidingMoves(Position(fromSquare), piece, bishopDirections)...)
		case PieceTypeRook:
			moves = append(moves, b.generateSlidingMoves(Position(fromSquare), piece, rookDirections)...)
		case PieceTypeQueen:
			moves = append(moves, b.generateSlidingMoves(Position(fromSquare), piece, queenDirections)...)
		}
	}

	return moves
}

// generatePawnMoves generates all pawn moves including captures, pushes, and en passant
func (b *Board) generatePawnMoves(sideIdx int, occupancy Bitboard) []Move {
	var moves []Move
	isWhite := sideIdx == 0
	pawnBitboard := b.PieceBitboards[sideIdx][PieceTypePawn-1]
	enemyPieces := b.BlackPieces

	if !isWhite {
		enemyPieces = b.WhitePieces
	}

	if isWhite {
		moves = append(moves, b.generateWhitePawnMoves(pawnBitboard, occupancy, enemyPieces)...)
	} else {
		moves = append(moves, b.generateBlackPawnMoves(pawnBitboard, occupancy, enemyPieces)...)
	}

	return moves
}

// generateWhitePawnMoves handles white pawn movement
func (b *Board) generateWhitePawnMoves(pawns, occupancy, enemyPieces Bitboard) []Move {
	var moves []Move

	// Single pushes (one square forward)
	singlePushes := (pawns << 8) &^ occupancy

	// Double pushes (two squares from starting rank)
	doublePushes := ((singlePushes & rank3) << 8) &^ occupancy

	// Captures
	leftCaptures := (pawns << 7) &^ fileH & enemyPieces   // Capture up-left, avoid H->A wrap
	rightCaptures := (pawns << 9) &^ fileA & enemyPieces  // Capture up-right, avoid A->H wrap

	// Generate en passant captures
	moves = append(moves, b.generateEnPassantMoves(pawns, true)...)

	// Generate regular moves
	moves = append(moves, b.generatePawnPushMoves(singlePushes, -8, PieceWhitePawn, false)...)
	moves = append(moves, b.generatePawnPushMoves(doublePushes, -16, PieceWhitePawn, false)...)
	moves = append(moves, b.generatePawnCaptureMoves(leftCaptures, -7, PieceWhitePawn)...)
	moves = append(moves, b.generatePawnCaptureMoves(rightCaptures, -9, PieceWhitePawn)...)

	return moves
}

// generateBlackPawnMoves handles black pawn movement
func (b *Board) generateBlackPawnMoves(pawns, occupancy, enemyPieces Bitboard) []Move {
	var moves []Move

	// Single pushes (one square forward for black = down the board)
	singlePushes := (pawns >> 8) &^ occupancy

	// Double pushes (two squares from starting rank)
	doublePushes := ((singlePushes & rank6) >> 8) &^ occupancy

	// Captures (diagonal attacks)
	leftCaptures := (pawns >> 7) &^ fileA & enemyPieces   // Capture down-right, avoid A->H wrap
	rightCaptures := (pawns >> 9) &^ fileH & enemyPieces  // Capture down-left, avoid H->A wrap

	// Generate en passant captures
	moves = append(moves, b.generateEnPassantMoves(pawns, false)...)

	// Generate regular moves
	moves = append(moves, b.generatePawnPushMoves(singlePushes, 8, PieceBlackPawn, false)...)
	moves = append(moves, b.generatePawnPushMoves(doublePushes, 16, PieceBlackPawn, false)...)
	moves = append(moves, b.generatePawnCaptureMoves(leftCaptures, 7, PieceBlackPawn)...)
	moves = append(moves, b.generatePawnCaptureMoves(rightCaptures, 9, PieceBlackPawn)...)

	return moves
}

// generateEnPassantMoves generates en passant capture moves
func (b *Board) generateEnPassantMoves(pawns Bitboard, isWhite bool) []Move {
	var moves []Move

	if b.EnPassantTarget >= 64 {
		return moves
	}

	epSquare := bb(b.EnPassantTarget)
	var epCaptures Bitboard
	var fromOffset int
	piece := PieceBlackPawn

	if isWhite {
		piece = PieceWhitePawn
		epLeft := (pawns << 7) &^ fileH & epSquare
		epRight := (pawns << 9) &^ fileA & epSquare
		epCaptures = epLeft | epRight

		for epCaptures != 0 {
			to := popLSB(&epCaptures)
			// Determine which pawn made the capture
			if b.Pieces[to-7] == piece && (to%8) != 0 {
				fromOffset = -7
			} else {
				fromOffset = -9
			}
			moves = append(moves, Move{
				From:  Position(to + fromOffset),
				To:    Position(to),
				Piece: piece,
			})
		}
	} else {
		epLeft := (pawns >> 7) &^ fileA & epSquare
		epRight := (pawns >> 9) &^ fileH & epSquare
		epCaptures = epLeft | epRight

		for epCaptures != 0 {
			to := popLSB(&epCaptures)
			// Determine which pawn made the capture
			if b.Pieces[to+7] == piece && (to%8) != 7 {
				fromOffset = 7
			} else {
				fromOffset = 9
			}
			moves = append(moves, Move{
				From:  Position(to + fromOffset),
				To:    Position(to),
				Piece: piece,
			})
		}
	}

	return moves
}

// generatePawnPushMoves generates pawn push moves (single and double)
func (b *Board) generatePawnPushMoves(pushes Bitboard, fromOffset int, piece Piece, isCapture bool) []Move {
	var moves []Move

	// Separate promotion moves from regular moves
	var promotionRank Bitboard
	if isWhite(piece) {
		promotionRank = rank8
	} else {
		promotionRank = rank1
	}

	// Regular pushes (non-promotion)
	regularPushes := pushes &^ promotionRank
	for regularPushes != 0 {
		to := popLSB(&regularPushes)
		moves = append(moves, Move{
			From:  Position(to + fromOffset),
			To:    Position(to),
			Piece: piece,
		})
	}

	// Promotion pushes
	promotionPushes := pushes & promotionRank
	for promotionPushes != 0 {
		to := popLSB(&promotionPushes)
		from := Position(to + fromOffset)
		moves = append(moves, generatePromotionMoves(from, Position(to), piece)...)
	}

	return moves
}

// generatePawnCaptureMoves generates pawn capture moves
func (b *Board) generatePawnCaptureMoves(captures Bitboard, fromOffset int, piece Piece) []Move {
	var moves []Move

	// Separate promotion captures from regular captures
	var promotionRank Bitboard
	if isWhite(piece) {
		promotionRank = rank8
	} else {
		promotionRank = rank1
	}

	// Regular captures (non-promotion)
	regularCaptures := captures &^ promotionRank
	for regularCaptures != 0 {
		to := popLSB(&regularCaptures)
		moves = append(moves, Move{
			From:  Position(to + fromOffset),
			To:    Position(to),
			Piece: piece,
		})
	}

	// Promotion captures
	promotionCaptures := captures & promotionRank
	for promotionCaptures != 0 {
		to := popLSB(&promotionCaptures)
		from := Position(to + fromOffset)
		moves = append(moves, generatePromotionMoves(from, Position(to), piece)...)
	}

	return moves
}

// generateKnightMoves generates all knight moves from a square
func (b *Board) generateKnightMoves(from Position, piece Piece) []Move {
	var moves []Move
	targets := knightMask(from) &^ b.getOwnPieces(isWhite(piece))

	for targets != 0 {
		to := popLSB(&targets)
		moves = append(moves, Move{
			From:  from,
			To:    Position(to),
			Piece: piece,
		})
	}

	return moves
}

// generateSlidingMoves generates moves for sliding pieces (bishop, rook, queen)
func (b *Board) generateSlidingMoves(from Position, piece Piece, directions []int) []Move {
	var moves []Move
	ownPieces := b.getOwnPieces(isWhite(piece))

	for _, direction := range directions {
		current := int(from)

		for {
			next, valid := step(current, direction)
			if !valid {
				break
			}

			current = next
			squareBB := bb(Position(current))

			// Blocked by own piece
			if (squareBB & ownPieces) != 0 {
				break
			}

			// Valid move
			moves = append(moves, Move{
				From:  from,
				To:    Position(current),
				Piece: piece,
			})

			// Capture stops sliding
			if b.Pieces[current] != PieceNone {
				break
			}
		}
	}

	return moves
}

// generateKingMoves generates king moves including castling
func (b *Board) generateKingMoves(sideIdx int) []Move {
	var moves []Move
	kingBitboard := b.PieceBitboards[sideIdx][PieceTypeKing-1]

	if kingBitboard == 0 {
		return moves
	}

	from := Position(bitscanLSB(kingBitboard))
	piece := b.Pieces[from]

	// Regular king moves
	targets := kingMask(from) &^ b.getOwnPieces(isWhite(piece))
	for targets != 0 {
		to := popLSB(&targets)
		moves = append(moves, Move{
			From:  from,
			To:    Position(to),
			Piece: piece,
		})
	}

	// Castling moves
	moves = append(moves, b.generateCastlingMoves(from, piece)...)

	return moves
}

// generateCastlingMoves generates castling moves for the king
func (b *Board) generateCastlingMoves(from Position, piece Piece) []Move {
	var moves []Move

	if piece == PieceWhiteKing && from == 4 {
		// White kingside castling
		if (b.CastlingRights&CastlingWhiteKingside) != 0 &&
			b.Pieces[5] == PieceNone && b.Pieces[6] == PieceNone &&
			!b.isSquareAttacked(4, false) && !b.isSquareAttacked(5, false) && !b.isSquareAttacked(6, false) {
			moves = append(moves, Move{From: 4, To: 6, Piece: PieceWhiteKing})
		}

		// White queenside castling
		if (b.CastlingRights&CastlingWhiteQueenside) != 0 &&
			b.Pieces[3] == PieceNone && b.Pieces[2] == PieceNone && b.Pieces[1] == PieceNone &&
			!b.isSquareAttacked(4, false) && !b.isSquareAttacked(3, false) && !b.isSquareAttacked(2, false) {
			moves = append(moves, Move{From: 4, To: 2, Piece: PieceWhiteKing})
		}
	} else if piece == PieceBlackKing && from == 60 {
		// Black kingside castling
		if (b.CastlingRights&CastlingBlackKingside) != 0 &&
			b.Pieces[61] == PieceNone && b.Pieces[62] == PieceNone &&
			!b.isSquareAttacked(60, true) && !b.isSquareAttacked(61, true) && !b.isSquareAttacked(62, true) {
			moves = append(moves, Move{From: 60, To: 62, Piece: PieceBlackKing})
		}

		// Black queenside castling
		if (b.CastlingRights&CastlingBlackQueenside) != 0 &&
			b.Pieces[59] == PieceNone && b.Pieces[58] == PieceNone && b.Pieces[57] == PieceNone &&
			!b.isSquareAttacked(60, true) && !b.isSquareAttacked(59, true) && !b.isSquareAttacked(58, true) {
			moves = append(moves, Move{From: 60, To: 58, Piece: PieceBlackKing})
		}
	}

	return moves
}

// ============================================================================
// MOVE VALIDATION HELPERS
// ============================================================================

// isPawnMoveLegal validates pawn moves including en passant
func (b *Board) isPawnMoveLegal(move Move) bool {
	from, to := move.From, move.To
	piece := move.Piece
	isWhitePlayer := isWhite(piece)

	if isWhitePlayer {
		// Single push forward
		if to == from+8 && b.Pieces[to] == PieceNone {
			return true
		}

		// Double push from starting rank
		if from/8 == 1 && to == from+16 &&
			b.Pieces[from+8] == PieceNone && b.Pieces[to] == PieceNone {
			return true
		}

		// Diagonal captures
		if (to == from+7 && from%8 != 0) || (to == from+9 && from%8 != 7) {
			// Regular capture
			if b.Pieces[to] != PieceNone && !sameColor(b.Pieces[to], piece) {
				return true
			}
			// En passant capture
			if Position(to) == b.EnPassantTarget {
				return true
			}
		}
	} else {
		// Black pawn moves (mirror of white logic)
		if to == from-8 && b.Pieces[to] == PieceNone {
			return true
		}

		if from/8 == 6 && to == from-16 &&
			b.Pieces[from-8] == PieceNone && b.Pieces[to] == PieceNone {
			return true
		}

		if (to == from-7 && from%8 != 7) || (to == from-9 && from%8 != 0) {
			if b.Pieces[to] != PieceNone && !sameColor(b.Pieces[to], piece) {
				return true
			}
			if Position(to) == b.EnPassantTarget {
				return true
			}
		}
	}

	return false
}

// isSliderMoveLegal validates sliding piece moves (bishop, rook, queen)
func (b *Board) isSliderMoveLegal(from, to Position, directions []int) bool {
	for _, direction := range directions {
		current := int(from)

		for {
			next, valid := step(current, direction)
			if !valid {
				break
			}

			current = next
			if Position(current) == to {
				return true
			}

			// Path blocked
			if b.Pieces[current] != PieceNone {
				break
			}
		}
	}

	return false
}

// isKingMoveLegal validates king moves including castling structure
func (b *Board) isKingMoveLegal(move Move) bool {
	// Regular king move
	if kingMask(move.From)&bb(move.To) != 0 {
		return true
	}

	// Castling moves
	if move.IsCastleKingside() || move.IsCastleQueenside() {
		return b.isCastlingStructureValid(move)
	}

	return false
}

// isCastlingStructureValid checks castling rights and empty squares
func (b *Board) isCastlingStructureValid(move Move) bool {
	piece := move.Piece
	from, to := move.From, move.To

	if piece == PieceWhiteKing && from == 4 {
		if to == 6 { // Kingside
			return (b.CastlingRights&CastlingWhiteKingside) != 0 &&
				b.Pieces[5] == PieceNone && b.Pieces[6] == PieceNone
		}
		if to == 2 { // Queenside
			return (b.CastlingRights&CastlingWhiteQueenside) != 0 &&
				b.Pieces[1] == PieceNone && b.Pieces[2] == PieceNone && b.Pieces[3] == PieceNone
		}
	}

	if piece == PieceBlackKing && from == 60 {
		if to == 62 { // Kingside
			return (b.CastlingRights&CastlingBlackKingside) != 0 &&
				b.Pieces[61] == PieceNone && b.Pieces[62] == PieceNone
		}
		if to == 58 { // Queenside
			return (b.CastlingRights&CastlingBlackQueenside) != 0 &&
				b.Pieces[57] == PieceNone && b.Pieces[58] == PieceNone && b.Pieces[59] == PieceNone
		}
	}

	return false
}

// ============================================================================
// ATTACK DETECTION
// ============================================================================

// isKingInCheck determines if the king of the specified color is in check
func (b *Board) isKingInCheck(isWhiteKing bool) bool {
	kingSquare := b.findKingSquare(isWhiteKing)
	if kingSquare >= 64 {
		return false // King not found (shouldn't happen in valid position)
	}
	return b.isSquareAttacked(kingSquare, !isWhiteKing)
}

// isSquareAttacked checks if a square is attacked by the specified side
func (b *Board) isSquareAttacked(square Position, byWhite bool) bool {
	attackingSide := sideIndex(byWhite)

	// Check pawn attacks
	if b.isPawnAttacking(square, byWhite) {
		return true
	}

	// Check knight attacks
	if b.isKnightAttacking(square, attackingSide) {
		return true
	}

	// Check sliding piece attacks
	if b.isSlidingPieceAttacking(square, attackingSide, bishopDirections, PieceTypeBishop) ||
		b.isSlidingPieceAttacking(square, attackingSide, rookDirections, PieceTypeRook) {
		return true
	}

	// Check king attacks
	if b.isKingAttacking(square, attackingSide) {
		return true
	}

	return false
}

// isPawnAttacking checks if any pawn of the specified color attacks the square
func (b *Board) isPawnAttacking(square Position, byWhite bool) bool {
	attackingSide := sideIndex(byWhite)
	pawnBitboard := b.PieceBitboards[attackingSide][PieceTypePawn-1]

	if byWhite {
		// White pawns attack diagonally upward (+7, +9)
		// So we check if there are white pawns diagonally below the square
		if square >= 9 && (square%8) != 0 { // Check bounds and file wrap
			if pawnBitboard&bb(square-9) != 0 {
				return true
			}
		}
		if square >= 7 && (square%8) != 7 { // Check bounds and file wrap
			if pawnBitboard&bb(square-7) != 0 {
				return true
			}
		}
	} else {
		// Black pawns attack diagonally downward (-7, -9)
		// So we check if there are black pawns diagonally above the square
		if square <= 55 && (square%8) != 7 { // Check bounds and file wrap
			if pawnBitboard&bb(square+7) != 0 {
				return true
			}
		}
		if square <= 54 && (square%8) != 0 { // Check bounds and file wrap
			if pawnBitboard&bb(square+9) != 0 {
				return true
			}
		}
	}

	return false
}

// isKnightAttacking checks if any knight attacks the square
func (b *Board) isKnightAttacking(square Position, attackingSide int) bool {
	knightMask := knightMask(square)
	knightBitboard := b.PieceBitboards[attackingSide][PieceTypeKnight-1]
	return (knightMask & knightBitboard) != 0
}

// isSlidingPieceAttacking checks if any sliding piece attacks the square
func (b *Board) isSlidingPieceAttacking(square Position, attackingSide int, directions []int, pieceType Piece) bool {
	for _, direction := range directions {
		current := int(square)

		for {
			next, valid := step(current, direction)
			if !valid {
				break
			}

			current = next
			piece := b.Pieces[current]

			if piece == PieceNone {
				continue
			}

			// Found a piece - check if it's an attacking piece
			if sideIndex(isWhite(piece)) == attackingSide {
				currentPieceType := piece & PieceTypeMask
				if currentPieceType == pieceType || currentPieceType == PieceTypeQueen {
					return true
				}
			}

			// Path blocked by any piece
			break
		}
	}

	return false
}

// isKingAttacking checks if the enemy king attacks the square
func (b *Board) isKingAttacking(square Position, attackingSide int) bool {
	kingMask := kingMask(square)
	kingBitboard := b.PieceBitboards[attackingSide][PieceTypeKing-1]
	return (kingMask & kingBitboard) != 0
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// findKingSquare efficiently finds the king's position using bitboards
func (b *Board) findKingSquare(isWhiteKing bool) Position {
	sideIdx := sideIndex(isWhiteKing)
	kingBitboard := b.PieceBitboards[sideIdx][PieceTypeKing-1]

	if kingBitboard == 0 {
		return 64 // King not found
	}

	return Position(bitscanLSB(kingBitboard))
}

// getOwnPieces returns bitboard of own pieces for the specified color
func (b *Board) getOwnPieces(isWhitePlayer bool) Bitboard {
	if isWhitePlayer {
		return b.WhitePieces
	}
	return b.BlackPieces
}

// generatePromotionMoves creates all four promotion moves for a pawn
func generatePromotionMoves(from, to Position, pawn Piece) []Move {
	if isWhite(pawn) {
		return []Move{
			{From: from, To: to, Piece: pawn, Promotion: PieceWhiteQueen},
			{From: from, To: to, Piece: pawn, Promotion: PieceWhiteRook},
			{From: from, To: to, Piece: pawn, Promotion: PieceWhiteBishop},
			{From: from, To: to, Piece: pawn, Promotion: PieceWhiteKnight},
		}
	}
	return []Move{
		{From: from, To: to, Piece: pawn, Promotion: PieceBlackQueen},
		{From: from, To: to, Piece: pawn, Promotion: PieceBlackRook},
		{From: from, To: to, Piece: pawn, Promotion: PieceBlackBishop},
		{From: from, To: to, Piece: pawn, Promotion: PieceBlackKnight},
	}
}

// ============================================================================
// BITBOARD OPERATIONS
// ============================================================================

// bb creates a bitboard with a single bit set at the specified position
func bb(position Position) Bitboard {
	return Bitboard(1) << position
}

// popLSB pops the least significant bit and returns its index
func popLSB(bitboard *Bitboard) int {
	b := *bitboard
	if b == 0 {
		return -1
	}

	lsb := b & -b
	index := bitscanLSB(b)
	*bitboard = b &^ lsb

	return index
}

// bitscanLSB finds the index of the least significant bit using de Bruijn multiplication
func bitscanLSB(bitboard Bitboard) int {
	if bitboard == 0 {
		return -1
	}

	const deBruijn64 = Bitboard(0x03f79d71b4cb0a89)
	index64 := [64]int{
		0, 1, 48, 2, 57, 49, 28, 3,
		61, 58, 50, 42, 38, 29, 17, 4,
		62, 55, 59, 36, 53, 51, 43, 22,
		45, 39, 33, 30, 24, 18, 12, 5,
		63, 47, 56, 27, 60, 41, 37, 16,
		54, 35, 52, 21, 44, 32, 23, 11,
		46, 26, 40, 15, 34, 20, 31, 10,
		25, 14, 19, 9, 13, 8, 7, 6,
	}

	return index64[((bitboard&-bitboard)*deBruijn64)>>58]
}

// ============================================================================
// PIECE MOVEMENT MASKS
// ============================================================================

// knightMask generates attack mask for a knight on the given square
func knightMask(square Position) Bitboard {
	file := int(square % 8)
	rank := int(square / 8)
	var mask Bitboard

	// All 8 possible knight moves
	knightMoves := []struct{ fileOffset, rankOffset int }{
		{1, 2}, {2, 1}, {-1, 2}, {-2, 1},
		{1, -2}, {2, -1}, {-1, -2}, {-2, -1},
	}

	for _, move := range knightMoves {
		targetFile := file + move.fileOffset
		targetRank := rank + move.rankOffset

		if targetFile >= 0 && targetFile < 8 && targetRank >= 0 && targetRank < 8 {
			mask |= bb(Position(targetRank*8 + targetFile))
		}
	}

	return mask
}

// kingMask generates attack mask for a king on the given square
func kingMask(square Position) Bitboard {
	file := int(square % 8)
	rank := int(square / 8)
	var mask Bitboard

	// All 8 directions a king can move
	for fileOffset := -1; fileOffset <= 1; fileOffset++ {
		for rankOffset := -1; rankOffset <= 1; rankOffset++ {
			if fileOffset == 0 && rankOffset == 0 {
				continue // Skip the current square
			}

			targetFile := file + fileOffset
			targetRank := rank + rankOffset

			if targetFile >= 0 && targetFile < 8 && targetRank >= 0 && targetRank < 8 {
				mask |= bb(Position(targetRank*8 + targetFile))
			}
		}
	}

	return mask
}

// step moves one square in a direction, with bounds and wrap-around checking
func step(current int, direction int) (int, bool) {
	next := current + direction

	// Check board boundaries
	if next < 0 || next >= 64 {
		return 0, false
	}

	currentFile := current % 8
	nextFile := next % 8
	currentRank := current / 8
	nextRank := next / 8

	// Validate move based on direction to prevent wrapping
	switch direction {
	case +1, -1: // Horizontal moves
		return next, nextRank == currentRank

	case +8, -8: // Vertical moves
		return next, nextFile == currentFile

	case +9, -9, +7, -7: // Diagonal moves
		fileChange := abs(nextFile - currentFile)
		rankChange := abs(nextRank - currentRank)
		return next, fileChange == 1 && rankChange == 1

	default:
		return 0, false
	}
}

// ============================================================================
// PIECE HELPER FUNCTIONS
// ============================================================================

// isWhite checks if a piece is white
func isWhite(piece Piece) bool {
	return (piece&PieceColorMask)>>3 == 0
}

// sameColor checks if two pieces are the same color
func sameColor(piece1, piece2 Piece) bool {
	if piece1 == PieceNone || piece2 == PieceNone {
		return false
	}
	return (piece1 & PieceColorMask) == (piece2 & PieceColorMask)
}

// sideIndex converts color boolean to array index (0=white, 1=black)
func sideIndex(isWhitePlayer bool) int {
	if isWhitePlayer {
		return 0
	}
	return 1
}

// ============================================================================
// CONSTANTS AND DIRECTION VECTORS
// ============================================================================

// File and rank bitboard constants
var (
	fileA = Bitboard(0x0101010101010101)
	fileB = fileA << 1
	fileG = fileA << 6
	fileH = fileA << 7

	rank1 = Bitboard(0xFF)
	rank2 = rank1 << (8 * 1)
	rank3 = rank1 << (8 * 2)
	rank6 = rank1 << (8 * 5)
	rank7 = rank1 << (8 * 6)
	rank8 = rank1 << (8 * 7)
)

// Movement direction vectors
var (
	rookDirections   = []int{+1, -1, +8, -8}           // Horizontal and vertical
	bishopDirections = []int{+9, -9, +7, -7}           // Diagonal
	queenDirections  = []int{+1, -1, +8, -8, +9, -9, +7, -7} // All directions
)
