// Copyright © 2024 Rak Laptudirm <rak@laptudirm.com>
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

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"laptudirm.com/x/arbiter/pkg/tournament"
	"laptudirm.com/x/mess/pkg/board"
	"laptudirm.com/x/mess/pkg/board/piece"
)

func Tournament() *cobra.Command {
	return &cobra.Command{
		Use:   "tournament details-file",
		Short: "Run a tournament with different engines",
		Args:  cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			engine1, err := tournament.NewEngine(tournament.EngineConfig{
				Name:   "Mess1",
				Cmd:    "./engines/mess",
				Logger: os.Stdout,
			})

			if err != nil {
				os.Exit(1)
			}

			engine2, err := tournament.NewEngine(tournament.EngineConfig{
				Name:   "Mess2",
				Cmd:    "./engines/mess",
				Logger: os.Stdout,
			})

			if err != nil {
				os.Exit(1)
			}

			game := tournament.Game{
				Board: board.New(board.FEN(board.StartFEN)),

				Engines: [piece.ColorN]*tournament.Player{
					engine1, engine2,
				},

				TotalTime: [piece.ColorN]time.Duration{
					8 * time.Second, 8 * time.Second,
				},
			}

			fmt.Println("debug: starting game")
			fmt.Println(game.Play())

			return nil
		},
	}
}
