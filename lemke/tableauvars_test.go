package lemke

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestVariableAssignments(t *testing.T) {

	vars := newTableauVariables(4)

	for i := 0; i <= vars.n; i++ {
		zi := vars.z(i)
		assert.Equal(t, i, zi.idx, "z(i) should equal i")
		assert.Equal(t, i, zi.col(), "col(z(i)) should equal i")
		assert.Equal(t, zi.idx, vars.fromCol(i).idx, "var(col(z(i))) = var(i) should equal z(i)")
		assert.Equal(t, false, zi.isBasic(), "z(i) should NOT be basic")
	}

	for i := 1; i <= vars.n; i++ {
		wi := vars.w(i)
		assert.Equal(t, i+vars.n, wi.idx, "w(i) should equal i + n")
		assert.Equal(t, i-1, wi.row(), "row(w(i)) should equal i - 1 (w is 1-indexed)")
		assert.Equal(t, wi.idx, vars.fromRow(i-1).idx, "var(row(w(i - 1))) = var(i - 1) should equal w(i)")
		assert.Equal(t, true, wi.isBasic(), "w(i) should be basic")
	}
}

func TestSwap(t *testing.T) {

	vars := newTableauVariables(4)

	leaveVar := vars.w(1) // first row, w(1) = 4
	assert.Equal(t, true, leaveVar.isBasic(), "leaving var should be basic")
	assert.Equal(t, 0, leaveVar.row(), "w(1) points to row 0")

	enterVar := vars.z(0) // first col, z(0) = 0
	assert.Equal(t, false, enterVar.isBasic(), "entering var should be cobasic")
	assert.Equal(t, 0, enterVar.col(), "z(0) points to col 0")

	row, col := vars.swap(enterVar, leaveVar)

	assert.Equal(t, 0, col, "col should be original col(w1)")
	assert.Equal(t, 0, row, "row should be original row(z0)")

	assert.Equal(t, leaveVar, vars.fromCol(col), "w1 should be var(col)")
	assert.Equal(t, enterVar, vars.fromRow(row), "z0 should be var(row)")
}

func TestComplement(t *testing.T) {

	vars := newTableauVariables(4)

	for i := 1; i <= vars.n; i++ {
		compVar := vars.z(i).complement()
		assert.Equal(t, vars.w(i).idx, compVar.idx, "w(i) should be complement of z(i)")
	}

	// TODO: assert panic?
	// _, err := lcp.complement(lcp.z(0))
	// assert.NotNil(t, err, "Should not be able to get complement of z0")
}
