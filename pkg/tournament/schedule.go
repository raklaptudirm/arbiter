package tournament

import (
	"fmt"
	"slices"
)

type Scheduler interface {
	Initialize(*Tournament)
	NextPair(game_number int) (int, int)
	TotalGames() int
}

type RoundRobin struct {
	tournament *Tournament

	player_count int

	pair_number         int
	games_per_encounter int

	player1, player2          int
	circle_top, circle_bottom []int
}

func (rr *RoundRobin) Initialize(tour *Tournament) {
	rr.tournament = tour
	rr.player_count = len(tour.Config.Engines)
	rounded_total := rr.player_count + rr.player_count%2

	rr.games_per_encounter = 2 * tour.Config.GamePairs

	rr.circle_top = make([]int, rounded_total/2)
	rr.circle_bottom = make([]int, rounded_total/2)

	for i := 0; i < rounded_total; i++ {
		if i < rounded_total/2 {
			rr.circle_top[i] = i
		} else {
			rr.circle_bottom[rounded_total-i-1] = i
		}
	}

	fmt.Println(rr.circle_top)
	fmt.Println(rr.circle_bottom)
}

func (rr *RoundRobin) NextPair(game_num int) (int, int) {
	if game_num%rr.games_per_encounter != 0 {
		rr.player1, rr.player2 = rr.player2, rr.player1
		return rr.player1, rr.player2
	}

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

	return rr.NextPair(game_num)
}

func (rr *RoundRobin) TotalGames() int {
	return (rr.player_count * (rr.player_count - 1) / 2) * rr.games_per_encounter
}
