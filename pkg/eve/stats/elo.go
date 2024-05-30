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

// SPRT does a statistical probability ratio test calculation on the given
// number of wins, draws, and losses from the tournament and returns the
// log-likelihood ratio (llr) for whether elo0 or elo1 is more likely to
// be correct. It only calculates when at least one of each result is there.
func SPRT(ws, ds, ls float64, elo0, elo1 float64) (llr float64) {
	// Implement Dirichlet([0.5, 0.5, 0.5]) prior
	ws += 0.5
	ds += 0.5
	ls += 0.5

	N := ws + ds + ls // total number of games
	_, dlo := wdlToElo(ws/N, ds/N, ls/N)

	w0, d0, l0 := eloToWDL(elo0, dlo) // elo0 WDL probabilities
	w1, d1, l1 := eloToWDL(elo1, dlo) // elo1 WDL probabilities

	// log-likelihood ratio (llr)
	return ws*math.Log(w1/w0) +
		ds*math.Log(d1/d0) +
		ls*math.Log(l1/l0)
}

// Elo returns the likely elo of the target player along with its p < 0.05
// upper bound and lower bound, called mu, muMax, and muMin respectively.
func Elo(ws, ds, ls int) (muMin float64, mu float64, muMax float64) {
	N := float64(ws + ds + ls) // total number of games

	if N == 0 {
		return 0, 0, 0
	}

	w := float64(ws) / N // measured win probability
	d := float64(ds) / N // measured draw probability
	l := float64(ls) / N // measured loss probability

	// empirical mean of random variable
	mu = w + d/2

	// standard deviation of the random variable
	sigma := math.Sqrt(w*math.Pow(1-mu, 2)+d*math.Pow(0.5-mu, 2)+l*math.Pow(0-mu, 2)) / math.Sqrt(N)

	muMax = clampElo(mu + phiInv(0.025)*sigma) // upper bound
	muMin = clampElo(mu + phiInv(0.975)*sigma) // lower bound

	return muMin, clampElo(mu), muMax
}

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
