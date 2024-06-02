package stats

import "math"

// PentaSPRT takes the results of the game pairs and the two elo hypotheses and
// returns a log-likelihood ratio which compares the fit of the two hypotheses
// to the provided game-pair data using a pentanomial model. In an SPRT test,
// either hypothesis might be accepted based on the llr and the calculated
// stopping bounds for the test, which are based on the desired type I and type
// II error probabilities.
func PentaSPRT(lls, lds, dds, wds, wws int, elo0, elo1 float64) (llr float64) {
	N := float64(lls+lds+dds+wds+wws) + 2.5 // total number of pairs

	ll := (float64(lls) + 0.5) / N // measured loss-loss probability
	ld := (float64(lds) + 0.5) / N // measured loss-draw probability
	dd := (float64(dds) + 0.5) / N // measured win-loss/draw-draw probability
	wd := (float64(wds) + 0.5) / N // measured win-draw probability
	ww := (float64(wws) + 0.5) / N // measured win-win probability

	// empirical mean of random variable
	mu := ww + 0.75*wd + 0.5*dd + 0.25*ld

	// standard deviation (multiplied by sqrt of N) of the random variable
	r := math.Sqrt(
		ww*math.Pow(1-mu, 2) +
			wd*math.Pow(0.75-mu, 2) +
			dd*math.Pow(0.50-mu, 2) +
			ld*math.Pow(0.25-mu, 2) +
			ll*math.Pow(0.00-mu, 2),
	)

	// convert elo bounds to score
	mu0 := nEloToScore(elo0, r)
	mu1 := nEloToScore(elo1, r)

	// deviation to the score bounds
	r0 := ww*math.Pow(1-mu0, 2) +
		wd*math.Pow(0.75-mu0, 2) +
		dd*math.Pow(0.50-mu0, 2) +
		ld*math.Pow(0.25-mu0, 2) +
		ll*math.Pow(0.00-mu0, 2)
	r1 := ww*math.Pow(1-mu1, 2) +
		wd*math.Pow(0.75-mu1, 2) +
		dd*math.Pow(0.50-mu1, 2) +
		ld*math.Pow(0.25-mu1, 2) +
		ll*math.Pow(0.00-mu1, 2)

	if r0 == 0 || r1 == 0 {
		return 0
	}

	// log-likelihood ratio (llr)
	// note: this is not the exact llr formula but rather a simplified yet
	// very accurate approximation. see http://hardy.uhasselt.be/Fishtest/support_MLE_multinomial.pdf
	return 0.5 * N * math.Log(r0/r1)
}

// PentaElo calculates the best fit elo for the given game pair results using a
// pentanomial model. It also calculates the maximum and minimum values of that
// elo estimate (the error bounds) with p < 0.05.
func PentaElo(lls, lds, dds, wds, wws int) (muMin float64, mu float64, muMax float64) {
	N := float64(lls+lds+dds+wds+wws) + 2.5 // total number of pairs

	ll := (float64(lls) + 0.5) / N // measured loss-loss probability
	ld := (float64(lds) + 0.5) / N // measured loss-draw probability
	dd := (float64(dds) + 0.5) / N // measured win-loss/draw-draw probability
	wd := (float64(wds) + 0.5) / N // measured win-draw probability
	ww := (float64(wws) + 0.5) / N // measured win-win probability

	// empirical mean of random variable
	mu = ww + 0.75*wd + 0.5*dd + 0.25*ld

	// standard deviation of the random variable
	sigma := math.Sqrt(
		ww*math.Pow(1-mu, 2)+
			wd*math.Pow(0.75-mu, 2)+
			dd*math.Pow(0.50-mu, 2)+
			ld*math.Pow(0.25-mu, 2)+
			ll*math.Pow(0.00-mu, 2),
	) / math.Sqrt(N)

	muMax = mu + phiInv(0.025)*sigma // upper bound
	muMin = mu + phiInv(0.975)*sigma // lower bound

	return clampElo(muMin), clampElo(mu), clampElo(muMax)
}
