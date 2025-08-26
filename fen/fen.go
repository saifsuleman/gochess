package fen

import (
	"fmt"
	"gochess/core"
	"strings"
)

func DefaultFEN() string {
	return "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
}

func squareFromString(s string) core.Position {
	if s == "-" {
		return 64
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	return core.Position(file + rank*8)
}

func BoardToFEN(board *core.Board) string {
	var sb strings.Builder

	// 1. Piece placement
	for rank := 7; rank >= 0; rank-- {
		emptyCount := 0
		for file := range 8 {
			sq := core.Position(rank*8 + file)
			piece := board.Pieces[sq]
			if piece == core.PieceNone {
				emptyCount++
				continue
			}
			if emptyCount > 0 {
				sb.WriteString(fmt.Sprintf("%d", emptyCount))
				emptyCount = 0
			}
			switch piece {
			case core.PieceWhitePawn:
				sb.WriteByte('P')
			case core.PieceWhiteKnight:
				sb.WriteByte('N')
			case core.PieceWhiteBishop:
				sb.WriteByte('B')
			case core.PieceWhiteRook:
				sb.WriteByte('R')
			case core.PieceWhiteQueen:
				sb.WriteByte('Q')
			case core.PieceWhiteKing:
				sb.WriteByte('K')
			case core.PieceBlackPawn:
				sb.WriteByte('p')
			case core.PieceBlackKnight:
				sb.WriteByte('n')
			case core.PieceBlackBishop:
				sb.WriteByte('b')
			case core.PieceBlackRook:
				sb.WriteByte('r')
			case core.PieceBlackQueen:
				sb.WriteByte('q')
			case core.PieceBlackKing:
				sb.WriteByte('k')
			}
		}
		if emptyCount > 0 {
			sb.WriteString(fmt.Sprintf("%d", emptyCount))
		}
		if rank > 0 {
			sb.WriteByte('/')
		}
	}

	// 2. Active color
	if board.WhiteToMove {
		sb.WriteString(" w ")
	} else {
		sb.WriteString(" b ")
	}

	// 3. Castling rights
	if board.CastlingRights == core.CastlingRightsNone {
		sb.WriteString("- ")
	} else {
		if board.CastlingRights&core.CastlingWhiteKingside != 0 {
			sb.WriteByte('K')
		}
		if board.CastlingRights&core.CastlingWhiteQueenside != 0 {
			sb.WriteByte('Q')
		}
		if board.CastlingRights&core.CastlingBlackKingside != 0 {
			sb.WriteByte('k')
		}
		if board.CastlingRights&core.CastlingBlackQueenside != 0 {
			sb.WriteByte('q')
		}
		sb.WriteByte(' ')
	}
	// 4. En passant target
	if board.EnPassantTarget == 64 {
		sb.WriteString("-")
	} else {
		file := board.EnPassantTarget & 7
		rank := board.EnPassantTarget >> 3
		sb.WriteByte('a' + byte(file))
		sb.WriteByte('1' + byte(rank))
	}

	// Note: We omit halfmove clock and fullmove number for simplicity
	return sb.String()
}

func LoadFromFEN(fen string) (*core.Board, error) {
	parts := strings.Split(fen, " ")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid FEN; not enough fields %s", fen)
	}

	board := core.NewBoard()
	// 1. Piece placement
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("invalid FEN: expected 8 ranks")
	}

	for rank := 7; rank >= 0; rank-- {
		row := ranks[7-rank] // FEN goes rank8 -> rank1
		file := 0
		for _, ch := range row {
			if ch >= '1' && ch <= '8' {
				file += int(ch - '0')
				continue
			}
			var piece core.Piece
			switch ch {
			case 'P':
				piece = core.PieceWhitePawn
			case 'N':
				piece = core.PieceWhiteKnight
			case 'B':
				piece = core.PieceWhiteBishop
			case 'R':
				piece = core.PieceWhiteRook
			case 'Q':
				piece = core.PieceWhiteQueen
			case 'K':
				piece = core.PieceWhiteKing
			case 'p':
				piece = core.PieceBlackPawn
			case 'n':
				piece = core.PieceBlackKnight
			case 'b':
				piece = core.PieceBlackBishop
			case 'r':
				piece = core.PieceBlackRook
			case 'q':
				piece = core.PieceBlackQueen
			case 'k':
				piece = core.PieceBlackKing
			default:
				return nil, fmt.Errorf("invalid piece char: %c", ch)
			}
			sq := rank*8 + file
			board.AddPiece(core.Position(sq), piece)
			file++
		}
		if file != 8 {
			return nil, fmt.Errorf("invalid FEN rank: %s", row)
		}
	}

	// 2. Active color
	switch parts[1] {
	case "w":
		board.WhiteToMove = true
	case "b":
		board.WhiteToMove = false
	default:
		return nil, fmt.Errorf("invalid active color: %s", parts[1])
	}

	// 3. Castling rights
	board.CastlingRights = core.CastlingRightsNone
	if parts[2] != "-" {
		for _, ch := range parts[2] {
			switch ch {
			case 'K':
				board.CastlingRights |= core.CastlingWhiteKingside
			case 'Q':
				board.CastlingRights |= core.CastlingWhiteQueenside
			case 'k':
				board.CastlingRights |= core.CastlingBlackKingside
			case 'q':
				board.CastlingRights |= core.CastlingBlackQueenside
			default:
				return nil, fmt.Errorf("invalid castling right: %c", ch)
			}
		}
	}

	// 4. En passant target
	board.EnPassantTarget = squareFromString(parts[3])
	fmt.Printf("en passant target: %d\n", board.EnPassantTarget)

	return board, nil
}

