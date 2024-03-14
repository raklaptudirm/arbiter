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

package arbiter

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/internal/util"
	"laptudirm.com/x/arbiter/pkg/data"
)

type Identifier struct {
	Name    string
	Source  string
	Version string
	Path    string

	IsCore bool
}

func NewIdentifier(engine string) *Identifier {
	// <repository-identifier>[@<version>]
	source, version, found := strings.Cut(engine, "@")
	var identifier Identifier

	identifier.Name = filepath.Base(source)
	identifier.Path = filepath.Join(util.SourceDirectory, strings.ToLower(identifier.Name))
	identifier.Version = version
	if !found {
		// By-default try to install the latest stable release.
		identifier.Version = "stable"
	}

	switch strings.Count(source, "/") {
	case 0:
		// Arbiter-core Engine: <engine-name>
		identifier.Source = data.Engines[source].Source
		identifier.IsCore = true
	case 1:
		// Github Engine: <owner>/<engine-name>
		identifier.Source = "https://github.com/" + source
	default:
		// Git Repository Engine: <full-repository-url>
		identifier.Source = source
	}

	return &identifier
}

func Fetch(engine *Identifier) error {
	// If the repository has been cloned previously, just pull any new changes.
	if _, err := os.Stat(engine.Path); !errors.Is(err, fs.ErrNotExist) {
		return execute_info(
			"Pulling latest changes to the Engine's source repository...",
			"Error encountered while pulling repository",
			"git", "-C", engine.Path, "pull",
		)
	}

	// If the repository hasn't been cloned before, clone it into the machine.
	return execute_info(
		"Fetching the Engine's source repository...",
		"Error encountered while fetching repository",
		"git", "clone", engine.Source, engine.Path,
	)
}

func execute_info(info, errStr, command string, args ...string) error {
	s := spinner.New(spinner.CharSets[SPIN], 100*time.Millisecond)
	logrus.Info(info)
	s.Start()
	defer s.Stop()
	return execute(errStr, command, args...)
}

func execute(errStr, command string, args ...string) error {
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

	fmt.Print("\x1b[33m")
	err := cmd.Run()
	fmt.Print("\x1b[0m")

	// Close the pipes.
	_ = ow.Close()
	_ = ew.Close()

	if err != nil {
		// Dump command's stdout and stderr in case of failure.
		if !logrus.IsLevelEnabled(logrus.TraceLevel) {
			fmt.Print("\x1b[31m")
			_, _ = io.Copy(os.Stdout, or)
			_, _ = io.Copy(os.Stderr, er)
			fmt.Print("\x1b[0m")
		}
		if errStr == "" {
			return err
		}
		return errors.New(errStr)
	}

	return nil
}

func output(command string, args ...string) (string, error) {
	ba, err := exec.Command(command, args...).Output()
	return strings.Trim(string(ba), " \t\n\r"), err
}
