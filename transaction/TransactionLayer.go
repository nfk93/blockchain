package transaction

import (
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
)

type TLNode struct {
	block Block
	state State
}

type State struct {
	ledger     map[PublicKey]int
	parentHash string
}

type Tree struct {
	treeMap map[string]TLNode
	head    string
}

func StartTransactionLayer(blockInput chan Block, stateReturn chan State, finalizeChan chan string) {
	tree := Tree{make(map[string]TLNode), ""}
	gen := createGenesis()
	_ = processBlock(gen, tree)

	go func() {
		for {
			b := <-blockInput
			processBlock(b, tree)
		}
	}()

	for {
		finalize := <-finalizeChan
		finalState := tree.treeMap[finalize].state
		stateReturn <- finalState
	}

}

func processBlock(b Block, t Tree) State {
	s := State{}
	s.parentHash = b.ParentPointer
	s.ledger = t.treeMap[s.parentHash].state.ledger
	if s.ledger == nil {
		s.ledger = make(map[PublicKey]int)
	}

	// Update state
	if b.Slot != 0 {
		for _, tr := range b.BlockData.Trans {
			s.addTransaction(tr)
		}
	}

	// Update head
	t.head = b.BlockHash

	// Create new node in the tree
	createNewNode(b, s, t)
	return s
}

func createNewNode(b Block, s State, t Tree) {
	n := TLNode{b, s}
	t.treeMap[b.BlockHash] = n
}

func (s *State) addTransaction(t Transaction) {
	//TODO: Handle checks of legal transactions
	s.ledger[t.To] += t.Amount
	s.ledger[t.From] -= t.Amount
}

func createGenesis() Block {
	sk, _ := KeyGen(256)
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

//func createNewBlock(transactions []Transaction) Block {
//	//s := State{}
//	//var addedTransactions []Transaction
//	//for i:=0; i<10; i++  {
//	//	newTrans := transactions[0]
//	//	transactions = transactions[1:]
//	//	s.addTransaction(newTrans)
//	//	addedTransactions = append(addedTransactions, newTrans)
//	//}
//
//	//TODO: Make proper way of creating a new block
//	b := Block{43,
//		"",
//		43,
//		"",
//		43,
//		"",
//		Data{transactions[0:min(10, len(transactions))]},
//		"",
//		""}
//
//	return b
//}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
