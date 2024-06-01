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
	"github.com/spf13/cobra"
	"laptudirm.com/x/arbiter/internal/arbiter/cmd/restart"
)

func Restart() *cobra.Command {
	cmd := cobra.Command{
		Use:   "restart",
		Short: "Restart a stopped tournament by name",
	}

	cmd.AddCommand(restart.SPRT())
	cmd.AddCommand(restart.Tournament())
	return &cmd
}
