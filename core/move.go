package core

import "slices"

import "math/bits"

var (
	magicRook = NewSlidingPiece([]Direction{
		{dr: 1, df: 0}, // right
		{dr: -1, df: 0}, // left
		{dr: 0, df: 1}, // up
		{dr: 0, df: -1}, // down
	})

	magicBishop = NewSlidingPiece([]Direction{
		{dr: 1, df: 1}, // up-right
		{dr: -1, df: -1}, // down-left
		{dr: 1, df: -1}, // up-left
		{dr: -1, df: 1}, // down-right
	})

	magicQueen = NewSlidingPiece([]Direction{
		{dr: 1, df: 0}, // right
		{dr: -1, df: 0}, // left
		{dr: 0, df: 1}, // up
		{dr: 0, df: -1}, // down
		{dr: 1, df: 1}, // up-right
		{dr: -1, df: -1}, // down-left
		{dr: 1, df: -1}, // up-left
		{dr: -1, df: 1}, // down-right
	})
)

var kingAttacks [64]Bitboard = computeKingAttacks()
var knightAttacks [64]Bitboard = computeKnightAttacks()
var pawnAttacks [2][64]Bitboard = computePawnAttacks()

type Direction struct {
	dr, df int
}

type SlidingPiece struct {
	directions []Direction
	masks [64]Bitboard
	attacks [64][]Bitboard
}

type Move struct {
	From Position
	To Position
	Promotion Piece
}

func (b *Board) IsMoveLegal(move Move) bool {
	pseudoLegalMoves := b.generatePseudoLegalMoves()
	return slices.Contains(pseudoLegalMoves, move)
}

func (b *Board) generatePseudoLegalMoves() []Move {
	moves := []Move{}

	myPieces := b.WhitePieces
	theirPieces := b.BlackPieces
	myColor := PieceColorWhite >> 3
	if !b.WhiteToMove {
		myPieces = b.BlackPieces
		theirPieces = b.WhitePieces
		myColor = PieceColorBlack >> 3
	}

	occupied := b.AllPieces

	pawns := b.PieceBitboards[myColor][PieceTypePawn - 1]
	for _, sq := range pawns.Squares() {
		if myColor == PieceColorWhite {
			oneStep := sq + 8
			if oneStep < 64 && (occupied >> oneStep) & 1 == 0 {
				if rankOf(oneStep) == 7 {
					for _, promo := range []Piece{PieceWhiteQueen, PieceWhiteRook, PieceWhiteBishop, PieceWhiteKnight} {
						moves = append(moves, Move{From: Position(sq), To: Position(oneStep), Promotion: promo})
					}
				} else {
					moves = append(moves, Move{From: Position(sq), To: Position(oneStep)})
					if rankOf(sq) == 1 {
						twoStep := sq + 16
						if (occupied>>twoStep)&1 == 0 {
							moves = append(moves, Move{From: Position(sq), To: Position(twoStep)})
						}
					}
				}
			}
			// captures
			for _, target := range pawnAttacks[0][sq].Squares() {
				if (theirPieces >> target) & 1 != 0 || Position(target-8) == b.EnPassantTarget {
					if rankOf(target) == 7 {
						for _, promo := range []Piece{PieceWhiteQueen, PieceWhiteRook, PieceWhiteBishop, PieceWhiteKnight} {
							moves = append(moves, Move{From: Position(sq), To: Position(target), Promotion: promo})
						}
					} else {
						moves = append(moves, Move{From: Position(sq), To: Position(target)})
					}
				}
			}
		} else {
			// black pawn
			oneStep := sq - 8
			if sq >= 8 && (occupied>>oneStep)&1 == 0 {
				if rankOf(oneStep) == 0 {
					for _, promo := range []Piece{PieceBlackQueen, PieceBlackRook, PieceBlackBishop, PieceBlackKnight} {
						moves = append(moves, Move{From: Position(sq), To: Position(oneStep), Promotion: promo})
					}
				} else {
					moves = append(moves, Move{From: Position(sq), To: Position(oneStep)})
					// double push
					if rankOf(sq) == 6 {
						twoStep := sq - 16
						if (occupied>>twoStep)&1 == 0 {
							moves = append(moves, Move{From: Position(sq), To: Position(twoStep)})
						}
					}
				}
			}
			// captures
			for _, target := range pawnAttacks[1][sq].Squares() {
				if (theirPieces>>target)&1 != 0 || Position(target+8) == b.EnPassantTarget {
					if rankOf(target) == 0 {
						for _, promo := range []Piece{PieceBlackQueen, PieceBlackRook, PieceBlackBishop, PieceBlackKnight} {
							moves = append(moves, Move{From: Position(sq), To: Position(target), Promotion: promo})
						}
					} else {
						moves = append(moves, Move{From: Position(sq), To: Position(target)})
					}
				}
			}
		}
	}

	knights := b.PieceBitboards[myColor][PieceTypeKnight - 1]
	for _, sq := range knights.Squares() {
		for _, target := range knightAttacks[sq].Squares() {
			if (myPieces>>target)&1 == 0 {
				moves = append(moves, Move{From: Position(sq), To: Position(target)})
			}
		}
	}

	bishops := b.PieceBitboards[myColor][PieceTypeBishop - 1]
	for _, sq := range bishops.Squares() {
		mask := magicBishop.masks[sq]
		index := compress(occupied, mask)
		for _, target := range magicBishop.attacks[sq][index].Squares() {
			if (myPieces>>target)&1 == 0 {
				moves = append(moves, Move{From: Position(sq), To: Position(target)})
			}
		}
	}

	rooks := b.PieceBitboards[myColor][PieceTypeRook - 1]
	for _, sq := range rooks.Squares() {
		mask := magicRook.masks[sq]
		index := compress(occupied, mask)
		for _, target := range magicRook.attacks[sq][index].Squares() {
			if (myPieces>>target)&1 == 0 {
				moves = append(moves, Move{From: Position(sq), To: Position(target)})
			}
		}
	}

	queens := b.PieceBitboards[myColor][PieceTypeQueen - 1]
	for _, sq := range queens.Squares() {
		mask := magicQueen.masks[sq]
		index := compress(occupied, mask)
		for _, target := range magicQueen.attacks[sq][index].Squares() {
			if (myPieces>>target)&1 == 0 {
				moves = append(moves, Move{From: Position(sq), To: Position(target)})
			}
		}
	}

	king := b.PieceBitboards[myColor][PieceTypeKing - 1]
	for _, sq := range king.Squares() {
		for _, target := range kingAttacks[sq].Squares() {
			if (myPieces>>target)&1 == 0 {
				moves = append(moves, Move{From: Position(sq), To: Position(target)})
			}
		}
	}

	return moves
}

func rankOf(sq uint8) uint8 { return sq >> 3 }
func fileOf(sq uint8) uint8 { return sq & 7 }

func compress(occ, mask Bitboard) int {
	index := 0
	bitIndex := 0
	for mask != 0 {
		lsb := mask & -mask;
		if occ & lsb != 0 {
			index |= (1 << bitIndex)
		}
		mask &= mask - 1
		bitIndex++
	}
	return index
}

// reconstructs the occupancy from a given index
func decompress(index int, mask Bitboard) Bitboard {
	var occ Bitboard
	bitPosition := 0
	for mask != 0 {
		lsb := mask & -mask
		if (index >> bitPosition) & 1 != 0 {
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
			mask |= 1 << (r * 8 + f)
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
			if block & (1 << sq) != 0 {
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
