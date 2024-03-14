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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/internal/util"
	"laptudirm.com/x/arbiter/pkg/data"
)

type Identifier struct {
	Name      string
	SourceURL string
	Version   string
	LocalPath string

	IsCore bool
}

func NewIdentifier(engine string) *Identifier {
	// <repository-identifier>[@<version>]
	source, version, found := strings.Cut(engine, "@")
	var identifier Identifier

	identifier.Name = filepath.Base(source)
	identifier.LocalPath = filepath.Join(util.SourceDirectory, strings.ToLower(identifier.Name))
	identifier.Version = version
	if !found {
		// By-default try to install the latest stable release.
		identifier.Version = "stable"
	}

	switch strings.Count(source, "/") {
	case 0:
		// Arbiter-core Engine: <engine-name>
		identifier.SourceURL = data.Engines[source].Source
		identifier.IsCore = true
	case 1:
		// Github Engine: <owner>/<engine-name>
		identifier.SourceURL = "https://github.com/" + source
	default:
		// Git Repository Engine: <full-repository-url>
		identifier.SourceURL = source
	}

	return &identifier
}

func Fetch(engine *Identifier) (*git.Repository, error) {
	s := spinner.New(spinner.CharSets[SPIN], 100*time.Millisecond)
	defer s.Stop()

	// If the repository has been cloned previously, just pull any new changes.
	if r, err := git.PlainOpen(engine.LocalPath); err == nil {
		logrus.Info("Pulling from the Engine's source repository...")
		s.Start()
		if w, err := r.Worktree(); err == nil {
			// Try and pull latest changes to the repository.
			err := w.Pull(&git.PullOptions{RemoteURL: engine.SourceURL})
			// If there are no errors, or the branch is already upto date, return.
			if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
				return r, nil
			}

			logrus.Debug(err)
		}

		s.Stop()

		// Fallback to cloning since the current repository is unusable.
		logrus.Error("Pulling repository failed, making a fresh clone")
		_ = os.RemoveAll(engine.LocalPath) // Remove the repository
	}

	s.Start()

	// If the repository hasn't been cloned previously or is corrupted, clone it.
	logrus.Info("Fetching the Engine's source repository...")
	return git.PlainClone(engine.LocalPath, false, &git.CloneOptions{URL: engine.SourceURL})
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

	s := spinner.New(spinner.CharSets[SPIN], 100*time.Millisecond)

	// Pre-run stuff
	fmt.Print("\x1b[33m") // Make the outputs yellow.
	s.Start()             // Start the ~working~ spinner.

	err := cmd.Run()

	// Post-run stufff
	s.Stop()             // Stop the ~working~ spinner.
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
