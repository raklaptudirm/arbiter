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

package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func Execute(errStr, command string, args ...string) error {
	logrus.Debugf("\x1b[34m%s\x1b[0m %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)

	// Creates pipes for stdout and stderr.
	or, ow, _ := os.Pipe()
	er, ew, _ := os.Pipe()

	cmd.Stdout = ow
	cmd.Stderr = ew

	// Show the commands output if logging level is Trace.
	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Pre-run stuff
	fmt.Print("\x1b[33m") // Make the outputs yellow.
	StartSpinner()        // Start the ~working~ spinner.

	err := cmd.Run()

	// Post-run stuff
	PauseSpinner()       // Stop the ~working~ spinner.
	fmt.Print("\x1b[0m") // Reset the terminal's color.

	// Close the pipes.
	_ = ow.Close()
	_ = ew.Close()

	if err != nil {
		// Dump command's stdout and stderr in case of failure.
		if !logrus.IsLevelEnabled(logrus.TraceLevel) {
			fmt.Print("==== \x1b[31mERROR\x1b[0m ====\n\x1b[31m")
			_, _ = io.Copy(os.Stdout, or)
			_, _ = io.Copy(os.Stderr, er)
			fmt.Print("\x1b[0m===============\n")
		}
		if errStr == "" {
			return err
		}
		return errors.New(errStr)
	}

	return nil
}
