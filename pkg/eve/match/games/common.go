package games

func GetOracle(name string) Oracle {
	switch name {
	case "ataxx":
		return &AtaxxOracle{}
	case "chess":
		return &ChessOracle{}
	default:
		return nil
	}
}

type Oracle interface {
	Initialize(fen string)
	SideToMove() Color
	MakeMove(mov string) error
	FEN() string
	GameResult() (Result, string)
	ZeroMoves() bool
}

type Color uint8

const (
	White Color = iota
	Black       = iota
)

type Result uint8

const (
	Ongoing Result = iota
	StmWins
	XtmWins
	Draw
)
