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

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/pkg/internal/util"
)

// Download fetches and builds the given version of the engine, and then moves it
// to the ARBITER_BINARY_DIR under the binary name <engine-name>-<version-name>.
func (engine *Engine) Download(version Version) error {
	binary := VersionBinary(engine, version)    // Name of version's binary file
	new_version := !Downloaded(engine, version) // Is the version a new download ?

	// Build the given version of the engine and move the file to binary.
	if err := engine.Build(version, binary); err != nil {
		return err
	}

	// Check if the Engine's binary was successfully built and moved.
	if _, err := os.Stat(binary); err != nil {
		return errors.New("Installer \x1b[31mfailed\x1b[0m in building the engine binary")
	}

	// Register the version with the manager if it is new.
	if new_version {
		Engines.AddVersion(engine, version.Name)
	}

	fmt.Printf("\nInstalled engine \x1b[92m%s %s\x1b[0m.\n", engine.Name, version.Name)
	return nil
}

// Build builds the binary of the given Version of the Engine and move it to dst.
func (engine *Engine) Build(version Version, dst string) error {
	// Reset repository state after building has been done.
	head, _ := engine.Head()
	defer func() {
		logrus.Debugf("Checking out to HEAD or %s", head.Name().Short())
		if err := engine.Checkout(&git.CheckoutOptions{
			Branch: head.Name(),
		}); err != nil {
			logrus.Error(err)
		}
	}()

	// Debug Logging: Log the reference to checkout back to.
	logrus.WithField("target", head.Name().Short()).
		Debug("Repository will be checked back after installation")

	// Fetch the git objects associated with the given version,
	// and checkout to its patch in preparation for building.
	if err := engine.FetchVersion(version); err != nil {
		return err
	}
	if err := engine.Checkout(&git.CheckoutOptions{
		// Checkout to a detached-HEAD.
		Hash: version.Ref.Hash(),
	}); err != nil {
		return err
	}

	// Some Engines registered in with the manager have custom installation scripts.
	// If a custom build script is available, use that to build the Engine's binary.
	if engine.Info != nil && engine.Info.BuildScript != "" {
		return script_build(engine.Path, dst, engine.Info.BuildScript)
	}

	// The default build method is to use an OpenBench-compliant Makefile.
	return makefile_build(engine.Path, dst)
}

// The default installation pathway. An OpenBench-compliant Makefile is used to
// build the Engine at a particular location, from which it is moved to the bin.
func makefile_build(src, dst string) error {
	logrus.Info("Trying to build using an \x1b[33mOpenBench-compliant Makefile\x1b[0m...")

	// Start the spinner in preparation for work :)
	util.StartSpinner()
	defer util.PauseSpinner()

	// Find the shallowest Makefile in the Engine's repository.
	var makefile_dir, makefile_depth = "", 10_000
	_ = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		// Makefile names are case-insensitive. If a new Makefile is found
		// check if it is shallower than the previous Makefile and then replace.
		if strings.EqualFold(filepath.Base(path), "makefile") &&
			// Shallowness of a Makefile is determined by the number of path
			// elements in its file path, separated by filepath.Separator.
			strings.Count(path, string(filepath.Separator)) < makefile_depth {
			makefile_dir = filepath.Dir(path)
			makefile_depth = strings.Count(path, string(filepath.Separator))
		}
		return nil
	})

	// No Makefile was found in the Engine's source.
	if makefile_dir == "" {
		return errors.New("Makefile \x1b[31mnot found\x1b[0m in engine's git")
	}

	// Debug Logging: Log the discovered Makefile's directory.
	logrus.WithField("makefile-directory", makefile_dir).Debug("makefile found in git")

	// Run the command necessary to build a binary with an OpenBench-compliant Makefile.
	// make -j EXE=engine-binary # The binary will be moved from here later.
	err := util.Execute(
		makefile_dir, // Run the command in the makefile's directory.
		"Makefile failed to build the engine binary",
		"make", "-j", "EXE=engine-binary",
	)

	if err != nil {
		return err
	}

	// Move the engine binary to the destination provided by the caller.
	if err := os.Rename(filepath.Join(makefile_dir, "engine-binary"), dst); err != nil {
		// Log the error as DEBUG level and return a human-readable error message.
		logrus.Debug(err)
		return errors.New("Discovered Makefile is \x1b[31mnot Openbench-compliant\x1b[0m.")
	}

	return nil
}

// Installation using a custom script recorded in arbiter's core engine records.
func script_build(src, dst, build_script string) error {
	logrus.Info("Trying to build using an \x1b[33mIn-built Installation Script\x1b[0m...")

	// Start the spinner in preparation for work (:
	util.StartSpinner()
	defer util.PauseSpinner()

	// Shell command to run the build-script.
	// TODO: Make this standard Windows compatible.
	script := exec.Command("sh")

	// Run the script in the source directory of the Engine.
	script.Dir = src

	// Pipe the build-script into the shell command.
	script.Stdin = strings.NewReader(build_script)

	// Show the command's output if logging level is Trace.
	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		script.Stdout = os.Stdout
		script.Stderr = os.Stderr
	}

	if err := script.Run(); err != nil {
		return errors.New("Build script failed; Check requirements or open an issue")
	}

	// Move the engine binary to the destination provided by the caller.
	if err := os.Rename(filepath.Join(src, "engine-binary"), dst); err != nil {
		// Log the error as DEBUG level and return a human-readable error message.
		return errors.New("Build script failed; Check requirements or open an issue")
	}

	return nil
}
