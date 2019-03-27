package transaction

import (
	"fmt"
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
	treeMap       map[string]TLNode
	head          string
	lastFinalized string
	hardness      int
}

func StartTransactionLayer(blockInput chan Block, stateReturn chan State, finalizeChan chan string, blockReturn chan Block, newBlockChan chan CreateBlockData) {
	tree := Tree{make(map[string]TLNode), "", "", 0}

	// Process a block coming from the consensus layer
	go func() {
		for {
			b := <-blockInput
			tree.processBlock(b)
		}
	}()

	// Finalize a given block
	go func() {
		for {
			finalize := <-finalizeChan
			tree.lastFinalized = finalize
			stateReturn <- tree.treeMap[finalize].state
		}
	}()

	// A new block should be created from the transactions in transList
	for {
		newBlockData := <-newBlockChan
		blockReturn <- tree.createNewBlock(newBlockData)
	}

}

func (t *Tree) processBlock(b Block) {
	s := State{}
	s.parentHash = b.ParentPointer
	s.ledger = copyMap(t.treeMap[s.parentHash].state.ledger)
	if s.ledger == nil {
		s.ledger = make(map[PublicKey]int)
	}

	// Update head
	t.head = b.CalculateBlockHash()

	// Update state
	if len(b.BlockData.Trans) != 0 {
		for _, tr := range b.BlockData.Trans {
			s.addTransaction(tr)
		}
	}

	// Create new node in the tree
	t.createNewNode(b, s)
}

func (t *Tree) createNewNode(b Block, s State) {
	n := TLNode{b, s}
	t.treeMap[b.CalculateBlockHash()] = n
}

func (s *State) addTransaction(t Transaction) {
	//TODO: Handle checks of legal transactions

	if !t.VerifyTransaction() {
		fmt.Println("The transactions didn't verify", t)
		return
	}

	//if s.ledger[t.From] < t.Amount { //TODO: remove comment such that it checks the balance
	//	fmt.Println("Not enough money on senders account")
	//	return
	//}
	s.ledger[t.To] += t.Amount
	s.ledger[t.From] -= t.Amount
}

func CreateGenesis() Block {
	genBlock := Block{0,
		"",
		PublicKey{},
		"",
		BlockNonce{},
		"",
		Data{[]Transaction{}}, //TODO: GENESISDATA should be proper created
		""}

	return genBlock
}

func (t Tree) createNewBlock(blockData CreateBlockData) Block {
	s := State{}
	s.ledger = copyMap(t.treeMap[t.head].state.ledger)

	var addedTransactions []Transaction

	noOfTrans := len(blockData.TransList)

	for i := 0; i < min(10, noOfTrans); i++ { //TODO: Change to only run i X time
		newTrans := blockData.TransList[i]
		//transactions = transactions[1:]
		s.addTransaction(newTrans)
		addedTransactions = append(addedTransactions, newTrans)
	}

	prevBlockNonce := t.treeMap[t.lastFinalized].block.BlockNonce

	newBlockNonce := CreateNewBlockNonce(prevBlockNonce, blockData.SlotNo, blockData.Sk, blockData.Pk)

	b := CreateNewBlock(blockData, t.head, newBlockNonce, addedTransactions)

	return b
}

// Helpers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyMap(originalMap map[PublicKey]int) map[PublicKey]int {
	newMap := make(map[PublicKey]int)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}
