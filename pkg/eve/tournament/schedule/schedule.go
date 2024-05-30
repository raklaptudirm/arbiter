package schedule

import (
	"fmt"
)

func New(name string) (Scheduler, error) {
	switch name {
	case "round-robin", "":
		return &RoundRobin{}, nil
	case "gauntlet":
		return &Gauntlet{}, nil
	default:
		return nil, fmt.Errorf("new tour: invalid scheduler %s", name)
	}
}

type Scheduler interface {
	Initialize(int)
	NextEncounter() (int, int)
	TotalEncounters() int
}
