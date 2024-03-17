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
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/pkg/internal/util"
)

// ResolveVersion resolves a version string for an Engine into a Version.
// The following formats for the version string are supported:
//
// stable: Resolves to the latest tagged patch of the Engine.
// latest: Resolves to the latest patch of the Engine.
// <name>: Resolves to the patch with the given name.
func (engine *Engine) ResolveVersion(v string) (Version, error) {
	var err error
	var version Version
	switch v {
	case "stable":
		// Find the latest tagged patch of the Engine.
		version.Ref, err = engine.FindStable()

	case "latest":
		// Find the latest patch of the Engine.
		version.Ref, err = engine.FindLatest()

	default:
		// Find the patch corresponding to the given version string.
		version.Ref, err = engine.FindTag(v)
	}

	if err != nil || version.Ref == nil {
		// Print the actual error at DEBUG level, and return a human-readable error instead.
		logrus.Debug(err)
		return version, fmt.Errorf("Unable to find version \x1b[31m%s\x1b[0m", v)
	}

	// Determine the version's name from the reference.
	version.Name = version.Ref.Name().Short()

	return version, nil
}

// Version represents an installable version of an Engine.
type Version struct {
	Name string              // Human-readable name of the version
	Ref  *plumbing.Reference // Git object reference of the version
}

// FindStable finds the reference of the latest tagged patch to the Engine.
func (engine *Engine) FindStable() (*plumbing.Reference, error) {
	logrus.Debug("Looking for the latest stable release...")

	// Find the Engine's remote repository.
	remote, err := engine.Remote(git.DefaultRemoteName)
	if err != nil {
		return nil, err
	}

	// Get a list of git objects from the Engine's remote repository.
	refs, err := remote.List(&git.ListOptions{PeelingOption: git.AppendPeeled})
	if err != nil {
		return nil, err
	}

	var stable *plumbing.Reference
	for _, ref := range refs {
		// Iterate through the objects found in the Engine's remote repository,
		// and find the latest tagged patch among them. Which tag is the latest
		// is determined by the Alphanum sorting algorithm. Due to the fact that
		// an engine's versioning system usually follows an internal format,
		// the Alphanum algorithm is able to correctly identify the latest version.
		// https://web.archive.org/web/20210803201519/http://www.davekoelle.com/alphanum.html
		if ref.Name().IsTag() {
			// Replace the current reference if it or it is an older tag, or it is nil.
			if util.AlphanumCompare(stable.Name().Short(), ref.Name().Short()) || stable == nil {
				stable = ref
			}
		}
	}

	// If we didn't find any tagged patches, fallback to downloading @latest.
	if stable == nil {
		return engine.FindLatest()
	}

	return stable, nil
}

// FindLatest finds the reference of the latest patch to the Engine.
func (engine *Engine) FindLatest() (*plumbing.Reference, error) {
	return engine.Head()
}

// FindTag finds the reference to the patch tagged with the given name in the Engine.
func (engine *Engine) FindTag(tag string) (*plumbing.Reference, error) {
	// Find the Engine's remote repository.
	remote, err := engine.Remote(git.DefaultRemoteName)
	if err != nil {
		return nil, err
	}

	// Get a list of git objects from the Engine's remote repository.
	refs, err := remote.List(&git.ListOptions{PeelingOption: git.AppendPeeled})
	if err != nil {
		return nil, err
	}

	// Iterate through the objects found in the Engine's remote repository,
	// and return the reference to the tag which has the required name.
	for _, ref := range refs {
		if ref.Name().IsTag() && ref.Name().Short() == tag {
			return ref, nil
		}
	}

	// We were unable to find a tag with the required name, return an error.
	return nil, fmt.Errorf("Unable to find version \x1b[31m%s\x1b[0m", tag)
}
