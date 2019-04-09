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

type Tree struct {
	treeMap       map[string]TLNode
	head          string
	lastFinalized string
	hardness      float64
}

var tree Tree

func StartTransactionLayer(blockInput chan Block, stateReturn chan State, finalizeChan chan string, blockReturn chan Block, newBlockChan chan CreateBlockData, sk SecretKey) {
	tree = Tree{make(map[string]TLNode), "", "", 0.0}

	// Process a NodeBlock coming from the consensus layer
	go func() {
		for {
			b := <-blockInput
			if len(tree.treeMap) == 0 && b.Slot == 0 && b.ParentPointer == "" {
				tree.lastFinalized = b.CalculateBlockHash()
				tree.createNewNode(b, b.BlockData.GenesisData.InitialState)
			} else if len(tree.treeMap) > 0 {
				tree.processBlock(b)
			} else {
				fmt.Println("Tree not initialized. Please send Genesis Node!! ")
			}
		}
	}()

	// Finalize a given NodeBlock
	go func() {
		for {
			finalize := <-finalizeChan
			if finalizedNode, ok := tree.treeMap[finalize]; ok {
				tree.lastFinalized = finalize
				stateReturn <- finalizedNode.state
			} else {
				fmt.Println("Couldn't finalize")
				stateReturn <- State{}
			}
		}
	}()

	// A new NodeBlock should be created from the transactions in transList
	for {
		newBlockData := <-newBlockChan
		newBlock := tree.createNewBlock(newBlockData)
		blockReturn <- newBlock
	}
}

func (t *Tree) processBlock(b Block) {
	s := State{}
	s.ParentHash = b.ParentPointer
	s.Ledger = copyMap(t.treeMap[s.ParentHash].state.Ledger)
	if s.Ledger == nil {
		s.Ledger = make(map[PublicKey]int)
	}

	// Update state
	if len(b.BlockData.Trans) != 0 {
		for _, tr := range b.BlockData.Trans {
			s.AddTransaction(tr)
		}
	}

	// Create new node in the tree
	t.createNewNode(b, s)

	// Update head
	t.head = b.CalculateBlockHash()
}

func (t *Tree) createNewNode(b Block, s State) {
	n := TLNode{b, s}
	t.treeMap[b.CalculateBlockHash()] = n
}

func (t *Tree) createNewBlock(blockData CreateBlockData) Block {
	s := State{}
	s.Ledger = copyMap(t.treeMap[t.head].state.Ledger)

	var addedTransactions []Transaction

	noOfTrans := len(blockData.TransList)

	for i := 0; i < min(10, noOfTrans); i++ { //TODO: Change to only run i X time
		newTrans := blockData.TransList[i]
		//transactions = transactions[1:]
		s.AddTransaction(newTrans)
		addedTransactions = append(addedTransactions, newTrans)
	}

	b := Block{blockData.SlotNo,
		t.head,
		blockData.Pk,
		blockData.Draw,
		BlockNonce{},
		t.lastFinalized,
		Data{addedTransactions, GenesisData{}},
		""}

	b.SignBlock(blockData.Sk)
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

//func PreviousStatesAsString(currentBlock Block) string {
//	var buf bytes.Buffer
//
//	currentBlock = tree.treeMap[tree.head].block
//
//	parentBlock := currentBlock.ParentPointer
//	for currentBlock.LastFinalized != parentBlock {
//		buf.WriteString(string(tree.treeMap[parentBlock].state.StateAsString()))
//		parentBlock = tree.treeMap[parentBlock].block.ParentPointer
//	}
//
//	return buf.String()
//}
//func (t *Tree) CreateNewBlockNonce(finalizeThisBlock string, slot int, sk SecretKey, pk PublicKey) BlockNonce {
//
//	blockToFinalize := t.treeMap[finalizeThisBlock].block
//
//	var buf bytes.Buffer
//	buf.WriteString("NONCE")
//	buf.WriteString(PreviousStatesAsString(blockToFinalize))
//	buf.WriteString(strconv.Itoa(slot))
//
//	newNonceString := buf.String()
//	newNonce := HashSHA(newNonceString)
//	bn := BlockNonce{newNonce, "", pk}
//	bn.SignBlockNonce(sk)
//	return bn
//}
