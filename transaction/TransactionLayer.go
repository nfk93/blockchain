package transaction

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
)

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

func StartTransactionLayer(blockInput chan Block, stateReturn chan State, finalizeChan chan string, blockReturn chan Block, newBlockChan chan CreateBlockData, sk SecretKey) {
	tree := Tree{make(map[string]TLNode), "", "", BlockNonce{}, 0.0}

	// Process a NodeBlock coming from the consensus layer
	go func() {
		for {
			b := <-blockInput
			if b.Slot == 0 { // TODO: REMOVE when we have a proper initialization of GENESIS block / data
				tree.lastFinalized = b.CalculateBlockHash()
				tree.currentNonce = b.BlockNonce
			}
			tree.processBlock(b)
		}
	}()

	// Finalize a given NodeBlock
	go func() {
		for {
			finalize := <-finalizeChan
			if finalizedNode, ok := tree.treeMap[finalize]; ok {
				tree.lastFinalized = finalize
				tree.currentNonce = tree.CreateNewBlockNonce(finalize, finalizedNode.block.Slot, sk, finalizedNode.block.BakerID)
				stateReturn <- finalizedNode.state
			} else {
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

func CreateTestGenesis() Block {
	sk, pk := KeyGen(2048)
	genBlock := Block{0,
		"",
		PublicKey{},
		"",
		BlockNonce{"GENESIS", Sign("GENESIS", sk), pk},
		"",
		Data{[]Transaction{}, GenesisData{}}, //TODO: GENESISDATA should be proper created
		""}

	genBlock.LastFinalized = genBlock.CalculateBlockHash()
	return genBlock
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

	b := CreateNewBlock(blockData, t.head, t.currentNonce, addedTransactions)

	return b
}

func (t *Tree) CreateNewBlockNonce(finalizeThisBlock string, slot int, sk SecretKey, pk PublicKey) BlockNonce {

	blockToFinalize := t.treeMap[finalizeThisBlock].block

	var buf bytes.Buffer
	buf.WriteString("NONCE")
	buf.WriteString(previousStatesAsString(blockToFinalize, *t))
	//buf.WriteString(nonce.Nonce) //Old block nonce //TODO: Should also contain new states
	buf.WriteString(strconv.Itoa(slot))

	newNonceString := buf.String()
	newNonce := HashSHA(newNonceString)
	signature := Sign(string(newNonce), sk)

	return BlockNonce{newNonce, signature, pk}
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

func previousStatesAsString(currentBlock Block, t Tree) string {
	var buf bytes.Buffer

	currentBlock = t.treeMap[t.head].block

	parentBlock := currentBlock.ParentPointer
	for currentBlock.LastFinalized != parentBlock {
		buf.WriteString(string(t.treeMap[parentBlock].state.StateAsString()))
		parentBlock = t.treeMap[parentBlock].block.ParentPointer
	}

	return buf.String()
}
