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
	"path/filepath"

	"github.com/adrg/xdg"
)

const FilePermissions = 0755

var Directory = filepath.Join(xdg.Home, "arbiter")

func TryMkdir(dir string) {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		_ = os.Mkdir(dir, FilePermissions)
	}
}

func TryCreate(file string, data []byte) {
	if _, err := os.Stat(file); errors.Is(err, fs.ErrNotExist) {
		_ = os.WriteFile(file, data, FilePermissions)
	}
}

func init() {
	TryMkdir(Directory)
	TryMkdir(filepath.Join(Directory, "paused"))
	TryMkdir(filepath.Join(Directory, "paused", "sprt"))
	TryMkdir(filepath.Join(Directory, "paused", "tour"))
}
