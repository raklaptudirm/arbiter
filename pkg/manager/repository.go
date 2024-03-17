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

package manager

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/pkg/internal/util"
)

func (engine *Engine) FetchRepository() error {
	var err error

	logrus.Info("Fetching the engine's source repository...")
	util.StartSpinner()
	defer util.PauseSpinner()

	logrus.Debug("Fetching tags from repository origin...")

	// Check if we already have a repository for this engine.
	logrus.Debug("Trying to open an existing repository...")
	if engine.Repository, err = git.PlainOpen(engine.Path); err == nil {
		if engine.Worktree, err = engine.Repository.Worktree(); err == nil {
			err = engine.Pull(&git.PullOptions{
				RemoteURL: engine.URL,
			})
			if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
				return nil
			}
		}

		_ = os.RemoveAll(engine.Path)
	}

	util.PauseSpinner()

	logrus.Debug("Trying to clone the engine to a new repository...")
	fmt.Printf("\x1b[33m")
	if engine.Repository, err = git.PlainClone(engine.Path, false, &git.CloneOptions{
		URL:   engine.URL,
		Depth: 1, SingleBranch: true, Tags: git.NoTags,
		Progress: os.Stdout,
	}); err == nil {
		engine.Worktree, err = engine.Repository.Worktree()
	}
	fmt.Printf("\x1b[0m")

	return err
}

func (engine *Engine) FetchVersion(version Version) error {
	name := version.Ref.Name()
	if name.IsTag() {
		logrus.WithField("refspec", name.String()+":"+name.String()).Debug("Fetching required tag")
		fmt.Printf("\x1b[33m")
		err := engine.Repository.Fetch(&git.FetchOptions{
			Depth: 1,
			RefSpecs: []config.RefSpec{
				config.RefSpec("+" + name.String() + ":" + name.String()),
			},
			Progress: os.Stdout,
		})
		fmt.Printf("\x1b[0m")

		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}

	return nil
}
