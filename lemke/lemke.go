package lemke

import (
	"fmt"
	"math/big"
)

// Lemke solves the linear complementarity probelm via Lemke's algorithm.
// Returns nil with an error if ray termination
func (lcp *LCP) Lemke(d []*big.Rat) ([]*big.Rat, error) {
	return lcp.LemkeWithPivotMax(d, 0)
}

// LemkeWithPivotMax solves the linear complementarity probelm via Lemke's algorithm.
// It will only perform up to maxCount pivots before exiting.
func (lcp *LCP) LemkeWithPivotMax(d []*big.Rat, maxCount int) ([]*big.Rat, error) {

	tableau, scaleFactors, err := lcp.createTableau(d)
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
			break  // ray termination...
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
		z[i] = result(tableau.vars.z(i + 1), den, tableau, scaleFactors)
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
