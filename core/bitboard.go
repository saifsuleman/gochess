package core

import "math/bits"

type Bitboard uint64

func (bb *Bitboard) PopCount() int {
	return bits.OnesCount64(uint64(*bb))
}

func (bb *Bitboard) LSB() int {
	return bits.TrailingZeros64(uint64(*bb))
}

func (bb *Bitboard) PopLSB() int {
	b := *bb
	lsb := b & -b
	*bb = b & (b - 1)
	return bits.TrailingZeros64(uint64(lsb))
}
