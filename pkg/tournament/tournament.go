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
	"os"

	"github.com/sirupsen/logrus"
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

	tour.games = make(chan *Game)
	tour.results = make(chan Result)
	tour.complete = make(chan bool)

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

	games    chan *Game
	results  chan Result
	complete chan bool

	Games  int
	Scores []struct {
		Wins, Losses, Draws int
	}
}

func (tour *Tournament) Start() error {
	// 1 Tournament = {ROUNDS} Rounds
	// 1 Round      = {SOME_N} Encounters
	// 1 Encounter  = {GAME_P} Game Pairs
	// 1 Game Pair  = 2 Games

	go tour.ResultHandler()
	for i := 0; i < tour.Config.Concurrency; i++ {
		go tour.Thread()
	}

	for round := 0; round < tour.Config.Rounds; round++ {
		tour.Scheduler.Initialize(len(tour.Config.Engines))

		for encounter := 0; encounter < tour.Scheduler.TotalEncounters(); encounter++ {
			p1, p2 := tour.Scheduler.NextEncounter()

			for pair := 0; pair < tour.Config.GamePairs; pair++ {
				for game := 0; game < 2; game++ {
					game, err := NewGame(tour.Config.Engines[p1], tour.Config.Engines[p2], tour.openings.Current())
					if err != nil {
						return err
					}

					game.Round, game.Number = round+1, encounter+1
					game.Player1, game.Player2 = p1, p2

					switch tour.Config.Game {
					case "chess":
						game.Oracle = &games.ChessOracle{}
					case "ataxx":
						game.Oracle = &games.AtaxxOracle{}
					}

					tour.games <- game
					p1, p2 = p2, p1
				}

				if encounter%2 == 1 {
					tour.openings.Next()
				}
			}
		}
	}

	close(tour.games)
	<-tour.complete

	return nil
}

func (tour *Tournament) Thread() {
	for game := range tour.games {
		if err := tour.RunGame(game); err != nil {
			logrus.Error(err)
		}
	}
}

func (tour *Tournament) RunGame(game *Game) error {
	fmt.Printf(
		"Round #%d Game #%d: %s vs %s (%s)\n",
		game.Round,
		game.Number,
		game.Engines[0].config.Name,
		game.Engines[1].config.Name,
		tour.openings.Current(),
	)

	score, reason := game.Play()

	tour.results <- Result{
		Game: game,

		Player1: game.Player1,
		Player2: game.Player2,

		Result: score,
		Reason: reason,
	}

	return nil
}

func (tour *Tournament) ResultHandler() {
	result_count := 0
	result_target := tour.Config.Rounds * tour.Scheduler.TotalEncounters() * tour.Config.GamePairs * 2
	for result := range tour.results {
		result_count++

		switch result.Result {
		case common.Player1Wins:
			tour.Scores[result.Player1].Wins++
			tour.Scores[result.Player2].Losses++

		case common.Player2Wins:
			tour.Scores[result.Player2].Wins++
			tour.Scores[result.Player1].Losses++

		case common.Draw:
			tour.Scores[result.Player1].Draws++
			tour.Scores[result.Player2].Draws++
		}

		fmt.Fprintf(os.Stderr,
			"Round #%d Game #%d: %s vs %s: %s\n",
			result.Game.Round,
			result.Game.Number,
			result.Game.Engines[0].config.Name,
			result.Game.Engines[1].config.Name,
			result,
		)

		if result_count%5 == 0 {
			fmt.Println("╔══════════════════════════════════════════════════════════╗")
			fmt.Println("║    Name               Elo Error   Wins Loss Draw   Total ║")
			fmt.Println("╠══════════════════════════════════════════════════════════╣")
			for i, engine := range tour.Config.Engines {
				score := tour.Scores[i]
				lower, elo, upper := stats.Elo(score.Wins, score.Draws, score.Losses)
				fmt.Printf(
					"║ %2d. %-15s   %+4.0f %4.0f   %4d %4d %4d   %5d ║\n",
					i+1, engine.Name,
					elo, math.Abs(math.Max(upper-elo, elo-lower)),
					score.Wins, score.Losses, score.Draws,
					score.Wins+score.Losses+score.Draws)
			}
			fmt.Println("╚══════════════════════════════════════════════════════════╝")
		}

		if result_count == result_target {
			close(tour.results)
			tour.complete <- true
			return
		}
	}

}

type Result struct {
	Game *Game

	Player1, Player2 int

	Result common.Score
	Reason string
}

func (result Result) String() string {
	switch result.Result {
	case common.Player1Wins:
		return fmt.Sprintf("%s wins by %s", result.Game.Engines[0].config.Name, result.Reason)
	case common.Player2Wins:
		return fmt.Sprintf("%s wins by %s", result.Game.Engines[1].config.Name, result.Reason)
	case common.Draw:
		return fmt.Sprintf("Draw by %s", result.Reason)
	}

	return "illegal result"
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
