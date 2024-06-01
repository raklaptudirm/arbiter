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

package sprt

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	arbiter "laptudirm.com/x/arbiter/pkg/common"
	"laptudirm.com/x/arbiter/pkg/eve/match"
	"laptudirm.com/x/arbiter/pkg/eve/stats"
)

func NewTournament(config Config) (*SPRT, error) {
	var sprt SPRT
	sprt.Config = config

	var err error
	sprt.openings, err = match.NewBook(config.Openings)
	if err != nil {
		return nil, err
	}

	sprt.results = make(chan PairResult)
	sprt.complete = make(chan bool)

	return &sprt, nil
}

type SPRT struct {
	Config

	openings *match.OpeningBook

	results  chan PairResult
	complete chan bool

	number int
	ended  bool

	a, b float64
}

func (sprt *SPRT) Start() error {
	sprt.a, sprt.b = stats.StoppingBounds(sprt.Config.Alpha, sprt.Config.Beta)

	go sprt.ResultHandler()
	for i := 0; i < sprt.Config.Concurrency; i++ {
		go sprt.Thread()
	}

	<-sprt.complete
	return nil
}

func (sprt *SPRT) Thread() {
	for !sprt.ended {
		sprt.openings.Next()
		opening := sprt.openings.Current()

		var pair PairResult

		p1, p2 := 0, 1
		for game := 0; game < 2; game++ {
			sprt.number++

			match := Match{
				Config: match.Config{
					Game:        sprt.Config.Game,
					PositionFEN: opening,
					Engines: [2]match.EngineConfig{
						sprt.Config.Engines[p1],
						sprt.Config.Engines[p2],
					},
				},

				Number: sprt.number,

				Player1: p1,
				Player2: p2,
			}

			result, err := sprt.RunGame(&match)
			if err != nil {
				logrus.Error(err)
			}

			pair.Matches[game] = result

			p1, p2 = p2, p1
		}

		pair.Result = match.GetPairResult(
			pair.Matches[0].Result,
			pair.Matches[1].Result,
		)

		sprt.results <- pair
	}
}

type Match struct {
	match.Config
	Number int

	Player1, Player2 int
}

func (sprt *SPRT) RunGame(game *Match) (Result, error) {
	logrus.Infof(
		"\x1b[33mStarting\x1b[0m Game #%d: %s vs %s (\x1b[33m%s\x1b[0m)\n",
		game.Number,
		game.Engines[0].Name,
		game.Engines[1].Name,
		sprt.openings.Current(),
	)

	score, reason := match.Run(&game.Config)
	if game.Player2 == 0 {
		score = -score
	}

	return Result{
		Match:  game,
		Result: score,
		Reason: reason,
	}, nil
}

func (sprt *SPRT) ResultHandler() {
	result_count := 0
	for pair := range sprt.results {
		switch pair.Result {
		case match.WinWin:
			sprt.State.WinWin++
		case match.WinDraw:
			sprt.State.WinDraw++
		case match.DrawDraw:
			sprt.State.DrawDraw++
		case match.DrawLoss:
			sprt.State.DrawLoss++
		case match.LossLoss:
			sprt.State.LossLoss++
		}

		result_count++

		for _, result := range pair.Matches {
			switch result.Result {
			case match.Win:
				sprt.State.Wins++
			case match.Loss:
				sprt.State.Losses++
			case match.Draw:
				sprt.State.Draws++
			}

			logrus.Infof(
				"\x1b[32mFinished\x1b[0m Game #%d: %s vs %s: %s\n",
				result.Match.Number,
				result.Match.Engines[0].Name,
				result.Match.Engines[1].Name,
				result,
			)
		}

		if result_count%5 == 0 {
			sprt.Report()
		}

		if llr := sprt.LLR(); llr <= sprt.a {
			fmt.Println("\n\x1b[31mH0 Accepted")
		} else if llr >= sprt.b {
			fmt.Println("\n\x1b[32mH1 Accepted")
		} else {
			continue
		}

		sprt.Report()

		fmt.Print("\x1b[0m")
		close(sprt.results)
		sprt.complete <- true
		return
	}

}

func (sprt *SPRT) Report() {
	data, _ := yaml.Marshal(sprt.Wrap())
	os.WriteFile(filepath.Join(arbiter.Directory, "paused", "sprt", sprt.Name), data, 0777)

	lower, elo, upper := stats.Elo(sprt.State.Wins, sprt.State.Draws, sprt.State.Losses)
	err := math.Abs(math.Max(upper-elo, elo-lower))

	n := sprt.State.Wins + sprt.State.Losses + sprt.State.Draws

	llr := sprt.LLR()

	elo_str := fmt.Sprintf("║ ELO   | %.2f +- %.2f (95%%)", elo, err)
	llr_str := fmt.Sprintf("║ LLR   | %.2f (%.2f, %.2f) [%.2f, %.2f]", llr, sprt.a, sprt.b, sprt.Config.Elo0, sprt.Config.Elo1)
	gam_str := fmt.Sprintf("║ GAMES | N: %d W: %d L: %d D: %d", n, sprt.State.Wins, sprt.State.Losses, sprt.State.Draws)

	fmt.Println("╔═════════════════════════════════════════════════╗")
	fmt.Printf("%-50s║\n", elo_str)
	fmt.Printf("%-50s║\n", llr_str)
	fmt.Printf("%-50s║\n", gam_str)
	if !sprt.Config.Legacy {
		penta_str := fmt.Sprintf(
			"║ PENTA | [%d, %d, %d, %d, %d]",
			sprt.State.LossLoss, sprt.State.DrawLoss,
			sprt.State.DrawDraw,
			sprt.State.WinDraw, sprt.State.WinWin,
		)
		fmt.Printf("%-50s║\n", penta_str)
	}
	fmt.Println("╚═════════════════════════════════════════════════╝")
}

func (sprt *SPRT) LLR() float64 {
	if sprt.Config.Legacy {
		return stats.SPRT(
			sprt.State.Wins,
			sprt.State.Draws,
			sprt.State.Losses,
			sprt.Config.Elo0,
			sprt.Config.Elo1,
		)
	}

	return stats.PentaSPRT(
		sprt.State.LossLoss,
		sprt.State.DrawLoss,
		sprt.State.DrawDraw,
		sprt.State.WinDraw,
		sprt.State.WinWin,
		sprt.Config.Elo0,
		sprt.Config.Elo1,
	)
}

func (sprt *SPRT) Wrap() Config {
	config := sprt.Config
	sprt.Openings = sprt.openings.Wrap()
	return config
}

type PairResult struct {
	Result  match.PairResult
	Matches [2]Result
}

type Result struct {
	Match *Match

	Result match.Result
	Reason string
}

func (result Result) String() string {
	switch result.Result {
	case match.Win:
		return fmt.Sprintf("%s wins by %s", result.Match.Engines[result.Match.Player1].Name, result.Reason)
	case match.Loss:
		return fmt.Sprintf("%s wins by %s", result.Match.Engines[result.Match.Player2].Name, result.Reason)
	case match.Draw:
		return fmt.Sprintf("Draw by %s", result.Reason)
	}

	return "illegal result"
}

type Config struct {
	Name string

	// The engines participating in the tournament.
	Engines [2]match.EngineConfig `yaml:"engines"`

	// The game that will be played.
	Game string `yaml:"game"`

	// Number of games that will be played concurrently.
	Concurrency int `yaml:"concurrency"`

	Legacy bool `yaml:"legacy"`

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

	Elo0, Elo1  float64 // The null and the alternate elo hypotheses.
	Alpha, Beta float64 // Confidence bounds for Error types I and II.

	Openings match.OpeningConfig

	// PGNOut string // File to store the game PGNs at.
	// EPDOut string // File to store the game ends EPD at.

	State struct {
		Wins, Losses, Draws                           int
		WinWin, WinDraw, DrawDraw, DrawLoss, LossLoss int
	}
}
