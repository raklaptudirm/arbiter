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
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/pkg/internal/util"
)

const SPIN = 31

type Repository struct {
	Engine *Engine

	BuildScript string

	Path string
	*git.Repository
	*git.Worktree
}

func NewBareRepository(engine_ident string) (*Repository, error) {
	var err error

	var repo Repository
	if repo.Engine, err = NewEngine(engine_ident); err != nil {
		return nil, err
	}

	repo.Path = filepath.Join(util.SourceDirectory, strings.ToLower(repo.Engine.Name))

	if repo.Engine.Info != nil {
		repo.BuildScript = repo.Engine.Info.BuildScript
	}

	return &repo, nil
}

func (repo *Repository) InstallEngine() error {
	var err error
	logrus.WithField("engine", repo.Engine.Name).Debug("Installing Engine")

	fmt.Printf("\x1b[92mInstalling Player:\x1b[0m %s by %s\n\n", repo.Engine.Name, repo.Engine.Author)

	if repo.Repository == nil {
		logrus.Debug("Fetching repository into a local path...")
		if err := repo.Fetch(); err != nil {
			return err
		}
	}

	logrus.Debug("Getting engine version to install...")

	version, err := repo.NewVersion(repo.Engine.Version)
	if err != nil {
		return err
	}

	if err := repo.Build(version); err != nil {
		return err
	}

	// Check if the binary directory exists, build it if not.
	if _, err := os.Stat(util.BinaryDirectory); errors.Is(err, fs.ErrNotExist) {
		if err := os.Mkdir(util.BinaryDirectory, 0755); err != nil {
			return err
		}
	}

	engine_binary := filepath.Join(util.BinaryDirectory, strings.ToLower(repo.Engine.Name))
	version_binary := engine_binary + "-" + version.Name

	// Move the engine binary to the binary directory.
	if err := os.Rename("engine-binary", version_binary); err != nil {
		return errors.New("Installer \x1b[31mfailed\x1b[0m in building the engine binary")
	}

	// Hardlink the engine binary to the latest installation.
	_ = os.Remove(engine_binary)
	_ = os.Link(version_binary, engine_binary)

	fmt.Printf("\nInstalled engine \x1b[92m%s %s\x1b[0m.\n", repo.Engine.Name, version.Name)
	return nil
}

func (repo *Repository) Build(version Version) error {
	// Reset repository state after stuff has been done.
	head, _ := repo.Head()
	defer util.Ignore(repo.Checkout(&git.CheckoutOptions{Hash: head.Hash()}))
	logrus.WithField("target", head.Hash().String()[0:7]).
		Debug("Repository will be checked back after installation")
	if err := repo.Checkout(&git.CheckoutOptions{
		Hash: version.Hash,
	}); err != nil {
		return err
	}

	// Building the engine is done with the repo root as the current
	// working directory. Any build script can assume that this fact is true.
	// A proper build script will build the engine and put it in ./engine-bin.

	// Reset directory state after stuff has been done.
	current_dir, _ := os.Getwd()
	defer util.Ignore(os.Chdir(current_dir))
	if err := os.Chdir(repo.Path); err != nil {
		return err
	}

	// Some engines registered in arbiter core have custom installation scripts.
	if repo.Engine.Info != nil && repo.BuildScript != "" {
		return script_build(repo.BuildScript)
	}

	return makefile_build()
}

// The default installation pathway. An OpenBench-compliant Makefile is used to
// build the engine at a particular location, from which it is moved to the bin.
func makefile_build() error {
	s := spinner.New(spinner.CharSets[SPIN], 100*time.Millisecond)
	logrus.Info("Trying to build using an \x1b[33mOpenBench-compliant Makefile\x1b[0m...")
	s.Start()
	defer s.Stop()

	src, _ := os.Getwd()

	var makeDir, makeDepth = "", 10_000
	_ = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if strings.EqualFold(filepath.Base(path), "makefile") &&
			strings.Count(path, string(filepath.Separator)) < makeDepth {
			makeDir = filepath.Dir(path)
			makeDepth = strings.Count(path, string(filepath.Separator))
		}
		return nil
	})

	if makeDir == "" {
		return errors.New("Makefile \x1b[31mnot found\x1b[0m in engine's git")
	}

	logrus.WithField("makefile-directory", makeDir).Debug("makefile found in git")

	// Change the working directory to the git containing the makefile.
	if err := os.Chdir(makeDir); err != nil {
		return err
	}

	err := util.Execute(
		"Makefile failed to build the engine binary",
		"make", "-j", "EXE=engine-binary",
	)

	if err != nil {
		return err
	}

	// Move the engine binary to the expected place so that it can be found by the installer.
	if err := os.Rename("engine-binary", filepath.Join(src, "engine-binary")); err != nil {
		return errors.New("Discovered Makefile is \x1b[31mnot Openbench-compliant\x1b[0m.")
	}

	// Change the working directory back to the root of the source code.
	return os.Chdir(src)
}

// Installation using a custom script recorded in arbiter's core engine records.
func script_build(build_script string) error {
	s := spinner.New(spinner.CharSets[SPIN], 100*time.Millisecond)
	logrus.Info("Trying to build using an \x1b[33mIn-built Installation Script\x1b[0m...")
	s.Start()
	defer s.Stop()

	// Pipe the build script into a shell.
	script := exec.Command("sh")
	script.Stdin = strings.NewReader(build_script)

	// Show the commands output if logging level is Trace.
	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		script.Stdout = os.Stdout
		script.Stderr = os.Stderr
	}

	if err := script.Run(); err != nil {
		return errors.New("Build script failed; Check requirements or open an issue")
	}

	return nil
}

func (repo *Repository) Fetch() error {
	var err error

	// If the repo has been cloned previously, just pull any new changes.
	if repo.Repository, err = git.PlainOpen(repo.Path); err == nil {
		logrus.Info("Pulling from the Player's source repo...")
		util.StartSpinner()

		if repo.Worktree, err = repo.Repository.Worktree(); err == nil {
			// Try and pull latest changes to the repo from engine source.
			err := repo.Pull(&git.PullOptions{
				// This option is necessary to ensure some other repository isn't
				// cloned instead of the current engine in its repository directory.
				RemoteURL: repo.Engine.SourceURL,
			})

			util.PauseSpinner()

			// If there are no errors, or the branch is already upto date, return.
			if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
				return nil
			}

			logrus.Debug(err)
		}

		util.PauseSpinner()

		// Fallback to cloning since the current repo is unusable.
		logrus.Error("Pulling repo failed, making a fresh clone")
	}

	// Remove any existing stuff in the path.
	_ = os.RemoveAll(repo.Path)

	// If the repo hasn't been cloned previously or is corrupted, clone it.
	logrus.Info("Fetching the Player's source repo...")

	util.StartSpinner()
	repo.Repository, err = git.PlainClone(repo.Path, false, &git.CloneOptions{
		URL: repo.Engine.SourceURL,
	})
	util.PauseSpinner()

	if err == nil {
		repo.Worktree, err = repo.Repository.Worktree()
	}

	return err
}

//func (repo *Repository) ResolveVersion(version string) (plumbing.Hash, error) {
//	switch version {
//	// Find the latest stable(tagged) version of the engine.
//	case "stable":
//		var stable *plumbing.Reference
//		var stable_date time.Time
//
//		logrus.Debug("Looking for the latest stable release...")
//		tags, err := repo.Tags()
//		if err != nil {
//			goto fallback
//		}
//
//		err = tags.ForEach(func(tag_ref *plumbing.Reference) error {
//			revision := plumbing.Revision(tag_ref.Name().String())
//			commit_hash, err := repo.ResolveRevision(revision)
//			if err != nil {
//				return err
//			}
//
//			commit, err := repo.CommitObject(*commit_hash)
//			if err != nil {
//				return err
//			}
//
//			logrus.WithFields(logrus.Fields{
//				"commit": commit.Hash.String()[0:7], "time": commit.Committer.When,
//			}).Debug("Checking tag for time")
//
//			if stable == nil || commit.Committer.When.After(stable_date) {
//				stable = tag_ref
//				stable_date = commit.Committer.When
//			}
//			return nil
//		})
//
//		if stable != nil && err == nil {
//			return stable.Hash(), err
//		}
//
//		// Some error encountered while looking for latest stable version.
//		// Fallback to finding the latest development version of the engine.
//	fallback:
//		fallthrough
//
//	// Find the latest development version of the engine.
//	case "latest":
//		latest, err := repo.Head()
//		return latest.Hash(), err
//
//	// Find the version corresponding to the given tag.
//	default:
//		tag, err := repo.Tag(repo.Engine.Version)
//		return tag.Hash(), err
//	}
//
//	panic("reached unreachable statement")
//}
