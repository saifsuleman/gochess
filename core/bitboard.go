package core

import "math/bits"

type Bitboard uint64

func (bb Bitboard) Squares() []uint8 {
	squares := []uint8{}
	b := bb
	for b != 0 {
		lsb := b & -b
		sq := uint8(bits.TrailingZeros64(uint64(lsb)))
		squares = append(squares, sq)
		b &= b - 1
	}
	return squares
}

func (bb *Bitboard) PopLSB() int {
	sq := bits.TrailingZeros64(uint64(*bb))
	*bb &= *bb - 1
	return sq
}
