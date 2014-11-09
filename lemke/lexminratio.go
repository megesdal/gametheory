package lemke

import "fmt"

/*
 * minVar
 * ===========================================================
 * @return the leaving variable in  VARS, given by lexmin row,
 * when  enter  in VARS is entering variable
 * only positive entries of entering column tested
 * boolean  *z0leave  indicates back that  z0  can leave the
 * basis, but the lex-minratio test is performed fully,
 * so the returned value might not be the index of  z0
 */
func lexminratio(lcp *LCP, enter bascobas) (bascobas, bool, error) {

	var err error
	z0leave := false
	leaveCandidateRows := make([]int, 0, lcp.n)

	if lcp.isBasicVar(enter) {
		return enter, z0leave, fmt.Errorf("Variable %v is already in basis. Must be cobasic to enter.", lcp.var2str(enter))
	}

	enterCol := lcp.var2col(enter)

	// start with  leavecand = { i | A[i][col] > 0 }
	for i := 0; i < lcp.n; i++ {
		if isPositive(lcp.a.entry(i, enterCol)) {
			leaveCandidateRows = append(leaveCandidateRows, i)
		}
	}

	if len(leaveCandidateRows) == 0 {
		return enter, z0leave, fmt.Errorf("Ray termination when trying to enter %s", enter)
	}

	/*else if (numcand == 1) {
		RecordStats(0, numcand);
		z0leave = IsLeavingRowZ0(leavecand[0]);
	}*/

	leaveCandidateRows, z0leave = processCandidates(lcp, enterCol, leaveCandidateRows)

	return lcp.row2var(leaveCandidateRows[0]), z0leave, err
}

/*
 * processCandidates
 * ================================================================
 * as long as there is more than one leaving candidate perform
 * a minimum ratio test for the columns of  j  in RHS, W(1),... W(n)
 * in the tableau.  That test has an easy known result if
 * the test column is basic or equal to the entering variable.
 */
func processCandidates(lcp *LCP, enterCol int, leaveCandidateRows []int) ([]int, bool) {

	var z0leave bool
	leaveCandidateRows, z0leave = processRHS(lcp, enterCol, leaveCandidateRows)
	for j := 1; len(leaveCandidateRows) > 1; j++ {
		//if j >= A.RHS() {                                             /* impossible, perturbed RHS should have full rank */
		//    throw new RuntimeException("lex-minratio test failed"); //TODO
		//}

		wj := lcp.w(j)
		if lcp.isBasicVar(wj) { /* testcol < 0: W(j) basic, Eliminate its row from leavecand */
			leaveCandidateRows = remove(leaveCandidateRows, lcp.var2row(wj))
		} else { // not a basic testcolumn: perform minimum ratio tests
			testCol := lcp.var2col(wj) /* since testcol is the  jth  unit column                    */
			if testCol != enterCol {   /* otherwise nothing will change */
				leaveCandidateRows = minRatioTest(lcp.a, enterCol, testCol, leaveCandidateRows)
			}
		}
	}

	return leaveCandidateRows, z0leave
}

func remove(slice []int, value int) []int {
	for i := 0; i < len(slice); i++ {
		if slice[i] == value {
			slice[i] = slice[len(slice)-1] // shuffling of leavecand allowed
			slice = slice[:len(slice)-1]
			break
		}
	}
	return slice
}

func processRHS(lcp *LCP, enterCol int, leaveCandidateRows []int) ([]int, bool) {

	leaveCandidateRows = minRatioTest(lcp.a, enterCol, lcp.n+1, leaveCandidateRows)

	z0leave := false

	for i := 0; i < len(leaveCandidateRows); i++ { // seek  z0  among the first-col leaving candidates
		z0leave = lcp.row2var(leaveCandidateRows[i]) == lcp.z(0)
		if z0leave {
			break
		}
		/* alternative, to force z0 leaving the basis:
		 * return whichvar[leavecand[i]];
		 */
	}

	return leaveCandidateRows, z0leave
}

func minRatioTest(A *tableau, enterCol int, testCol int, candidateRows []int) []int {

	numCandidates := 0
	for i := 1; i < len(candidateRows); i++ { /* investigate remaining candidates                  */

		// sign of  A[l_0,t] / A[l_0,col] - A[l_i,t] / A[l_i,col]
		// note only positive entries of entering column considered
		sgn := A.ratioTest(
			candidateRows[0], candidateRows[i], enterCol, testCol)

		if sgn == 0 {
			// new ratio is the same as before
			numCandidates++
			candidateRows[numCandidates] = candidateRows[i]
		} else if sgn == 1 {
			// new smaller ratio detected
			numCandidates = 0
			candidateRows[numCandidates] = candidateRows[i]
		}
	}

	return candidateRows[:numCandidates+1]
}
