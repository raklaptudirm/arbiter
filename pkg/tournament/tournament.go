// Copyright © 2023 Rak Laptudirm <rak@laptudirm.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tournament

import (
	"fmt"
	"math"

	"laptudirm.com/x/arbiter/pkg/stats"
	"laptudirm.com/x/arbiter/pkg/tournament/common"
	"laptudirm.com/x/arbiter/pkg/tournament/games"
)

func NewTournament(config Config) (*Tournament, error) {
	var tour Tournament
	tour.Config = config
	tour.Scores = make([]struct {
		Wins   int
		Losses int
		Draws  int
	}, len(config.Engines))

	var err error
	tour.openings, err = NewBook(config.Openings.File, config.Openings.Order)
	if err != nil {
		return nil, err
	}

	switch config.Scheduler {
	case "round-robin", "":
		tour.Scheduler = &RoundRobin{}
	default:
		return nil, fmt.Errorf("new tour: invalid scheduler %s", config.Scheduler)
	}

	return &tour, nil
}

type Tournament struct {
	Config Config

	Scheduler Scheduler
	openings  *Book

	Games  int
	Scores []struct {
		Wins, Losses, Draws int
	}
}

func (tour *Tournament) Start() error {
	for round := 0; round < tour.Config.Rounds; round++ {
		tour.Scheduler.Initialize(tour)

		for game_num := 0; game_num < tour.Scheduler.TotalGames(); game_num++ {
			p1, p2 := tour.Scheduler.NextPair(game_num)
			fmt.Printf(
				"Round #%d Game #%d: %s vs %s (%s)\n",
				round+1,
				game_num+1,
				tour.Config.Engines[p1].Name,
				tour.Config.Engines[p2].Name,
				tour.openings.Current(),
			)

			game, err := NewGame(tour.Config.Engines[p1], tour.Config.Engines[p2], tour.openings.Current())
			if err != nil {
				return err
			}

			switch tour.Config.Game {
			case "chess":
				game.Oracle = &games.ChessOracle{}
			case "ataxx":
				game.Oracle = &games.ChessOracle{}
			}

			score, err := game.Play()
			if err != nil {
				return err
			}

			switch score {
			case common.Player1Wins:
				tour.Scores[p1].Wins++
				tour.Scores[p2].Losses++

			case common.Player2Wins:
				tour.Scores[p2].Wins++
				tour.Scores[p1].Losses++

			case common.Draw:
				tour.Scores[p1].Draws++
				tour.Scores[p2].Draws++
			}

			fmt.Println("    Name               Elo Err   Wins Loss Draw   Total")
			for i, engine := range tour.Config.Engines {
				score := tour.Scores[i]
				lower, elo, upper := stats.Elo(score.Wins, score.Draws, score.Losses)
				fmt.Printf(
					"%2d. %-15s   %+4.0f %3.0f   %4d %4d %4d   %5d\n",
					i+1, engine.Name,
					elo, math.Max(upper-elo, elo-lower),
					score.Wins, score.Losses, score.Draws,
					score.Wins+score.Losses+score.Draws)
			}

			if game_num%2 == 1 {
				tour.openings.Next()
			}
		}
	}

	return nil
}

type Config struct {
	// The engines participating in the tournament.
	Engines []EngineConfig `yaml:"engines"`

	// The game that will be played.
	Game string `yaml:"game"`

	// Number of games that will be played concurrently.
	Concurrency int `yaml:"concurrency"`

	// Game adjudication stuff.
	// Draw        struct {
	// 	MoveNumber int
	// 	MoveCount  int
	// 	Score      int
	// }
	// Resign struct {
	// 	MoveCount int
	// 	Score     int
	// }
	// MaxMoves int

	Event string `yaml:"event"` // Event field of the PGN.
	Site  string `yaml:"site"`  // Site field of the PGN.

	Scheduler string `yaml:"scheduler"`

	// 1 Tournament = {ROUNDS} Rounds
	// 1 Round      = {SOME_N} Encounters
	// 1 Encounter  = {GAME_P} Game Pairs
	// 1 Game Pair  = 2 Games
	Rounds    int `yaml:"rounds"`     // Number of rounds to run the tournament for.
	GamePairs int `yaml:"game-pairs"` // Number of games per encounter in every round.

	Sprt struct {
		Elo0, Elo1  int // The null and the alternate elo hypotheses.
		Alpha, Beta int // Confidence bounds for Error types I and II.
	}

	Openings struct {
		File string
		// Format string // only EPD opening files supported.
		Order string
		Start int

		// When to switch to a new opening in a tournament:
		// default:
		Policy string

		// Number of times to play each opening. Only functional if the opening
		// policy is set to default (or is unset, same difference).
		Repeat int
	}

	PGNOut string // File to store the game PGNs at.
	EPDOut string // File to store the game ends EPD at.

	// Restart a crashed engine instead of stopping the match.
	Recover bool
}
