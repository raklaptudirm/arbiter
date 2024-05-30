package tournament

import (
	"slices"
)

type Scheduler interface {
	Initialize(int)
	NextEncounter() (int, int)
	TotalEncounters() int
}

type RoundRobin struct {
	player_count int

	pair_number int

	player1, player2          int
	circle_top, circle_bottom []int
}

func (rr *RoundRobin) Initialize(n int) {
	rr.player_count = n
	rounded_total := rr.player_count + rr.player_count%2

	rr.circle_top = make([]int, rounded_total/2)
	rr.circle_bottom = make([]int, rounded_total/2)

	for i := 0; i < rounded_total; i++ {
		if i < rounded_total/2 {
			rr.circle_top[i] = i
		} else {
			rr.circle_bottom[rounded_total-i-1] = i
		}
	}

	rr.pair_number = 0
}

func (rr *RoundRobin) NextEncounter() (int, int) {
	if rr.pair_number >= len(rr.circle_top) {
		rr.pair_number = 0

		last_idx := len(rr.circle_top) - 1
		last_elem := rr.circle_top[last_idx]

		rr.circle_top = slices.Insert(rr.circle_top, 1, rr.circle_bottom[0])[:last_idx+1]
		rr.circle_bottom = append(rr.circle_bottom, last_elem)[1:]
	}

	rr.player1 = rr.circle_top[rr.pair_number]
	rr.player2 = rr.circle_bottom[rr.pair_number]

	rr.pair_number++

	if rr.player1 < rr.player_count && rr.player2 < rr.player_count {
		return rr.player1, rr.player2
	}

	return rr.NextEncounter()
}

func (rr *RoundRobin) TotalEncounters() int {
	return rr.player_count * (rr.player_count - 1) / 2
}

type Gauntlet struct {
	player_count int
	game_number  int
}

func (g *Gauntlet) Initialize(n int) {
	g.player_count = n
	g.game_number = 0
}

func (g *Gauntlet) NextEncounter() (int, int) {
	g.game_number++
	return 0, g.game_number
}

func (g *Gauntlet) TotalEncounters() int {
	return g.player_count - 1
}
