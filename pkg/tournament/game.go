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
	"strings"
	"time"

	. "laptudirm.com/x/arbiter/pkg/tournament/common"
	"laptudirm.com/x/arbiter/pkg/tournament/games"
)

func NewGame(engine1Config, engine2Config EngineConfig, position [6]string) (*Game, error) {
	engine1, err := NewEngine(engine1Config)
	if err != nil {
		return nil, err
	}

	engine2, err := NewEngine(engine2Config)
	if err != nil {
		return nil, err
	}

	return &Game{
		StartFEN: position[0] + " " +
			position[1] + " " +
			position[2] + " " +
			position[3] + " " +
			position[4] + " " +
			position[5] + " ",

		Engines: [2]*Player{
			engine1, engine2,
		},
	}, nil
}

//func parseTime(time string) (int, time.Duration, time.Duration, error) {
//	movesStr, time, found := strings.Cut(time, "/")
//	var moveN int
//	var err error
//	if found {
//		moveN, err = strconv.Atoi(movesStr)
//		if err != nil {
//			return 0, 0, 0, err
//		}
//	}
//
//	perMove, increment, found := strings.Cut(time, "+")
//	if !found {
//		return 0, 0, 0, errors.New("parse tc: increment not found")
//	}
//
//	increment, err := strconv.Atoi(increment)
//}

type Game struct {
	StartFEN  string
	GameEndFn games.GameEndedFn

	moves string

	moveList []string

	Engines   [2]*Player
	TotalTime [2]time.Duration
	Increment [2]time.Duration
}

func (game *Game) Play() (Score, error) {
	fmt.Println("debug: initializing white")
	if err := game.Engines[0].NewGame(); err != nil {
		return BlackWins, err
	}

	fmt.Println("debug: initializing black")
	if err := game.Engines[1].NewGame(); err != nil {
		return WhiteWins, err
	}

	sideToMove := 0
	for {
		if game.GameEndFn != nil {
			ended, result := game.GameEndFn(game.StartFEN, strings.Fields(game.moves))
			if ended {
				return result, nil
			}
		}

		engine := game.Engines[sideToMove]

		if err := engine.Write("position fen %s moves %s", game.StartFEN, game.moves); err != nil {
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
		if err != nil {
			return GameLostBy[sideToMove], err
		}

		bestmove := strings.Fields(line)[1]

		game.TotalTime[sideToMove] -= timeSpent
		game.TotalTime[sideToMove] += game.Increment[sideToMove]

		game.moves += " " + bestmove

		sideToMove ^= 1
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
