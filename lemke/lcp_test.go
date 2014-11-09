package lemke

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariableAssignments(t *testing.T) {

	lcp := &LCP{n: 4}
	lcp.initVars()

	for i := 0; i <= lcp.n; i++ {
		zi := lcp.z(i)
		assert.Equal(t, i, zi, "z(i) should equal i")
		assert.Equal(t, i, lcp.var2col(zi), "col(z(i)) should equal i")
		assert.Equal(t, zi, lcp.col2var(i), "var(col(z(i))) = var(i) should equal z(i)")
		assert.Equal(t, false, lcp.isBasicVar(zi), "z(i) should NOT be basic")
	}

	for i := 1; i <= lcp.n; i++ {
		wi := lcp.w(i)
		assert.Equal(t, i+lcp.n, wi, "w(i) should equal i + n")
		assert.Equal(t, i-1, lcp.var2row(wi), "row(w(i)) should equal i - 1 (w is 1-indexed)")
		assert.Equal(t, wi, lcp.row2var(i-1), "var(row(w(i - 1))) = var(i - 1) should equal w(i)")
		assert.Equal(t, true, lcp.isBasicVar(wi), "w(i) should be basic")
	}
}

func TestSwap(t *testing.T) {

	lcp := &LCP{n: 4}
	lcp.initVars()

	leaveVar := lcp.w(1) // first row, w(1) = 4
	assert.Equal(t, true, lcp.isBasicVar(leaveVar), "leaving var should be basic")
	assert.Equal(t, 0, lcp.var2row(leaveVar), "w(1) points to row 0")

	enterVar := lcp.z(0) // first col, z(0) = 0
	assert.Equal(t, false, lcp.isBasicVar(enterVar), "entering var should be cobasic")
	assert.Equal(t, 0, lcp.var2col(enterVar), "z(0) points to col 0")

	row, col := lcp.swap(enterVar, leaveVar)

	assert.Equal(t, 0, col, "col should be original col(w1)")
	assert.Equal(t, 0, row, "row should be original row(z0)")

	assert.Equal(t, leaveVar, lcp.col2var(col), "w1 should be var(col)")
	assert.Equal(t, enterVar, lcp.row2var(row), "z0 should be var(row)")
}

func TestComplement(t *testing.T) {

	lcp := &LCP{n: 4}
	lcp.initVars()

	for i := 1; i <= lcp.n; i++ {
		compVar, _ := lcp.complement(lcp.z(i))
		assert.Equal(t, lcp.w(i), compVar, "w(i) should be complement of z(i)")
	}

	_, err := lcp.complement(lcp.z(0))
	assert.NotNil(t, err, "Should not be able to get complement of z0")
}
