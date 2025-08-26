package core

import (
	"fmt"
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
	white := b.WhiteToMove
	pseudoLegalMoves := b.GeneratePseudoLegalMoves()
	legalMoves := []Move{}

	for _, move := range pseudoLegalMoves {
		b.Push(&move)

		var king Position
		if white {
			king = Position(b.PieceBitboards[0][PieceTypeKing - 1].LSB())
		} else {
			king = Position(b.PieceBitboards[1][PieceTypeKing - 1].LSB())
		}

		enemyMoves := b.GeneratePseudoLegalMoves()
		inCheck := false

		for _, emove := range enemyMoves {
			if emove.To == king {
				inCheck = true
				break
			}
		}

		b.Pop()

		if !inCheck {
			legalMoves = append(legalMoves, move)
		}
	}

	return legalMoves
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

			// castling (TODO: check detection for when someone is checking in the path of our castling)
			if color == PieceColorWhite && b.WhiteCanCastleKingside() {
				if ((b.AllPieces >> 5) & 1) == 0 && ((b.AllPieces >> 6) & 1) == 0 {
					moves = append(moves, Move{ From: sq, To: 6 })
				}
			} else {
				if (((b.AllPieces >> 61) & 1) == 0) && (((b.AllPieces >> 62) & 1) == 0) {
					moves = append(moves, Move{ From: sq, To: 62 })
				}
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

func (b *Board) GetSlidingAttacks(sp *SlidingPiece, sq int) Bitboard {
	mask := sp.interiorMasks[sq]
	occ := b.AllPieces & mask
	index := (occ * sp.magic[sq]) >> sp.shifts[sq]
	result := sp.attacks[sq][index]
	return result
}

func MoveTest() {
	mask := computeInteriorMask(17, BishopDirections)
	blockers := computeBlockers(mask)
	fmt.Printf("Blocker: %d\n", blockers[7])
}
