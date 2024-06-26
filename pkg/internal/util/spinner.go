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

package util

import (
	"time"

	"github.com/briandowns/spinner"
)

var global_spinner = spinner.New(spinner.CharSets[31], 100*time.Millisecond)
var spinning = false

func StartSpinner() {
	if !spinning {
		spinning = true
		global_spinner.Start()
	}
}

func PauseSpinner() {
	if spinning {
		spinning = false
		global_spinner.Stop()
	}
}
