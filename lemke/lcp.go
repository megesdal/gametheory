package lemke

import (
	"fmt"
	"math/big"
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
type LCP struct {
	m []*big.Rat
	q []*big.Rat
	n int
}

func NewLCP(M []*big.Rat, q []*big.Rat) *LCP {

	n := len(q)
	if len(M)%n != 0 {
		panic("M.rows and q are not same dimensions")
	}

	if len(M)/n != n {
		panic(fmt.Sprintf("M must be a square matrix but was %dx%d", len(M)/n, n))
	}

	fmt.Printf("Creating LCP of dimenstion n=%d\n", n)

	return &LCP{m: M, q: q, n: n}
}

func (lcp *LCP) M(i int, j int) *big.Rat {
	return lcp.m[i*lcp.n+j]
}

func (lcp *LCP) Q(i int) *big.Rat {
	return lcp.q[i]
}
