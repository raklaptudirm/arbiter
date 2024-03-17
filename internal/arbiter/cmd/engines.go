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

	"github.com/spf13/cobra"

	arbiter "laptudirm.com/x/arbiter/pkg/manager"
)

func Engines() *cobra.Command {
	return &cobra.Command{
		Use:   "engines",
		Short: "Lists the installed engines and their versions",
		Args:  cobra.ExactArgs(0),

		RunE: func(cmd *cobra.Command, args []string) error {
			found_engine := false

			for engine, info := range arbiter.Engines {
				if len(info.Versions) == 0 {
					continue
				}

				if !found_engine {
					found_engine = true
					fmt.Println("\u001B[32mInstalled Engines\u001B[0m:\n")
				}

				versions := ""
				for _, version := range info.Versions {
					if version == info.Current {
						versions = "\x1b[33m" + version + "\x1b[0m " + versions
					} else {
						versions += version + " "
					}
				}

				name := fmt.Sprintf("\x1b[34m%s\x1b[0m:", engine)

				fmt.Printf("- %-20s %s\n", name, versions)
			}

			if !found_engine {
				fmt.Println("\x1b[31mNo Engines Downloaded.\x1b[0m")
				return nil
			}

			return nil
		},
	}
}
