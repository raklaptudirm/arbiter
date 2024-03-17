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
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

// Engine represents one of the game-engines managed by arbiter/manager.
// It contains metadata about the game-engine and its source repository.
type Engine struct {
	// Basic Information
	Name   string
	Author string
	Info   *EngineInfo

	// Source Repository Information
	URL  string // URL of the engine's remote repository
	Path string // Path to the engine's local repository
	*git.Repository
	*git.Worktree
}

// NewEngine creates an instance of *manager.Engine from the given engine
// identifier string. The identifier has to have one of the following formats:
//
// 1. <engine-name>                 - Core Engine Format
// 2. <engine-author>/<engine-name> - GitHub Engine Format
// 3. <full-source-git-url>         - Git Engine Format
//
// Only engines whose configuration are present in arbiter by default or have
// been previously installed can be identified using the format (1).
//
// Only engines whose repositories are hosted on GitHub can be identified by
// (2). github.com/<engine-author>/<engine-name> has to be the engine source.
func NewEngine(identifier string) (*Engine, error) {
	var engine Engine

	// In all formats, the engine name is the last part of the identifier:
	// [<stuff-depending-on-the-particular-format>/]<engine-name>
	engine.Name = filepath.Base(identifier)

	// The engine's repository will be stored at ARBITER_SOURCE/<engine-name>.
	engine.Path = filepath.Join(SourceDirectory, strings.ToLower(engine.Name))

	// The formats can be differentiated between using the number of '/' in
	// the identifier. (1) has 0, (2) has 1, and (3) has >= 2 '/'s.
	switch strings.Count(identifier, "/") {
	case 0:
		// Format (1): <engine-name>
		// The engine has to be found in the configuration.
		if info, found := Engines[identifier]; found {
			engine.Info = &info
			engine.URL = info.Source
			engine.Author = info.Author
		} else {
			return nil, fmt.Errorf("Engine %s not found in arbiter dataset", engine.Name)
		}

	case 1:
		// Format (2): <engine-author>/<engine-name>
		engine.URL = "https://github.com/" + identifier
		engine.Author, _, _ = strings.Cut(identifier, "/")

	default:
		// Format (3): <full-source-git-url>
		engine.URL = identifier
		engine.Author = filepath.Base(filepath.Dir(identifier))
	}

	// Debug logging: Engine's Details
	logrus.WithFields(logrus.Fields{
		"name":       engine.Name,
		"author":     engine.Author,
		"identifier": engine.URL,
	}).Debug("Figured out basic engine details")

	return &engine, nil
}

func (engine *Engine) Binary() string {
	return filepath.Join(BinaryDirectory, engine.Name)
}

func (engine *Engine) VersionBinary(version Version) string {
	return engine.Binary() + "-" + version.Name
}

func (engine *Engine) Downloaded(version Version) bool {
	_, err := os.Stat(engine.VersionBinary(version))
	return err == nil
}
