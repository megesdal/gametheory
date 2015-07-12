package gametheory

import (
	"fmt"
	"github.com/megesdal/gametheory/lemke"
	"math/big"
)

func (sf *SequenceForm) Lemke() {

	// 1. Find payment adjustment...

	// 2. Create LCP (M, q)

	// 3. Create prior beliefs and generate covering vector d

	// 4. Run Lemke

	// 5. Create Equilibrium and compute payoffs
}

func (sf *SequenceForm) CreateLCP() {

	if len(sf.plNames) != 2 {
		panic("Sequence Form LCP must have two and only two players")
	}

	// TODO: payment adjustment
	maxpay := maxPayoffs(sf)

	// preprocess priors here so that we can re-randomize the priors without having to reconstruct this object?
	//for (Player pl = firstPlayer; pl != null; pl = pl.next) {
	//		behavtorealprob(pl);
	//}

	// TODO: do I need the + 1 since I have the empty sequence already?
	pl1 := sf.plNames[0]
	pl2 := sf.plNames[1]
	maxpay1 := maxpay[pl1]
	maxpay2 := maxpay[pl2]
	seqs1 := sf.plSeqs[pl1]
	seqs2 := sf.plSeqs[pl2]
	isets1 := sf.plIsets[pl1]
	isets2 := sf.plIsets[pl2]
	n := len(seqs1) + len(isets2) + len(seqs2) + len(isets1)
	M := make([]*big.Rat, n*n)
	for i := 0; i < n*n; i++ {
		M[i] = new(big.Rat)
	}

	// -A and -B\T
	for i, pl1Seq := range seqs1 {
		for j, pl2Seq := range seqs2 {
			// TODO: normalize payments by payment adjust here?
			pl1Payoff := sf.plPayoffs[pl1][pl1Seq][pl2Seq]
			if pl1Payoff == nil {
				pl1Payoff = new(big.Rat)
			} else {
				tmp := new(big.Rat)
				pl1Payoff = tmp.Add(tmp, maxpay1).Sub(tmp, pl1Payoff)
			}
			M[i*n+j+len(seqs1)+len(isets2)] = pl1Payoff // -A

			pl2Payoff := sf.plPayoffs[pl2][pl2Seq][pl1Seq]
			if pl2Payoff == nil {
				pl2Payoff = new(big.Rat)
			} else {
				tmp := new(big.Rat)
				fmt.Println(pl2, maxpay2, pl2Payoff)
				pl2Payoff = tmp.Add(tmp, maxpay2).Sub(tmp, pl2Payoff)
			}
			M[(j+len(seqs1)+len(isets2))*n+i] = pl2Payoff // -B\T
		}
	}

	// -E\T and E
	for i, pl1Iset := range isets1 {
		for j, pl1Seq := range seqs1 {
			value := big.NewRat(sf.plConstraints[pl1][pl1Iset][pl1Seq], 1)
			M[j*n+i+len(seqs1)+len(isets2)+len(seqs2)] = new(big.Rat).Neg(value) // -E\T
			M[(i+len(seqs1)+len(isets2)+len(seqs2))*n+j] = value                 // E
		}
	}

	// F and -F\T
	for i, pl2Iset := range isets2 {
		for j, pl2Seq := range seqs2 {
			value := big.NewRat(sf.plConstraints[pl2][pl2Iset][pl2Seq], 1)
			M[(i+len(seqs1))*n+j+len(seqs1)+len(isets2)] = value                   // F
			M[(j+len(seqs1)+len(isets2))*n+i+len(seqs1)] = new(big.Rat).Neg(value) // -F\T
		}
	}

	// define RHS q,  using special shape of SF constraints RHS e,f
	q := make([]*big.Rat, n)
	for i := 0; i < n; i++ {
		if i == len(seqs1) || i == len(seqs1)+len(isets2)+len(seqs2) {
			q[i] = big.NewRat(-1, 1)
		} else {
			q[i] = new(big.Rat)
		}
	}

	d := sf.coveringVector(M, q)

	z, err := lemke.Solve(lemke.NewLCP(M, q), d)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("SUCCESS...", z)
		sf.parseLemkeSolution(z)
	}
}

func (sf *SequenceForm) coveringVector(M []*big.Rat, q []*big.Rat) []*big.Rat {

	n := len(q)
	pl1 := sf.plNames[0]
	pl2 := sf.plNames[1]
	offset := len(sf.plSeqs[pl1]) + len(sf.plIsets[pl2])

	plPriors := make(map[string]map[string]*big.Rat)
	plPriors[pl1] = make(map[string]*big.Rat)
	plPriors[pl2] = make(map[string]*big.Rat)

	sf.assignPriors(pl1, plPriors[pl1])
	sf.assignPriors(pl2, plPriors[pl2])

	d := make([]*big.Rat, n)
	/* covering vector  = -rhsq */
	for i, qi := range q {
		d[i] = new(big.Rat).Neg(qi)
	}

	/* first blockrow += -Aq    */
	for j, seq2 := range sf.plSeqs[pl2] {
		prob := plPriors[pl2][seq2]
		if prob.Sign() != 0 {
			for i := 0; i < len(sf.plSeqs[pl1]); i++ {
				d[i].Add(d[i], new(big.Rat).Mul(M[i*n+offset+j], prob))
			}
		}
	}

	/* third blockrow += -B\T p */
	for j, seq1 := range sf.plSeqs[pl1] {
		prob := plPriors[pl1][seq1]
		if prob.Sign() != 0 {
			for i := offset; i < offset+len(sf.plSeqs[pl2]); i++ {
				d[i].Add(d[i], new(big.Rat).Mul(M[i*n+j], prob))
			}
		}
	}

	return d
}

// I could just do a depth first visit on the tree again... using chance if it exists and assigning otherwise
// pl -> move -> Pr (pl -> (empty) -> 1)

func (sf *SequenceForm) assignPriors(pl string, priors map[string]*big.Rat) {
	//moves := sf.plSeqs[pl]
	priors["\u2205"] = big.NewRat(1, 1)

	for _, iset := range sf.plIsets[pl] {
		//seqIn := sf.plIsetDefSeqs[iset]
		//probIn := priors[pl][seqIn]
		sf.assignIsetPriors(pl, iset, priors)
		// TODO: c.realprob = c.prob.multiply(seqin.get(c.iset()).realprob);
	}
}

func (sf *SequenceForm) assignIsetPriors(pl string, iset string, priors map[string]*big.Rat) {

	probToGive := big.NewRat(1, 1)
	var nNeedProb int64
	var probIn *big.Rat
	for seq, val := range sf.plConstraints[pl][iset] {
		if val == 1 {
			//if child.reachedby.prob == null) {
			nNeedProb++
			/*} else {
							probToGive = probToGive.subtract(child.reachedby.prob);
							if (probToGive.compareTo(0) < 0) {
								     child.reachedby.prob = child.reachedby.prob.add(probToGive);
								     probToGive = Rational.ZERO;
							    }
						   }
			      }*/
		} else if val == -1 {
			probIn = priors[seq] // there should only be one... if isets sorted by depth it should be filled in
		}
	}

	if probIn == nil {
		fmt.Println("probIn was nil...", pl, iset)
		probIn = big.NewRat(1, 1)
	}

	if nNeedProb > 0 {
		/*if (prng != null) {
			Rational[] probs = Rational.probVector(nNeedProb, prng);
			int i = 0;
			for (Node child = h.firstNode().firstChild(); child != null; child = child.sibling()) {
				if (child.reachedby.prob == null) {
					randomPriors.put(child.reachedby, probToGive.multiply(probs[i]));
					++i;
				}
			}
		} else {*/
		prob := probToGive.Mul(probToGive, big.NewRat(1, nNeedProb)).Mul(probToGive, probIn)
		for seq, val := range sf.plConstraints[pl][iset] {
			if priors[seq] == nil && val == 1 {
				//if child.reachedby.prob == null) {
				priors[seq] = prob
				fmt.Println("realprob", pl, seq, prob)
				/*} else {
								probToGive = probToGive.subtract(child.reachedby.prob);
								if (probToGive.compareTo(0) < 0) {
									     child.reachedby.prob = child.reachedby.prob.add(probToGive);
									     probToGive = Rational.ZERO;
								    }
							   }
				      }*/
			}
		}
	}
}

// TODO: this doesn't need to be 2-player only...
func maxPayoffs(sf *SequenceForm) map[string]*big.Rat {

	pl1 := sf.plNames[0]
	pl2 := sf.plNames[1]
	var maxpay1 *big.Rat
	var maxpay2 *big.Rat
	for _, pl1Seq := range sf.plSeqs[pl1] {
		for _, pl2Seq := range sf.plSeqs[pl2] {
			payoff1 := sf.plPayoffs[pl1][pl1Seq][pl2Seq]
			if payoff1 != nil && (maxpay1 == nil || maxpay1.Cmp(payoff1) < 0) {
				maxpay1 = payoff1
			}

			payoff2 := sf.plPayoffs[pl2][pl2Seq][pl1Seq]
			if payoff2 != nil && (maxpay2 == nil || maxpay2.Cmp(payoff2) < 0) {
				maxpay2 = payoff2
			}
		}
	}

	maxpay := make(map[string]*big.Rat)
	maxpay[pl1] = maxpay1
	maxpay[pl2] = maxpay2
	return maxpay
}

// pl -> seq -> prob
func (sf *SequenceForm) parseLemkeSolution(z []*big.Rat) map[string]map[string]*big.Rat {

	pl1 := sf.plNames[0]
	pl2 := sf.plNames[1]
	probs := make(map[string]map[string]*big.Rat)
	offset := len(sf.plSeqs[pl1]) + len(sf.plIsets[pl2])

	probs1 := make([]*big.Rat, len(sf.plSeqs[pl1]))
	for i := 0; i < len(probs1); i++ {
		probs1[i] = z[i]
	}

	// how to find expected payoffs... traverse the tree after?
	probs[pl1] = sf.createMoveMap(probs1, pl1)

	probs2 := make([]*big.Rat, len(sf.plSeqs[pl2]))
	for i := 0; i < len(probs2); i++ {
		probs2[i] = z[i+offset]
	}

	// how to find expected payoffs... traverse the tree after?
	probs[pl2] = sf.createMoveMap(probs2, pl2)
	return probs
}

// TODO...
// these are realization probabilities, not behavior probs...
// need to convert back...
func (sf *SequenceForm) createMoveMap(probs []*big.Rat, pl string) map[string]*big.Rat {

	realProbs := make(map[string]*big.Rat)
	behvProbs := make(map[string]*big.Rat)

	for i := 1; i < len(probs); i++ { // skip the empty seq...
		realProb := probs[i]
		if probs[i].Sign() != 0 {

			// because of how seqs are ordered we know that the closer-to-root probs
			// will be filled in when we need them
			seq := sf.plSeqs[pl][i]
			iset := sf.plSeqIsets[pl][seq]
			parentSeq := sf.plIsetDefSeqs[pl][iset]
			parentRealProb := realProbs[parentSeq]

			var behvProb *big.Rat
			if parentRealProb == nil {
				behvProb = big.NewRat(1, 1)
			} else {
				behvProb = new(big.Rat).Inv(parentRealProb)
			}
			behvProb.Mul(behvProb, realProb)
			realProbs[seq] = realProb
			behvProbs[seq] = behvProb
			fmt.Println("move", seq, realProb, behvProb)
		}
	}
	return behvProbs
}
