package gametheory

import (
  "bytes"
	"fmt"
	"github.com/megesdal/matrixprinter"
	"math/big"
  "strings"
)

type SequenceForm struct {
	seed int64

	plPayoffs     map[string]map[string]map[string]*big.Rat // pl -> own move -> (other moves tuple) -> payoff
	plConstraints map[string]map[string]map[string]int      // pl -> own iset -> own move -> 1 if move in iset, -1 if move into iset, 0 otherwise

	// used to build constraints...
	// TODO: needed?
	//seqIndex  map[string]int // move/seq name -> col ??
	//isetIndex map[string]int // iset name -> row  ??

	plIsets map[string][]string // pl -> [own iset]
	plSeqs  map[string][]string // pl -> [own move/seq]
	plNames []string            // [pl]

	//isets  []*InformationSet
	priors map[*Move]*big.Rat
}

func NewSequenceForm(nf *NodeFactory) *SequenceForm {
	sf := new(SequenceForm)
	//sf.seqIndex = make(map[string]int)
	//sf.isetIndex = make(map[string]int)
	sf.plIsets = make(map[string][]string)
	sf.plPayoffs = make(map[string]map[string]map[string]*big.Rat)
  sf.plConstraints = make(map[string]map[string]map[string]int)
  sf.plNames = make([]string, 0)
  sf.plSeqs = make(map[string][]string)
  fmt.Println("=====START seqform====")
	sf.recVisitNode(0, big.NewRat(1, 1), nf, make(map[string]*MoveFactory))
	return sf
}

func (sf *SequenceForm) recVisitNode(depth int, prob *big.Rat, nf *NodeFactory, sequences map[string]*MoveFactory) {
	fmt.Println(depth, prob, "node", nf.Player, nf.Iset)

  if !nf.Chance {
    plExists := false
    for _, pl := range sf.plNames {
      if pl == nf.Player {
        plExists = true
        break
      }
    }

    if !plExists {
      fmt.Println("Adding name: ", nf.Player)
      sf.plNames = append(sf.plNames, nf.Player)
      sf.plSeqs[nf.Player] = make([]string, 1)
      sf.plSeqs[nf.Player][0] = "\u2205"  // empty sequence
      sf.plIsets[nf.Player] = make([]string, 0)
    }
  }

	lastMove := sequences[nf.Player]

  if !nf.Chance {
    isetsForPl := sf.plIsets[nf.Player]
    isetExists := false
    for _, plIset := range isetsForPl {
      if plIset == nf.Iset {
        isetExists = true
        break
      }
    }
  	if !isetExists {
      fmt.Println("adding iset", nf.Player, nf.Iset)
  		isetsForPl = append(isetsForPl, nf.Iset)
  		sf.plIsets[nf.Player] = isetsForPl // needed?
  	}
  }

	// iset -> lastMove -> -1
	for _, move := range nf.Moves {

    if move.Name == "\u2205" {
      panic("Use of reserved name for empty sequence...")
    }

    seqExists := false
    for _, seq := range sf.plSeqs[nf.Player] {
      if seq == move.Name {
        seqExists = true
        break
      }
    }

    if !seqExists {
      sf.plSeqs[nf.Player] = append(sf.plSeqs[nf.Player], move.Name)
    }

		// iset -> move -> 1
    _, exists := sf.plConstraints[nf.Player]
    if !exists {
      sf.plConstraints[nf.Player] = make(map[string]map[string]int)
    }
    _, exists = sf.plConstraints[nf.Player][nf.Iset]
    if !exists {
      sf.plConstraints[nf.Player][nf.Iset] = make(map[string]int)
    }
		sf.plConstraints[nf.Player][nf.Iset][move.Name] = 1
		sequences[nf.Player] = move
		fmt.Println("seq", nf.Player, nf.Iset, move.Name)
		sf.followMove(depth, prob, move, sequences)
	}

  lastMoveName := "\u2205"
	if lastMove != nil {
    lastMoveName = lastMove.Name
  }
  sf.plConstraints[nf.Player][nf.Iset][lastMoveName] = -1
	fmt.Println("seq revert", nf.Player, nf.Iset, lastMoveName)

	sequences[nf.Player] = lastMove
}

func (sf *SequenceForm) payoffSeqKey(sequences map[string]*MoveFactory, except string) string {
  key := ""
  first := true
  for _, pl := range sf.plNames {
    if pl != except {
      if first {
        first = false
      } else {
        key += ":"
      }
      key += sequences[pl].Name
    }
  }
  return key
}

func (sf *SequenceForm) followMove(depth int, prob *big.Rat, mf *MoveFactory, sequences map[string]*MoveFactory) {
	fmt.Println(depth, prob, "move", mf.String())

	var nextProb *big.Rat
	if mf.Prob != 0 {
		nextProb = new(big.Rat).SetFloat64(mf.Prob)
		nextProb.Mul(nextProb, prob)
	} else {
		nextProb = prob
	}

	if mf.Next != nil {
		sf.recVisitNode(depth+1, nextProb, mf.Next, sequences)
	} else if mf.Outcome != nil {
		// looking
		for _, outcome := range mf.Outcome {
			pl := outcome.Player
			seq := sequences[pl]
			payoff := outcome.Payoff
			if sf.plPayoffs[pl] == nil {
				sf.plPayoffs[pl] = make(map[string]map[string]*big.Rat)
			}
      if sf.plPayoffs[pl][seq.Name] == nil {
        sf.plPayoffs[pl][seq.Name] = make(map[string]*big.Rat)
      }
      othersKey := sf.payoffSeqKey(sequences, pl)
      payoffRat := new(big.Rat).SetFloat64(payoff)
      fmt.Println("payoff", pl, "own move", seq.Name, "others move tuple", othersKey, "payoff", payoffRat)
			sf.plPayoffs[pl][seq.Name][othersKey] = payoffRat
		}
	}
}

func (sf *SequenceForm) String() string {

	table := matrixprinter.NewTable()

  fmt.Println("seqform::String")
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
        if (td == 0) {
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

    otherMoves := make([]string,0,len(sf.plNames) - 1)
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
      fmt.Println("Looking for payoff", pl, plSeq, key, payoff)
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
    sf.recAppendRow(table, pl, idx + 1, sequences)
    return
  }

  // for each move for player
  for _, seq := range sf.plSeqs[otherPl] {
    lastSeq := sequences[otherPl]
    sequences[otherPl] = seq
    sf.recAppendRow(table, pl, idx + 1, sequences)
    sequences[otherPl] = lastSeq
  }
}
