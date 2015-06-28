package lemke

import (
	"errors"
	"fmt"
	"math/big"
)

// LCP (aka. Linear Complementarity Problem)
// =============================================================================
// (1) Mz + q >= 0
// (2) z >= 0
// (3) z'(Mz + q) = 0
//
// (1) and (2) are feasibility conditions.
// (3) is complementarity condition (also written as w = Mz + q where w and z are orthogonal)
// Lemke algorithm takes this (M, q) and a covering vector (d) and outputs a solution
//
type LCP struct {
	m []*big.Rat
	q []*big.Rat
	n int
}

func NewLCP(M []*big.Rat, q []*big.Rat) *LCP {

	n := len(q)
	if len(M)%n != 0 {
		panic("M.rows and q are not same dimensions")
	}

	if len(M) / n != n {
		panic(fmt.Sprintf("M must be a square matrix but was %dx%d", len(M) / n, n))
	}

	fmt.Printf("Creating LCP of dimenstion n=%d\n", n)

	return &LCP{m: M, q: q, n: n}
}

func (lcp *LCP) M(i int, j int) *big.Rat {
	return lcp.m[i*lcp.n+j]
}

func (lcp *LCP) Q(i int) *big.Rat {
	return lcp.q[i]
}

/* scale factors for variables z
* scfa[Z(0)]   for  d,  scfa[RHS] for  q
* scfa[Z(1..n)] for cols of  M
* result variables to be multiplied with these
*/
func (lcp *LCP) createTableau(d []*big.Rat) (*tableau, []*big.Int, error) {

	err := checkInputs(lcp.q, d)
	if err != nil {
		return nil, nil, err
	}

	tableau := newTableau(lcp.n)
	scfa := make([]*big.Int, tableau.ncols)

	for j := 0; j < tableau.ncols; j++ {

		fnVec := func(i int) *big.Rat {
			if j == 0 {
				return d[i]
			}
			if j == lcp.n+1 {
				return lcp.q[i]
			}
			return lcp.m[i*lcp.n+(j-1)]
		}

		// TODO: store scaleFactor on tableauVariable struct?
		scaleFactor := computeScaleFactor(lcp.n, fnVec)
		scfa[j] = scaleFactor

		for i := 0; i < tableau.nrows; i++ {

			rat := fnVec(i)

			/* cols 0..n of  A  contain LHS cobasic cols of  Ax = b     */
			/* where the system is here         -Iw + dz_0 + Mz = -q    */
			/* cols of  q  will be negated after first min ratio test   */
			/* A[i][j] = num * (scfa[j] / den),  fraction is integral       */

			value := new(big.Int).Mul(rat.Num(), scaleFactor)
			value.Div(value, rat.Denom())
			tableau.set(i, j, value)
		}
	}

	return tableau, scfa, nil
}

/*
 * compute lcm  of denominators for  col  j  of  A
 * Necessary for converting fractions to integers and back again
 */
func computeScaleFactor(n int, vec func(i int) *big.Rat) *big.Int {

	lcm := big.NewInt(1)
	for i := 0; i < n; i++ {

		denom := vec(i).Denom()

		tmp := new(big.Int).GCD(nil, nil, lcm, denom)
		tmp.Div(lcm, tmp)
		lcm.Mul(tmp, denom)
	}
	return lcm
}

/*
 * asserts that  d >= 0  and not  q >= 0  (o/w trivial sol)
 * and that q[i] < 0  implies  d[i] > 0
 */
func checkInputs(q []*big.Rat, d []*big.Rat) error {

	isQPos := true
	for i := 0; i < len(q); i++ {
		if d[i].Sign() < 0 {
			return fmt.Errorf("Covering vector  d[%d] = %s negative. Cannot start Lemke.", i+1, d[i])
		} else if q[i].Sign() < 0 {
			isQPos = false
			if d[i].Sign() == 0 {
				return fmt.Errorf("Covering vector  d[%d] = 0  where  q[%d] = %s  is negative. Cannot start Lemke.", i+1, i+1, q[i])
			}
		}
	}

	if isQPos {
		return errors.New("No need to start Lemke since  q>=0. Trivial solution  z=0.")
	}

	return nil
}
