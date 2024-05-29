package common

type Score int

const (
	Player1Wins Score = +1
	Draw        Score = 0
	Player2Wins Score = -1
)

var GameLostBy = [2]Score{
	0: Player2Wins,
	1: Player1Wins,
}

func (result Score) String() string {
	switch result {
	case Player1Wins:
		return "1-0"
	case Draw:
		return "1/2-1/2"
	case Player2Wins:
		return "0-1"
	default:
		return "?-?"
	}
}
