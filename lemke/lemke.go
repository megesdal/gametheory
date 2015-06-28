package lemke

import (
	"errors"
	"fmt"
	"math/big"
)

// Solve the linear complementarity probelm via Lemke's algorithm.
// Returns nil with an error if ray termination
func Solve(lcp *LCP, d []*big.Rat) ([]*big.Rat, error) {
	return SolveWithPivotMax(lcp, d, 0)
}

// SolveWithPivotMax will only perform up to maxCount pivots before exiting.
func SolveWithPivotMax(lcp *LCP, d []*big.Rat, maxCount int) ([]*big.Rat, error) {

	tableau, scaleFactors, err := createTableau(lcp, d)
	if err != nil {
		return nil, err
	}

	nextLeavingVar := func(enter *tableauVariable) (*tableauVariable, bool, error) {
		return lexminratio(tableau, enter)
	}

	// z0 enters the basis to obtain lex-feasible solution
	enter := tableau.vars.z(0)
	leave, z0leave, err := nextLeavingVar(enter)

	// now give the entering q-col its correct sign
	tableau.negateCol(tableau.rhsCol())

	pivotCount := 1
	for {

		fmt.Printf("%d LCP:\n%v", pivotCount, tableau)
		fmt.Println(pivotCount, "entering", enter, "leaving", leave)

		tableau.pivot(leave, enter)

		if z0leave {
			break // z0 will have a value of zero but may still be basic... amend?
		}

		// selectpivot
		enter = leave.complement()
		fmt.Printf("Complement of %s is %s\n", leave, enter)

		leave, z0leave, err = nextLeavingVar(enter)
		if err != nil {
			break // ray termination...
		}

		if pivotCount == maxCount { /* maxcount == 0 is equivalent to infinity since pivotcount starts at 1 */
			//log.warning(String.format("------- stop after %d pivoting steps --------", maxcount));
			break
		}

		pivotCount++
	}

	fmt.Printf("LCP (final):\n%v", tableau)
	return solution(tableau, scaleFactors), err // LCP solution = z  vector
}

/*
 * LCP result
 * current basic solution turned into  solz [0..n-1]
 * note that Z(1)..Z(n)  become indices  0..n-1
 */
func solution(tableau *tableau, scaleFactors []*big.Int) []*big.Rat {

	z := make([]*big.Rat, tableau.vars.n)
	den := new(big.Int).Mul(tableau.det, scaleFactors[tableau.rhsCol()]) // TODO: okay? or do I need a fresh copy each time?
	for i := 0; i < len(z); i++ {
		// skip z0... just z(1)..z(n)
		z[i] = result(tableau.vars.z(i+1), den, tableau, scaleFactors)
	}
	return z
}

/*
 * Z(i):  scfa[i]*rhs[row] / (scfa[RHS]*det)
 * W(i):  rhs[row] / (scfa[RHS]*det)
 */
func result(tvar *tableauVariable, den *big.Int, tableau *tableau, scaleFactors []*big.Int) *big.Rat {

	rv := big.NewRat(0, 1)
	if tvar.isBasic() {

		var num *big.Int
		if tvar.isZ() {
			num = scaleFactors[tvar.idx]
		} else {
			num = big.NewInt(1)
		}

		row := tvar.row()
		num.Mul(num, tableau.rhsEntry(row))

		rv.SetFrac(num, den)
	}

	return rv
}

/* scale factors for variables z
* scfa[Z(0)]   for  d,  scfa[RHS] for  q
* scfa[Z(1..n)] for cols of  M
* result variables to be multiplied with these
 */
func createTableau(lcp *LCP, d []*big.Rat) (*tableau, []*big.Int, error) {

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
