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
func lexminratio(tableau *tableau, enter *tableauVariable) (*tableauVariable, bool, error) {

	var err error
	z0leave := false
	leaveCandidateRows := make([]int, 0, tableau.vars.n)

	if enter.isBasic() {
		panic(fmt.Sprintf("Variable %v is already in basis. Must be cobasic to enter.", enter))
	}

	enterCol := enter.col()

	// start with  leavecand = { i | A[i][col] > 0 }
	for i := 0; i < tableau.vars.n; i++ {
		if tableau.entry(i, enterCol).Sign() > 0 {
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

	leaveCandidateRows, z0leave = processCandidates(tableau, enterCol, leaveCandidateRows)

	return tableau.vars.fromRow(leaveCandidateRows[0]), z0leave, err
}

/*
 * processCandidates
 * ================================================================
 * as long as there is more than one leaving candidate perform
 * a minimum ratio test for the columns of  j  in RHS, W(1),... W(n)
 * in the tableau.  That test has an easy known result if
 * the test column is basic or equal to the entering variable.
 */
func processCandidates(tableau *tableau, enterCol int, leaveCandidateRows []int) ([]int, bool) {

	leaveCandidateRows = minRatioTest(tableau, enterCol, tableau.ncols-1, leaveCandidateRows)
	z0leave := checkForZ0(tableau.vars, leaveCandidateRows)
	/* alternative, to force z0 leaving the basis:
	* return whichvar[leavecand[i]];
	 */

	for j := 1; len(leaveCandidateRows) > 1; j++ {
		//if j >= A.RHS() {                                             /* impossible, perturbed RHS should have full rank */
		//    throw new RuntimeException("lex-minratio test failed"); //TODO
		//}

		wj := tableau.vars.w(j)
		fmt.Printf("Checking leave candidate %s\n", wj)
		if wj.isBasic() { /* testcol < 0: W(j) basic, Eliminate its row from leavecand */
			fmt.Printf("%s is basic... removing from candidate\n", wj)
			leaveCandidateRows = remove(leaveCandidateRows, wj.row())
		} else { // not a basic testcolumn: perform minimum ratio tests
			testCol := wj.col()      /* since testcol is the  jth  unit column                    */
			if testCol != enterCol { /* otherwise nothing will change */
				leaveCandidateRows = minRatioTest(tableau, enterCol, testCol, leaveCandidateRows)
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

func checkForZ0(vars *tableauVariables, leaveCandidateRows []int) bool {
	for i := 0; i < len(leaveCandidateRows); i++ { // seek  z0  among the first-col leaving candidates
		if vars.fromRow(leaveCandidateRows[i]).isZ0() {
			return true
		}
	}
	return false
}

func minRatioTest(tableau *tableau, enterCol int, testCol int, candidateRows []int) []int {

	numCandidates := 0
	for i := 1; i < len(candidateRows); i++ { /* investigate remaining candidates                  */

		// sign of  A[l_0,t] / A[l_0,col] - A[l_i,t] / A[l_i,col]
		// note only positive entries of entering column considered
		sgn := tableau.ratioTest(
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
