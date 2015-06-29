package gametheory

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type Move struct {
	name string
	iset *InformationSet // where this move emanates from
	prob *big.Rat        // behavior probability... for chance nodes
}

func NewMove(label string) *Move {
	return &Move{name: label}
}

type Node struct {
	sibling    *Node // to the "right"
	nextInIset *Node
	firstChild *Node
	lastChild  *Node

	parent *Node
	iset   *InformationSet

	//?public boolean terminal;

	reachedBy *Move // move of edge from parent
	outcome   *Outcome
}

/*func NewNode(name string) *Node {
	return &Node{
		id,
	}
}*/

type Outcome struct {
	node *Node
	// TODO: just put payoffs on Node?
}

func NewOutcome(node *Node) *Outcome {
	return &Outcome{node: node}
}

type Player struct {
	name   string
	chance bool
}

func NewPlayer(name string) *Player {
	return &Player{name, false}
}

func ChancePlayer() *Player {
	return &Player{"!", true}
}

func (self *Player) isChance() bool {
	return self.chance
}

func (self *Player) String() string {
	return self.name
}

type InformationSet struct {
	id        uint
	name      string
	player    *Player
	firstNode *Node
	lastNode  *Node
	moves     []*Move
}

/*func NewInformationSet(id uint, pl *Player) *InformationSet {
	return InformationSet{
		id,
		str(id),
		pl,
	}
}*/

/*fn move_count(&self) -> usize {
    	let mut count = 0;
        let mut child = self.first_node;
        loop {
            if child == None {  // can I do this?  Or do I need a match...
                break;
            }
            count += 1;
            child = child.sibling;
        }
    	count
	}*/

/*
 On a move obj, prob property only given by chance players.
 On a move obj, either a `next` property or an `outcome` property is given.
 json form:
 {
   player: "name"|null
   iset: "name"|null
   chance: true|false
   moves: [
     {
       name: "name"
       prob: <0.0..1.0>
       next: { ... }
       outcome: [
        {
          pl:
          pay:
        },
        ...
       ]
     },
     ...
   ]
 }
*/
type ExtensiveForm struct {
	players []*Player
	isets   []*InformationSet
	root    *Node

	moveIndex map[string]int
}

func (self *ExtensiveForm) UnmarshalJSON(bytes []byte) error {
	fmt.Println("Called UnmarshalJSON")
	var rootFactory NodeFactory
	err := json.Unmarshal(bytes, &rootFactory)
	if err != nil {
		return err
	}
	fmt.Println(rootFactory.String())

	// TODO: playerIndex := make(map[string]int)
	// TODO: isetIndex := make(map[string]int)
	self.moveIndex = make(map[string]int)

	// get all the players...
	// recursivly visit the tree and construct the ex

	// at the leaves I am interested in probability it took to get there
	// also the defining sequence for each player to get there
	// with perfect recall we can take define this as the last move taken by each player
	// so the input sequence is the (move_p1, move_p2, ...)
	// the expectation of the payoffs are taken
	seqform := NewSequenceForm(&rootFactory)

	// TODO: delete this...
	if seqform != nil {
		return nil
	}

	return nil
}

func (tree *ExtensiveForm) recVisitNode(depth int, prob *big.Rat, nf *NodeFactory) {
	fmt.Println(depth, prob, "node", nf.Player, nf.Iset)
	for _, move := range nf.Moves {
		tree.followMove(depth, prob, move)
	}
}

func (tree *ExtensiveForm) followMove(depth int, prob *big.Rat, mf *MoveFactory) {
	fmt.Println(depth, prob, "move", mf.Name, mf.Prob)

	var nextProb *big.Rat
	if mf.Prob != 0 {
		nextProb = new(big.Rat).SetFloat64(mf.Prob)
		nextProb.Mul(nextProb, prob)
	} else {
		nextProb = prob
	}

	if mf.Next != nil {
		tree.recVisitNode(depth+1, nextProb, mf.Next)
	} else if mf.Outcome != nil {
		leafStr := "$["
		for _, plPayoff := range mf.Outcome {
			leafStr += plPayoff.String() + ","
		}
		leafStr += "]"
		fmt.Println("leaf!", leafStr)
	}
}

type NodeFactory struct {
	Player string
	Iset   string
	Chance bool
	Moves  []*MoveFactory
}

func (n *NodeFactory) String() string {
	rv := fmt.Sprintf("%v %v %v [", n.Player, n.Chance, n.Iset)
	for _, move := range n.Moves {
		rv += "\n  " + move.String()
	}
	rv += "\n]"
	return rv
}

type MoveFactory struct {
	Name    string
	Prob    float64
	Next    *NodeFactory
	Outcome []*OutcomeFactory
}

func (m *MoveFactory) String() string {

	rv := m.Name
	if m.Prob != 0 {
		rv += fmt.Sprintf("[Pr=%v]", m.Prob)
	}
	if m.Next != nil {
		rv += "->" + m.Next.String()
	} else if m.Outcome != nil {
		rv += "->$["
		for _, plPayoff := range m.Outcome {
			rv += plPayoff.String() + ","
		}
		rv += "]"
	}
	return rv
}

type OutcomeFactory struct {
	Player string
	Payoff float64
}

func (o *OutcomeFactory) String() string {
	return fmt.Sprintf("%v=%v", o.Player, o.Payoff)
}
