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

	"github.com/sirupsen/logrus"
	"laptudirm.com/x/arbiter/pkg/eve/match"
	"laptudirm.com/x/arbiter/pkg/eve/stats"
	"laptudirm.com/x/arbiter/pkg/eve/tournament/schedule"
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

	tour.games = make(chan *Match)
	tour.results = make(chan Result)
	tour.complete = make(chan bool)

	tour.Scheduler, err = schedule.New(config.Scheduler)
	if err != nil {
		return nil, err
	}

	return &tour, nil
}

type Tournament struct {
	Config Config

	Scheduler schedule.Scheduler
	openings  *Book

	games    chan *Match
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
					tour.games <- &Match{
						Config: match.Config{
							Game:        tour.Config.Game,
							PositionFEN: tour.openings.Current(),
							Engines: [2]match.EngineConfig{
								tour.Config.Engines[p1],
								tour.Config.Engines[p2],
							},
						},

						Round:  round + 1,
						Number: encounter*tour.Scheduler.TotalEncounters() + pair*2 + game + 1,

						Player1: p1,
						Player2: p2,
					}

					// Switch turn.
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

type Match struct {
	match.Config

	Round, Number    int
	Player1, Player2 int
}

func (tour *Tournament) RunGame(game *Match) error {
	logrus.Infof(
		"\x1b[33mStarting\x1b[0m Round #%d Game #%d: %s vs %s (\x1b[33m%s\x1b[0m)\n",
		game.Round,
		game.Number,
		game.Engines[0].Name,
		game.Engines[1].Name,
		tour.openings.Current(),
	)

	score, reason := match.Run(&game.Config)

	tour.results <- Result{
		Match:  game,
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
		case match.Player1Wins:
			tour.Scores[result.Match.Player1].Wins++
			tour.Scores[result.Match.Player2].Losses++

		case match.Player2Wins:
			tour.Scores[result.Match.Player2].Wins++
			tour.Scores[result.Match.Player1].Losses++

		case match.Draw:
			tour.Scores[result.Match.Player1].Draws++
			tour.Scores[result.Match.Player2].Draws++
		}

		logrus.Infof(
			"\x1b[32mFinished\x1b[0m Round #%d Game #%d: %s vs %s: %s\n",
			result.Match.Round,
			result.Match.Number,
			result.Match.Engines[0].Name,
			result.Match.Engines[1].Name,
			result,
		)

		if result_count%5 == 0 {
			tour.Report()
		}

		if result_count == result_target {
			close(tour.results)
			tour.complete <- true
			return
		}
	}

}

func (tour *Tournament) Report() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║    Name               Elo Error   Wins Loss Draw   Total ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	for i, engine := range tour.Config.Engines {
		score := tour.Scores[i]
		lower, elo, upper := stats.Elo(score.Wins, score.Draws, score.Losses)

		format := "║ %2d. %-15s   %+4.0f %4.0f   %4d %4d %4d   %5d ║\n"
		if tour.Config.Scheduler == "gauntlet" && i == 0 {
			if elo >= 0 {
				format = "║ \x1b[32m%2d. %-15s   %+4.0f %4.0f   %4d %4d %4d   %5d\x1b[0m ║\n"
			} else {
				format = "║ \x1b[31m%2d. %-15s   %+4.0f %4.0f   %4d %4d %4d   %5d\x1b[0m ║\n"
			}
		}

		fmt.Printf(
			format,
			i+1, engine.Name,
			elo, math.Abs(math.Max(upper-elo, elo-lower)),
			score.Wins, score.Losses, score.Draws,
			score.Wins+score.Losses+score.Draws)
	}
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}

type Result struct {
	Match *Match

	Result match.Result
	Reason string
}

func (result Result) String() string {
	switch result.Result {
	case match.Player1Wins:
		return fmt.Sprintf("%s wins by %s", result.Match.Engines[0].Name, result.Reason)
	case match.Player2Wins:
		return fmt.Sprintf("%s wins by %s", result.Match.Engines[1].Name, result.Reason)
	case match.Draw:
		return fmt.Sprintf("Draw by %s", result.Reason)
	}

	return "illegal result"
}

type Config struct {
	// The engines participating in the tournament.
	Engines []match.EngineConfig `yaml:"engines"`

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
