package lemke

import "fmt"

type tableauVariable struct {
	idx int
	s   *tableauVariables
}

func (tvar *tableauVariable) row() int {
	// TODO: panic if not basic?
	return tvar.s.toRowCol[tvar.idx]
}

func (tvar *tableauVariable) col() int {
	// TODO: panic if not co-basic?
	return tvar.s.toRowCol[tvar.idx] - tvar.s.n
}

func (tvar *tableauVariable) isBasic() bool {
	return tvar.s.toRowCol[tvar.idx] < tvar.s.n
}

func (tvar *tableauVariable) isZ() bool {
	return tvar.idx <= tvar.s.n
}

func (tvar *tableauVariable) isZ0() bool {
	return tvar.idx == 0
}

/*
 * complement of  v  in VARS, error if  v==Z(0).
 * this is  W(i) for Z(i)  and vice versa, i=1...n
 */
func (tvar *tableauVariable) complement() *tableauVariable {

	if tvar.idx == 0 {
		panic("Attempt to find complement of z0.")
	}

	if tvar.isZ() {
		return &tvar.s.lookup[tvar.idx+tvar.s.n]
	}
	return &tvar.s.lookup[tvar.idx-tvar.s.n]
}

func (tvar *tableauVariable) String() string {
	if tvar.isZ() {
		return fmt.Sprintf("z%d", tvar.idx)
	}
	return fmt.Sprintf("w%d", tvar.idx-tvar.s.n)
}

/* tableauVariables
 * init tableau variables:
 * Z(0)...Z(n)  nonbasic,  W(1)...W(n) basic
 * This is for setting up a complementary basis/cobasis
 *
 */
type tableauVariables struct {
	/*  v in VARS, v cobasic:  TABCOL(v) is v's tableau col */
	/*  v  basic:  TABCOL(v) < 0,  TABCOL(v)+n   is v's row */
	/* VARS   = 0..2n = Z(0) .. Z(n) W(1) .. W(n)           */
	/* ROWCOL = 0..2n,  0 .. n-1: tabl rows (basic vars)    */
	/*                  n .. 2n:  tabl cols  0..n (cobasic) */
	lookup     []tableauVariable
	toRowCol   []int
	fromRowCol []int
	n          int
}

func newTableauVariables(n int) *tableauVariables {

	vars := tableauVariables{
		lookup:     make([]tableauVariable, 2*n+1),
		toRowCol:   make([]int, 2*n+1),
		fromRowCol: make([]int, 2*n+1),
		n:          n,
	}

	for i := 0; i <= n; i++ {
		vars.lookup[i] = tableauVariable{i, &vars}
		vars.toRowCol[i] = i + n
		vars.fromRowCol[i+n] = i
	}

	for i := 1; i <= n; i++ {
		vars.lookup[i+n] = tableauVariable{i + n, &vars}
		vars.toRowCol[i+n] = i - 1
		vars.fromRowCol[i-1] = i + n
	}

	return &vars
}

func (vars *tableauVariables) z(subscript int) *tableauVariable {
	return &vars.lookup[subscript]
}

func (vars *tableauVariables) w(subscript int) *tableauVariable {
	return &vars.lookup[subscript+vars.n]
}

func (vars *tableauVariables) fromRow(row int) *tableauVariable {
	return &vars.lookup[vars.fromRowCol[row]]
}

func (vars *tableauVariables) fromCol(col int) *tableauVariable {
	return &vars.lookup[vars.fromRowCol[col+vars.n]]
}

func (vars *tableauVariables) swap(enter *tableauVariable, leave *tableauVariable) (int, int) {

	leaveRow := leave.row() // basic var is leaving
	enterCol := enter.col() // cobasic var is entering

	vars.toRowCol[leave.idx] = enterCol + vars.n
	vars.fromRowCol[enterCol+vars.n] = leave.idx

	vars.toRowCol[enter.idx] = leaveRow
	vars.fromRowCol[leaveRow] = enter.idx

	return leaveRow, enterCol
}
