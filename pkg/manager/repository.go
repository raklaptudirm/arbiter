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

// FetchRepository fetches the given remote repository into the given path. If
// the repository was previously cloned, it tries to get it up to date with the
// remote repository.
func FetchRepository(url, path string) (*git.Repository, error) {
	logrus.Info("Fetching the engine's source repository...")
	util.StartSpinner()
	defer util.PauseSpinner()

	// Check if we already have a repository for this engine.
	logrus.Debug("Trying to open an existing repository...")
	if repository, err := git.PlainOpen(path); err == nil {
		if worktree, err := repository.Worktree(); err == nil {
			// Pull any changes made to the remote repository.
			err = worktree.Pull(&git.PullOptions{
				RemoteURL: url,
			})
			if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
				return repository, nil
			}
		}

		_ = os.RemoveAll(path)
	}

	util.PauseSpinner()

	// Repository wasn't previously cloned or is corrupt, so clone from scratch.

	logrus.Debug("Trying to clone the engine to a new repository...")
	fmt.Printf("\x1b[33m")
	defer fmt.Printf("\x1b[0m")
	return git.PlainClone(path, false, &git.CloneOptions{
		URL:   url,
		Depth: 1, SingleBranch: true, Tags: git.NoTags,
		Progress: os.Stdout,
	})
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
