package gametheory

import (
	"bytes"
	"fmt"
	"github.com/megesdal/matrixprinter"
	"strings"
)

func (sf *SequenceForm) String() string {

	table := matrixprinter.NewTable()

	for _, pl := range sf.plNames {
		table.Append(pl)
		for _, seq := range sf.plSeqs[pl] {
			table.Append(seq)
		}
		/*table.Append("\u2205")
		  for _, iset := range sf.plIsets[pl] {
		    table.Append(iset)
			}*/
		table.EndRow()

		sf.recAppendRow(table, pl, 0, make(map[string]string))

		// then constraint rows... those are easy
		for _, iset := range sf.plIsets[pl] {
			// TODO: empty iset?
			table.Append(iset)
			for _, seq := range sf.plSeqs[pl] {
				td := sf.plConstraints[pl][iset][seq]
				if td == 0 {
					table.Append(".")
				} else {
					table.AppendInt(td)
				}
			}
			table.EndRow()
		}
		table.EndRow()
	}

	/*for i := 0; i < A.nrows; i++ {
		table.Append(A.vars.fromRow(i).String())
		for j := 0; j < A.ncols; j++ {
			table.Append(A.entry(i, j).String())
		}
		table.EndRow()
	}*/

	var buffer bytes.Buffer
	table.Print(&buffer)
	return buffer.String()
}

func (sf *SequenceForm) recAppendRow(table *matrixprinter.Table, pl string, idx int, sequences map[string]string) {

	if idx == len(sf.plNames) {

		otherMoves := make([]string, 0, len(sf.plNames)-1)
		for seqPl, seq := range sequences {
			if seqPl != pl {
				otherMoves = append(otherMoves, seq)
			}
		}

		// base case...
		if len(otherMoves) == 1 {
			table.Append(otherMoves[0])
		} else {
			fmt.Println(strings.Join(otherMoves, ":"))
			panic("not impl yet...")
		}

		key := strings.Join(otherMoves, ":")

		for _, plSeq := range sf.plSeqs[pl] {
			payoff := sf.plPayoffs[pl][plSeq][key]
			//fmt.Println("Looking for payoff", pl, plSeq, key, payoff)
			if payoff != nil {
				if payoff.IsInt() {
					table.Append(payoff.Num().String())
				} else {
					table.Append(payoff.String())
				}
			} else {
				table.Append(".")
			}
		}
		table.EndRow()
		return
	}

	otherPl := sf.plNames[idx]
	if pl == otherPl {
		sf.recAppendRow(table, pl, idx+1, sequences)
		return
	}

	// for each move for player
	for _, seq := range sf.plSeqs[otherPl] {
		lastSeq := sequences[otherPl]
		sequences[otherPl] = seq
		sf.recAppendRow(table, pl, idx+1, sequences)
		sequences[otherPl] = lastSeq
	}
}
