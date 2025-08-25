package core

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
	return true
}

func (b *Board) generatePseudoLegalMoves() []Move {
	moves := []Move{}

	for sq := range 64 {
		piece := b.Pieces[sq]

		if piece == PieceNone {
			continue
		}

		switch piece {
		case PieceWhitePawn:
			oneStep := sq + 8
			if oneStep < 64 && b.Pieces[oneStep] == PieceNone {
				if rankOf(oneStep) == 7 {
					for _, promo := range []Piece{PieceWhiteQueen, PieceWhiteRook, PieceWhiteBishop, PieceWhiteKnight} {
						moves = append(moves, Move{From: Position(sq), To: Position(oneStep), Promotion: promo} )
					}
				} else {
					moves = append(moves, Move{From: Position(sq), To: Position(oneStep)})
					if rankOf(sq) == 1 {
						twoStep := sq + 16
						if b.Pieces[twoStep] == PieceNone {
							moves = append(moves, Move{From: Position(sq), To: Position(twoStep)})
						}
					}
				}
			}

			for _, target := range pawnAttacks[0][sq].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if (enemyPiece != PieceNone && enemyColor == PieceColorBlack) || Position(target - 8) == b.EnPassantTarget {
					if rankOf(int(target)) == 7 {
						for _, promo := range []Piece{PieceWhiteQueen, PieceWhiteRook, PieceWhiteBishop, PieceWhiteKnight} {
							moves = append(moves, Move{From: Position(sq), To: Position(target), Promotion: promo} )
						}
					} else {
						moves = append(moves, Move{From: Position(sq), To: Position(target)})
					}
				}
			}
			continue
		case PieceBlackPawn:
			oneStep := sq - 8
			if oneStep >= 0 && b.Pieces[oneStep] == PieceNone {
				if rankOf(oneStep) == 0 {
					for _, promo := range []Piece{PieceBlackQueen, PieceBlackRook, PieceBlackBishop, PieceBlackKnight} {
						moves = append(moves, Move{From: Position(sq), To: Position(oneStep), Promotion: promo} )
					}
				} else {
					moves = append(moves, Move{From: Position(sq), To: Position(oneStep)})
					if rankOf(sq) == 6 {
						twoStep := sq - 16
						if b.Pieces[twoStep] == PieceNone {
							moves = append(moves, Move{From: Position(sq), To: Position(twoStep)})
						}
					}
				}
			}

			for _, target := range pawnAttacks[1][sq].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if (enemyPiece != PieceNone && enemyColor == PieceColorWhite) || Position(target + 8) == b.EnPassantTarget {
					if rankOf(int(target)) == 0 {
						for _, promo := range []Piece{PieceBlackQueen, PieceBlackRook, PieceBlackBishop, PieceBlackKnight} {
							moves = append(moves, Move{From: Position(sq), To: Position(target), Promotion: promo} )
						}
					} else {
						moves = append(moves, Move{From: Position(sq), To: Position(target)})
					}
				}
			}

			continue
		}

		color := piece.Color()
		type_ := piece.Type()

		switch type_ {
		case PieceTypeKnight:
			for _, target := range knightAttacks[sq].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if enemyPiece == PieceNone || enemyColor != color {
					moves = append(moves, Move{From: Position(sq), To: Position(target)})
				}
			}
		case PieceTypeBishop:
			occ := b.Occupancy()
			mask := magicBishop.masks[sq]
			index := compress(occ, mask)
			for _, target := range magicBishop.attacks[sq][index].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if enemyPiece == PieceNone || enemyColor != color {
					moves = append(moves, Move{From: Position(sq), To: Position(target)})
				}
			}
		case PieceTypeRook:
			occ := b.Occupancy()
			mask := magicRook.masks[sq]
			index := compress(occ, mask)
			for _, target := range magicRook.attacks[sq][index].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if enemyPiece == PieceNone || enemyColor != color {
					moves = append(moves, Move{From: Position(sq), To: Position(target)})
				}
			}
		case PieceTypeQueen:
			occ := b.Occupancy()
			mask := magicQueen.masks[sq]
			index := compress(occ, mask)
			for _, target := range magicQueen.attacks[sq][index].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if enemyPiece == PieceNone || enemyColor != color {
					moves = append(moves, Move{From: Position(sq), To: Position(target)})
				}
			}
		case PieceTypeKing:
			for _, target := range kingAttacks[sq].Squares() {
				enemyPiece := b.Pieces[target]
				enemyColor := enemyPiece.Color()
				if enemyPiece == PieceNone || enemyColor != color {
					moves = append(moves, Move{From: Position(sq), To: Position(target)})
				}
			}
		}
	}

	return moves
}

func (b *Board) Occupancy() Bitboard {
	return 01
}

func rankOf(sq int) int { return sq >> 3 }
func fileOf(sq int) int { return sq & 7 }

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
	rank := rankOf(sq)
	file := fileOf(sq)
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
	rank := rankOf(sq)
	file := fileOf(sq)
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
		rank := rankOf(sq)
		file := fileOf(sq)
		for _, o := range offsets {
			r := rank + o[0]
			f := file + o[1]
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
		rank := rankOf(sq)
		file := fileOf(sq)
		for dr := -1; dr <= 1; dr++ {
			for df := -1; df <= 1; df++ {
				if dr == 0 && df == 0 {
					continue
				}
				r := rank + dr
				f := file + df
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
		r := rankOf(sq)
		f := fileOf(sq)

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
