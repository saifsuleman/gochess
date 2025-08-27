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
	Move       core.Move
	Score      int16
	depth      int8
	bound      Bound
	gen        uint8
}

type TTBucket struct {
	Entries [4]TTEntry
}

type TranspositionalTable struct {
	Buckets []TTBucket
	Mask    uint64
	Gen     uint8
}

func NewTranspositionalTable(sizeMB int) *TranspositionalTable {
	numBuckets := (sizeMB * 1024 * 1024) / (4 * 16)
	if numBuckets <= 0 {
		numBuckets = 1
	}
	return &TranspositionalTable{
		Buckets: make([]TTBucket, numBuckets),
		Mask:    uint64(numBuckets - 1),
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

func (tt *TranspositionalTable) Probe(key uint64) (bit bool, e TTEntry) {
	idx := tt.index(key)
	pk := partialKey(key)
	b := &tt.Buckets[idx]

	for i := range b.Entries {
		if b.Entries[i].PartialKey == pk {
			return true, b.Entries[i]
		}
	}

	return false, TTEntry{}
}

func (tt *TranspositionalTable) Store(key uint64, depth int, score int, bound Bound, move core.Move, ply int) {
	idx := tt.index(key)
	pk := partialKey(key)
	b := &tt.Buckets[idx]

	for i := range b.Entries {
		if b.Entries[i].PartialKey == pk {
			if depth >= int(b.Entries[i].depth) || bound == FlagExact {
				b.Entries[i].Score = ToTTScore(score, ply)
				b.Entries[i].depth = int8(depth)
				b.Entries[i].bound = bound
				if move != (core.Move{}) {
					b.Entries[i].Move = move
				}
				b.Entries[i].gen = tt.Gen
			} else {
				if move != (core.Move{}) {
					b.Entries[i].Move = move
				}

				b.Entries[i].gen = tt.Gen
			}

			return
		}
	}

	victim := 0
	bestScore := 1<<31 - 1
	for i := range b.Entries {
		if b.Entries[i].PartialKey == 0 {
			victim = i
			break
		}

		penalty := int(b.Entries[i].depth)
		if b.Entries[i].gen == tt.Gen {
			penalty += 8
		}

		if penalty < bestScore {
			bestScore = penalty
			victim = i
		}
	}

	b.Entries[victim] = TTEntry{
		PartialKey: pk,
		Score:      ToTTScore(score, ply),
		depth:      int8(depth),
		bound:      bound,
		Move:       move,
		gen:        tt.Gen,
	}
}

func (tt *TranspositionalTable) ProbeCut(key uint64, depth, alpha, beta, ply int) (ok bool, score int, flag Bound, move core.Move) {
	hit, e := tt.Probe(key)
	if !hit {
		return false, 0, FlagExact, core.Move{}
	}
	s := FromTTScore(e.Score, ply)
	if int(e.depth) >= depth {
		switch e.bound {
		case FlagExact:
			return true, s, e.bound, e.Move
		case FlagLower:
			if s >= beta {
				return true, s, e.bound, e.Move
			}
		case FlagUpper:
			if s <= alpha {
				return true, s, e.bound, e.Move
			}
		}
	}

	return false, s, e.bound, e.Move
}
