package games

import (
	"errors"
	"strings"

	"laptudirm.com/x/mess/pkg/board"
	"laptudirm.com/x/mess/pkg/board/move"
	"laptudirm.com/x/mess/pkg/formats/fen"
)

type ChessOracle struct {
	board *board.Board
	moves []move.Move
}

func (oracle *ChessOracle) Initialize(fenstr string) {
	oracle.board = board.New(board.FEN(fen.FromString(fenstr)))
	oracle.moves = oracle.board.GenerateMoves(false)
}

func (oracle *ChessOracle) MakeMove(mov_str string) error {
	found, index := false, 0
	for i, mov := range oracle.moves {
		if strings.EqualFold(mov.String(), mov_str) {
			found = true
			index = i
			break
		}
	}

	if !found {
		return errors.New("illegal move")
	}

	oracle.board.MakeMove(oracle.moves[index])
	oracle.moves = oracle.board.GenerateMoves(false)
	return nil
}

func (oracle *ChessOracle) FEN() string {
	fen := [6]string(oracle.board.FEN())
	return strings.Join(fen[:], " ")
}

func (oracle *ChessOracle) ZeroMoves() bool {
	return oracle.board.DrawClock == 0
}

func (oracle *ChessOracle) GameResult() Result {
	switch {
	case len(oracle.moves) == 0:
		if oracle.board.IsInCheck(oracle.board.SideToMove) {
			return Draw
		}

		fallthrough

	case oracle.board.DrawClock >= 100,
		oracle.board.IsThreefoldRepetition(),
		oracle.board.IsInsufficientMaterial():
		return Draw
	}

	return Ongoing
}
