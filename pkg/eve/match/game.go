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

package match

import (
	"strings"
	"time"

	"laptudirm.com/x/arbiter/pkg/eve/match/games"
)

type Config struct {
	Game, PositionFEN string

	Engines [2]EngineConfig
}

func Run(config *Config) (Result, string) {
	engines := [2]*Engine{}
	remaining_time := [2]TimeControl{}

	var err error

	if remaining_time[0], err = ParseTime(config.Engines[0].TimeC); err != nil {
		return Loss, err.Error()
	}

	if remaining_time[1], err = ParseTime(config.Engines[1].TimeC); err != nil {
		return Win, err.Error()
	}

	if engines[0], err = StartEngine(config.Engines[0]); err != nil {
		return Loss, err.Error()
	}

	if engines[1], err = StartEngine(config.Engines[1]); err != nil {
		return Win, err.Error()
	}

	defer engines[0].Kill()
	defer engines[1].Kill()

	oracle := games.GetOracle(config.Game)
	if oracle != nil {
		oracle.Initialize(config.PositionFEN)
	}

	moves := ""
	// : EngineIndex
	whiteEngine := uint8(oracle.SideToMove())
	engineToMove := 0
	for {
		engine := engines[engineToMove]

		if err := engine.Write("position fen %s moves%s", config.PositionFEN, moves); err != nil {
			return GameLostBy[engineToMove], err.Error()
		}

		if err := engine.Synchronize(); err != nil {
			return GameLostBy[engineToMove], err.Error()
		}

		if err := engine.Write(
			"go wtime %d btime %d winc %d binc %d",
			remaining_time[whiteEngine].Base.Milliseconds(),
			remaining_time[1^whiteEngine].Base.Milliseconds(),
			remaining_time[whiteEngine].Inc.Milliseconds(),
			remaining_time[1^whiteEngine].Inc.Milliseconds(),
		); err != nil {
			return GameLostBy[engineToMove], err.Error()
		}

		startTime := time.Now()
		line, err := engine.Await(
			"bestmove .*",
			remaining_time[engineToMove].Base,
		)
		timeSpent := time.Since(startTime)
		remaining_time[engineToMove].Base -= timeSpent
		remaining_time[engineToMove].Base += remaining_time[engineToMove].Inc

		if err != nil {
			return GameLostBy[engineToMove], err.Error()
		}

		bestmove := strings.Fields(line)[1]
		moves += " " + bestmove

		engineToMove ^= 1

		if oracle != nil {
			err := oracle.MakeMove(bestmove)
			if err != nil {
				return GameLostBy[engineToMove], err.Error()
			}

			result, reason := oracle.GameResult()
			switch result {
			case games.StmWins:
				return Win - Result(2*engineToMove), reason
			case games.XtmWins:
				return Loss + Result(2*engineToMove), reason
			case games.Draw:
				return Draw, reason
			}

			if oracle.ZeroMoves() {
				config.PositionFEN = oracle.FEN()
				moves = ""
			}
		}
	}
}
