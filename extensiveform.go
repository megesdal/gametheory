package gametheory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
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
	root    *NodeFactory
}

func (tree *ExtensiveForm) UnmarshalJSON(bytes []byte) error {
	
	var rootFactory NodeFactory
	err := json.Unmarshal(bytes, &rootFactory)
	if err != nil {
		return err
	}

	// get all the players...
	// recursivly visit the tree and construct the ex

	// at the leaves I am interested in probability it took to get there
	// also the defining sequence for each player to get there
	// with perfect recall we can take define this as the last move taken by each player
	// so the input sequence is the (move_p1, move_p2, ...)
	// the expectation of the payoffs are taken

	tree.root = &rootFactory;
	return nil
}

func (tree *ExtensiveForm) String() string {
	var buffer bytes.Buffer
	recString(&buffer, 0, tree.root, make(map[int]bool))
	return buffer.String()
}

// return num lines used
func recString(w io.Writer, depth int, n *NodeFactory, openDepths map[int]bool) int {

	numLines := 1
	if !n.Chance {
		for i := 0; i < depth - 1; i++ {
			if (openDepths[i]) {
				io.WriteString(w, "| ")
			} else {
				io.WriteString(w, "  ")
			}
		}
		if depth != 0 {
			io.WriteString(w, "\\-")
		}
	  io.WriteString(w, n.Player)
		io.WriteString(w, "::")
		io.WriteString(w, n.Iset)
	  io.WriteString(w, "\n")
		depth++
	} else {
		numLines = 0
	}

	for i, move := range n.Moves {
		if i != len(n.Moves) - 1 {
			openDepths[depth - 1] = true
		} else {
			openDepths[depth - 1] = false
		}
		numLines += moveString(w, depth, move, openDepths)
	}
	return numLines
}

func moveString(w io.Writer, depth int, m *MoveFactory, openDepths map[int]bool) int {

	for i := 0; i < depth - 1; i++ {
		if (openDepths[i]) {
			io.WriteString(w, "| ")
		} else {
			io.WriteString(w, "  ")
		}
	}

	if !openDepths[depth - 1] {
		io.WriteString(w, "\\-")
	} else {
		io.WriteString(w, "+-")
	}
	if m.Prob != 0 {
		io.WriteString(w, fmt.Sprintf("?[%v]", m.Prob))
	} else {
		io.WriteString(w, m.Name)
	}

	numLines := 1
	if m.Next != nil {
		io.WriteString(w, "\n")
		numLines += recString(w, depth + 1, m.Next, openDepths)
	} else if m.Outcome != nil {
		outcomeStrs := make([]string, len(m.Outcome))
		for i, plPayoff := range m.Outcome {
			outcomeStrs[i] = plPayoff.String()
		}
		leafStr := "{" + strings.Join(outcomeStrs, ",") + "}\n"
		/*for i := 0; i < depth - 1; i++ {
			io.WriteString(w, "  ")
		}
		if isLast {
			io.WriteString(w, "  ")
		} else {
			io.WriteString(w, "| ")
		}*/
		io.WriteString(w, "->")
		io.WriteString(w, leafStr)
	}
	return numLines
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
		outcomeStrs := make([]string, len(m.Outcome))
		for i, plPayoff := range m.Outcome {
			outcomeStrs[i] = plPayoff.String()
		}
		leafStr := "$[" + strings.Join(outcomeStrs, ",") + "]"
		rv += leafStr
	}
	return rv
}

type OutcomeFactory struct {
	Player string
	Payoff float64
}

func (o *OutcomeFactory) String() string {
	return fmt.Sprintf("%v:%v", o.Player, o.Payoff)
}
