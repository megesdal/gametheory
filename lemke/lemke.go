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

	err := checkInputs(lcp.q, d)
	if err != nil {
		return nil, err
	}

	lcp.addCoveringVector(d)

	nextLeavingVar := func(enter *tableauVariable) (*tableauVariable, bool, error) {
		return lexminratio(lcp.a, lcp.vars, enter)
	}

	enter := lcp.vars.z(0) // z0 enters the basis to obtain lex-feasible solution
	leave, z0leave, err := nextLeavingVar(enter)

	lcp.negateRHS() // now give the entering q-col its correct sign

	pivotCount := 1
	for {

		fmt.Printf("%d LCP:\n%v", pivotCount, lcp)
		fmt.Println(pivotCount, "entering", enter, "leaving", leave)

		lcp.pivot(leave, enter)

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

	fmt.Printf("LCP (final):\n%v", lcp)
	return lcp.solution(), err // LCP solution = z  vector
}
