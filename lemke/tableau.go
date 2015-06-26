package lemke

import (
	"math/big"
)

type tableau struct {
	matrix []*big.Int
	ncols  int
	nrows  int
	det    *big.Int // determinant
}

func newTableau(nrows int, ncols int) *tableau {
	tableau := &tableau{}
	tableau.nrows = nrows
	tableau.ncols = ncols
	tableau.matrix = make([]*big.Int, nrows*ncols)

	tableau.det = big.NewInt(-1) // TODO: how do I know this? Specific to LCP?
	return tableau
}

func (A *tableau) set(row int, col int, value *big.Int) {
	A.matrix[row*A.ncols+col] = value
}

func (A *tableau) entry(row int, col int) *big.Int {
	return A.matrix[row*A.ncols+col]
}

func isPositive(num *big.Int) bool {
	return num.Sign() > 0
}

func isNegative(num *big.Int) bool {
	return num.Sign() < 0
}

func isZero(num *big.Int) bool {
	return num.Sign() == 0
}

func (A *tableau) pivotOnRowCol(row int, col int) {

	pivelt := A.entry(row, col) /* pivelt anyhow later new determinant  */

	if pivelt.Sign() == 0 {
		panic("Trying to pivot on a zero")
	}

	negpiv := false
	if pivelt.Sign() < 0 {
		negpiv = true
		pivelt.Neg(pivelt)
	}

	for i := 0; i < A.nrows; i++ {
		if i != row { // A[row][..] remains unchanged
			entry := A.entry(i, col)
			nonzero := entry.Sign() != 0
			for j := 0; j < A.ncols; j++ {
				if j != col {

					//A[i,j] = (A[i,j] A[row,col] - A[i,col] A[row,j]) / det
					entryIJ := A.entry(i, j)
					entryIJ.Mul(entryIJ, pivelt)
					if nonzero {
						tmp := new(big.Int).Mul(entry, A.entry(row, j))
						if negpiv {
							entryIJ.Add(entryIJ, tmp)
						} else {
							entryIJ.Sub(entryIJ, tmp)
						}
					}
					entryIJ.Div(entryIJ, A.det)
				}
			}
			if nonzero && !negpiv {
				/* row  i  has been dealt with, update  A[i][col]  safely   */
				entry.Neg(entry)
			}
		}
	}

	A.set(row, col, A.det)
	if negpiv {
		A.negateRow(row)
	}

	A.det = pivelt //by construction always positive
}

/*
 * sign of  A[a,testcol] / A[a,col] - A[b,testcol] / A[b,col]
 * (assumes only positive entries of col are considered)
 */
func (A *tableau) ratioTest(rowA int, rowB int, colA int, colB int) int {
	a := new(big.Int).Mul(A.entry(rowA, colB), A.entry(rowB, colA))
	b := new(big.Int).Mul(A.entry(rowB, colB), A.entry(rowA, colA))
	return a.Cmp(b)
}

func (A *tableau) negateRow(row int) {
	for j := 0; j < A.ncols; j++ {
		entry := A.entry(row, j)
		if entry.Sign() != 0 {
			entry.Neg(entry)
		}
	}
}

func (A *tableau) negateCol(col int) {
	for i := 0; i < A.nrows; i++ {
		entry := A.entry(i, col)
		if entry.Sign() != 0 {
			entry.Neg(entry)
		}
	}
}
