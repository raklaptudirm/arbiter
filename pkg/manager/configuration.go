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

	"gopkg.in/yaml.v2"

	arbiter "laptudirm.com/x/arbiter/pkg/common"
)

type EngineInfoList map[string]EngineInfo

func (list EngineInfoList) TryAddEngine(engine *Engine) {
	if _, found := list[engine.Name]; !found {
		list[engine.Name] = EngineInfo{
			Author: engine.Author,
			Source: engine.URL,
		}
	}

	list.Dump()
}

func (list EngineInfoList) AddVersion(engine *Engine, version string) {
	list.TryAddEngine(engine)
	info := list[engine.Name]
	info.Versions = append(info.Versions, version)
	list[engine.Name] = info
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
	_ = os.WriteFile(EnginesFile, file, arbiter.FilePermissions)
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
	arbiter.TryMkdir(SourceDirectory)
	arbiter.TryMkdir(BinaryDirectory)

	arbiter.TryCreate(EnginesFile, BaseEngineFile)

	file, _ := os.ReadFile(EnginesFile)
	_ = yaml.Unmarshal(file, &Engines)
}
