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
	"errors"
	"strconv"
	"strings"
	"time"

	. "laptudirm.com/x/arbiter/pkg/tournament/common"
	"laptudirm.com/x/arbiter/pkg/tournament/games"
)

func NewGame(engine1Config, engine2Config EngineConfig, position string) (*Game, error) {
	engine1, err := NewEngine(engine1Config)
	if err != nil {
		return nil, err
	}

	engine2, err := NewEngine(engine2Config)
	if err != nil {
		return nil, err
	}

	_, time_1, inc_1, err := parseTime(engine1Config.TimeC)
	if err != nil {
		return nil, err
	}
	_, time_2, inc_2, err := parseTime(engine1Config.TimeC)
	if err != nil {
		return nil, err
	}

	return &Game{
		StartFEN: position,

		Engines: [2]*Player{
			engine1, engine2,
		},

		TotalTime: [2]time.Duration{
			time_1, time_2,
		},

		Increment: [2]time.Duration{
			inc_1, inc_2,
		},
	}, nil
}

// movestogo/time+increment, where time seconds or minutes:seconds
func parseTime(time_str string) (int, time.Duration, time.Duration, error) {
	moves_str, time_str, found := strings.Cut(time_str, "/")
	movestogo := -1
	var err error
	if found {
		movestogo, err = strconv.Atoi(moves_str)
		if err != nil {
			return 0, 0, 0, err
		}
	} else {
		time_str = moves_str
	}

	time_str, inc_str, found := strings.Cut(time_str, "+")
	if !found {
		return 0, 0, 0, errors.New("parse tc: increment not found")
	}

	incs, err := strconv.ParseFloat(inc_str, 32)
	if err != nil {
		return 0, 0, 0, err
	}

	increment := time.Millisecond * time.Duration(incs*1000)

	movetime := time.Duration(0)
	min_str, sec_str, found := strings.Cut(time_str, ":")
	if found {
		mins, err := strconv.ParseFloat(min_str, 32)
		if err != nil {
			return 0, 0, 0, err
		}

		movetime += time.Minute * time.Duration(mins)
	} else {
		sec_str = min_str
	}

	secs, err := strconv.ParseFloat(sec_str, 32)
	if err != nil {
		return 0, 0, 0, err
	}

	movetime += time.Millisecond * time.Duration(secs*1000)
	return movestogo, movetime, increment, nil
}

type Game struct {
	StartFEN string
	Oracle   games.Oracle

	moves string

	moveList []string

	Engines   [2]*Player
	TotalTime [2]time.Duration
	Increment [2]time.Duration
}

func (game *Game) Play() (Score, error) {
	if err := game.Engines[0].NewGame(); err != nil {
		return Player2Wins, err
	}

	if err := game.Engines[1].NewGame(); err != nil {
		return Player1Wins, err
	}

	defer game.Engines[0].Kill()
	defer game.Engines[1].Kill()

	if game.Oracle != nil {
		game.Oracle.Initialize(game.StartFEN)
	}

	sideToMove := 0
	for {
		engine := game.Engines[sideToMove]

		if err := engine.Write("position fen %s moves%s", game.StartFEN, game.moves); err != nil {
			return GameLostBy[sideToMove], err
		}

		if err := engine.Synchronize(); err != nil {
			return GameLostBy[sideToMove], err
		}

		if err := engine.Write(
			"go wtime %d btime %d winc %d binc %d",
			game.TotalTime[0].Milliseconds(),
			game.TotalTime[1].Milliseconds(),
			game.Increment[0].Milliseconds(),
			game.Increment[1].Milliseconds(),
		); err != nil {
			return GameLostBy[sideToMove], err
		}

		startTime := time.Now()
		line, err := engine.Await(
			"bestmove .*",
			game.TotalTime[sideToMove],
		)
		timeSpent := time.Since(startTime)
		game.TotalTime[sideToMove] -= timeSpent
		game.TotalTime[sideToMove] += game.Increment[sideToMove]

		if err != nil {
			return GameLostBy[sideToMove], err
		}

		bestmove := strings.Fields(line)[1]
		game.moves += " " + bestmove

		sideToMove ^= 1

		if game.Oracle != nil {
			err := game.Oracle.MakeMove(bestmove)
			if err != nil {
				return GameLostBy[sideToMove], err
			}

			switch game.Oracle.GameResult() {
			case games.StmWins:
				return Player1Wins - Score(2*sideToMove), nil
			case games.XtmWins:
				return Player2Wins + Score(2*sideToMove), nil
			case games.Draw:
				return Draw, nil
			}

			if game.Oracle.ZeroMoves() {
				game.StartFEN = game.Oracle.FEN()
				game.moves = ""
			}
		}
	}
}

type Result struct {
	Score  Score
	Reason string
}

const (
	DrawByAdjudication         = "Draw by adjudication"
	DrawByFiftyMoveRule        = "Draw by 50 move rule"
	DrawByInsufficientMaterial = "Draw by insufficient material"
	DrawByStalemate            = ""
)
