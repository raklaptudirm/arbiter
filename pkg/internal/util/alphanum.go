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
	"regexp"
	"strconv"
)

var chunkifyRegexp = regexp.MustCompile(`(\d+|\D+)`)

func chunkify(s string) []string {
	return chunkifyRegexp.FindAllString(s, -1)
}

// AlphanumCompare returns true if the first string precedes the second one according to natural order
func AlphanumCompare(a, b string) bool {
	chunks_a := chunkify(a)
	chunks_b := chunkify(b)

	for i := range chunks_a {
		if i >= len(chunks_b) {
			return false
		}

		aInt, aErr := strconv.Atoi(chunks_a[i])
		bInt, bErr := strconv.Atoi(chunks_b[i])

		// If both chunks are numeric, compare them as integers
		if aErr == nil && bErr == nil {
			if aInt == bInt {
				if i == len(chunks_a)-1 {
					// We reached the last chunk of A, thus B is greater than A
					return true
				} else if i == len(chunks_b)-1 {
					// We reached the last chunk of B, thus A is greater than B
					return false
				}

				continue
			}

			return aInt < bInt
		}

		// So far both strings are equal, continue to next chunk
		if chunks_a[i] == chunks_b[i] {
			if i == len(chunks_a)-1 {
				// We reached the last chunk of A, thus B is greater than A
				return true
			} else if i == len(chunks_b)-1 {
				// We reached the last chunk of B, thus A is greater than B
				return false
			}

			continue
		}

		return chunks_a[i] < chunks_b[i]
	}

	return false
}
