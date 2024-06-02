// Copyright © 2023 Rak Laptudirm <rak@laptudirm.com>
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

package stats

import "math"

func StoppingBounds(alpha, beta float64) (lower float64, upper float64) {
	lower = math.Log(beta / (1 - alpha))
	upper = math.Log((1 - beta) / alpha)
	return
}

func clampElo(x float64) float64 {
	switch {
	case x <= 0, x >= 1:
		return 0

	default:
		return -400 * math.Log10(1/x-1)
	}
}

// eloToWDL converts the bayesian elo to its wdl probabilities.
func eloToWDL(elo, dlo float64) (w float64, d float64, l float64) {
	w = 1 / (1 + math.Pow(10, (-elo+dlo)/400)) // win probability sigmoid
	l = 1 / (1 + math.Pow(10, (+elo+dlo)/400)) // loss probability sigmoid
	d = 1 - w - l                              // draw probability curve
	return w, d, l
}

// wdlToElo converts the wdl probabilities to it's bayesian elo.
func wdlToElo(w, d, l float64) (elo float64, dlo float64) {
	elo = 200 * math.Log10((w/l)*((1-l)/(1-w)))
	dlo = 200 * math.Log10(((1-l)/l)*((1-w)/w))
	return elo, dlo
}

func phiInv(p float64) float64 {
	return math.Sqrt2 * math.Erfinv(2*p-1)
}

func nEloToScore(nelo, r float64) float64 {
	return nelo*math.Sqrt2*r/(800/math.Ln10) + 0.5
}
