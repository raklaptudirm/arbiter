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

func NewTournament(config Config) (*Tournament, error) {
	var tour Tournament
	tour.Engines = make([]*Player, len(config.Engines))

	for i, engine := range config.Engines {
		var err error
		tour.Engines[i], err = NewEngine(engine)

		if err != nil {
			return nil, err
		}
	}

	return &tour, nil
}

type Tournament struct {
	Engines             []*Player
	Wins, Draws, Losses int
}

func (tour *Tournament) Start() error {
	return nil
}

type Config struct {
	Engines []EngineConfig

	Concurrency int
	Draw        struct {
		MoveNumber int
		MoveCount  int
		Score      int
	}
	Resign struct {
		MoveCount int
		Score     int
	}

	MaxMoves int

	Event string
	Site  string

	Games int

	Rounds int

	Sprt struct {
		Elo0, Elo1  int
		Alpha, Beta int
	}

	Openings struct {
		File   string
		Format string
		Order  string
		Plies  int
		Start  int
		Policy string
	}

	PGNOut string
	EPDOut string

	Recover bool
}
