package games

type GameEndedFn = func(string, []string) Result

type Oracle interface {
	Initialize(fen string)
	MakeMove(mov string) error
	FEN() string
	GameResult() (Result, string)
	ZeroMoves() bool
}

type Result uint8

const (
	Ongoing Result = iota
	StmWins
	XtmWins
	Draw
)
