package core

import (
	"fmt"
	"math/bits"
	"slices"
)

var (
	magicRook = NewSlidingPiece([]Direction{
		{dr: 1, df: 0}, {dr: -1, df: 0}, {dr: 0, df: 1}, {dr: 0, df: -1},
	})

	magicBishop = NewSlidingPiece([]Direction{
		{dr: 1, df: 1}, {dr: -1, df: -1}, {dr: 1, df: -1}, {dr: -1, df: 1},
	})

	magicQueen = NewSlidingPiece([]Direction{
		{dr: 1, df: 0}, {dr: -1, df: 0}, {dr: 0, df: 1}, {dr: 0, df: -1},
		{dr: 1, df: 1}, {dr: -1, df: -1}, {dr: 1, df: -1}, {dr: -1, df: 1},
	})

	// Rank and file masks
	RANK_1 Bitboard = 0x00000000000000FF
	RANK_2 Bitboard = 0x000000000000FF00
	RANK_3 Bitboard = 0x0000000000FF0000
	RANK_7 Bitboard = 0x00FF000000000000
	RANK_8 Bitboard = 0xFF00000000000000
	FILE_A Bitboard = 0x0101010101010101
	FILE_B Bitboard = 0x0202020202020202
	FILE_G Bitboard = 0x4040404040404040
	FILE_H Bitboard = 0x8080808080808080
)

var kingAttacks [64]Bitboard = computeKingAttacks()
var knightAttacks [64]Bitboard = computeKnightAttacks()
var pawnAttacks [2][64]Bitboard = computePawnAttacks()

type Direction struct {
	dr, df int
}

type SlidingPiece struct {
	directions []Direction
	masks      [64]Bitboard
	attacks    [64][]Bitboard
}

type Move struct {
	From      Position
	To        Position
	Promotion Piece
}

type CastleRights struct {
	WhiteKingside  bool
	WhiteQueenside bool
	BlackKingside  bool
	BlackQueenside bool
}

// TODO: optimize later
func (b *Board) IsMoveLegal(move Move) bool {
	ourMoves := b.generatePseudoLegalMoves()
	if !slices.Contains(ourMoves, move) {
		return false
	}

	b.Push(&move)

	kingPos := Position(0)
	if b.WhiteToMove {
		kingBB := b.PieceBitboards[0][PieceTypeKing-1]
		if kingBB != 0 {
			kingPos = Position(bits.TrailingZeros64(uint64(kingBB)))
		}
	} else {
		kingBB := b.PieceBitboards[1][PieceTypeKing-1]
		if kingBB != 0 {
			kingPos = Position(bits.TrailingZeros64(uint64(kingBB)))
		}
	}

	enemyMoves := b.generatePseudoLegalMoves()
	isInCheck := false

	for _, em := range enemyMoves {
		if em.To == kingPos {
			isInCheck = true
			break
		}
	}

	b.Pop()

	return !isInCheck
}

func (b *Board) GenerateLegalMoves() []Move {
	pseudoLegalMoves := b.generatePseudoLegalMoves()
	legalMoves := []Move{}
	for _, m := range pseudoLegalMoves {
		b.Push(&m)
		fmt.Println("0")

		kingPos := Position(0)
		if b.WhiteToMove {
			kingBB := b.PieceBitboards[0][PieceTypeKing-1]
			if kingBB != 0 {
				kingPos = Position(bits.TrailingZeros64(uint64(kingBB)))
			}
		} else {
			kingBB := b.PieceBitboards[1][PieceTypeKing-1]
			if kingBB != 0 {
				kingPos = Position(bits.TrailingZeros64(uint64(kingBB)))
			}
		}

		enemyMoves := b.generatePseudoLegalMoves()
		isInCheck := false
		for _, em := range enemyMoves {
			if em.To == kingPos {
				isInCheck = true
				break
			}
		}

		b.Pop()
		fmt.Println("1")

		if !isInCheck {
			legalMoves = append(legalMoves, m)
		}
	}
	return legalMoves
}

func (b *Board) generatePseudoLegalMoves() []Move {
	moves := []Move{}

	myPieces := b.WhitePieces
	theirPieces := b.BlackPieces
	myColor := uint8(PieceColorWhite >> 3)
	isWhite := b.WhiteToMove

	if !isWhite {
		myPieces = b.BlackPieces
		theirPieces = b.WhitePieces
		myColor = PieceColorBlack >> 3
	}

	occupied := b.AllPieces
	empty := ^occupied

	b.generatePawnMoves(&moves, myColor, isWhite, theirPieces, empty)
	b.generateKnightMoves(&moves, myColor, myPieces)
	b.generateBishopMoves(&moves, myColor, myPieces, occupied)
	b.generateRookMoves(&moves, myColor, myPieces, occupied)
	b.generateQueenMoves(&moves, myColor, myPieces, occupied)
	b.generateKingMoves(&moves, myColor, myPieces, isWhite)

	return moves
}

func (b *Board) generatePawnMoves(moves *[]Move, myColor uint8, isWhite bool, theirPieces, empty Bitboard) {
	pawns := b.PieceBitboards[myColor][PieceTypePawn-1]

	if isWhite {
		// White pawn moves
		// Single pushes
		singlePush := (pawns << 8) & empty
		nonPromoPush := singlePush &^ RANK_8
		promoPush := singlePush & RANK_8

		// Double pushes (only from rank 2)
		doublePush := ((singlePush & RANK_3) << 8) & empty

		// Captures
		capturesLeft := ((pawns << 7) &^ FILE_H) & theirPieces
		capturesRight := ((pawns << 9) &^ FILE_A) & theirPieces

		// Promotion captures
		promoCapLeft := ((pawns << 7) &^ FILE_H) & theirPieces & RANK_8
		promoCapRight := ((pawns << 9) &^ FILE_A) & theirPieces & RANK_8

		// Non-promotion captures
		capturesLeft &^= RANK_8
		capturesRight &^= RANK_8

		// En passant captures
		if b.EnPassantTarget != 0 {
			epTarget := Bitboard(1) << uint(b.EnPassantTarget)
			epCapturesLeft := ((pawns << 7) &^ FILE_H) & epTarget
			epCapturesRight := ((pawns << 9) &^ FILE_A) & epTarget
			addMovesFromBitboard(moves, epCapturesLeft, -7)
			addMovesFromBitboard(moves, epCapturesRight, -9)
		}

		// Add regular moves
		addMovesFromBitboard(moves, nonPromoPush, -8)
		addMovesFromBitboard(moves, doublePush, -16)
		addMovesFromBitboard(moves, capturesLeft, -7)
		addMovesFromBitboard(moves, capturesRight, -9)

		// Add promotion moves
		addPromotionMoves(moves, promoPush, -8, true)
		addPromotionMoves(moves, promoCapLeft, -7, true)
		addPromotionMoves(moves, promoCapRight, -9, true)

	} else {
		// Black pawn moves
		// Single pushes
		singlePush := (pawns >> 8) & empty
		nonPromoPush := singlePush &^ RANK_1
		promoPush := singlePush & RANK_1

		// Double pushes (only from rank 7)
		doublePush := ((singlePush & (RANK_7 >> 8)) >> 8) & empty

		// Captures
		capturesLeft := ((pawns >> 9) &^ FILE_H) & theirPieces
		capturesRight := ((pawns >> 7) &^ FILE_A) & theirPieces

		// Promotion captures
		promoCapLeft := ((pawns >> 9) &^ FILE_H) & theirPieces & RANK_1
		promoCapRight := ((pawns >> 7) &^ FILE_A) & theirPieces & RANK_1

		// Non-promotion captures
		capturesLeft &^= RANK_1
		capturesRight &^= RANK_1

		// En passant captures
		if b.EnPassantTarget != 0 {
			epTarget := Bitboard(1) << uint(b.EnPassantTarget)
			epCapturesLeft := ((pawns >> 9) &^ FILE_H) & epTarget
			epCapturesRight := ((pawns >> 7) &^ FILE_A) & epTarget
			addMovesFromBitboard(moves, epCapturesLeft, 9)
			addMovesFromBitboard(moves, epCapturesRight, 7)
		}

		// Add regular moves
		addMovesFromBitboard(moves, nonPromoPush, 8)
		addMovesFromBitboard(moves, doublePush, 16)
		addMovesFromBitboard(moves, capturesLeft, 9)
		addMovesFromBitboard(moves, capturesRight, 7)

		// Add promotion moves
		addPromotionMoves(moves, promoPush, 8, false)
		addPromotionMoves(moves, promoCapLeft, 9, false)
		addPromotionMoves(moves, promoCapRight, 7, false)
	}
}

func (b *Board) generateKnightMoves(moves *[]Move, myColor uint8, myPieces Bitboard) {
	knights := b.PieceBitboards[myColor][PieceTypeKnight-1]

	for knights != 0 {
		from := bits.TrailingZeros64(uint64(knights))
		knights &= knights - 1 // Clear LSB

		targets := knightAttacks[from] &^ myPieces
		addMovesFromSquare(moves, uint8(from), targets)
	}
}

func (b *Board) generateBishopMoves(moves *[]Move, myColor uint8, myPieces, occupied Bitboard) {
	bishops := b.PieceBitboards[myColor][PieceTypeBishop-1]

	for bishops != 0 {
		from := bits.TrailingZeros64(uint64(bishops))
		bishops &= bishops - 1

		mask := magicBishop.masks[from]
		index := compress(occupied, mask)
		targets := magicBishop.attacks[from][index] &^ myPieces

		addMovesFromSquare(moves, uint8(from), targets)
	}
}

func (b *Board) generateRookMoves(moves *[]Move, myColor uint8, myPieces, occupied Bitboard) {
	rooks := b.PieceBitboards[myColor][PieceTypeRook-1]

	for rooks != 0 {
		from := bits.TrailingZeros64(uint64(rooks))
		rooks &= rooks - 1

		mask := magicRook.masks[from]
		index := compress(occupied, mask)
		targets := magicRook.attacks[from][index] &^ myPieces

		addMovesFromSquare(moves, uint8(from), targets)
	}
}

func (b *Board) generateQueenMoves(moves *[]Move, myColor uint8, myPieces, occupied Bitboard) {
	queens := b.PieceBitboards[myColor][PieceTypeQueen-1]

	for queens != 0 {
		from := bits.TrailingZeros64(uint64(queens))
		queens &= queens - 1

		mask := magicQueen.masks[from]
		index := compress(occupied, mask)
		targets := magicQueen.attacks[from][index] &^ myPieces

		addMovesFromSquare(moves, uint8(from), targets)
	}
}

func (b *Board) generateKingMoves(moves *[]Move, myColor uint8, myPieces Bitboard, isWhite bool) {
	king := b.PieceBitboards[myColor][PieceTypeKing-1]

	if king != 0 {
		from := bits.TrailingZeros64(uint64(king))
		targets := kingAttacks[from] &^ myPieces
		addMovesFromSquare(moves, uint8(from), targets)
	}

	b.generateCastlingMoves(moves, isWhite)
}

func (b *Board) generateCastlingMoves(moves *[]Move, isWhite bool) {
	if isWhite {
		// White kingside castling
		if (b.CastlingRights & CastlingWhiteKingside) != 0 {
			// Check if squares between king and rook are empty (f1, g1)
			if (b.AllPieces & (Bitboard(1)<<5 | Bitboard(1)<<6)) == 0 {
				*moves = append(*moves, Move{From: Position(4), To: Position(6)})
			}
		}

		// White queenside castling
		if (b.CastlingRights & CastlingWhiteQueenside) != 0 {
			// Check if squares between king and rook are empty (b1, c1, d1)
			if (b.AllPieces & (Bitboard(1)<<1 | Bitboard(1)<<2 | Bitboard(1)<<3)) == 0 {
				*moves = append(*moves, Move{From: Position(4), To: Position(2)})
			}
		}
	} else {
		// Black kingside castling
		if (b.CastlingRights & CastlingBlackKingside) != 0 {
			// Check if squares between king and rook are empty (f8, g8)
			if (b.AllPieces & (Bitboard(1)<<61 | Bitboard(1)<<62)) == 0 {
				*moves = append(*moves, Move{From: Position(60), To: Position(62)})
			}
		}

		// Black queenside castling
		if (b.CastlingRights & CastlingBlackQueenside) != 0 {
			// Check if squares between king and rook are empty (b8, c8, d8)
			if (b.AllPieces & (Bitboard(1)<<57 | Bitboard(1)<<58 | Bitboard(1)<<59)) == 0 {
				*moves = append(*moves, Move{From: Position(60), To: Position(58)})
			}
		}
	}
}

func addMovesFromBitboard(moves *[]Move, targets Bitboard, offset int) {
	for targets != 0 {
		to := bits.TrailingZeros64(uint64(targets))
		targets &= targets - 1
		from := to + offset
		*moves = append(*moves, Move{From: Position(from), To: Position(to)})
	}
}

func addMovesFromSquare(moves *[]Move, from uint8, targets Bitboard) {
	for targets != 0 {
		to := bits.TrailingZeros64(uint64(targets))
		targets &= targets - 1
		*moves = append(*moves, Move{From: Position(from), To: Position(to)})
	}
}

func addPromotionMoves(moves *[]Move, targets Bitboard, offset int, isWhite bool) {
	promotions := []Piece{PieceWhiteQueen, PieceWhiteRook, PieceWhiteBishop, PieceWhiteKnight}
	if !isWhite {
		promotions = []Piece{PieceBlackQueen, PieceBlackRook, PieceBlackBishop, PieceBlackKnight}
	}

	for targets != 0 {
		to := bits.TrailingZeros64(uint64(targets))
		targets &= targets - 1
		from := to + offset

		for _, promo := range promotions {
			*moves = append(*moves, Move{
				From: Position(from),
				To: Position(to),
				Promotion: promo,
			})
		}
	}
}

// Keep existing helper functions
func rankOf(sq uint8) uint8 { return sq >> 3 }
func fileOf(sq uint8) uint8 { return sq & 7 }

func compress(occ, mask Bitboard) int {
	index := 0
	bitIndex := 0
	for mask != 0 {
		lsb := mask & -mask
		if occ&lsb != 0 {
			index |= (1 << bitIndex)
		}
		mask &= mask - 1
		bitIndex++
	}
	return index
}

func decompress(index int, mask Bitboard) Bitboard {
	var occ Bitboard
	bitPosition := 0
	for mask != 0 {
		lsb := mask & -mask
		if (index>>bitPosition)&1 != 0 {
			occ |= lsb
		}
		mask &= mask - 1
		bitPosition++
	}
	return occ
}

func computeMask(sq int, directions []Direction) Bitboard {
	var mask Bitboard
	rank := int(rankOf(uint8(sq)))
	file := int(fileOf(uint8(sq)))
	for _, d := range directions {
		r, f := rank, file
		for {
			r += d.dr
			f += d.df

			if r <= 0 || r >= 7 || f <= 0 || f >= 7 {
				break
			}

			mask |= 1 << (r*8 + f)
		}
	}
	return mask
}

func computeSlidingSet(sq int, block Bitboard, directions []Direction) Bitboard {
	var attacks Bitboard
	rank := int(rankOf(uint8(sq)))
	file := int(fileOf(uint8(sq)))
	for _, d := range directions {
		r, f := rank, file
		for {
			r += d.dr
			f += d.df
			if r < 0 || r > 7 || f < 0 || f > 7 {
				break
			}
			sq := r*8 + f
			attacks |= 1 << sq
			if block&(1<<sq) != 0 {
				break
			}
		}
	}
	return attacks
}

func NewSlidingPiece(directions []Direction) *SlidingPiece {
	p := &SlidingPiece{
		directions: directions,
	}
	for sq := range 64 {
		mask := computeMask(sq, directions)
		p.masks[sq] = mask
		relevantBits := bits.OnesCount64(uint64(mask))
		entries := 1 << relevantBits
		p.attacks[sq] = make([]Bitboard, entries)
		for index := range entries {
			occ := decompress(int(index), mask)
			attack := computeSlidingSet(sq, occ, directions)
			p.attacks[sq][index] = attack
		}
	}
	return p
}

func computeKnightAttacks() [64]Bitboard {
	knightAttacks := [64]Bitboard{}
	offsets := [8][2]int{
		{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
		{1, 2}, {1, -2}, {-1, 2}, {-1, -2},
	}
	for sq := range 64 {
		rank := rankOf(uint8(sq))
		file := fileOf(uint8(sq))
		for _, o := range offsets {
			r := int(rank) + o[0]
			f := int(file) + o[1]
			if r >= 0 && r < 8 && f >= 0 && f < 8 {
				knightAttacks[sq] |= 1 << (r*8 + f)
			}
		}
	}
	return knightAttacks
}

func computeKingAttacks() [64]Bitboard {
	kingAttacks := [64]Bitboard{}
	for sq := range 64 {
		rank := rankOf(uint8(sq))
		file := fileOf(uint8(sq))
		for dr := -1; dr <= 1; dr++ {
			for df := -1; df <= 1; df++ {
				if dr == 0 && df == 0 {
					continue
				}
				r := int(rank) + dr
				f := int(file) + df
				if r >= 0 && r < 8 && f >= 0 && f < 8 {
					kingAttacks[sq] |= 1 << (r*8 + f)
				}
			}
		}
	}
	return kingAttacks
}

func computePawnAttacks() [2][64]Bitboard {
	pawnAttacks := [2][64]Bitboard{}
	for sq := range 64 {
		r := int(rankOf(uint8(sq)))
		f := int(fileOf(uint8(sq)))

		// White pawns attack diagonally up
		if r+1 < 8 && f-1 >= 0 {
			pawnAttacks[0][sq] |= 1 << ((r+1)*8 + (f-1))
		}
		if r+1 < 8 && f+1 < 8 {
			pawnAttacks[0][sq] |= 1 << ((r+1)*8 + (f+1))
		}

		// Black pawns attack diagonally down
		if r-1 >= 0 && f-1 >= 0 {
			pawnAttacks[1][sq] |= 1 << ((r-1)*8 + (f-1))
		}
		if r-1 >= 0 && f+1 < 8 {
			pawnAttacks[1][sq] |= 1 << ((r-1)*8 + (f+1))
		}
	}
	return pawnAttacks
}

func (b *Board) IsMoveKingsideCastle(move *Move) bool {
	if b.WhiteToMove {
		if b.PieceBitboards[0][PieceTypeKing-1] == 0 {
			return false
		}
		return move.From == 4 && move.To == 6
	} else {
		if b.PieceBitboards[1][PieceTypeKing-1] == 0 {
			return false
		}
		return move.From == 60 && move.To == 62
	}
}

func IsMoveQueensideCastle(move *Move, isWhite bool) bool {
	if isWhite {
		return move.From == 4 && move.To == 2
	} else {
		return move.From == 60 && move.To == 58
	}
}
