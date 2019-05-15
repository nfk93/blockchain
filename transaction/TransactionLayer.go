package transaction

import (
	"fmt"
	. "github.com/nfk93/blockchain/objects"
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

const transactionGas = 1
const blockReward = 100
const gasLimit = 1000 //(Gas limit for blocks) TODO What is good numbers?

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

	accumulatedRewards := blockReward
	s := State{}
	s.ParentHash = b.ParentPointer
	s.Ledger = copyMap(t.treeMap[s.ParentHash].state.Ledger)
	s.ConAccounts = copyContMap(t.treeMap[s.ParentHash].state.ConAccounts)
	s.ConStake = copyMap(t.treeMap[s.ParentHash].state.ConStake)
	s.TotalStake = t.treeMap[s.ParentHash].state.TotalStake

	// Remove expired contracts from ledger in TL and from ConLayer
	expiredContracts := s.CleanContractLedger()
	ExpireAtConLayer(expiredContracts)

	// Update state
	if len(b.BlockData.Trans) != 0 {
		for _, td := range b.BlockData.Trans {

			gasCost := s.HandleTransData(td, transactionGas)
			accumulatedRewards += gasCost

		}
	}

	// Collection storageCosts
	noOfSlots := b.Slot - t.treeMap[b.ParentPointer].block.Slot
	collectedStorageCosts := s.CollectStorageCost(noOfSlots)
	accumulatedRewards += collectedStorageCosts

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
	t.head = b.CalculateBlockHash()

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

	accumulatedGasUse := 0
	for _, td := range blockData.TransList {

		if accumulatedGasUse < gasLimit {
			gasUse := s.HandleTransData(td, transactionGas)
			accumulatedGasUse += gasUse
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
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyMap(originalMap map[string]int) map[string]int {
	if originalMap == nil {
		return make(map[string]int)
	}
	newMap := make(map[string]int)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func copyContMap(originalMap map[string]ContractAccount) map[string]ContractAccount {
	if originalMap == nil {
		return make(map[string]ContractAccount)
	}
	newMap := make(map[string]ContractAccount)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func GetCurrentLedger() map[string]int {
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
		fmt.Printf("Amount %v is owned by %v\n", ledger[k], k[4:14])
	}
}
