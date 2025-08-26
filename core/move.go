package core

import (
	"log"
	"slices"
)

var (
	RookDirections = []Direction{ {1, 0}, {-1, 0}, {0, 1}, {0, -1} }
	BishopDirections = []Direction{ {1, 1}, {1, -1}, {-1, 1}, {-1, -1} }

	slidingRook 		= NewSlidingPiece([]Direction{ {1, 0}, {-1, 0}, {0, 1}, {0, -1} }, rookMagics, rookShifts)
	slidingBishop 	= NewSlidingPiece([]Direction{ {1, 1}, {1, -1}, {-1, 1}, {-1, -1} }, bishopMagics, bishopShifts)

	kingAttacks   = computeKingAttacks()
	knightAttacks = computeKnightAttacks()
	pawnAttacks	 = computePawnAttacks()

	zeroOccupancyRookAttacks [64]Bitboard = computeFixedAttacks(slidingRook, 0)
	zeroOccupancyBishopAttacks [64]Bitboard = computeFixedAttacks(slidingBishop, 0)
)

type SlidingPiece struct {
	interiorMasks [64]Bitboard
	magic [64]Bitboard
	shifts [64]int
	attacks [64][]Bitboard
	directions []Direction
}

type Move struct {
	From      Position
	To        Position
	Promotion Piece
}

type Direction struct {
	dr, df int
}

// TODO: optimize later
func (b *Board) IsMoveLegal(move Move) bool {
	moves := b.GenerateLegalMoves()
	return slices.Contains(moves, move)
}

func NewSlidingPiece(directions []Direction, magic [64]Bitboard, shifts [64]int) *SlidingPiece {
	interiorMasks := computeInteriorMasks(directions)
	attacks := [64][]Bitboard{}

	for sq := range Position(64) {
		interiorMask := interiorMasks[sq]
		attacks[sq] = computeAttacks(sq, directions, interiorMask, magic[sq], shifts[sq])
	}

	return &SlidingPiece{
		directions: directions,
		magic: magic,
		shifts: shifts,
		interiorMasks: interiorMasks,
		attacks: attacks,
	}
}

func (b *Board) InCheck(white bool) bool {
	kingSq := b.KingSquare(white)
	if kingSq >= 64 {
		return false
	}
	attackerIsWhite := !white
	attacking := b.GetAttackingBitboard(kingSq, attackerIsWhite)
	return (attacking & (1 << kingSq)) != 0
}

func (b *Board) GenerateLegalCaptures() []Move {
	pseudoLegalMoves := b.GeneratePseudoLegalMoves()
	captures := []Move{}
	for _, move := range pseudoLegalMoves {
		if (b.AllPieces & (1 << move.To)) != 0 {
			captures = append(captures, move)
		}
	}
	legalCaptures := b.FilterLegality(captures)
	return legalCaptures
}

func (b *Board) GenerateLegalMoves() []Move {
	pseudoLegalMoves := b.GeneratePseudoLegalMoves()
	legalMoves := b.FilterLegality(pseudoLegalMoves)
	return legalMoves
}

func (b *Board) FilterLegality(pseudoLegalMoves []Move) []Move {
	var friendlyBit int
	var enemyBit int

	if b.WhiteToMove {
		friendlyBit = 0
		enemyBit = 1
	} else {
		friendlyBit = 1
		enemyBit = 0
	}

	legalMoves := []Move{}
	kingSq := b.KingSquare(b.WhiteToMove)

	if kingSq >= 64 {
		return legalMoves
	}

	knightCheckers := knightAttacks[kingSq] & b.PieceBitboards[enemyBit][PieceTypeKnight - 1]
	pawnCheckers := pawnAttacks[friendlyBit][kingSq] & b.PieceBitboards[enemyBit][PieceTypePawn - 1]
	rookCheckers := b.GetSlidingAttacks(slidingRook, int(kingSq)) & (b.PieceBitboards[enemyBit][PieceTypeRook - 1] | b.PieceBitboards[enemyBit][PieceTypeQueen - 1])
	bishopCheckers := b.GetSlidingAttacks(slidingBishop, int(kingSq)) & (b.PieceBitboards[enemyBit][PieceTypeBishop - 1] | b.PieceBitboards[enemyBit][PieceTypeQueen - 1])
	checkers := knightCheckers | pawnCheckers | rookCheckers | bishopCheckers
	checkersN := checkers.PopCount()

	// This needs to operate under a hypothetical occupancy of 0
	rookLikePinners := zeroOccupancyRookAttacks[kingSq] & (b.PieceBitboards[enemyBit][PieceTypeRook-1] | b.PieceBitboards[enemyBit][PieceTypeQueen-1])
	bishopLikePinners := zeroOccupancyBishopAttacks[kingSq] & (b.PieceBitboards[enemyBit][PieceTypeBishop-1] | b.PieceBitboards[enemyBit][PieceTypeQueen-1])

	pinners := rookLikePinners | bishopLikePinners
	occ := b.AllPieces

	var pinnedBB Bitboard = 0
	var pinAllowedMask [64]Bitboard

	for p := pinners; p != 0; {
		pinnerSq := Position(p.PopLSB())
		between := betweenMask(kingSq, pinnerSq)
		blockers := between & occ
		if blockers.PopCount() == 1 {
			pinnedSq := Position(blockers.LSB())
			isOurPiece := false
			if friendlyBit == 0 {
				isOurPiece = ((b.WhitePieces)>>pinnedSq)&1 == 1
			} else {
				isOurPiece = ((b.BlackPieces)>>pinnedSq)&1 == 1
			}
			if isOurPiece {
				pinnedBB |= 1 << pinnedSq
				pinAllowedMask[pinnedSq] = between | (1 << pinnerSq)
			}
		}
	}

	var checkMask Bitboard = 0
	switch checkersN {
	case 0:
		checkMask = ^Bitboard(0)
	case 1:
		checkerSq := Position(checkers.LSB())
		checkMask = (1 << checkerSq)
		if rookCheckers != 0 || bishopCheckers != 0 {
			checkMask |= betweenMask(kingSq, checkerSq)
		}
	default:
		checkMask = 0
	}

	attackerIsWhite := !b.WhiteToMove

	for _, move := range pseudoLegalMoves {
		from, to := move.From, move.To
		type_ := b.Pieces[from] & PieceTypeMask

		if from == kingSq {
			occ_ := (occ &^ (1 << from)) | (1 << to)

			var enemyAfter [6]Bitboard
			for i := range 6 {
				enemyAfter[i] = b.PieceBitboards[enemyBit][i] &^ (1 << to)
			}

			if !squareIsAttackedUnderOcc(to, attackerIsWhite, occ_, &enemyAfter) {
				legalMoves = append(legalMoves, move)
			}

			continue
		}

		if checkersN >= 2 {
			continue
		}

		if (pinnedBB & (1 << from)) != 0 {
			if pinAllowedMask[from] & (1 << to) == 0 {
				continue
			}
		}

		if checkersN == 1 {
			if (checkMask & (1 << to)) == 0 {
				continue
			}
		}

		// special case here that i lost my mind over: en passant can reveal a check
		if type_ == PieceTypePawn && to == b.EnPassantTarget {
			var captureSq Position
			if b.WhiteToMove {
				captureSq = to - 8
			} else {
				captureSq = to + 8
			}

			occ_ := (occ &^ (1 << from) &^ (1 << captureSq)) | (1 << to)

			var enemyAfter [6]Bitboard
			for i := range 6 {
				enemyAfter[i] = b.PieceBitboards[enemyBit][i]
			}
			enemyAfter[PieceTypePawn - 1] &^= 1 << captureSq

			if squareIsAttackedUnderOcc(kingSq, attackerIsWhite, occ_, &enemyAfter) {
				continue
			}

			legalMoves = append(legalMoves, move)
			continue
		}

		legalMoves = append(legalMoves, move)
	}

	return legalMoves
}

func betweenMask(a, b Position) Bitboard {
	if a == b {
		return 0
	}

	ar := int(a >> 3)
	af := int(a & 7)
	br := int(b >> 3)
	bf := int(b & 7)

	dr := br - ar
	df := bf - af

	var stepR, stepF int
	if dr == 0 {
		stepR = 0
	} else if dr > 0 {
		stepR = 1
	} else {
		stepR = -1
	}
	if df == 0 {
		stepF = 0
	} else if df > 0 {
		stepF = 1
	} else {
		stepF = -1
	}

	if !(dr == 0 || df == 0 || abs(dr) == abs(df)) {
		return 0
	}

	var mask Bitboard = 0
	step := 1
	for {
		r := ar + step*stepR
		f := af + step*stepF
		if r < 0 || r > 7 || f < 0 || f > 7 {
			break
		}
		pos := Position(r*8 + f)
		if pos == b {
			break
		}
		mask |= Bitboard(1) << pos
		step++
	}
	return mask
}

func getSlidingAttacksWithOcc(sp *SlidingPiece, sq int, occ Bitboard) Bitboard {
	mask := sp.interiorMasks[sq]
	occMasked := occ & mask
	index := (occMasked * sp.magic[sq]) >> sp.shifts[sq]
	return sp.attacks[sq][index]
}

func squareIsAttackedUnderOcc(pos Position, attackerIsWhite bool, occ Bitboard, enemyPieces *[6]Bitboard) bool {
	var pawnIndex int

	// pawn attacks are inverted because this is from the perspective of the king being attacked...
	if attackerIsWhite {
		pawnIndex = 1
	} else {
		pawnIndex = 0
	}

	if (pawnAttacks[pawnIndex][pos] & enemyPieces[PieceTypePawn-1]) != 0 {
		return true
	}

	if (knightAttacks[pos] & enemyPieces[PieceTypeKnight-1]) != 0 {
		return true
	}

	if (kingAttacks[pos] & enemyPieces[PieceTypeKing-1]) != 0 {
		return true
	}

	rookAtts := getSlidingAttacksWithOcc(slidingRook, int(pos), occ)
	if (rookAtts & (enemyPieces[PieceTypeRook-1] | enemyPieces[PieceTypeQueen-1])) != 0 {
		return true
	}
	bishopAtts := getSlidingAttacksWithOcc(slidingBishop, int(pos), occ)
	if (bishopAtts & (enemyPieces[PieceTypeBishop-1] | enemyPieces[PieceTypeQueen-1])) != 0 {
		return true
	}

	return false
}

func (b *Board) GeneratePseudoLegalMoves() []Move {
	moves := []Move{}

	var ourBitboard Bitboard
	var enemyBitboard Bitboard

	if b.WhiteToMove {
		ourBitboard = b.WhitePieces
		enemyBitboard = b.BlackPieces
	} else {
		ourBitboard = b.BlackPieces
		enemyBitboard = b.WhitePieces
	}

	for bb := ourBitboard; bb != 0; {
		sq := Position(bb.PopLSB())
		piece := b.Pieces[sq]
		type_ := piece.Type()
		color := piece.Color()

		switch type_ {
		case PieceTypeKnight:
			attacks := knightAttacks[sq] & ^ourBitboard
			for attacks != 0 {
				to := Position(attacks.PopLSB())
				moves = append(moves, Move{ From: sq, To: to })
			}
		case PieceTypeKing:
			attacks := kingAttacks[sq] & ^ourBitboard
			for attacks != 0 {
				to := Position(attacks.PopLSB())
				moves = append(moves, Move{ From: sq, To: to })
			}

			if color == PieceColorWhite && b.WhiteCanCastleKingside() {
				if ((b.AllPieces >> 5) & 1) == 0 && ((b.AllPieces >> 6) & 1) == 0 && !b.IsSquareAttacked(4, false) && !b.IsSquareAttacked(5, false) && !b.IsSquareAttacked(6, false) {
					moves = append(moves, Move{ From: sq, To: 6 })
				}
			}

			if color == PieceColorWhite && b.WhiteCanCastleQueenside() {
				if ((b.AllPieces >> 1) & 1) == 0 && ((b.AllPieces >> 2) & 1) == 0 && ((b.AllPieces >> 3) & 1) == 0 && !b.IsSquareAttacked(4, false) && !b.IsSquareAttacked(3, false) && !b.IsSquareAttacked(2, false) {
					moves = append(moves, Move{ From: sq, To: 2 })
				}
			}

			if color == PieceColorBlack && b.BlackCanCastleKingside() {
				if ((b.AllPieces >> 61) & 1) == 0 && ((b.AllPieces >> 62) & 1) == 0 && !b.IsSquareAttacked(60, true) && !b.IsSquareAttacked(61, true) && !b.IsSquareAttacked(62, true) {
					moves = append(moves, Move{ From: sq, To: 62 })
				}
			}

			if color == PieceColorBlack && b.BlackCanCastleQueenside() {
				if ((b.AllPieces >> 57) & 1) == 0 && ((b.AllPieces >> 58) & 1) == 0 && ((b.AllPieces >> 59) & 1) == 0 && !b.IsSquareAttacked(60, true) && !b.IsSquareAttacked(59, true) && !b.IsSquareAttacked(58, true) {
					moves = append(moves, Move{ From: sq, To: 58 })
				}
			}
		case PieceTypePawn:
			var attacks Bitboard
			if color == PieceColorWhite {
				attacks = pawnAttacks[0][sq] & enemyBitboard
			} else {
				attacks = pawnAttacks[1][sq] & enemyBitboard
			}

			var promotionRank Position
			if color == PieceColorWhite {
				promotionRank = 6
			} else {
				promotionRank = 1
			}

			for attacks != 0 {
				to := Position(attacks.PopLSB())
				if (sq >> 3) == promotionRank {
					for _, promoType := range []uint8{ PieceTypeQueen, PieceTypeRook, PieceTypeBishop, PieceTypeKnight } {
						moves = append(moves, Move{ From: sq, To: to, Promotion: Piece(color | promoType) })
					}
				} else {
					moves = append(moves, Move{ From: sq, To: to })
				}
			}

			// single / double pushes
			var forward Position
			var startRank Position
			if color == PieceColorWhite {
				forward = sq + 8
				startRank = 1
			} else {
				forward = sq - 8
				startRank = 6
			}

			if (sq >> 3) == promotionRank {
				var promotionMoves []Move
				for _, promoType := range []uint8{ PieceTypeQueen, PieceTypeRook, PieceTypeBishop, PieceTypeKnight } {
					// forward promotion
					if (forward < 64) && ((b.AllPieces >> forward) & 1) == 0 {
						promotionMoves = append(promotionMoves, Move{ From: sq, To: forward, Promotion: Piece(color | promoType) })
					}
				}
				moves = append(moves, promotionMoves...)
			} else if (forward < 64) && ((b.AllPieces >> forward) & 1) == 0 {
				moves = append(moves, Move{ From: sq, To: forward })
				if (sq >> 3) == startRank {
					doubleForward := forward + (forward - sq)
					if (doubleForward < 64) && ((b.AllPieces >> doubleForward) & 1) == 0 {
						moves = append(moves, Move{ From: sq, To: doubleForward })
					}
				}
			}

			// en passant
			if b.EnPassantTarget != 64 {
				if color == PieceColorWhite {
					if (sq + 7 == b.EnPassantTarget && sq & 7 > 0) || (sq + 9 == b.EnPassantTarget && (sq & 7 < 7)) {
						moves = append(moves, Move{ From: sq, To: b.EnPassantTarget })
					}
				} else {
					if (sq - 7 == b.EnPassantTarget && sq & 7 < 7) || (sq - 9 == b.EnPassantTarget && sq & 7 > 0) {
						moves = append(moves, Move{ From: sq, To: b.EnPassantTarget })
					}
				}
			}
		case PieceTypeBishop:
			attacks := b.GetSlidingAttacks(slidingBishop, int(sq)) & ^ourBitboard
			for attacks != 0 {
				to := Position(attacks.PopLSB())
				moves = append(moves, Move{ From: sq, To: to })
			}
		case PieceTypeRook:
			attacks := b.GetSlidingAttacks(slidingRook, int(sq)) & ^ourBitboard
			for attacks != 0 {
				to := Position(attacks.PopLSB())
				moves = append(moves, Move{ From: sq, To: to })
			}
		case PieceTypeQueen:
			rookAttacks := b.GetSlidingAttacks(slidingRook, int(sq))
			bishopAttacks := b.GetSlidingAttacks(slidingBishop, int(sq))
			attacks := (rookAttacks | bishopAttacks) & ^ourBitboard
			for attacks != 0 {
				to := Position(attacks.PopLSB())
				moves = append(moves, Move{ From: sq, To: to })
			}
		}
	}

	return moves
}

func computeAttacks(sq Position, directions []Direction, interiorMask Bitboard, magic Bitboard, shift int) []Bitboard {
	bitsN := 64 - shift
	lookupSize := 1 << bitsN
	table := make([]Bitboard, lookupSize)
	blockers := computeBlockers(interiorMask)

	seen := make(map[Bitboard]bool)
	for _, b := range blockers {
		if seen[b] {
			log.Fatalf("Duplicate blocker pattern detected for square %d", sq)
		}
		seen[b] = true
	}

	for _, blocker := range blockers {
		index := (blocker * magic) >> shift
		moves := computeLegalMoves(sq, blocker, directions)

		if table[index] != 0 && table[index] != moves {
			log.Fatalf("Collision detected for square %d at index %d", sq, index)
		}

		table[index] = moves
	}
	return table
}

func computeLegalMoves(sq Position, blockers Bitboard, directions []Direction) Bitboard {
	var bitboard Bitboard = 0

	for _, d := range directions {
		dr, df := d.dr, d.df
		for step := 1; step < 8; step++ {
			rank := int(sq >> 3) + step * dr
			file := int(sq & 7) + step * df
			coord := int(sq) + step * (dr * 8 + df)
			if rank < 0 || rank > 7 || file < 0 || file > 7 {
				break
			}
			bitboard |= 1 << coord
			if ((blockers >> coord) & 1) == 1 {
				break
			}
		}
	}
	return bitboard
}

func computeBlockers(mask Bitboard) []Bitboard {
	squares := []int{}
	for i := range 64 {
		if (mask>>i)&1 != 0 {
			squares = append(squares, i)
		}
	}

	numSquares := len(squares)
	numPatterns := 1 << numSquares
	blockers := make([]Bitboard, numPatterns)

	for pattern := range numPatterns {
		var bb Bitboard = 0
		for i, sq := range squares {
			if (pattern>>i)&1 != 0 {
				bb |= 1 << sq
			}
		}
		blockers[pattern] = bb
	}

	return blockers
}


func computeInteriorMask(sq Position, directions []Direction) Bitboard {
	var mask Bitboard = 0
	for _, d := range directions {
		delta := d.dr * 8 + d.df
		for step := 1; step < 8; step++ {
			nextRank := int(sq >> 3) + (step + 1) * d.dr
			nextFile := int(sq & 7) + (step + 1) * d.df
			coord := int(sq) + step * delta
			if nextRank < 0 || nextRank > 7 || nextFile < 0 || nextFile > 7 {
				break
			}
			mask |= 1 << coord
		}
	}
	return mask
}

func computeInteriorMasks(directions []Direction) [64]Bitboard {
	masks := [64]Bitboard{}
	for sq := range Position(64) {
		masks[sq] = computeInteriorMask(sq, directions)
	}
	return masks
}

func computeKnightAttacks() [64]Bitboard {
	bitboards := [64]Bitboard{}

	for sq := range Position(64) {
		var attacks Bitboard = 0
		rank := sq >> 3
		file := sq & 7
		directions := []Direction{ {2, 1}, {2, -1}, {-2, 1}, {-2, -1}, {1, 2}, {1, -2}, {-1, 2}, {-1, -2} }

		for _, d := range directions {
			dr, df := d.dr, d.df
			r, f := int(rank) + dr, int(file) + df
			if r >= 0 && r <= 7 && f >= 0 && f <= 7 {
				s := Position(r * 8 + f)
				attacks |= 1 << s
			}
		}

		bitboards[sq] = attacks
	}
	return bitboards
}

func computeKingAttacks() [64]Bitboard {
	bitboards := [64]Bitboard{}

	for sq := range Position(64) {
		var attacks Bitboard = 0
		rank := sq >> 3
		file := sq & 7
		directions := []Direction{ {1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1} }
		for _, d := range directions {
			dr, df := d.dr, d.df
			r, f := int(rank) + dr, int(file) + df
			if r >= 0 && r <= 7 && f >= 0 && f <= 7 {
				s := Position(r * 8 + f)
				attacks |= 1 << s
			}
		}
		bitboards[sq] = attacks
	}

	return bitboards
}

func computePawnAttacks() [2][64]Bitboard {
	bitboards := [2][64]Bitboard{}

	for sq := range Position(64) {
		var whiteAttacks, blackAttacks Bitboard = 0, 0
		rank := sq >> 3
		file := sq & 7
		whiteDirections := []Direction{ {1, 1}, {1, -1} }
		blackDirections := []Direction{ {-1, 1}, {-1, -1} }
		for _, d := range whiteDirections {
			dr, df := d.dr, d.df
			r, f := int(rank) + dr, int(file) + df
			if r >= 0 && r <= 7 && f >= 0 && f <= 7 {
				s := Position(r * 8 + f)
				whiteAttacks |= 1 << s
			}
		}

		for _, d := range blackDirections {
			dr, df := d.dr, d.df
			r, f := int(rank) + dr, int(file) + df
			if r >= 0 && r <= 7 && f >= 0 && f <= 7 {
				s := Position(r * 8 + f)
				blackAttacks |= 1 << s
			}
		}
		bitboards[0][sq] = whiteAttacks
		bitboards[1][sq] = blackAttacks
	}

	return bitboards
}

func computeFixedAttacks(sp *SlidingPiece, occ Bitboard) [64]Bitboard {
	var attacks [64]Bitboard
	for sq := range Position(64) {
		attacks[sq] = getSlidingAttacksWithOcc(sp, int(sq), occ)
	}
	return attacks
}

func (b *Board) GetSlidingAttacks(sp *SlidingPiece, sq int) Bitboard {
	mask := sp.interiorMasks[sq]
	occ := b.AllPieces & mask
	index := (occ * sp.magic[sq]) >> sp.shifts[sq]
	result := sp.attacks[sq][index]
	return result
}

// TODO: remove these shitty functions with something better
func (b *Board) GetAttackingBitboard(pos Position, byWhite bool) Bitboard {
	var attacking Bitboard = 0
	var enemyPieces Bitboard
	if byWhite {
		enemyPieces = b.WhitePieces
	} else {
		enemyPieces = b.BlackPieces
	}

	var pawnIndex int
	if byWhite {
		pawnIndex = 0
	} else {
		pawnIndex = 1
	}

	for enemyPieces != 0 {
		sq := enemyPieces.PopLSB()
		piece := b.Pieces[sq]
		type_ := piece & PieceTypeMask
		switch type_ {
		case PieceTypePawn:
			pawnAttacking := pawnAttacks[pawnIndex][sq]
			attacking |= pawnAttacking
		case PieceTypeKnight:
			knightAttacking := knightAttacks[sq]
			attacking |= knightAttacking
		case PieceTypeKing:
			kingAttacking := kingAttacks[sq]
			attacking |= kingAttacking
		case PieceTypeBishop:
			bishopAttacking := b.GetSlidingAttacks(slidingBishop, sq)
			attacking |= bishopAttacking
		case PieceTypeQueen:
			bishopAttacking := b.GetSlidingAttacks(slidingBishop, sq)
			rookAttacking := b.GetSlidingAttacks(slidingRook, sq)
			attacking |= bishopAttacking | rookAttacking
		case PieceTypeRook:
			rookAttacking := b.GetSlidingAttacks(slidingRook, sq)
			attacking |= rookAttacking
		}
	}

	return attacking
}

func (b *Board) IsSquareAttacked(pos Position, byWhite bool) bool {
	attacking := b.GetAttackingBitboard(pos, byWhite)
	return (attacking & (1 << pos)) != 0
}

func (b *Board) KingSquare(white bool) Position {
	var board Bitboard
	if white {
		board = b.PieceBitboards[0][PieceTypeKing - 1]
	} else {
		board = b.PieceBitboards[1][PieceTypeKing - 1]
	}
	return Position(board.LSB())
}
