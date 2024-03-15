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
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/pkg/common"
)

type Engine struct {
	Name    string
	Author  string
	Version string

	SourceURL string

	Info *arbiter.EngineInfo
}

func NewEngine(ident string) (*Engine, error) {
	// <git-engine>[@<version>]
	source, version, found := strings.Cut(ident, "@")
	var engine Engine

	engine.Name = filepath.Base(source)
	engine.Version = version
	if !found {
		// By-default try to install the latest stable release.
		engine.Version = "stable"
	}

	switch strings.Count(source, "/") {
	case 0:
		// Arbiter-core Player: <engine-name>
		if info, found := arbiter.Engines[source]; found {
			engine.Info = &info
			engine.Author = info.Author
			engine.SourceURL = info.Source
		} else {
			return nil, fmt.Errorf("Engine %s not found in arbiter dataset", engine.Name)
		}

	case 1:
		// Github Player: <owner>/<engine-name>
		engine.SourceURL = "https://github.com/" + source
		engine.Author, _, _ = strings.Cut(source, "/")

	default:
		// Git Repository Player: <full-git-url>
		engine.SourceURL = source
		engine.Author = "Anonymous"
	}

	logrus.WithFields(logrus.Fields{
		"name":    engine.Name,
		"author":  engine.Author,
		"version": engine.Version,
		"source":  engine.SourceURL,
	}).Debug("Created new manager.Engine")

	return &engine, nil
}

type Version struct {
	Name string
	Hash plumbing.Hash
}

func (repo *Repository) NewVersion(v string) (Version, error) {
	var version Version
	switch v {
	// Find the latest stable(tagged) version of the engine.
	case "stable":
		stable, err := repo.Stable()
		if err != nil {
			return version, err
		}

		version.Name = stable.Name().Short()
		version.Hash = stable.Hash()

	// Find the latest development version of the engine.
	case "latest":
		latest, err := repo.Head()
		if err != nil {
			return version, errors.New("Unable to find version \x1b[31mstable\x1b[0m")
		}

		version.Name = latest.Hash().String()[0:7]
		version.Hash = latest.Hash()

	// Find the version corresponding to the given tag.
	default:
		tag, err := repo.Tag(v)
		if err != nil {
			return version, fmt.Errorf("Unable to find version \x1b[31m%s\x1b[0m", v)
		}

		version.Name = tag.Name().Short()
		version.Hash = tag.Hash()
	}

	return version, nil
}

func (repo *Repository) Stable() (*plumbing.Reference, error) {
	var stable *plumbing.Reference
	var stable_date time.Time

	logrus.Debug("Looking for the latest stable release...")
	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	err = tags.ForEach(func(tag_ref *plumbing.Reference) error {
		revision := plumbing.Revision(tag_ref.Name().String())
		commit_hash, err := repo.ResolveRevision(revision)
		if err != nil {
			return err
		}

		commit, err := repo.CommitObject(*commit_hash)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"commit": commit.Hash.String()[0:7], "time": commit.Committer.When,
		}).Debug("Checking tag for time")

		if stable == nil || commit.Committer.When.After(stable_date) {
			stable = tag_ref
			stable_date = commit.Committer.When
		}
		return nil
	})

	if stable == nil || err != nil {
		return nil, errors.New("Unable to find version \x1b[31mstable\x1b[0m")
	}

	return stable, nil
}

func (engine *Engine) Binary() string {
	return filepath.Join(arbiter.BinaryDirectory, engine.Name)
}

func (engine *Engine) VersionBinary(version Version) string {
	return engine.Binary() + "-" + version.Name
}

func (engine *Engine) Installed(version Version) bool {
	_, err := os.Stat(engine.VersionBinary(version))
	return err == nil
}
