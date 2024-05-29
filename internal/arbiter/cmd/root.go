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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	root := &cobra.Command{
		Use:  "arbiter",
		Args: cobra.NoArgs,

		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// If --trace flag is provided, set logging level to Trace.
			if cmd.Flag("trace").Changed {
				logrus.SetLevel(logrus.TraceLevel)
			}
		},
	}

	// global flags
	root.PersistentFlags().BoolP("help", "h", false, "Show Help Information")
	root.PersistentFlags().BoolP("version", "v", false, "Show Arbiter's Version")
	root.PersistentFlags().BoolP("trace", "t", false, "Show Trace Information")

	// TODO: properly set version
	versionStr := "v0.0.0\n"
	root.SetVersionTemplate(versionStr)
	root.Version = versionStr

	// Register the various commands.
	root.AddCommand(Engines())
	root.AddCommand(Install())
	root.AddCommand(Completion())
	root.AddCommand(Remove())

	return root
}
