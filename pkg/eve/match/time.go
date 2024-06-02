package match

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// TimeControl stores the time control configuration for an engine.
type TimeControl struct {
	MovesToGo int
	Base, Inc time.Duration
}

// ParseTime parses the given time-control configuration string into a
// TimeControl object. The string should have a format of
// movestogo/time+increment, where both time and increment in seconds. The
// movestogo part is optional and maybe omitted for a non-cyclic time control.
func ParseTime(time_str string) (TimeControl, error) {
	var tc TimeControl

	// Split the string into 'movestogo' and 'time+increment' parts.
	moves_str, time_str, found := strings.Cut(time_str, "/")
	tc.MovesToGo = -1
	var err error
	if found {
		// Parse the given movestogo.
		tc.MovesToGo, err = strconv.Atoi(moves_str)
		if err != nil {
			return TimeControl{}, err
		}
	} else {
		// If there is no movestogo part, moves_str is the time_str.
		time_str = moves_str
	}

	// Split the time string into 'time' and 'increment' parts.
	time_str, inc_str, found := strings.Cut(time_str, "+")
	if !found {
		// Both the time and increment are required fields.
		return TimeControl{}, errors.New("parse tc: increment not found")
	}

	// Parse the increment string.
	incs, err := strconv.ParseFloat(inc_str, 32)
	if err != nil {
		return TimeControl{}, err
	}

	// Parse the base time string.
	secs, err := strconv.ParseFloat(time_str, 32)
	if err != nil {
		return TimeControl{}, err
	}

	tc.Inc = time.Millisecond * time.Duration(incs*1000)
	tc.Base = time.Millisecond * time.Duration(secs*1000)
	return tc, nil
}
