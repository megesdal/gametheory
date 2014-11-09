package lemke

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLemkeRun2(t *testing.T) {

	M := ints2rats([]int{2, 1, 1, 3})
	q := ints2rats([]int{-1, -1})
	d := ints2rats([]int{2, 1})

	lcp, err := NewLCP(M, q)
	assert.Nil(t, err)

	z, err := lcp.Lemke(d)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(z))
	assert.Equal(t, false, z[0].IsInt())
	assert.Equal(t, int64(2), z[0].Num().Int64())
	assert.Equal(t, int64(5), z[0].Denom().Int64())
	assert.Equal(t, false, z[1].IsInt())
	assert.Equal(t, int64(1), z[1].Num().Int64())
	assert.Equal(t, int64(5), z[1].Denom().Int64())
}

func TestLemkeRun3(t *testing.T) {

	M := ints2rats([]int{0, -1, 2, 2, 0, -2, -1, 1, 0})
	q := ints2rats([]int{-3, 6, -1})
	d := ints2rats([]int{1, 1, 1})

	lcp, err := NewLCP(M, q)
	assert.Nil(t, err)

	z, err := lcp.Lemke(d)
	assert.Nil(t, err)

	//StringWriter output = new StringWriter();
	//LemkeWriter lemkeWriter = new LemkeWriter(output, output);
	//lemkeWriter.outtabl(algo.A);
	//Assert.True(false, output.ToString());

	assert.Equal(t, 3, len(z))
	assert.Equal(t, true, z[0].IsInt())
	assert.Equal(t, int64(0), z[0].Num().Int64())
	assert.Equal(t, true, z[1].IsInt())
	assert.Equal(t, int64(1), z[1].Num().Int64())
	assert.Equal(t, true, z[2].IsInt())
	assert.Equal(t, int64(3), z[2].Num().Int64())
}

func ints2rats(ints []int) []*big.Rat {
	rats := make([]*big.Rat, len(ints))
	for i := 0; i < len(ints); i++ {
		rats[i] = big.NewRat(int64(ints[i]), int64(1))
	}
	return rats
}
