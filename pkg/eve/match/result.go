package match

// PairResult represents the result of a single game pair.
type PairResult int

const (
	WinWin   = PairResult(Win + Win)   // Player 1 Double kills
	WinDraw  = PairResult(Win + Draw)  // Player 1 Wins and Holds
	DrawDraw = PairResult(Draw + Draw) // Win-Loss or Draw-Draw
	DrawLoss = PairResult(Draw + Loss) // Player 2 Wins and Holds
	LossLoss = PairResult(Loss + Loss) // Player 2 Double kills
)

// GetPairResult returns the PairResult given the Result of each game in the
// pair. Game 1 should be Player 1 vs 2 and Game 2 should be Player 2 vs 1.
func GetPairResult(result1, result2 Result) PairResult {
	return PairResult(result1 + result2)
}

// Result represents the result of a single match.
type Result int

const (
	Win  Result = +1
	Draw Result = 0
	Loss Result = -1
)

// GameLostBy maps the losing player to the match's Result.
var GameLostBy = [2]Result{
	0: Loss,
	1: Win,
}

// String returns a string representation of the given Result.
func (result Result) String() string {
	switch result {
	case Win:
		return "1-0"
	case Draw:
		return "1/2-1/2"
	case Loss:
		return "0-1"
	default:
		return "?-?"
	}
}
