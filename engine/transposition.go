package engine

import "gochess/core"

type TTEntry struct {
	depth int
	score int
	flag int
	move core.Move
}

type TranspositionalTable struct {

}
