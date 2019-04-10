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
}

var tree Tree
var transactionFee = 1
var blockReward = 100
var systemAccount PublicKey

func StartTransactionLayer(channels ChannelStruct) {
	tree = Tree{make(map[string]TLNode), "", ""}

	// Process a NodeBlock coming from the consensus layer
	go func() {
		for {
			b := <-channels.BlockToTrans
			if len(tree.treeMap) == 0 && b.Slot == 0 && b.ParentPointer == "" {
				tree.lastFinalized = b.CalculateBlockHash()
				tree.createNewNode(b, b.BlockData.GenesisData.InitialState)
				tree.head = b.CalculateBlockHash()
				systemAccount = b.BlockData.GenesisData.SystemAccount
			} else if len(tree.treeMap) > 0 {
				if _, exist := tree.treeMap[b.CalculateBlockHash()]; !exist {
					tree.processBlock(b)
				}
			} else {
				fmt.Println("Tree not initialized. Please send Genesis Node!! ")
			}
		}
	}()

	// Finalize a given NodeBlock
	go func() {
		for {
			finalize := <-channels.FinalizeToTrans
			if finalizedNode, ok := tree.treeMap[finalize]; ok {
				tree.lastFinalized = finalize
				channels.StateFromTrans <- finalizedNode.state
			} else {
				fmt.Println("Couldn't finalize")
				channels.StateFromTrans <- State{}
			}
		}
	}()

	// A new NodeBlock should be created from the transactions in transList
	for {
		newBlockData := <-channels.TransToTrans
		newBlock := tree.createNewBlock(newBlockData)
		channels.BlockFromTrans <- newBlock
	}
}

func (t *Tree) processBlock(b Block) {
	successfulTransactions := 0
	s := State{}
	s.ParentHash = b.ParentPointer
	s.Ledger = copyMap(t.treeMap[s.ParentHash].state.Ledger)
	if s.Ledger == nil {
		s.Ledger = make(map[PublicKey]int)
	}

	// Update state
	if len(b.BlockData.Trans) != 0 {
		for _, tr := range b.BlockData.Trans {
			transSuccess := s.AddTransaction(tr, transactionFee)
			if transSuccess {
				successfulTransactions += 1
			}
		}
	}

	// Verify our new state matches the state of the block creator to ensure he has also done the same work
	if s.VerifyStateHash(b.StateHash, b.BakerID) {
		// Create new node in the tree
		t.createNewNode(b, s)

		// Pay the block creator
		s.AddBlockRewardAndTransFees(b.BakerID, blockReward+(successfulTransactions*transactionFee))

		// Update head
		t.head = b.CalculateBlockHash()

	} else {
		s.AddBlockRewardAndTransFees(systemAccount, blockReward+(successfulTransactions*transactionFee))
		fmt.Println("Proof of work in block didn't match...")
	}

}

func (t *Tree) createNewNode(b Block, s State) {
	n := TLNode{b, s}
	t.treeMap[b.CalculateBlockHash()] = n
}

func (t *Tree) createNewBlock(blockData CreateBlockData) Block {
	s := State{}
	s.Ledger = copyMap(t.treeMap[t.head].state.Ledger)
	s.ParentHash = t.head
	var addedTransactions []Transaction

	noOfTrans := len(blockData.TransList)

	for i := 0; i < min(10, noOfTrans); i++ { //TODO: Change to only run i X time
		newTrans := blockData.TransList[i]
		s.AddTransaction(newTrans, transactionFee)
		addedTransactions = append(addedTransactions, newTrans)
	}

	b := Block{blockData.SlotNo,
		t.head,
		blockData.Pk,
		blockData.Draw,
		BlockNonce{},
		t.lastFinalized,
		Data{addedTransactions, GenesisData{}},
		s.CreateStateHash(blockData.Sk),
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
