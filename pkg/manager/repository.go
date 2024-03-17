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
	"github.com/go-git/go-git/v5/config"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/pkg/common"
	"laptudirm.com/x/arbiter/pkg/internal/util"
)

const SPIN = 31

func (engine *Engine) InstallEngine(version Version) error {
	if err := engine.Build(version); err != nil {
		return err
	}

	reinstall := engine.Installed(version)

	engine_binary := filepath.Join(arbiter.BinaryDirectory, strings.ToLower(engine.Name))
	version_binary := engine_binary + "-" + version.Name

	// Move the engine binary to the binary directory.
	if err := os.Rename("engine-binary", version_binary); err != nil {
		return errors.New("Installer \x1b[31mfailed\x1b[0m in building the engine binary")
	}

	// Register the engine and the new version if they previously weren't.
	if !reinstall {
		arbiter.Engines.TryAddEngine(engine.Name, engine.Author, engine.SourceURL)
		arbiter.Engines.InstallVersion(engine.Name, version.Name)
	}

	fmt.Printf("\nInstalled engine \x1b[92m%s %s\x1b[0m.\n", engine.Name, version.Name)
	return nil
}

func (engine *Engine) Build(version Version) error {
	if err := engine.Fetch(version); err != nil {
		return err
	}

	// Reset repository state after stuff has been done.
	head, _ := engine.Head()
	defer util.Ignore(engine.Checkout(&git.CheckoutOptions{Hash: head.Hash()}))
	logrus.WithField("target", head.Hash().String()[0:7]).
		Debug("Repository will be checked back after installation")
	if err := engine.Checkout(&git.CheckoutOptions{
		Hash: version.Ref.Hash(),
	}); err != nil {
		return err
	}

	// Building the engine is done with the engine root as the current
	// working directory. Any build script can assume that this fact is true.
	// A proper build script will build the engine and put it in ./engine-bin.

	// Reset directory state after stuff has been done.
	current_dir, _ := os.Getwd()
	defer util.Ignore(os.Chdir(current_dir))
	if err := os.Chdir(engine.Path); err != nil {
		return err
	}

	// Some engines registered in arbiter core have custom installation scripts.
	if engine.Info != nil && engine.Info.BuildScript != "" {
		return script_build(engine.Info.BuildScript)
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

func (engine *Engine) EfficientFetch() error {
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
				RemoteURL: engine.SourceURL,
			})
			if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
				return nil
			}
		}

		_ = os.RemoveAll(engine.Path)
	}

	logrus.Debug("Trying to clone the engine to a new repository...")
	if engine.Repository, err = git.PlainClone(engine.Path, false, &git.CloneOptions{
		URL:   engine.SourceURL,
		Depth: 1, SingleBranch: true, Tags: git.NoTags,
		Progress: os.Stdout,
	}); err == nil {
		engine.Worktree, err = engine.Repository.Worktree()
	}

	return err
}

func (engine *Engine) Fetch(version Version) error {
	name := version.Ref.Name()
	if name.IsTag() {
		logrus.WithField("refspec", name.String()+":"+name.String()).Debug("Fetching required tag")
		err := engine.Repository.Fetch(&git.FetchOptions{
			Depth: 1,
			RefSpecs: []config.RefSpec{
				config.RefSpec("+" + name.String() + ":" + name.String()),
			},
			Progress: os.Stdout,
		})

		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}

	return nil
}
