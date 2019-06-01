package transaction

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/smart"
	"log"
	"sort"
	"sync"
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
var tLock sync.RWMutex
var logToFile bool

const transactionGas = uint64(1)
const blockReward = uint64(100)
const gasLimit = uint64(1000) //(Gas limit for blocks) TODO What is good numbers?

func StartTransactionLayer(channels ChannelStruct, log_ bool) {
	tree = Tree{make(map[string]TreeNode), ""}
	logToFile = log_
	// Process a Block coming from the consensus layer
	go func() {
		for {
			b := <-channels.BlockToTrans
			if len(tree.treeMap) == 0 && b.Slot == 0 && b.ParentPointer == "" {
				tree.createNewNode(b, b.BlockData.GenesisData.InitialState)
				smart.StartSmartContractLayer(tree.head, log_)
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
			blockHash := <-channels.FinalizeToTrans
			channels.StateFromTrans <- tree.finalize(blockHash)
		}
	}()

	// A new NodeBlock should be created from the transactions in transList
	for {
		newBlockData := <-channels.TransToTrans
		channels.BlockFromTrans <- tree.createNewBlock(newBlockData)
	}
}

func (t *Tree) processBlock(b Block) {

	blockHash := b.CalculateBlockHash()
	accumulatedRewards := blockReward
	s := State{}
	s.ParentHash = b.ParentPointer

	tLock.RLock()
	s.Ledger = copyMap(t.treeMap[s.ParentHash].state.Ledger)
	s.ConStake = copyMap(t.treeMap[s.ParentHash].state.ConStake)
	s.ConOwners = copyContMap(t.treeMap[s.ParentHash].state.ConOwners)
	s.TotalStake = t.treeMap[s.ParentHash].state.TotalStake
	tLock.RUnlock()

	// Remove expired contracts from ledger in TL and from ConLayer
	// Collection storageCosts
	expiring, reward := smart.NewBlockTreeNode(blockHash, b.ParentPointer, b.Slot)
	s.CleanExpiredContract(expiring)
	accumulatedRewards += reward

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

	// Verify our new state matches the state of the block creator to ensure he has also done the same work
	if s.VerifyHashedState(b.StateHash, b.BakerID) {
		// Pay the block creator
		s.PayBlockRewardOrRemainGas(b.BakerID, accumulatedRewards)

	} else {
		fmt.Println("Proof of work in block didn't match...")
	}
	// Create new node in the tree
	t.createNewNode(b, s)

	if logToFile {
		// TODO: logToFile the ledger state to filepath out/slotno_blockhash[:6]
	}

}

func (t *Tree) finalize(blockHash string) State {
	tLock.RLock()
	defer tLock.RUnlock()
	if finalizedNode, ok := tree.treeMap[blockHash]; ok {
		smart.FinalizeBlock(blockHash)
		return finalizedNode.state
	} else {
		fmt.Println("Couldn't finalize")
		return State{}
	}
}

func (t *Tree) createNewNode(b Block, s State) {
	tLock.Lock()
	defer tLock.Unlock()
	blockHash := b.CalculateBlockHash()
	t.treeMap[blockHash] = TreeNode{b, s}

	// Update head
	t.head = blockHash
}

func (t *Tree) createNewBlock(blockData CreateBlockData) Block {
	s := State{}

	tLock.RLock()
	s.Ledger = copyMap(t.treeMap[t.head].state.Ledger)
	s.ParentHash = t.head
	s.TotalStake = t.treeMap[s.ParentHash].state.TotalStake
	tLock.RUnlock()

	var addedTransactions []TransData

	expiring, err := smart.SetStartingPointForNewBlock(t.head, blockData.SlotNo)
	if err != nil {
		// this should not happen
		log.Fatal("trying to create a new block which parent doesn't exist")
	}
	// Remove expired contracts from ledger in TL and from ConLayer
	s.CleanExpiredContract(expiring)

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

	smart.DoneCreatingNewBlock()
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
	tLock.RLock()
	defer tLock.RUnlock()
	return tree.treeMap[tree.head].state.Ledger
}

func PrintCurrentLedger() {
	tLock.RLock()
	ledger := tree.treeMap[tree.head].state.Ledger
	tLock.RUnlock()

	var keyList []string
	for k := range ledger {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	for _, k := range keyList {
		fmt.Printf("Amount %v is owned by %v\n", ledger[k], k[:10])
	}
}
