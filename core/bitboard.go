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

//
// func (bb *Bitboard) PopCount() int {
// 	return bits.OnesCount64(uint64(*bb))
// }
//
// func (bb *Bitboard) LSB() int {
// 	return bits.TrailingZeros64(uint64(*bb))
// }
//
// func (bb *Bitboard) PopLSB() int {
// 	b := *bb
// 	lsb := b & -b
// 	*bb = b & (b - 1)
// 	return bits.TrailingZeros64(uint64(lsb))
// }
