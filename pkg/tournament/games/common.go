package games

type GameEndedFn = func(string, []string) Result

type Result uint8

const (
	Ongoing Result = iota
	StmWins
	XtmWins
	Draw
)
