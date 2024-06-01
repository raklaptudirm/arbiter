package stats

import "math"

// SPRT does a statistical probability ratio test calculation on the given
// number of wins, draws, and losses from the tournament and returns the
// log-likelihood ratio (llr) for whether elo0 or elo1 is more likely to
// be correct. It only calculates when at least one of each result is there.
func SPRT(ws, ds, ls int, elo0, elo1 float64) (llr float64) {
	w := float64(ws) + 0.5
	d := float64(ds) + 0.5
	l := float64(ls) + 0.5

	N := w + d + l // total number of games
	_, dlo := wdlToElo(w/N, d/N, l/N)

	w0, d0, l0 := eloToWDL(elo0, dlo) // elo0 WDL probabilities
	w1, d1, l1 := eloToWDL(elo1, dlo) // elo1 WDL probabilities

	// log-likelihood ratio (llr)
	return w*math.Log(w1/w0) +
		d*math.Log(d1/d0) +
		l*math.Log(l1/l0)
}

// Elo returns the likely elo of the target player along with its p < 0.05
// upper bound and lower bound, called mu, muMax, and muMin respectively.
func Elo(ws, ds, ls int) (muMin float64, mu float64, muMax float64) {
	N := float64(ws+ds+ls) + 1.5 // total number of games

	w := (float64(ws) + 0.5) / N // measured win probability
	d := (float64(ds) + 0.5) / N // measured draw probability
	l := (float64(ls) + 0.5) / N // measured loss probability

	// empirical mean of random variable
	mu = w + d/2

	// standard deviation of the random variable
	sigma := math.Sqrt(
		w*math.Pow(1-mu, 2)+
			d*math.Pow(0.5-mu, 2)+
			l*math.Pow(0-mu, 2),
	) / math.Sqrt(N)

	muMax = mu + phiInv(0.025)*sigma // upper bound
	muMin = mu + phiInv(0.975)*sigma // lower bound

	return clampElo(muMin), clampElo(mu), clampElo(muMax)
}
