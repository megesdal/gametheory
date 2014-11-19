package lemke

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/megesdal/gametheory/util"
)

type bascobas int

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

	/*  v in VARS, v cobasic:  TABCOL(v) is v's tableau col */
	/*  v  basic:  TABCOL(v) < 0,  TABCOL(v)+n   is v's row */
	/* VARS   = 0..2n = Z(0) .. Z(n) W(1) .. W(n)           */
	/* ROWCOL = 0..2n,  0 .. n-1: tabl rows (basic vars)    */
	/*                  n .. 2n:  tabl cols  0..n (cobasic) */
	vars2rowcol []int
	rowcol2vars []bascobas
}

func NewLCP(M []*big.Rat, q []*big.Rat) (*LCP, error) {

	if len(M)%len(q) != 0 {
		return nil, errors.New("M and q are not right dimensions")
	}

	ncols := len(M) / len(q)
	nrows := len(q)
	log.Printf("Creating LCP with matrix M [%dx%d] and vector q [%dx1]\n", nrows, ncols, nrows)

	if ncols != nrows {
		return nil, fmt.Errorf("M must be a square matrix but was %dx%d", nrows, ncols)
	}

	lcp := &LCP{m: M, q: q, n: len(q)}
	lcp.initVars()
	lcp.initTableau()

	return lcp, nil
}

func (lcp *LCP) M(i int, j int) *big.Rat {
	return lcp.m[i*lcp.n+j]
}

func (lcp *LCP) Q(i int) *big.Rat {
	return lcp.q[i]
}

/*
 * init tableau variables:
 * Z(0)...Z(n)  nonbasic,  W(1)...W(n) basic
 * This is for setting up a complementary basis/cobasis
 */
func (lcp *LCP) initVars() {

	lcp.vars2rowcol = make([]int, 2*lcp.n+1)
	lcp.rowcol2vars = make([]bascobas, 2*lcp.n+1)

	for i := 0; i <= lcp.n; i++ {
		lcp.vars2rowcol[i] = lcp.n + i
		lcp.rowcol2vars[lcp.n+i] = bascobas(i)
	}

	for i := 1; i <= lcp.n; i++ {
		lcp.vars2rowcol[i+lcp.n] = i - 1
		lcp.rowcol2vars[i-1] = bascobas(i + lcp.n)
	}
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

func (lcp *LCP) swap(enter bascobas, leave bascobas) (int, int) {

	leaveRow := lcp.var2row(leave) // basic var is leaving
	enterCol := lcp.var2col(enter) // cobasic var is entering

	lcp.vars2rowcol[leave] = enterCol + lcp.n
	lcp.rowcol2vars[enterCol+lcp.n] = leave

	lcp.vars2rowcol[enter] = leaveRow
	lcp.rowcol2vars[leaveRow] = enter

	return leaveRow, enterCol
}

/*
 * Pivot tableau on the element  A[row][col] which must be nonzero
 * afterwards tableau normalized with positive determinant
 * and updated tableau variables
 * @param leave (r) VAR defining row of pivot element
 * @param enter (s) VAR defining col of pivot element
 */
func (lcp *LCP) pivot(leave bascobas, enter bascobas) error {

	if !lcp.isBasicVar(leave) {
		return fmt.Errorf("%v is not in the basis", lcp.var2str(leave))
	}

	if lcp.isBasicVar(enter) {
		return fmt.Errorf("%v is already in the basis", lcp.var2str(leave))
	}

	row, col := lcp.swap(enter, leave) /* update tableau variables                                  */
	log.Println("pivoting (", row, ",", col, ")")
	return lcp.a.pivotOnRowCol(row, col)
}

/*
 * LCP result
 * current basic solution turned into  solz [0..n-1]
 * note that Z(1)..Z(n)  become indices  0..n-1
 */
func (lcp *LCP) solution() []*big.Rat {

	z := make([]*big.Rat, lcp.n)
	for i := 1; i <= lcp.n; i++ {
		z[i-1] = lcp.result(lcp.z(i))
	}
	return z
}

/*
 * Z(i):  scfa[i]*rhs[row] / (scfa[RHS]*det)
 * W(i):  rhs[row] / (scfa[RHS]*det)
 */
func (lcp *LCP) result(varIdx bascobas) *big.Rat {

	rv := big.NewRat(0, 1)
	if lcp.isBasicVar(varIdx) {

		var scaleFactor *big.Int
		if lcp.isZVar(varIdx) {
			scaleFactor = lcp.scfa[int(varIdx)]
		} else {
			scaleFactor = big.NewInt(1)
		}

		row := lcp.var2row(varIdx)
		num := new(big.Int).Mul(scaleFactor, lcp.rhs(row))
		den := new(big.Int).Mul(lcp.a.det, lcp.scfa[lcp.n+1])

		rv.SetFrac(num, den)
	}

	return rv
}

func (lcp *LCP) z(idx int) bascobas {
	return bascobas(idx)
}

func (lcp *LCP) w(idx int) bascobas {
	return bascobas(idx + lcp.n)
}

func (lcp *LCP) row2var(row int) bascobas {
	return lcp.rowcol2vars[row]
}

func (lcp *LCP) col2var(col int) bascobas {
	return lcp.rowcol2vars[col+lcp.n]
}

func (lcp *LCP) var2row(varIdx bascobas) int {
	return lcp.vars2rowcol[varIdx]
}

func (lcp *LCP) var2col(varIdx bascobas) int {
	return lcp.vars2rowcol[varIdx] - lcp.n
}

func (lcp *LCP) var2str(varIdx bascobas) string {
	if lcp.isZVar(varIdx) {
		return fmt.Sprintf("z%d", varIdx)
	}
	return fmt.Sprintf("w%d", int(varIdx)-lcp.n)
}

func (lcp *LCP) isBasicVar(varIdx bascobas) bool {
	return (lcp.vars2rowcol[varIdx] < lcp.n)
}

func (lcp *LCP) isZVar(varIdx bascobas) bool {
	return (int(varIdx) <= lcp.n)
}

/*
 * complement of  v  in VARS, error if  v==Z(0).
 * this is  W(i) for Z(i)  and vice versa, i=1...n
 */
func (lcp *LCP) complement(varIdx bascobas) (bascobas, error) {

	if varIdx == lcp.z(0) {
		return varIdx, errors.New("Attempt to find complement of z0.")
	}

	if lcp.isZVar(varIdx) {
		return bascobas(int(varIdx) + lcp.n), nil
	}
	return bascobas(int(varIdx) - lcp.n), nil
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
			matrixPrinter.Colpr(lcp.var2str(lcp.col2var(j)))
		}
	}
	matrixPrinter.Colnl()

	for i := 0; i < lcp.a.nrows; i++ {
		matrixPrinter.Colpr(lcp.var2str(lcp.row2var(i)))
		for j := 0; j < lcp.a.ncols; j++ {
			matrixPrinter.Colpr(lcp.a.entry(i, j).String())
		}
		matrixPrinter.Colnl()
	}

	var buffer bytes.Buffer
	matrixPrinter.Colout(&buffer)
	return buffer.String()
}
