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

package restart

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	arbiter "laptudirm.com/x/arbiter/pkg/common"
	"laptudirm.com/x/arbiter/pkg/eve/sprt"
)

func SPRT() *cobra.Command {
	return &cobra.Command{
		Use:   "sprt test-name",
		Short: "Restart a Sequential Probability Ratio Test",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.ReadFile(filepath.Join(arbiter.Directory, "paused", "sprt", args[0]))
			if err != nil {
				return err
			}

			var config sprt.Config
			err = yaml.Unmarshal(file, &config)
			if err != nil {
				return err
			}

			tour, err := sprt.NewTournament(config)
			if err != nil {
				return err
			}

			return tour.Start()
		},
	}
}
