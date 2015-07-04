package gametheory

import (
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
  for i := 0; i < n; i++{
    M[i] = make([]*big.Rat, n)
  }

    // -A and -B\T
    for i, pl1Seq := range seqs1 {
        for j, pl2Seq := range seqs2 {
            // TODO: normalize payments by payment adjust here...
            pl1Payoff := new(big.Rat).Neg(sf.plPayoffs[pl1][pl1Seq][pl2Seq])
            M[i][j + len(seqs1) + len(isets2) + 1] = pl1Payoff  // -A

            pl2Payoff := new (big.Rat).Neg(sf.plPayoffs[pl2][pl2Seq][pl1Seq])
            M[j + len(seqs1) + len(isets2) + 1][i] = pl2Payoff  // -B\T
        }
    }

        // -E\T and E
        for i, pl1Iset := range isets1  {
            for j, pl1Seq := range seqs1 {
              value := big.NewRat(sf.plConstraints[pl1][pl1Iset][pl1Seq], 1)
                M[j][i + len(seqs1) + len(isets2) + 1 + len(seqs2)] = new(big.Rat).Neg(value)  // -E\T
                M[i + len(seqs1) + len(isets2) + 1 + len(seqs2)][j] = value  // E
            }
        }

        // F and -F\T
        for i, pl2Iset := range isets2 {
            for j, pl2Seq := range seqs2 {
                value := big.NewRat(sf.plConstraints[pl2][pl2Iset][pl2Seq], 1)
                M[i + len(seqs1)][j + len(seqs1) + len(isets2) + 1] = value   // F
                M[j + len(seqs1) + len(isets2) + 1][i + len(seqs1)] = new(big.Rat).Neg(value)  // -F\T
            }
        }

        // define RHS q,  using special shape of SF constraints RHS e,f
        q := make([]*big.Rat, n)
        q[len(seqs1)] = big.NewRat(-1, 1)
        q[len(seqs1) + len(isets2) + 1 + len(seqs2)] = big.NewRat(-1, 1)

        //TODO: addCoveringVector(lcp)
}
