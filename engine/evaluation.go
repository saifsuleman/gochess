package engine

import (
	"gochess/core"
)

const (
	PhaseOpening Phase = iota
	PhaseMiddlegame
	PhaseEndgame
)

var phaseWeights = map[uint8]int{
	core.PieceTypeKnight: 1,
	core.PieceTypeBishop: 1,
	core.PieceTypeRook:   2,
	core.PieceTypeQueen:  4,
}

const maxPhaseScore = 24

type Phase int
type PST [64]int

func (e *Engine) Evaluate() int {
	score := 0
	mask := e.Board.AllPieces

	phase := determinePhase(e.Board)

	for mask != 0 {
		lsb := core.Position(mask.PopLSB())
		piece := e.Board.Pieces[lsb]
		value := e.PieceValue(piece)
		value += pstValue(piece, lsb, phase)
		if piece.Color() == core.PieceColorWhite {
			score += value
		} else {
			score -= value
		}
	}

	if e.Board.WhiteToMove {
		return score
	} else {
		return -score
	}
}

func determinePhase(b *core.Board) Phase {
	score := 0
	for sq := range 64 {
		piece := b.Pieces[sq]
		pt := piece.Type()
		if w, ok := phaseWeights[pt]; ok {
			score += w
		}
	}
	switch {
	case score > maxPhaseScore*2/3:
		return PhaseOpening
	case score > maxPhaseScore/3:
		return PhaseMiddlegame
	default:
		return PhaseEndgame
	}
}

func pstValue(piece core.Piece, sq core.Position, phase Phase) int {
	var normalized core.Position

	color := piece.Color()
	type_ := piece.Type()

	if color == core.PieceColorWhite {
		normalized = sq
	} else {
		rank := sq >> 3
		file := sq & 7
		normalized = (7-rank)<<3 | file
	}

	switch type_ {
	case core.PieceTypePawn:
		return pawnPST[normalized]
	case core.PieceTypeKnight:
		return knightPST[normalized]
	case core.PieceTypeBishop:
		return bishopPST[normalized]
	case core.PieceTypeRook:
		return rookPST[normalized]
	case core.PieceTypeQueen:
		return queenPST[normalized]
	case core.PieceTypeKing:
		if phase == PhaseEndgame {
			return kingPSTEndgame[normalized]
		} else {
			return kingPSTMiddlegame[normalized]
		}
	}

	return 0 // i mean this shouldn't happen
}

func (e *Engine) PieceValue(piece core.Piece) int {
	type_ := piece.Type()
	switch type_ {
	case core.PieceTypeNone:
		return 0
	case core.PieceTypePawn:
		return 100
	case core.PieceTypeKnight:
		return 320
	case core.PieceTypeBishop:
		return 330
	case core.PieceTypeRook:
		return 500
	case core.PieceTypeQueen:
		return 900
	case core.PieceTypeKing:
		return 20000
	default:
		return 0
	}
}
