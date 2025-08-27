package uci

import (
	"bufio"
	"fmt"
	"gochess/core"
	"gochess/engine"
	"gochess/fen"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UCIEngine struct {
	engine     *engine.Engine
	board      *core.Board
	searching  bool
	stopSearch chan struct{}
	mutex      sync.RWMutex
	options    map[string]UCIOption
}

type UCIOption struct {
	Name         string
	Type         string // "check", "spin", "combo", "button", "string"
	Default      interface{}
	Min          *int
	Max          *int
	ComboOptions []string
}

func NewUCIEngine() *UCIEngine {
	board, _ := fen.LoadFromFEN(fen.DefaultFEN())
	eng := engine.NewEngine(board)

	uci := &UCIEngine{
		engine:     eng,
		board:      board,
		searching:  false,
		stopSearch: make(chan struct{}),
		options:    make(map[string]UCIOption),
	}

	// Initialize default UCI options
	uci.initializeOptions()

	return uci
}

func (uci *UCIEngine) initializeOptions() {
	// Hash size option (transposition table size in MB)
	hashMin, hashMax := 1, 1024
	uci.options["Hash"] = UCIOption{
		Name:    "Hash",
		Type:    "spin",
		Default: 256,
		Min:     &hashMin,
		Max:     &hashMax,
	}

	// Clear Hash button
	uci.options["Clear Hash"] = UCIOption{
		Name: "Clear Hash",
		Type: "button",
	}

	// Ponder option (thinking on opponent's time)
	uci.options["Ponder"] = UCIOption{
		Name:    "Ponder",
		Type:    "check",
		Default: false,
	}
}

func (uci *UCIEngine) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		uci.processCommand(line)
	}
}

func (uci *UCIEngine) processCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	command := parts[0]

	switch command {
	case "uci":
		uci.handleUCI()
	case "debug":
		uci.handleDebug(parts[1:])
	case "isready":
		uci.handleIsReady()
	case "setoption":
		uci.handleSetOption(parts[1:])
	case "register":
		uci.handleRegister(parts[1:])
	case "ucinewgame":
		uci.handleNewGame()
	case "position":
		uci.handlePosition(parts[1:])
	case "go":
		uci.handleGo(parts[1:])
	case "stop":
		uci.handleStop()
	case "ponderhit":
		uci.handlePonderHit()
	case "quit":
		uci.handleQuit()
	default:
		// Unknown command - UCI engines should ignore unknown commands
	}
}

func (uci *UCIEngine) handleUCI() {
	fmt.Println("id name GoChess")
	fmt.Println("id author YourName")

	// Send options
	for _, option := range uci.options {
		uci.sendOption(option)
	}

	fmt.Println("uciok")
}

func (uci *UCIEngine) sendOption(option UCIOption) {
	switch option.Type {
	case "check":
		fmt.Printf("option name %s type check default %t\n",
			option.Name, option.Default.(bool))
	case "spin":
		fmt.Printf("option name %s type spin default %d min %d max %d\n",
			option.Name, option.Default.(int), *option.Min, *option.Max)
	case "combo":
		fmt.Printf("option name %s type combo default %s",
			option.Name, option.Default.(string))
		for _, combo := range option.ComboOptions {
			fmt.Printf(" var %s", combo)
		}
		fmt.Println()
	case "button":
		fmt.Printf("option name %s type button\n", option.Name)
	case "string":
		fmt.Printf("option name %s type string default %s\n",
			option.Name, option.Default.(string))
	}
}

func (uci *UCIEngine) handleDebug(args []string) {
	// Debug mode handling - can be used to enable/disable debug output
	if len(args) > 0 {
		if args[0] == "on" {
			// Enable debug mode
		} else if args[0] == "off" {
			// Disable debug mode
		}
	}
}

func (uci *UCIEngine) handleIsReady() {
	fmt.Println("readyok")
}

func (uci *UCIEngine) handleSetOption(args []string) {
	if len(args) < 4 || args[0] != "name" {
		return
	}

	// Find "value" keyword
	valueIndex := -1
	for i, arg := range args {
		if arg == "value" {
			valueIndex = i
			break
		}
	}

	if valueIndex == -1 {
		return
	}

	// Extract option name (everything between "name" and "value")
	optionName := strings.Join(args[1:valueIndex], " ")

	// Extract option value (everything after "value")
	optionValue := strings.Join(args[valueIndex+1:], " ")

	uci.setOption(optionName, optionValue)
}

func (uci *UCIEngine) setOption(name, value string) {
	option, exists := uci.options[name]
	if !exists {
		return
	}

	switch name {
	case "Hash":
		if hashSize, err := strconv.Atoi(value); err == nil {
			// Resize transposition table
			uci.engine.TT = engine.NewTranspositionalTable(hashSize)
		}
	case "Clear Hash":
		// Clear the transposition table
		uci.engine.TT.Clear()
	case "Ponder":
		// Handle ponder setting
		option.Default = (value == "true")
		uci.options[name] = option
	}
}

func (uci *UCIEngine) handleRegister(args []string) {
	// Registration handling - not needed for open source engines
	// Just send "registration checking" followed by "registration ok"
	fmt.Println("registration checking")
	fmt.Println("registration ok")
}

func (uci *UCIEngine) handleNewGame() {
	uci.mutex.Lock()
	defer uci.mutex.Unlock()

	// Reset to starting position
	board, _ := fen.LoadFromFEN(fen.DefaultFEN())
	uci.board = board
	uci.engine = engine.NewEngine(board)

	// Clear transposition table for new game
	uci.engine.TT.NewGeneration()
}

func (uci *UCIEngine) handlePosition(args []string) {
	uci.mutex.Lock()
	defer uci.mutex.Unlock()

	if len(args) == 0 {
		return
	}

	var board *core.Board
	var err error
	moveIndex := -1

	switch args[0] {
	case "startpos":
		board, err = fen.LoadFromFEN(fen.DefaultFEN())
		moveIndex = 1
	case "fen":
		if len(args) < 7 {
			return
		}
		// FEN string is typically 6 parts
		fenString := strings.Join(args[1:7], " ")
		board, err = fen.LoadFromFEN(fenString)
		moveIndex = 7
	default:
		return
	}

	if err != nil {
		return
	}

	uci.board = board

	// Apply moves if present
	if moveIndex < len(args) && args[moveIndex] == "moves" {
		for i := moveIndex + 1; i < len(args); i++ {
			move := uci.parseMove(args[i])
			if move != nil {
				uci.board.Push(move)
			}
		}
	}

	uci.engine.Board = uci.board.Clone()
}

func (uci *UCIEngine) parseMove(moveStr string) *core.Move {
	if len(moveStr) < 4 {
		return nil
	}

	fromFile := int(moveStr[0] - 'a')
	fromRank := int(moveStr[1] - '1')
	toFile := int(moveStr[2] - 'a')
	toRank := int(moveStr[3] - '1')

	if fromFile < 0 || fromFile > 7 || fromRank < 0 || fromRank > 7 ||
		toFile < 0 || toFile > 7 || toRank < 0 || toRank > 7 {
		return nil
	}

	from := core.Position(fromRank*8 + fromFile)
	to := core.Position(toRank*8 + toFile)

	var promotion core.Piece = core.PieceNone
	if len(moveStr) == 5 {
		// Handle promotion
		piece := uci.board.Pieces[from]
		color := piece.Color()

		switch moveStr[4] {
		case 'q':
			promotion = core.Piece(color | core.PieceTypeQueen)
		case 'r':
			promotion = core.Piece(color | core.PieceTypeRook)
		case 'b':
			promotion = core.Piece(color | core.PieceTypeBishop)
		case 'n':
			promotion = core.Piece(color | core.PieceTypeKnight)
		}
	}

	move := &core.Move{
		From:      from,
		To:        to,
		Promotion: promotion,
	}

	// Verify the move is legal
	if !uci.board.IsMoveLegal(*move) {
		return nil
	}

	return move
}

func (uci *UCIEngine) handleGo(args []string) {
	uci.mutex.Lock()
	if uci.searching {
		uci.mutex.Unlock()
		return
	}
	uci.searching = true
	uci.stopSearch = make(chan struct{})
	uci.mutex.Unlock()

	// Parse go command parameters
	searchParams := uci.parseGoCommand(args)

	go uci.search(searchParams)
}

type SearchParams struct {
	SearchMoves []string
	Ponder      bool
	WTime       *time.Duration
	BTime       *time.Duration
	WInc        *time.Duration
	BInc        *time.Duration
	MovesToGo   *int
	Depth       *int
	Nodes       *uint64
	Mate        *int
	MoveTime    *time.Duration
	Infinite    bool
}

func (uci *UCIEngine) parseGoCommand(args []string) SearchParams {
	params := SearchParams{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "searchmoves":
			// Collect moves until we hit another parameter
			j := i + 1
			for j < len(args) && !strings.HasPrefix(args[j], "wtime") &&
				!strings.HasPrefix(args[j], "btime") && args[j] != "winc" &&
				args[j] != "binc" && args[j] != "movestogo" &&
				args[j] != "depth" && args[j] != "nodes" &&
				args[j] != "mate" && args[j] != "movetime" &&
				args[j] != "infinite" && args[j] != "ponder" {
				params.SearchMoves = append(params.SearchMoves, args[j])
				j++
			}
			i = j - 1
		case "ponder":
			params.Ponder = true
		case "wtime":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					duration := time.Duration(ms) * time.Millisecond
					params.WTime = &duration
				}
				i++
			}
		case "btime":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					duration := time.Duration(ms) * time.Millisecond
					params.BTime = &duration
				}
				i++
			}
		case "winc":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					duration := time.Duration(ms) * time.Millisecond
					params.WInc = &duration
				}
				i++
			}
		case "binc":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					duration := time.Duration(ms) * time.Millisecond
					params.BInc = &duration
				}
				i++
			}
		case "movestogo":
			if i+1 < len(args) {
				if moves, err := strconv.Atoi(args[i+1]); err == nil {
					params.MovesToGo = &moves
				}
				i++
			}
		case "depth":
			if i+1 < len(args) {
				if depth, err := strconv.Atoi(args[i+1]); err == nil {
					params.Depth = &depth
				}
				i++
			}
		case "nodes":
			if i+1 < len(args) {
				if nodes, err := strconv.ParseUint(args[i+1], 10, 64); err == nil {
					params.Nodes = &nodes
				}
				i++
			}
		case "mate":
			if i+1 < len(args) {
				if mate, err := strconv.Atoi(args[i+1]); err == nil {
					params.Mate = &mate
				}
				i++
			}
		case "movetime":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					duration := time.Duration(ms) * time.Millisecond
					params.MoveTime = &duration
				}
				i++
			}
		case "infinite":
			params.Infinite = true
		}
	}

	return params
}

func (uci *UCIEngine) search(params SearchParams) {
	defer func() {
		uci.mutex.Lock()
		uci.searching = false
		uci.mutex.Unlock()
	}()

	// Calculate time budget
	timeBudget := uci.calculateTimeBudget(params)

	// Clone the board for searching
	searchBoard := uci.board.Clone()
	searchEngine := engine.NewEngine(searchBoard)

	// Perform the search
	bestMove := searchEngine.FindBestMove(timeBudget, true)

	if bestMove != nil {
		moveStr := uci.moveToString(*bestMove)
		fmt.Printf("bestmove %s\n", moveStr)
	} else {
		// No legal moves (checkmate or stalemate)
		fmt.Println("bestmove 0000")
	}
}

func (uci *UCIEngine) calculateTimeBudget(params SearchParams) time.Duration {
	// If movetime is specified, use it directly
	if params.MoveTime != nil {
		return *params.MoveTime
	}

	// If infinite search, use a very long time
	if params.Infinite {
		return time.Hour * 24 // 24 hours
	}

	// Simple time management based on remaining time
	var ourTime *time.Duration
	var ourInc *time.Duration

	if uci.board.WhiteToMove {
		ourTime = params.WTime
		ourInc = params.WInc
	} else {
		ourTime = params.BTime
		ourInc = params.BInc
	}

	if ourTime != nil {
		// Basic time management: use 1/30th of remaining time plus increment
		budget := *ourTime / 30
		if ourInc != nil {
			budget += *ourInc
		}

		// Don't use less than 1ms or more than remaining time
		if budget < time.Millisecond {
			budget = time.Millisecond
		}
		if budget > *ourTime {
			budget = *ourTime
		}

		return budget
	}

	// Default fallback
	return time.Second
}

func (uci *UCIEngine) moveToString(move core.Move) string {
	from := move.From
	to := move.To

	fromFile := byte('a' + (from & 7))
	fromRank := byte('1' + (from >> 3))
	toFile := byte('a' + (to & 7))
	toRank := byte('1' + (to >> 3))

	moveStr := string([]byte{fromFile, fromRank, toFile, toRank})

	if move.Promotion != core.PieceNone {
		switch move.Promotion.Type() {
		case core.PieceTypeQueen:
			moveStr += "q"
		case core.PieceTypeRook:
			moveStr += "r"
		case core.PieceTypeBishop:
			moveStr += "b"
		case core.PieceTypeKnight:
			moveStr += "n"
		}
	}

	return moveStr
}

func (uci *UCIEngine) handleStop() {
	uci.mutex.RLock()
	if uci.searching {
		close(uci.stopSearch)
	}
	uci.mutex.RUnlock()
}

func (uci *UCIEngine) handlePonderHit() {
	// Convert ponder search to normal search
	// This is a simplified implementation
}

func (uci *UCIEngine) handleQuit() {
	uci.handleStop()
	os.Exit(0)
}

// Main function for UCI mode
func RunUCI() {
	uciEngine := NewUCIEngine()
	uciEngine.Run()
}
