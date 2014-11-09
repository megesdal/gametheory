package nash

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/megesdal/gametheory/lemke"
)

// Equilibrium represents the information about a 2-player non-cooperative game
// nash equilibrium.
type Equilibrium struct {
	rowProbs []*big.Rat //length = #rows (1 per strategy)
	rowPay   *big.Rat

	colProbs []*big.Rat //length = #cols (1 per strategy)
	colPay   *big.Rat
}

func newEquilibrium(rowProbs []*big.Rat, colProbs []*big.Rat) *Equilibrium {
	return &Equilibrium{
		rowProbs: rowProbs,
		colProbs: colProbs,
	}
}

func LemkeEquilibrium(payoffs [][][]float64, seed int64) (*Equilibrium, error) {

	nrows := len(payoffs)
	if nrows == 0 {
		return nil, errors.New("Cannot have a payoff matrix with 0 rows")
	}

	ncols := len(payoffs[0])
	if ncols == 0 {
		return nil, errors.New("Cannot have a payoff matrix with 0 cols")
	}

	// set priors to randomly choose a strategy with Pr=1
	r := rand.New(rand.NewSource(seed))

	rowPriors := make([]*big.Rat, nrows)
	rowPr1 := r.Intn(nrows)
	for i := 0; i < nrows; i++ {
		if i == rowPr1 {
			rowPriors[i] = one()
		} else {
			rowPriors[i] = zero()
		}
	}

	colPriors := make([]*big.Rat, ncols)
	colPr1 := r.Intn(ncols)
	for j := 0; j < ncols; j++ {
		if j == colPr1 {
			colPriors[j] = one()
		} else {
			colPriors[j] = zero()
		}
	}

	return LemkeEquilibriumWithPriors(a, b, rowPriors, colPriors)
}

// LemkeEquilibriumWithPriors finds a single nash equilibrium by constructing
// an LCP and using Lemke's algorithm.
//
// The payoffs slice represents a 2d matrix with a payoff tuples as entries.
// The payoff tuples are just two sequencial slice entries with the first being
// the row payoff and the second being the column payoff.
//
// A 3 dimensional representation (the payoff tuple being the third dimension):
// P3 = [ [ [ -1, -1 ], [  0, -3 ] ],
//        [ [ -3,  0 ], [ -2, -2 ] ] ]
//
// Which gets translated to: P1 = [ -1, -1, 0, -3, -3, 0, -2, -2 ]
//
// The lookup would be P3[i][j][k] = P1[(i * ncols + j) * 2 + k], ncols = 2
//
func LemkeEquilibriumWithPriors(a []*big.Rat, b []*big.Rat, rowPriors []*big.Rat, colPriors []*big.Rat) (*Equilibrium, error) {

	nrows := len(rowPriors)
	ncols := len(colPriors)

	// 1. Adjust the payoffs to be strictly negative (max = -1)
	fnRevertA := correctPaymentsNeg(a, ncols)
	fnRevertB := correctPaymentsNeg(b, ncols)
	// TODO: consolidate above into a single method and single revert call

	// 2. Generate the LCP from the two payoff matrices and the priors
	lcp, err := generateLCP(a, b, ncols)
	if err != nil {
		return nil, err
	}
	d := generateCovVector(lcp, rowPriors, colPriors)

	// 3. Pass the combination of the two to the Lemke algorithm
	z, err := lcp.Lemke(d)
	if err != nil {
		return nil, err
	}

	// 4. Convert solution into a mixed strategy equilibrium
	pl1, pl2 := extractLCPSolution(z, nrows)
	eq := newEquilibrium(pl1, pl2)

	// 5. Get original payoffs back and compute expected payoffs
	fnRevertA()
	fnRevertB()

	eq.computeExpectedPayoffs(a, b)

	return eq
}

func correctPaymentsNeg(matrix []*big.Rat, ncols int) func() {

	max := matrix[0]
	nrows := len(matrix) / ncols
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			entry := matrix[i*ncols+j]
			if entry.Cmp(max) > 0 {
				max = entry
			}
		}
	}

	var correct *big.Rat
	if max.Sign() >= 0 {
		correct.Neg(max)
		correct.Sub(correct, one())
		applyPayCorrect(matrix, correct)
	} else {
		correct = zero()
	}

	return func() {
		if correct.Sign() < 0 {
			applyPayCorrect(matrix, correct.Neg(correct))
		}
	}
}

func applyPayCorrect(matrix []*big.Rat, ncols int, correct *big.Rat) {

	nrows := len(matrix) / ncols
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			entry := matrix[i*ncols+j]
			entry.Add(entry, correct)
		}
	}
}

// this assumes pays have been normalized to -1 as the max value
func generateLCP(a []*big.Rat, b []*big.Rat, ncols int) (*lemke.LCP, error) {

	nStratsA := len(a) / ncols
	nStratsB := ncols

	size := nStratsA + 1 + nStratsB + 1

	// fill  M
	M := make([]*big.Rat, size*size)

	// zeros
	for i := 0; i < nStratsA+1; i++ {
		for j := 0; j < nStratsA+1; j++ {
			M[i*size+j] = zero()
		}
	}

	for i := 0; i < nStratsB+1; i++ {
		for j := 0; j < nStratsB+1; j++ {
			M[(nStratsA+1+i)*size+nStratsA+1+j] = zero()
		}
	}

	M[nStratsA*size+nStratsA+1+nStratsB] = zero()
	M[(nStratsA+1+nStratsB)*size+nStratsA] = zero()

	// -A
	for i := 0; i < nStratsA; i++ {
		for j := 0; j < nStratsB; j++ {
			M[i*size+j+nStratsA+1] = new(big.Rat).Neg(a[i*ncols+j])
		}
	}

	// -E\T
	for i := 0; i < nStratsA; i++ {
		M[i*size+nStratsA+1+nStratsB] = negone()
	}

	// F
	for j := 0; j < nStratsB; j++ {
		M[nStratsA*size+nStratsA+1+j] = one()
	}

	// -B\T
	for i := 0; i < nStratsA; i++ {
		for j := 0; j < nStratsB; j++ {
			M[(j+nStratsA+1)*size+i] = new(big.Rat).Neg(b[i*ncols+j])
		}
	}

	// -F\T
	for j := 0; j < nStratsB; j++ {
		M[(nStratsA+1+j)*size+nStratsA] = negone()
	}

	// E
	for i := 0; i < nStratsA; i++ {
		M[(nStratsA+1+nStratsB)*size+i] = one()
	}

	// fill q (rhs)
	q := make([]*big.Rat, size)
	for i := 0; i < size; i++ {
		if i == nStratsA || i == nStratsA+1+nStratsB {
			q[i] = negone()
		} else {
			q[i] = zero()
		}
	}

	return lemke.NewLCP(M, q)
}

func generateCovVector(lcp *lemke.LCP, xPriors []*big.Rat, yPriors []*big.Rat) []*big.Rat {

	d := make([]*big.Rat, lcp.n)

	// covering vector  = -rhsq
	for i := 0; i < lcp.n; i++ {
		d[i] = new(big.Rat).Neg(lcp.q(i))
	}

	// first blockrow += -Aq
	offset := len(xPriors) + 1
	for i := 0; i < len(xPriors); i++ {
		for j := 0; j < yPriors.length; j++ {
			lcp.setd(i, lcp.d(i).add(lcp.M(i, offset+j).multiply(yPriors[j])))
		}
	}

	// third blockrow += -B\T p
	for i := offset; i < offset+len(yPriors); i++ {
		for j := 0; j < xPriors.length; j++ {
			lcp.setd(i, lcp.d(i).add(lcp.M(i, j).multiply(xPriors[j])))
		}
	}
}

func extractLCPSolution(z []*big.Rat, nrows int) ([]*big.Rat, []*big.Rat) {

	offset := nrows + 1
	pl1 := z[:offset-1]
	pl2 := z[offset : len(z)-1]

	return pl1, pl2
}

func negone() *big.Rat {
	return big.NewRat(int64(-1), int64(1))
}

func one() *big.Rat {
	return big.NewRat(int64(1), int64(1))
}

func zero() *big.Rat {
	return new(big.Rat)
}
