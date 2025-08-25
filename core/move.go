package core

import (
	"slices"
	"fmt"
)

var (
	slidingRook 		= NewSlidingPiece([]Direction{ {1, 0}, {-1, 0}, {0, 1}, {0, -1} }, rookMagics, rookShifts)
	slidingBishop 	= NewSlidingPiece([]Direction{ {1, 1}, {1, -1}, {-1, 1}, {-1, -1} }, bishopMagics, bishopShifts)

	kingAttacks   = computeKingAttacks()
	knightAttacks = computeKnightAttacks()
	pawnAttacks	 = computePawnAttacks()
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

func (b *Board) GenerateLegalMoves() []Move {
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
		case PieceTypePawn:
			var attacks Bitboard
			if color == PieceColorWhite {
				attacks = pawnAttacks[0][sq] & enemyBitboard
			} else {
				attacks = pawnAttacks[1][sq] & enemyBitboard
			}
			for attacks != 0 {
				to := Position(attacks.PopLSB())
				moves = append(moves, Move{ From: sq, To: to })
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

			if (forward < 64) && ((b.AllPieces >> forward) & 1) == 0 {
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
					if (sq + 7 == b.EnPassantTarget) || (sq + 9 == b.EnPassantTarget) {
						moves = append(moves, Move{ From: sq, To: b.EnPassantTarget })
					}
				} else {
					if (sq - 7 == b.EnPassantTarget) || (sq - 9 == b.EnPassantTarget) {
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
				fmt.Println("Queen move from", sq, "to", to)
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
	for _, blocker := range blockers {
		index := (blocker * magic) >> shift
		moves := computeLegalMoves(sq, blocker, directions)
		table[index] = moves
	}
	return table
}

func computeLegalMoves(sq Position, blockers Bitboard, directions []Direction) Bitboard {
	var bitboard Bitboard = 0

	for _, d := range directions {
		dr, df := d.dr, d.df
		for step := 1; step < 8; step++ {
			coord := int(sq) + step * (dr * 8 + df)
			if coord < 0 || coord >= 64 {
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

// I still don't entirely understand this voodoo shit
func computeBlockers(mask Bitboard) []Bitboard {
	moves := []int{}

	for i := range 64 {
		if ((mask >> i) & 1) == 1 {
			moves = append(moves, i)
		}
	}

	patterns := 1 << len(moves)
	attacks := make([]Bitboard, patterns)

	for pattern := range patterns {
		for bit := range len(moves) {
			bit := (pattern >> bit) & 1
			attacks[pattern] |= Bitboard(bit << moves[bit])
		}
	}

	return attacks
}

func computeInteriorMask(sq Position, directions []Direction) Bitboard {
	var mask Bitboard = 0
	for _, d := range directions {
		delta := d.dr * 8 + d.df
		for step := 1; step < 8; step++ {
			coord := int(sq) + step * delta // TODO: check this? maybe it should be split by rank/file to prevent file wrapping
			next := int(sq) + (step + 1) * delta
			if next < 0 || next >= 64 {
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

func (b *Board) GetSlidingAttacks(sp *SlidingPiece, sq int) Bitboard {
	mask := sp.interiorMasks[sq]
	occ := b.AllPieces & mask
	index := (occ * sp.magic[sq]) >> sp.shifts[sq]
	result := sp.attacks[sq][index]
	return result
}
