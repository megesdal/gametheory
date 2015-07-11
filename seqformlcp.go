package gametheory

import (
	"fmt"
	//"github.com/megesdal/gametheory/lemke"
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

	// preprocess priors here so that we can re-randomize the priors without having to reconstruct this object?
	//for (Player pl = firstPlayer; pl != null; pl = pl.next) {
	//		behavtorealprob(pl);
	//}

	// TODO: do I need the + 1 since I have the empty sequence already?
	pl1 := sf.plNames[0]
	pl2 := sf.plNames[1]
	seqs1 := sf.plSeqs[pl1]
	seqs2 := sf.plSeqs[pl2]
	isets1 := sf.plIsets[pl1]
	isets2 := sf.plIsets[pl2]
	n := len(seqs1) + len(isets2) + 1 + len(seqs2) + len(isets1) + 1
	M := make([][]*big.Rat, n)
	for i := 0; i < n; i++ {
		M[i] = make([]*big.Rat, n)
	}

	// -A and -B\T
	for i, pl1Seq := range seqs1 {
		for j, pl2Seq := range seqs2 {
			// TODO: normalize payments by payment adjust here...
			pl1Payoff := new(big.Rat).Neg(sf.plPayoffs[pl1][pl1Seq][pl2Seq])
			M[i][j+len(seqs1)+len(isets2)+1] = pl1Payoff // -A

			pl2Payoff := new(big.Rat).Neg(sf.plPayoffs[pl2][pl2Seq][pl1Seq])
			M[j+len(seqs1)+len(isets2)+1][i] = pl2Payoff // -B\T
		}
	}

	// -E\T and E
	for i, pl1Iset := range isets1 {
		for j, pl1Seq := range seqs1 {
			value := big.NewRat(sf.plConstraints[pl1][pl1Iset][pl1Seq], 1)
			M[j][i+len(seqs1)+len(isets2)+1+len(seqs2)] = new(big.Rat).Neg(value) // -E\T
			M[i+len(seqs1)+len(isets2)+1+len(seqs2)][j] = value                   // E
		}
	}

	// F and -F\T
	for i, pl2Iset := range isets2 {
		for j, pl2Seq := range seqs2 {
			value := big.NewRat(sf.plConstraints[pl2][pl2Iset][pl2Seq], 1)
			M[i+len(seqs1)][j+len(seqs1)+len(isets2)+1] = value                   // F
			M[j+len(seqs1)+len(isets2)+1][i+len(seqs1)] = new(big.Rat).Neg(value) // -F\T
		}
	}

	// define RHS q,  using special shape of SF constraints RHS e,f
	q := make([]*big.Rat, n)
	q[len(seqs1)] = big.NewRat(-1, 1)
	q[len(seqs1)+len(isets2)+1+len(seqs2)] = big.NewRat(-1, 1)

	//TODO: addCoveringVector(lcp)
}

func (sf *SequenceForm) coveringVector(M [][]*big.Rat, q []*big.Rat) []*big.Rat {

	pl1 := sf.plNames[0]
	pl2 := sf.plNames[1]
	dim1 := len(sf.plSeqs[pl1])
	dim2 := len(sf.plSeqs[pl2])
	offset := dim1 + 1 + len(sf.plIsets[pl2])

	priors := make(map[string]map[string]*big.Rat)
	priors[pl1] = make(map[string]*big.Rat)
	priors[pl2] = make(map[string]*big.Rat)

	d := make([]*big.Rat, len(q))
	/* covering vector  = -rhsq */
	for i, qi := range q {
		d[i] = new(big.Rat).Neg(qi)
	}

	/* first blockrow += -Aq    */
	for j := 0; j < dim2; j++ {
		prob := realprob(sf.plSeqs[pl2][j])
		if prob.Sign() != 0 {
			for i := 0; i < dim1; i++ {
				d[i].Add(d[i], new(big.Rat).Mul(M[i][offset+j], prob))
			}
		}
	}

	/* third blockrow += -B\T p */
	for j := 0; j < dim1; j++ {
		prob := realprob(sf.plSeqs[pl1][j])
		if prob.Sign() != 0 {
			for i := offset; i < offset+dim2; i++ {
				d[i].Add(d[i], new(big.Rat).Mul(M[i][j], prob))
			}
		}
	}
	return d
}

func realprob(seq string) *big.Rat {
	return nil // TODO...
}

// I could just do a depth first visit on the tree again... using chance if it exists and assigning otherwise
// pl -> move -> Pr (pl -> (empty) -> 1)

func (sf *SequenceForm) assignPriors(pl string, priors map[string]map[string]*big.Rat) {
	moves := sf.plSeqs[pl]
	priors[pl]["\u2205"] = big.NewRat(1, 1)

	for i, iset := range sf.plIsets[pl] {
		c := moves[i]
		fmt.Println("TODO", c, iset)
		// TODO: c.realprob = c.prob.multiply(seqin.get(c.iset()).realprob);
	}
}

func (sf *SequenceForm) assignIsetPriors(pl string, h string, priors map[string]map[string]*big.Rat) {
	probToGive := big.NewRat(1, 1)
	var nNeedProb int64
	var probIn *big.Rat
	for seq, val := range sf.plConstraints[pl][h] {
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
			probIn = priors[pl][seq] // there should only be one... if isets sorted by depth it should be filled in
		}
	}

	if probIn == nil {
		fmt.Println("probIn was nil...", pl, h)
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
		for seq, val := range sf.plConstraints[pl][h] {
			if val == 1 {
				//if child.reachedby.prob == null) {
				priors[pl][seq] = prob
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
