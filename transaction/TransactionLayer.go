package transaction

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
)

//var treeMap map[string]TLNode //Map of hash -> node

type TLNode struct {
	block Block
	state State
}

type State struct {
	ledger     map[string]int
	parentHash string
}

type Tree struct {
	treeMap map[string]TLNode
}

func StartTransactionLayer(blockInput chan Block, stateReturn chan State) {
	tree := Tree{make(map[string]TLNode)}
	_ = processBlock(createGenesis(), tree)

	for {
		b := <-blockInput
		s := processBlock(b, tree)
		stateReturn <- s
		fmt.Print()
	}

}

func processBlock(b Block, t Tree) State {
	s := State{}
	s.parentHash = b.ParentPointer
	s.ledger = t.treeMap[s.parentHash].state.ledger
	if s.ledger == nil {
		s.ledger = make(map[string]int)
	}

	// Update state
	if b.Slot != 0 {
		for _, tr := range b.BlockData.Trans {
			s.addTransaction(tr)
		}
	}

	// Create new node in the tree
	createNewNode(b, s, t)
	fmt.Println(s)
	return s
}

func createNewNode(b Block, s State, t Tree) {
	n := TLNode{b, s}
	//t.treeMap[b.BlockHash] = n //TODO: Change back to proper hash
	t.treeMap[strconv.Itoa(b.Slot)] = n
}

func (s *State) addTransaction(t Transaction) {
	s.ledger[t.To.String()] += t.Amount
	s.ledger[t.From.String()] -= t.Amount
}

func createGenesis() Block {
	sk, _ := Crypto.KeyGen(256)
	genBlock := Block{0,
		"",
		0,
		"VALID",
		0,
		"",
		Data{},
		"",
		""}

	genBlock.SignBlock(sk)
	genBlock.HashBlock()
	return genBlock
}
