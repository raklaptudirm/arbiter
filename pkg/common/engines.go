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
	"io/fs"
	"os"

	"gopkg.in/yaml.v2"
)

type EngineInfoList map[string]EngineInfo

func (list EngineInfoList) TryAddEngine(name, author, source string) {
	if _, found := list[name]; !found {
		list[name] = EngineInfo{
			Author: author,
			Source: source,
		}
	}

	list.Dump()
}

func (list EngineInfoList) InstallVersion(engine string, version string) {
	info := list[engine]
	info.Versions = append(info.Versions, version)
	list[engine] = info
	list.Dump()
}

func (list EngineInfoList) SetMainVersion(engine string, version string) {
	info := list[engine]
	info.Current = version
	list[engine] = info
	list.Dump()
}

func (list EngineInfoList) Dump() {
	file, _ := yaml.Marshal(list)
	_ = os.WriteFile(EnginesFile, file, Permissions)
}

type EngineInfo struct {
	Author string `yaml:"author"`
	Source string `yaml:"source"`

	// Installation Stuff
	Current     string   `yaml:"current"`
	Versions    []string `yaml:"versions,omitempty"`
	BuildScript string   `yaml:"build-script,omitempty"`
}

var Engines EngineInfoList

func init() {
	try_mkdir(ArbiterDirectory)
	try_mkdir(SourceDirectory)
	try_mkdir(BinaryDirectory)

	try_mkfile(EnginesFile, BaseEngineFile)

	file, _ := os.ReadFile(EnginesFile)
	_ = yaml.Unmarshal(file, &Engines)
}

func try_mkdir(dir string) {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		_ = os.Mkdir(dir, Permissions)
	}
}

func try_mkfile(file string, data []byte) {
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		_ = os.WriteFile(file, data, Permissions)
	}
}
