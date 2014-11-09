package lemke

import (
	"math/big"
	"testing"

	"github.com/bmizerany/assert"
)

func TestAssignment(t *testing.T) {

	n := 2
	A := newTableau(n, n+2)

	A.set(0, 0, big.NewInt(2))
	A.set(0, 1, big.NewInt(2))
	A.set(0, 2, big.NewInt(1))
	A.set(0, 3, big.NewInt(-1))
	A.set(1, 0, big.NewInt(1))
	A.set(1, 1, big.NewInt(1))
	A.set(1, 2, big.NewInt(3))
	A.set(1, 3, big.NewInt(-1))

	assert.Equal(t, big.NewInt(2), A.entry(0, 0), "A[0][0] = 2")
	assert.Equal(t, big.NewInt(2), A.entry(0, 1), "A[0][1] = 2")
	assert.Equal(t, big.NewInt(1), A.entry(0, 2), "A[0][2] = 2")
	assert.Equal(t, big.NewInt(-1), A.entry(0, 3), "A[0][3] = 2")
	assert.Equal(t, big.NewInt(1), A.entry(1, 0), "A[1][0] = 2")
	assert.Equal(t, big.NewInt(1), A.entry(1, 1), "A[1][1] = 1")
	assert.Equal(t, big.NewInt(3), A.entry(1, 2), "A[1][2] = 3")
	assert.Equal(t, big.NewInt(-1), A.entry(1, 3), "A[1][3] = -1")
}

func TestPosPivot(t *testing.T) {

	n := 2
	A := newTableau(n, n+2)
	for i := 0; i < A.nrows; i++ {
		for j := 0; j < A.ncols; j++ {
			value := big.NewInt(int64((i + 1) + j*10))
			A.set(i, j, value)
		}
	}

	assert.Equal(t, int64(1), A.entry(0, 0).Int64())
	assert.Equal(t, int64(11), A.entry(0, 1).Int64())
	assert.Equal(t, int64(2), A.entry(1, 0).Int64())
	assert.Equal(t, int64(12), A.entry(1, 1).Int64())

	A.pivotOnRowCol(0, 0)

	assert.Equal(t, int64(-1), A.entry(0, 0).Int64())
	assert.Equal(t, int64(11), A.entry(0, 1).Int64())
	assert.Equal(t, int64(-2), A.entry(1, 0).Int64())
	assert.Equal(t, int64(10), A.entry(1, 1).Int64())
}

func TestNegCol(t *testing.T) {

	n := 3
	A := newTableau(n, n+2)
	for i := 0; i < A.nrows; i++ {
		for j := 0; j < A.ncols; j++ {
			value := big.NewInt(int64(i + j*10))
			A.set(i, j, value)
		}
	}

	A.negateCol(1)

	assert.Equal(t, int64(20), A.entry(0, 2).Int64())
	assert.Equal(t, int64(-10), A.entry(0, 1).Int64())
	assert.Equal(t, int64(-11), A.entry(1, 1).Int64())
	assert.Equal(t, int64(-12), A.entry(2, 1).Int64())
}

func TestPositiveValuesRatioTest(t *testing.T) {

	n := 2
	A := newTableau(n, n+2)
	for i := 0; i < A.nrows; i++ {
		for j := 0; j < A.ncols; j++ {
			value := big.NewInt(int64((i + 1) + j*10))
			A.set(i, j, value)
		}
	}

	assert.Equal(t, 1, A.ratioTest(0, 1, 0, 1), "A[0,1] / A[0,0] - A[1,1] / A[1,0] = 5")
	assert.Equal(t, -1, A.ratioTest(1, 0, 0, 1), "A[1,1] / A[1,0] - A[0,1] / A[0,0] = -5")
}
