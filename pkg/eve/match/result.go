package match

type Result int

const (
	Player1Wins Result = +1
	Draw        Result = 0
	Player2Wins Result = -1
)

var GameLostBy = [2]Result{
	0: Player2Wins,
	1: Player1Wins,
}

func (result Result) String() string {
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
