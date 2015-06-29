package gametheory

import (
	"fmt"
	"math/big"
)

type SequenceForm struct {
	seed int64

	plPayoffs map[string]map[string]*big.Rat // pl -> move name -> payoff

	plConstraints map[string][][]int64 // TODO...
	isetSeqIn     map[*InformationSet]*Move
	nodeDefSeq    map[*Node][3]Move
	plSeqs        [3][]*Move

	// used to build constraints...
	seqIndex  map[string]int // move name -> col ??
	isetIndex map[string]int // iset name -> row  ??

	plIsets map[string][]string // pl -> [iset names]

	isets  []*InformationSet
	priors map[*Move]*big.Rat
}

func NewSequenceForm(nf *NodeFactory) *SequenceForm {
	sf := new(SequenceForm)
	sf.seqIndex = make(map[string]int)
	sf.isetIndex = make(map[string]int)
	sf.plIsets = make(map[string][]string)
	sf.plPayoffs = make(map[string]map[string]*big.Rat)
	sf.recVisitNode(0, big.NewRat(1, 1), nf, make(map[string]*MoveFactory))
	return sf
}

func (sf *SequenceForm) recVisitNode(depth int, prob *big.Rat, nf *NodeFactory, sequences map[string]*MoveFactory) {
	fmt.Println(depth, prob, "node", nf.Player, nf.Iset, sequences)
	lastMove := sequences[nf.Player]

	_, exists := sf.isetIndex[nf.Iset]
	if !exists {
		isetsForPl := sf.plIsets[nf.Player]
		isetsForPl = append(isetsForPl, nf.Iset)
		sf.isetIndex[nf.Iset] = len(isetsForPl)
		sf.plIsets[nf.Player] = isetsForPl // needed?
	}

	for _, move := range nf.Moves {
		sequences[nf.Player] = move
		fmt.Println("seq", nf.Player, move.String())
		sf.followMove(depth, prob, move, sequences)
	}
	if lastMove != nil {
		fmt.Println("seq revert", nf.Player, lastMove.String())
	}
	sequences[nf.Player] = lastMove
}

func (sf *SequenceForm) followMove(depth int, prob *big.Rat, mf *MoveFactory, sequences map[string]*MoveFactory) {
	fmt.Println(depth, prob, "move", mf.Name, mf.Prob)

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
		leafStr := "$["
		// looking
		for _, outcome := range mf.Outcome {
			pl := outcome.Player
			seq := sequences[pl]
			payoff := outcome.Payoff
			if sf.plPayoffs[pl] == nil {
				sf.plPayoffs[pl] = make(map[string]*big.Rat)
			}
			sf.plPayoffs[pl][seq.Name] = new(big.Rat).SetFloat64(payoff)
			leafStr += seq.String() + "->" + outcome.String() + ","
		}
		leafStr += "]"
		fmt.Println("leaf!", leafStr)
		// TODO: fill out payoff matrix entry...
	}
}
