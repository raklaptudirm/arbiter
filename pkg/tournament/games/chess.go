package games

import (
	"laptudirm.com/x/arbiter/pkg/tournament/common"
	"laptudirm.com/x/mess/pkg/board"
	"laptudirm.com/x/mess/pkg/formats/fen"
)

func HasChessGameEnded(fenstr string, moves []string) (bool, common.Score) {
	chessboard := board.New(board.FEN(fen.FromString(fenstr)))
	for _, mov := range moves {
		chessboard.MakeMove(chessboard.NewMoveFromString(mov))
	}

	movelist := chessboard.GenerateMoves(false)

	switch {
	case len(movelist) == 0:
		if chessboard.IsInCheck(chessboard.SideToMove) {
			return true, common.GameLostBy[int(chessboard.SideToMove)]
		}

		fallthrough

	case chessboard.DrawClock >= 100,
		chessboard.IsThreefoldRepetition():
		return true, common.Draw
	}

	return false, 0
}
