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

package data

import "github.com/MakeNowJust/heredoc/v2"

type EngineInfo struct {
	Source string
	Author string

	// Installation Stuff
	BuildScript string
}

var Engines = map[string]EngineInfo{
	"Stockfish": {
		Source: "https://github.com/official-stockfish/stockfish",
		Author: "the Stockfish Developers",
		BuildScript: heredoc.Doc(`
			cd src
			make -j profile-build
			mv stockfish ../engine-binary
		`),
	},

	"Ethereal": {Source: "https://github.com/AndyGrant/Ethereal", Author: "Andrew Grant"},
	"Stash":    {Source: "https://gitlab.com/mhouppin/stash-bot", Author: "Morgan Houppin"},
	"Mess":     {Source: "https://github.com/raklaptudirm/mess", Author: "Rak Laptudirm"},

	"Zataxx": {
		Source: "https://github.com/zzzzz151/Zataxx",
		Author: "zzzzz",
		BuildScript: heredoc.Doc(`
			cargo build --release
			mv ./target/release/zataxx ./engine-binary
		`),
	},
}
