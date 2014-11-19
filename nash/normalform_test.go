package nash

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLemke(t *testing.T) {

	var payMatrix [][][]float64
	json.Unmarshal([]byte(`
        [ [ [ 11, 3 ], [ 3, 0 ], [ 11, 3 ], [  3, 0 ] ],
          [ [  0, 2 ], [ 0, 7 ], [ 12, 0 ], [ 12, 5 ] ],
		  [ [  6, 0 ], [ 6, 0 ], [  0, 1 ], [  0, 1 ] ] ]`), &payMatrix)

	eq, err := LemkeEquilibrium(payMatrix, int64(1))
	assert.Nil(t, err)
	assert.Equal(t, "rows 0/1 1/3 2/3=4/1\ncols 0/1 2/3 0/1 1/3=7/3", eq.String())

	eq, err = LemkeEquilibrium(payMatrix, int64(2))
	assert.Nil(t, err)
	assert.Equal(t, "rows 1/1 0/1 0/1=11/1\ncols 1/1 0/1 0/1 0/1=3/1", eq.String())
}
