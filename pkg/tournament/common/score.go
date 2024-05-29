package common

type Score int

const (
	WhiteWins Score = +1
	Draw      Score = 0
	BlackWins Score = -1
)

var GameLostBy = [2]Score{
	0: BlackWins,
	1: WhiteWins,
}

func (result Score) String() string {
	switch result {
	case WhiteWins:
		return "1-0"
	case Draw:
		return "1/2-1/2"
	case BlackWins:
		return "0-1"
	default:
		return "?-?"
	}
}
