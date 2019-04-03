package transaction

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
)

type TransLayer struct {
	tree Tree
	sk   SecretKey
}

type TLNode struct {
	block Block
	state State
}

type Tree struct {
	treeMap       map[string]TLNode
	head          string
	lastFinalized string
	currentNonce  BlockNonce
	hardness      float64
}

func StartTransactionLayer(sk SecretKey, state State) TransLayer {

	tree := Tree{make(map[string]TLNode),
		"",
		"",
		BlockNonce{},
		0.0}
	return TransLayer{tree, sk}
}

func (tl *TransLayer) ReceiveBlock(b Block) bool {
	// TODO: TEST verification of the block

	if valid, msg := b.ValidateBlock(); !valid {
		fmt.Println(msg)
		return false
	}
	if b.Slot == 0 { // TODO: REMOVE when we have a proper initialization of GENESIS block / data
		tl.tree.lastFinalized = b.CalculateBlockHash()
		tl.tree.currentNonce = b.BlockNonce
	}
	tl.tree.processBlock(b)
	return true
}

func (tl *TransLayer) FinalizeBlock(blockHash string) State {
	if finalizedNode, ok := tl.tree.treeMap[blockHash]; ok {
		tl.tree.lastFinalized = blockHash
		tl.tree.currentNonce = CreateNewBlockNonce(blockHash, *tl)
		return finalizedNode.state
	} else {
		return State{}
	}
}

func (tl *TransLayer) CreateNewBlock(data CreateBlockData) Block {

	return tl.tree.createNewBlock(data, tl.sk)

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

func (t *Tree) createNewBlock(blockData CreateBlockData, sk SecretKey) Block {
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
		t.currentNonce,
		t.lastFinalized,
		Data{addedTransactions, GenesisData{}},
		""}

	b.SignBlock(sk)
	return b

}

func CreateNewBlockNonce(finalizeThisBlock string, tl TransLayer) BlockNonce {

	blockToFinalize := tl.tree.treeMap[finalizeThisBlock].block

	var buf bytes.Buffer
	buf.WriteString("NONCE")
	buf.WriteString(PreviousStatesAsString(blockToFinalize, tl.tree))
	buf.WriteString(strconv.Itoa(tl.tree.treeMap[finalizeThisBlock].block.Slot))

	newNonce := HashSHA(buf.String())
	newBlockNonce := BlockNonce{newNonce, "", tl.tree.treeMap[finalizeThisBlock].block.BakerID}
	newBlockNonce.SignBlockNonce(tl.sk)
	return newBlockNonce
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

func PreviousStatesAsString(currentBlock Block, t Tree) string {
	var buf bytes.Buffer

	currentBlock = t.treeMap[t.head].block

	parentBlock := currentBlock.ParentPointer
	for currentBlock.LastFinalized != parentBlock {
		buf.WriteString(string(t.treeMap[parentBlock].state.StateAsString()))
		parentBlock = t.treeMap[parentBlock].block.ParentPointer
	}

	return buf.String()
}

func (tl TransLayer) GetTree() Tree {
	return tl.tree
}

func (t Tree) GetState(state string) State {
	return t.treeMap[state].state
}
