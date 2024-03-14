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
	"io/fs"
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

const SPIN = 31

// IDENTIFIER:
// (1) core-engine[@version]
// (2) owner/github-engine[@version]
// (3) <full-repo-url>[@version]

func Install(engine *Identifier) error {
	logrus.WithField("engine", engine).Debug("Installing Engine")

	info := data.Engines[engine.Name]
	if engine.IsCore {
		fmt.Printf("\x1b[92mInstalling Engine:\x1b[0m %s by %s\n\n", engine.Name, info.Author)
	} else {
		fmt.Printf("\x1b[92mInstalling Engine:\x1b[0m %s\n\n", engine.Name)
	}

	// Fetch the engine repository.
	repository, err := Fetch(engine)
	if err != nil {
		return err
	}

	worktree, _ := repository.Worktree()

	// Reset repository state after stuff has been done.
	head, _ := repository.Head()
	defer worktree.Checkout(&git.CheckoutOptions{Hash: head.Hash()})
	logrus.WithField("old-head", head.Hash().String()).
		Debug("Fetched current head for rollback")

	// Version string to display.
	var install_tag string

	// Figure out which version to install and checkout to that tag.
	switch engine.Version {
	case "stable": // Install latest stable (tagged) version.
		// TODO: Add stable tag support
		fallthrough
	case "latest": // Install latest development version.
		install_tag = head.Hash().String()[:6]

	default: // Install the given version.
		tag, err := repository.Tag(engine.Version)
		if err != nil {
			return fmt.Errorf("Unable to find version \x1b[31m%s\x1b[0m", engine.Version)
		}

		// Checkout to the provided version.
		_ = worktree.Checkout(&git.CheckoutOptions{Hash: tag.Hash()})
		install_tag = engine.Version
	}

	// Building the engine is done with the repository root as the current
	// working directory. Any build script can assume that this fact is true.
	// A proper build script will build the engine and put it in ./engine-bin.
	if err := os.Chdir(engine.LocalPath); err != nil {
		return err
	}

	if engine.IsCore && info.BuildScript != "" {
		// Some engines registered in arbiter core have custom installation scripts.
		if err := script_install(info.BuildScript); err != nil {
			return errors.New("Build script failed; Check requirements or open an issue")
		}
	} else {
		// Default is OpenBench-compliant makefile installation.
		if err := makefile_install(); err != nil {
			return err
		}
	}

	// Check if the binary directory exists, build it if not.
	if _, err := os.Stat(util.BinaryDirectory); errors.Is(err, fs.ErrNotExist) {
		if err := os.Mkdir(util.BinaryDirectory, 0777); err != nil {
			return err
		}
	}

	// Move the engine binary to the binary directory.
	engine_binary := filepath.Join(util.BinaryDirectory, strings.ToLower(engine.Name))
	if err := os.Rename("engine-binary", engine_binary); err != nil {
		return errors.New("Installer \x1b[31mfailed\x1b[0m in building the engine binary")
	}

	fmt.Printf("\nInstalled engine \x1b[92m%s@%s\x1b[0m.\n", engine.Name, install_tag)
	return nil
}

// The default installation pathway. An OpenBench-compliant Makefile is used to
// build the engine at a particular location, from which it is moved to the bin.
func makefile_install() error {
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
		return errors.New("Makefile \x1b[31mnot found\x1b[0m in engine's repository")
	}

	logrus.WithField("makefile-directory", makeDir).Debug("makefile found in repository")

	// Change the working directory to the repository containing the makefile.
	if err := os.Chdir(makeDir); err != nil {
		return err
	}

	err := execute(
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
func script_install(build_script string) error {
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

	return script.Run()
}
