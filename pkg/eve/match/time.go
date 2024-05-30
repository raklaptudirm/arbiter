package match

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type TimeControl struct {
	MovesToGo int
	Base, Inc time.Duration
}

// movestogo/time+increment, both time and increment in seconds
func ParseTime(time_str string) (TimeControl, error) {
	var tc TimeControl

	moves_str, time_str, found := strings.Cut(time_str, "/")
	tc.MovesToGo = -1
	var err error
	if found {
		tc.MovesToGo, err = strconv.Atoi(moves_str)
		if err != nil {
			return TimeControl{}, err
		}
	} else {
		time_str = moves_str
	}

	time_str, inc_str, found := strings.Cut(time_str, "+")
	if !found {
		return TimeControl{}, errors.New("parse tc: increment not found")
	}

	incs, err := strconv.ParseFloat(inc_str, 32)
	if err != nil {
		return TimeControl{}, err
	}

	secs, err := strconv.ParseFloat(time_str, 32)
	if err != nil {
		return TimeControl{}, err
	}

	tc.Inc = time.Millisecond * time.Duration(incs*1000)
	tc.Base = time.Millisecond * time.Duration(secs*1000)
	return tc, nil
}
