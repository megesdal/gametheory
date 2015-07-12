package gametheory

import (
	"fmt"
	"math/big"
	"sort"
)

type SequenceForm struct {
	seed int64

	plPayoffs     map[string]map[string]map[string]*big.Rat // pl -> own move -> (other moves tuple) -> payoff

	// TODO: store an iset's moves and we don't need this...
	plConstraints map[string]map[string]map[string]int64    // pl -> own iset -> own move -> 1 if move from iset, -1 if move into iset, 0 otherwise
	plMaxPayoffs  map[string]*big.Rat

	plSeqDepths   map[string]map[string]int    // pl -> own move -> depth
	plIsetDefSeqs map[string]map[string]string // pl -> own iset -> move into iset (constraint == -1)
	plSeqIsets map[string]map[string]string // pl -> move from... -> iset

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
	sf.plConstraints = make(map[string]map[string]map[string]int64)
	sf.plMaxPayoffs = make(map[string]*big.Rat)
	sf.plSeqDepths = make(map[string]map[string]int)
	sf.plIsetDefSeqs = make(map[string]map[string]string)
	sf.plSeqIsets = make(map[string]map[string]string)
	sf.plNames = make([]string, 0)
	sf.plSeqs = make(map[string][]string)

	fmt.Println("=====START seqform====")
	sf.recVisitNode(0, big.NewRat(1, 1), nf, make(map[string]*MoveFactory))

	// sort plSeqs and plIsets
	for pl, seqs := range sf.plSeqs {
		seqDepths := sf.plSeqDepths[pl]
		seqDepth := func(seq1, seq2 string) bool {
			return seqDepths[seq1] < seqDepths[seq2]
		}
		By(seqDepth).Sort(seqs)
	}

	for pl, isets := range sf.plIsets {
		seqDepths := sf.plSeqDepths[pl]
		isetDefSeqs := sf.plIsetDefSeqs[pl]
		isetDepth := func(h1, h2 string) bool {
			fmt.Println("Comparing isets", h1, h2, isetDefSeqs[h1], isetDefSeqs[h2])
			return seqDepths[isetDefSeqs[h1]] < seqDepths[isetDefSeqs[h2]]
		}
		By(isetDepth).Sort(isets)
	}

	return sf
}

func (sf *SequenceForm) recVisitNode(depth int, prob *big.Rat, nf *NodeFactory, sequences map[string]*MoveFactory) {

	lastMove := sequences[nf.Player]

	if !nf.Chance {
		sf.addPlayerIfAbsent(nf.Player)
		sf.addOrVerifyInformationSet(nf.Player, nf.Iset, lastMove)
	}

	for _, move := range nf.Moves {

		if move.Name == "\u2205" {
			// TODO: quote it or replace it?
			panic("Use of reserved name for empty sequence...")
		}

		if !nf.Chance {
			seqExists := false
			for _, seq := range sf.plSeqs[nf.Player] {
				if seq == move.Name {
					seqExists = true
					break
				}
			}

			if !seqExists {
				sf.plSeqs[nf.Player] = append(sf.plSeqs[nf.Player], move.Name)
				sf.plSeqDepths[nf.Player][move.Name] = depth
			}

			// iset -> move -> 1
			sf.plSeqIsets[nf.Player][move.Name] = nf.Iset
			sf.plConstraints[nf.Player][nf.Iset][move.Name] = 1
		}
		sequences[nf.Player] = move
		sf.followMove(depth, prob, move, sequences)
	}

	// pop the seq stack...
	sequences[nf.Player] = lastMove
}

func (sf *SequenceForm) addPlayerIfAbsent(pl string) {
	plExists := false
	for _, plExisting := range sf.plNames {
		if pl == plExisting {
			plExists = true
			break
		}
	}

	if !plExists {
		sf.plNames = append(sf.plNames, pl)
		sf.plSeqs[pl] = make([]string, 1)
		sf.plSeqs[pl][0] = "\u2205" // empty sequence
		sf.plIsets[pl] = make([]string, 0)
		sf.plPayoffs[pl] = make(map[string]map[string]*big.Rat)
		sf.plSeqDepths[pl] = make(map[string]int)
		sf.plConstraints[pl] = make(map[string]map[string]int64)
		sf.plIsetDefSeqs[pl] = make(map[string]string)
		sf.plSeqIsets[pl] = make(map[string]string)
	}
}

func (sf *SequenceForm) addOrVerifyInformationSet(pl string, iset string, lastMove *MoveFactory) {

	isetsForPl := sf.plIsets[pl]
	isetExists := false
	for _, plIset := range isetsForPl {
		if plIset == iset {
			isetExists = true
			break
		}
	}

	if !isetExists {
		sf.plIsets[pl] = append(isetsForPl, iset)
	}

	// ensure constraints data structure
	_, exists := sf.plConstraints[pl][iset]
	if !exists {
		sf.plConstraints[pl][iset] = make(map[string]int64)
	}

	var lastMoveName string
	if lastMove == nil {
		lastMoveName = "\u2205"
	} else {
		lastMoveName = lastMove.Name
	}

	// if iset existed, this should already be set to -1, else we have imperfect recall..
	if isetExists && sf.plConstraints[pl][iset][lastMoveName] != -1 {
		panic("Imperfect Recall!!!") // TODO: switch to return an error...
	}

	// constraint: iset -> lastMove -> -1
	sf.plConstraints[pl][iset][lastMoveName] = -1
	sf.plIsetDefSeqs[pl][iset] = lastMoveName
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
	//fmt.Println(depth, prob, "move", mf.String())

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
			if sf.plPayoffs[pl][seq.Name] == nil {
				sf.plPayoffs[pl][seq.Name] = make(map[string]*big.Rat)
			}
			othersKey := sf.payoffSeqKey(sequences, pl)
			payoffRat := new(big.Rat).SetFloat64(payoff)
			maxPayoff := sf.plMaxPayoffs[pl]
			if maxPayoff == nil || payoffRat.Cmp(maxPayoff) > 0 {
				sf.plMaxPayoffs[pl] = payoffRat
			}
			//fmt.Println("payoff", pl, "own move", seq.Name, "others move tuple", othersKey, "payoff", payoffRat)
			sf.plPayoffs[pl][seq.Name][othersKey] = payoffRat
		}
	}
}

// sequence sorting...

// By is the type of a "less" function that defines the ordering of its Planet arguments.
type By func(seq1, seq2 string) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(seqs []string) {
	s := &seqSorter{
		seqs: seqs,
		by:   by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Stable(s)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type seqSorter struct {
	seqs []string
	by   func(seq1, seq2 string) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *seqSorter) Len() int {
	return len(s.seqs)
}

// Swap is part of sort.Interface.
func (s *seqSorter) Swap(i, j int) {
	s.seqs[i], s.seqs[j] = s.seqs[j], s.seqs[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *seqSorter) Less(i, j int) bool {
	return s.by(s.seqs[i], s.seqs[j])
}
