package lemke

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/megesdal/gametheory/util"
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
// TODO: Does M have to be square?  We seem to make that assumption?
type LCP struct {
	m []*big.Rat
	q []*big.Rat
	n int

	a *tableau

	/* scale factors for variables z
	 * scfa[Z(0)]   for  d,  scfa[RHS] for  q
	 * scfa[Z(1..n)] for cols of  M
	 * result variables to be multiplied with these
	 */
	scfa []*big.Int

	vars *tableauVariables
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

	lcp := &LCP{m: M, q: q, n: n}
	lcp.vars = newTableauVariables(n)
	lcp.initTableau()

	return lcp
}

func (lcp *LCP) M(i int, j int) *big.Rat {
	return lcp.m[i*lcp.n+j]
}

func (lcp *LCP) Q(i int) *big.Rat {
	return lcp.q[i]
}

func (lcp *LCP) initTableau() {

	lcp.a = newTableau(lcp.n, lcp.n+2)
	lcp.scfa = make([]*big.Int, lcp.a.ncols)

	for j := 1; j < lcp.a.ncols; j++ {

		fnVec := func(i int) *big.Rat {
			if j == lcp.n+1 {
				return lcp.q[i]
			}
			return lcp.m[i*lcp.n+(j-1)]
		}

		scaleFactor := computeScaleFactor(lcp.n, fnVec)
		lcp.scfa[j] = scaleFactor

		for i := 0; i < lcp.a.nrows; i++ {

			rat := fnVec(i)

			/* cols 0..n of  A  contain LHS cobasic cols of  Ax = b     */
			/* where the system is here         -Iw + dz_0 + Mz = -q    */
			/* cols of  q  will be negated after first min ratio test   */
			/* A[i][j] = num * (scfa[j] / den),  fraction is integral       */

			value := new(big.Int).Mul(rat.Num(), scaleFactor)
			value.Div(value, rat.Denom())
			lcp.a.set(i, j, value)
		}
	}
}

func (lcp *LCP) addCoveringVector(d []*big.Rat) {

	fnVec := func(i int) *big.Rat {
		return d[i]
	}

	scaleFactor := computeScaleFactor(lcp.n, fnVec)
	lcp.scfa[0] = scaleFactor

	for i := 0; i < lcp.a.nrows; i++ {
		rat := fnVec(i)
		value := new(big.Int).Mul(rat.Num(), scaleFactor)
		value.Div(value, rat.Denom())
		lcp.a.set(i, 0, value)
	}
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

/*
 * Pivot tableau on the element  A[row][col] which must be nonzero
 * afterwards tableau normalized with positive determinant
 * and updated tableau variables
 * @param leave (r) VAR defining row of pivot element
 * @param enter (s) VAR defining col of pivot element
 */
func (lcp *LCP) pivot(leave *tableauVariable, enter *tableauVariable) {

	if !leave.isBasic() {
		panic(fmt.Sprintf("%v is not in the basis", leave))
	}

	if enter.isBasic() {
		panic(fmt.Sprintf("%v is already in the basis", enter))
	}

	row, col := lcp.vars.swap(enter, leave) /* update tableau variables                                  */
	fmt.Println("pivoting (", row, ",", col, ")")
	lcp.a.pivotOnRowCol(row, col)
}

/*
 * LCP result
 * current basic solution turned into  solz [0..n-1]
 * note that Z(1)..Z(n)  become indices  0..n-1
 */
func (lcp *LCP) solution() []*big.Rat {

	z := make([]*big.Rat, lcp.n)
	for i := 1; i <= lcp.n; i++ {
		z[i-1] = lcp.result(lcp.vars.z(i))
	}
	return z
}

/*
 * Z(i):  scfa[i]*rhs[row] / (scfa[RHS]*det)
 * W(i):  rhs[row] / (scfa[RHS]*det)
 */
func (lcp *LCP) result(tvar *tableauVariable) *big.Rat {

	rv := big.NewRat(0, 1)
	if tvar.isBasic() {

		var scaleFactor *big.Int
		if tvar.isZ() {
			scaleFactor = lcp.scfa[tvar.idx]  // TODO: row?
		} else {
			scaleFactor = big.NewInt(1)
		}

		row := tvar.toRow()
		num := new(big.Int).Mul(scaleFactor, lcp.rhs(row))  // TODO: any BigInt here that I can overwrite at this point?
		den := new(big.Int).Mul(lcp.a.det, lcp.scfa[lcp.n+1]) // TODO: is this denom constant for the whole solution...

		rv.SetFrac(num, den)
	}

	return rv
}

func (lcp *LCP) negateRHS() {
	lcp.a.negateCol(lcp.a.ncols - 1)
}

func (lcp *LCP) rhs(row int) *big.Int {
	return lcp.a.entry(row, lcp.a.ncols-1)
}

func (lcp *LCP) String() string {

	matrixPrinter := util.NewMatrixPrinter()
	matrixPrinter.Colpr("")
	for j := 0; j < lcp.a.ncols; j++ {
		if j == lcp.a.ncols-1 {
			matrixPrinter.Colpr("rhs")
		} else {
			matrixPrinter.Colpr(lcp.vars.fromCol(j).String())
		}
	}
	matrixPrinter.Colnl()

	for i := 0; i < lcp.a.nrows; i++ {
		matrixPrinter.Colpr(lcp.vars.fromRow(i).String())
		for j := 0; j < lcp.a.ncols; j++ {
			matrixPrinter.Colpr(lcp.a.entry(i, j).String())
		}
		matrixPrinter.Colnl()
	}

	var buffer bytes.Buffer
	matrixPrinter.Colout(&buffer)
	return buffer.String()
}
