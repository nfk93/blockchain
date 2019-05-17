package transaction

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/smart"
	"sort"
)

type TreeNode struct {
	block Block
	state State
}

type Tree struct {
	treeMap map[string]TreeNode
	head    string
}

var tree Tree

const transactionGas = uint64(1)
const blockReward = uint64(100)
const gasLimit = uint64(1000) //(Gas limit for blocks) TODO What is good numbers?

func StartTransactionLayer(channels ChannelStruct) {
	tree = Tree{make(map[string]TreeNode), ""}
	// Process a Block coming from the consensus layer
	go func() {
		for {
			b := <-channels.BlockToTrans
			if len(tree.treeMap) == 0 && b.Slot == 0 && b.ParentPointer == "" {
				tree.createNewNode(b, b.BlockData.GenesisData.InitialState)
				tree.head = b.CalculateBlockHash()
			} else if len(tree.treeMap) > 0 {
				if _, exist := tree.treeMap[b.CalculateBlockHash()]; !exist {
					tree.processBlock(b)
				}
			} else {
				fmt.Println("Tree not initialized. Please send Genesis Node!! ")
			}
		}
	}()

	// Consensus layer asks for the state of a finalized block
	go func() {
		for {
			finalize := <-channels.FinalizeToTrans
			if finalizedNode, ok := tree.treeMap[finalize]; ok {
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

	blockHash := b.CalculateBlockHash()
	accumulatedRewards := blockReward
	s := State{}
	s.ParentHash = b.ParentPointer
	s.Ledger = copyMap(t.treeMap[s.ParentHash].state.Ledger)
	s.ConStake = copyMap(t.treeMap[s.ParentHash].state.ConStake)
	s.ConOwners = copyContMap(t.treeMap[s.ParentHash].state.ConOwners)
	s.TotalStake = t.treeMap[s.ParentHash].state.TotalStake

	// Remove expired contracts from ledger in TL and from ConLayer
	s.CleanExpiredContract(b.Slot)

	// Update state
	if len(b.BlockData.Trans) != 0 {
		for _, td := range b.BlockData.Trans {
			switch td.GetType() {
			case CONTRACTCALL:
				accumulatedRewards += s.HandleContractCall(td.ContractCall, blockHash, b.ParentPointer, b.Slot)
			case CONTRACTINIT:
				accumulatedRewards += s.HandleContractInit(td.ContractInit, blockHash, b.ParentPointer, b.Slot)
			case TRANSACTION:
				accumulatedRewards += s.AddTransaction(td.Transaction, transactionGas)
			}
		}
	}

	// Collection storageCosts
	accumulatedRewards += smart.StorageCost(blockHash)

	// Verify our new state matches the state of the block creator to ensure he has also done the same work
	if s.VerifyHashedState(b.StateHash, b.BakerID) {
		// Pay the block creator
		s.PayBlockRewardOrRemainGas(b.BakerID, accumulatedRewards)

	} else {
		fmt.Println("Proof of work in block didn't match...")
	}
	// Create new node in the tree
	t.createNewNode(b, s)

	// Update head
	t.head = blockHash

}

func (t *Tree) createNewNode(b Block, s State) {
	t.treeMap[b.CalculateBlockHash()] = TreeNode{b, s}
}

func (t *Tree) createNewBlock(blockData CreateBlockData) Block {
	s := State{}
	s.Ledger = copyMap(t.treeMap[t.head].state.Ledger)
	s.ParentHash = t.head
	s.TotalStake = t.treeMap[s.ParentHash].state.TotalStake
	var addedTransactions []TransData

	accumulatedGasUse := uint64(0)
	for _, td := range blockData.TransList {

		if accumulatedGasUse < gasLimit {
			switch td.GetType() {
			case CONTRACTCALL:
				accumulatedGasUse += s.HandleContractCall(td.ContractCall, "", t.head, blockData.SlotNo)
			case CONTRACTINIT:
				accumulatedGasUse += s.HandleContractInit(td.ContractInit, "", t.head, blockData.SlotNo)
			case TRANSACTION:
				accumulatedGasUse += s.AddTransaction(td.Transaction, transactionGas)
			}
			addedTransactions = append(addedTransactions, td)
		}

	}

	b := Block{blockData.SlotNo,
		t.head,
		blockData.Pk,
		blockData.Draw,
		BlockNonce{},
		blockData.LastFinalized,
		BlockData{addedTransactions, GenesisData{}},
		s.SignHashedState(blockData.Sk),
		""}

	b.SignBlock(blockData.Sk)
	return b

}

// Helpers
func copyMap(originalMap map[string]uint64) map[string]uint64 {
	if originalMap == nil {
		return make(map[string]uint64)
	}
	newMap := make(map[string]uint64)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func copyContMap(originalMap map[string]crypto.PublicKey) map[string]crypto.PublicKey {
	if originalMap == nil {
		return make(map[string]crypto.PublicKey)
	}
	newMap := make(map[string]crypto.PublicKey)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func GetCurrentLedger() map[string]uint64 {
	return tree.treeMap[tree.head].state.Ledger
}

func PrintCurrentLedger() {
	ledger := tree.treeMap[tree.head].state.Ledger

	var keyList []string
	for k := range ledger {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	for _, k := range keyList {
		fmt.Printf("Amount %v is owned by %v\n", ledger[k], k[:10])
	}
}
