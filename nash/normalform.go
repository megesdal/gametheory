package nash

import (
	"bytes"
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

func newEquilibrium(rowProbs []*big.Rat, colProbs []*big.Rat, fnPayoff func(int, int, int) *big.Rat) *Equilibrium {
	eq := &Equilibrium{
		rowProbs: rowProbs,
		colProbs: colProbs,
	}

	eq.computeExpectedPayoffs(len(rowProbs), len(colProbs), fnPayoff)
	return eq
}

func (eq *Equilibrium) computeExpectedPayoffs(nrows int, ncols int, fnPayoff func(int, int, int) *big.Rat) {

	eq.rowPay = zero()
	eq.colPay = zero()
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			prob := new(big.Rat).Mul(eq.rowProbs[i], eq.colProbs[j])

			rowPayPartial := new(big.Rat).Mul(prob, fnPayoff(i, j, 0))
			eq.rowPay.Add(eq.rowPay, rowPayPartial)

			colPayPartial := new(big.Rat).Mul(prob, fnPayoff(i, j, 1))
			eq.colPay.Add(eq.colPay, colPayPartial)
		}
	}
}

func (eq *Equilibrium) String() string {
	var buf bytes.Buffer

	buf.WriteString("rows")
	for i := 0; i < len(eq.rowProbs); i++ {
		buf.WriteString(" ")
		buf.WriteString(eq.rowProbs[i].String())
	}

	buf.WriteString("=")
	buf.WriteString(eq.rowPay.String())
	buf.WriteString("\n")

	buf.WriteString("cols")
	for i := 0; i < len(eq.colProbs); i++ {
		buf.WriteString(" ")
		buf.WriteString(eq.colProbs[i].String())
	}

	buf.WriteString("=")
	buf.WriteString(eq.colPay.String())
	return buf.String()
}

// LemkeEquilibrium finds a single nash equilibrium by constructing
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

	return LemkeEquilibriumWithPriors(convertToRats(payoffs), rowPriors, colPriors)
}

func convertToRats(payoffs64 [][][]float64) []*big.Rat {

	var payoffsRats []*big.Rat
	for i, nrows := 0, len(payoffs64); i < nrows; i++ {
		for j, ncols := 0, len(payoffs64[i]); j < ncols; j++ {
			for k, npl := 0, len(payoffs64[i][j]); k < npl; k++ {
				if payoffsRats == nil {
					payoffsRats = make([]*big.Rat, nrows*ncols*npl)
				}
				payoffsRats[(i*ncols+j)*npl+k] = new(big.Rat).SetFloat64(payoffs64[i][j][k])
			}
		}
	}
	return payoffsRats
}

// LemkeEquilibriumWithPriors runs the Lemke algorithm using the given prior
// beliefs to compute the covering vector.
func LemkeEquilibriumWithPriors(payoffs []*big.Rat, rowPriors []*big.Rat, colPriors []*big.Rat) (*Equilibrium, error) {

	nrows := len(rowPriors)
	ncols := len(colPriors)

	// 1. Adjust the payoffs to be strictly negative (max = -1)
	adjustedPayoffs := correctPaymentsNeg(payoffs)
	fnAdjustedPayoff := func(row int, col int, pl int) *big.Rat {
		return adjustedPayoffs[(row*ncols+col)*2+pl]
	}

	// 2. Generate the LCP from the two payoff matrices and the priors
	lcp := generateLCP(nrows, ncols, fnAdjustedPayoff)
	d := generateCovVector(lcp, rowPriors, colPriors)

	// 3. Pass the combination of the two to the Lemke algorithm
	z, err := lemke.Solve(lcp, d)
	if err != nil {
		return nil, err
	}

	// 4. Convert solution into a mixed strategy equilibrium
	pl1, pl2 := extractLCPSolution(z, nrows)

	// 5. Get original payoffs back and compute expected payoffs
	fnPayoff := func(row int, col int, pl int) *big.Rat {
		return payoffs[(row*ncols+col)*2+pl]
	}

	eq := newEquilibrium(pl1, pl2, fnPayoff)

	return eq, nil
}

func correctPaymentsNeg(payoffs []*big.Rat) []*big.Rat {

	max := [2]*big.Rat{payoffs[0], payoffs[1]}

	for i := 2; i < len(payoffs); i++ {
		entry := payoffs[i]
		pl := i % 2
		if entry.Cmp(max[pl]) > 0 {
			max[pl] = entry
		}
	}

	if max[0].Sign() < 0 && max[1].Sign() < 0 {
		return payoffs
	}

	corrections := [2]*big.Rat{zero(), zero()}
	for pl := 0; pl < 2; pl++ {
		if max[pl].Sign() >= 0 {
			corrections[pl].Neg(max[pl])
			corrections[pl].Sub(corrections[pl], one())
		}
	}

	return applyPayCorrect(payoffs, corrections)
}

func applyPayCorrect(payoffs []*big.Rat, corrections [2]*big.Rat) []*big.Rat {

	adjustedPayoffs := make([]*big.Rat, len(payoffs))

	for i := 0; i < len(payoffs); i++ {
		entry := payoffs[i]
		pl := i % 2
		adjustedPayoffs[i] = new(big.Rat).Add(entry, corrections[pl])
	}

	return adjustedPayoffs
}

// this assumes pays have been normalized to -1 as the max value
func generateLCP(nrows int, ncols int, fnPayoff func(int, int, int) *big.Rat) *lemke.LCP {

	size := nrows + ncols + 2

	// fill  M
	M := make([]*big.Rat, size*size)

	// zeros
	for i := 0; i < nrows+1; i++ {
		for j := 0; j < nrows+1; j++ {
			M[i*size+j] = zero()
		}
	}

	for i := 0; i < ncols+1; i++ {
		for j := 0; j < ncols+1; j++ {
			M[(nrows+1+i)*size+nrows+1+j] = zero()
		}
	}

	M[nrows*size+nrows+1+ncols] = zero()
	M[(nrows+1+ncols)*size+nrows] = zero()

	// -A
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			M[i*size+j+nrows+1] = new(big.Rat).Neg(fnPayoff(i, j, 0))
		}
	}

	// -E\T
	for i := 0; i < nrows; i++ {
		M[i*size+nrows+1+ncols] = negone()
	}

	// F
	for j := 0; j < ncols; j++ {
		M[nrows*size+nrows+1+j] = one()
	}

	// -B\T
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			M[(j+nrows+1)*size+i] = new(big.Rat).Neg(fnPayoff(i, j, 1))
		}
	}

	// -F\T
	for j := 0; j < ncols; j++ {
		M[(nrows+1+j)*size+nrows] = negone()
	}

	// E
	for i := 0; i < nrows; i++ {
		M[(nrows+1+ncols)*size+i] = one()
	}

	// fill q (rhs)
	q := make([]*big.Rat, size)
	for i := 0; i < size; i++ {
		if i == nrows || i == nrows+1+ncols {
			q[i] = negone()
		} else {
			q[i] = zero()
		}
	}

	return lemke.NewLCP(M, q)
}

func generateCovVector(lcp *lemke.LCP, rowPriors []*big.Rat, colPriors []*big.Rat) []*big.Rat {

	d := make([]*big.Rat, len(rowPriors)+len(colPriors)+2)

	// covering vector  = -rhsq
	for i := 0; i < len(d); i++ {
		d[i] = new(big.Rat).Neg(lcp.Q(i))
	}

	// first blockrow += -Aq
	offset := len(rowPriors) + 1
	for i := 0; i < len(rowPriors); i++ {
		for j := 0; j < len(colPriors); j++ {
			d[i].Add(d[i], new(big.Rat).Mul(lcp.M(i, offset+j), colPriors[j]))
		}
	}

	// third blockrow += -B\T p
	for i := offset; i < offset+len(colPriors); i++ {
		for j := 0; j < len(rowPriors); j++ {
			d[i].Add(d[i], new(big.Rat).Mul(lcp.M(i, j), rowPriors[j]))
		}
	}

	return d
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
