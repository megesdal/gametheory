package lemke

import (
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLexMinVar(t *testing.T) {

	n := 2
	a := newTableau(n)

	a.set(0, 0, big.NewInt(2))
	a.set(0, 1, big.NewInt(2))
	a.set(0, 2, big.NewInt(1))
	a.set(0, 3, big.NewInt(-1))
	a.set(1, 0, big.NewInt(1))
	a.set(1, 1, big.NewInt(1))
	a.set(1, 2, big.NewInt(3))
	a.set(1, 3, big.NewInt(-1))

	vars := newTableauVariables(n)
	// TODO: if I give lexmin a bascobas lookup function instead of full lcp, that should suffice
	leave, z0leave, err := lexminratio(a, a.vars.z(0))
	assert.Equal(t, vars.w(2), leave, "w2 = 4 is leaving")
	assert.Equal(t, false, z0leave, "z0 is not leaving")
	assert.Nil(t, err, "No ray termination")

	leave, z0leave, err = lexminratio(a, a.vars.z(1))
	assert.Equal(t, vars.w(2), leave, "w2 = 4 is leaving")
	assert.Equal(t, false, z0leave, "z0 is not leaving")
	assert.Nil(t, err, "No ray termination")

	leave, z0leave, err = lexminratio(a, a.vars.z(2))
	assert.Equal(t, vars.w(1), leave, "w1 = 3 is leaving")
	assert.Equal(t, false, z0leave, "z0 is not leaving")
	assert.Nil(t, err, "No ray termination")

	// TODO: expect panic...
	/*_, _, err = lexminratio(a, vars, vars.w(1))
	assert.NotNil(t, err, "Should be an error")
	assert.Equal(t, "Variable w1 is already in basis. Must be cobasic to enter.", err.Error())

	_, _, err = lexminratio(a, vars, vars.w(2))
	assert.NotNil(t, err, "Should be an error")
	assert.Equal(t, "Variable w2 is already in basis. Must be cobasic to enter.", err.Error())
	*/
}

// TODO: make this a benchmark test...
func Test1000LexMinVarOnLargeTableu(t *testing.T) {

	n := 1000
	a := newTableau(n)

	for i := 0; i < a.nrows; i++ {
		for j := 0; j < a.ncols; j++ {
			if j == 0 {
				a.set(i, j, big.NewInt(1))
			} else {
				a.set(i, j, big.NewInt(int64((i-(j-1))*((j*17)-(i*63)))))
			}
		}
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		leave, z0leave, err := lexminratio(a, a.vars.z(0))
		assert.Equal(t, 1001, leave.idx)
		assert.Equal(t, false, z0leave)
		assert.Nil(t, err)
	}

	duration := time.Since(start)
	log.Println("1000 lexmin took:", duration)
	assert.Equal(t, true, duration.Seconds() < 1) // less than 1 seconds
}

/*
 * Test row candidate elimination using sign of:
 * A[l_0,t] / A[l_0,col] - A[l_i,t] / A[l_i,col]
 */
func TestMinRatioTest(t *testing.T) {

	A := newTableau(2)
	A.set(0, 0, big.NewInt(int64(2)))
	A.set(0, 1, big.NewInt(int64(2)))
	A.set(0, 2, big.NewInt(int64(1)))
	A.set(0, 3, big.NewInt(int64(-1)))
	A.set(1, 0, big.NewInt(int64(1)))
	A.set(1, 1, big.NewInt(int64(1)))
	A.set(1, 2, big.NewInt(int64(3)))
	A.set(1, 3, big.NewInt(int64(-1)))

	candidates := []int{0, 1}

	col1 := 1
	testcol1 := 2
	sgn1 := A.ratioTest(candidates[0], candidates[1], col1, testcol1)
	assert.Equal(t, -1, sgn1, "A[0,2] / A[0,1] - A[1,2] / A[1,1] should be 1/2 - 3/1 = -5/2")

	newCandidates1 := minRatioTest(A, col1, testcol1, candidates)
	assert.Equal(t, 1, len(newCandidates1))
	assert.Equal(t, 0, newCandidates1[0])

	col2 := 2
	testcol2 := 1
	sgn2 := A.ratioTest(candidates[0], candidates[1], col2, testcol2)
	assert.Equal(t, 1, sgn2, "A[0,1] / A[0,2] - A[1,1] / A[1,2] should be 2/1 - 1/3 = 5/3")

	newCandidates2 := minRatioTest(A, col2, testcol2, candidates)
	assert.Equal(t, 1, len(newCandidates2))
	assert.Equal(t, 1, newCandidates2[0])
}
