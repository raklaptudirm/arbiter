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
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v2"

	arbiter "laptudirm.com/x/arbiter/pkg/common"
)

// Downloaded checks if the given Version of the given Engine has
// been downloaded previously by the manager.
func Downloaded(engine *Engine, version Version) bool {
	_, err := os.Stat(VersionBinary(engine, version.Name))
	return err == nil
}

// Binary returns the path to the main binary of the given Engine.
func Binary(engine *Engine) string {
	// ARBITER_BINARY_DIRECTORY/<engine-name>
	return filepath.Join(BinaryDirectory, engine.Name)
}

// VersionBinary returns the path to the given Version of the given Engine.
func VersionBinary(engine *Engine, version string) string {
	// ARBITER_BINARY_DIRECTORY/<engine-name>-<version-name>
	return Binary(engine) + "-" + version
}

// EngineInfoList maps Engine names to its EngineInfo.
type EngineInfoList map[string]EngineInfo

// TryAddEngine adds the given Engine to the EngineInfoList if
// it wasn't already included.
func (list EngineInfoList) TryAddEngine(engine *Engine) {
	if _, found := list[engine.Name]; !found {
		list[engine.Name] = EngineInfo{
			Author: engine.Author,
			Source: engine.URL,
		}
	}

	list.Dump()
}

func (list EngineInfoList) RemoveEngine(engine *Engine) {
	info := list[engine.Name]
	info.Current = ""
	info.Versions = []string{}
	list[engine.Name] = info
	list.Dump()
}

// AddVersion adds the given Version of the given Engine to the EngineInfoList.
func (list EngineInfoList) AddVersion(engine *Engine, version string) {
	list.TryAddEngine(engine)
	info := list[engine.Name]
	info.Versions = append(info.Versions, version)
	list[engine.Name] = info
	list.Dump()
}

func (list EngineInfoList) RemoveVersion(engine *Engine, version string) {
	info := list[engine.Name]
	versionIdx := slices.IndexFunc(info.Versions, func(v string) bool { return v == version })
	info.Versions = slices.Delete(info.Versions, versionIdx, versionIdx+1)
	list[engine.Name] = info
	list.Dump()
}

// SetMainVersion updates the version number of the main Engine binary.
func (list EngineInfoList) SetMainVersion(engine string, version string) {
	info := list[engine]
	info.Current = version
	list[engine] = info
	list.Dump()
}

// Dump writes the updated EngineInfoList to the configuration file.
func (list EngineInfoList) Dump() {
	file, _ := yaml.Marshal(list)
	_ = os.WriteFile(EnginesFile, file, arbiter.FilePermissions)
}

// EngineInfo stores information related to a single Engine.
type EngineInfo struct {
	Author string `yaml:"author"`
	Source string `yaml:"source"`

	// Installation Stuff
	Current     string   `yaml:"current"`
	BuildScript string   `yaml:"build-script,omitempty"`
	Versions    []string `yaml:"versions,omitempty"`
}

// Engines is the main EngineInfoList used by the manager.
var Engines EngineInfoList

func init() {
	// Create the source and binary directories if they don't yet exist.
	arbiter.TryMkdir(SourceDirectory)
	arbiter.TryMkdir(BinaryDirectory)

	// Create the engine configuration file if it doesn't exist.
	arbiter.TryCreate(EnginesFile, BaseEngineFile)

	// Load the engine configuration file into Engines.
	file, _ := os.ReadFile(EnginesFile)
	_ = yaml.Unmarshal(file, &Engines)
}
