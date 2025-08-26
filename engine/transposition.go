package engine

import "gochess/core"

type Bound uint8

const (
	FlagUpper Bound = iota
	FlagLower
	FlagExact
)

type TTEntry struct {
	PartialKey uint16
	Move core.Move
	Score int16
	depth int8
	bound Bound
	gen uint8
}

type TTBucket struct {
	Entries [4]TTEntry
}

type TranspositionalTable struct {
	Buckets []TTBucket
	Mask uint64
	Gen uint8
}

func NewTranspositionalTable(sizeMB int) *TranspositionalTable {
	numBuckets := (sizeMB * 1024 * 1024) / (4 * 16)
	if numBuckets <= 0 {
		numBuckets = 1
	}
	return &TranspositionalTable{
		Buckets: make([]TTBucket, numBuckets),
		Mask: uint64(numBuckets - 1),
	}
}

func (tt *TranspositionalTable) Clear() {
	for i := range tt.Buckets {
		var zero TTBucket
		tt.Buckets[i] = zero
	}

	tt.Gen = 1
}

func (tt *TranspositionalTable) NewGeneration() {
	tt.Gen++
}

func partialKey(full uint64) uint16 {
	return uint16(full >> 48) // fuck them 48 bits
}

func (tt *TranspositionalTable) index(full uint64) int {
	return int(full & tt.Mask)
}

func ToTTScore(score, ply int) int16 {
	if score >= MateThreshold {
		return int16(score + ply)
	}

	if score <= -MateThreshold {
		return int16(score - ply)
	}

	return int16(score)
}

func FromTTScore(stored int16, ply int) int {
	s := int(stored)
	if s >= MateThreshold {
		return s - ply
	}
	if s <= -MateThreshold {
		return s + ply
	}
	return s
}


