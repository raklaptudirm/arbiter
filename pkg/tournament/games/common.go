package games

import (
	. "laptudirm.com/x/arbiter/pkg/tournament/common"
)

type GameEndedFn = func(string, []string) (bool, Score)
