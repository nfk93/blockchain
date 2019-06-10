package transaction

import (
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
const blockReward = uint64(10000)
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
				log.Println("Tree not initialized. Please send Genesis Node!! ")
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
	expiring, storageReward := smart.NewBlockTreeNode(blockHash, b.ParentPointer, b.Slot)
	s.CleanExpiredContract(expiring)

	// Update state
	accumulatedGas := uint64(0)
	if len(b.BlockData.Trans) != 0 {
		for _, td := range b.BlockData.Trans {
			switch td.GetType() {
			case CONTRACTCALL:
				accumulatedGas += s.HandleContractCall(td.ContractCall, blockHash, b.ParentPointer, b.Slot)
			case CONTRACTINIT:
				accumulatedGas += s.HandleContractInit(td.ContractInit, blockHash, b.ParentPointer, b.Slot)
			case TRANSACTION:
				accumulatedGas += s.AddTransaction(td.Transaction, transactionGas)
			}
		}
	}

	// Verify our new state matches the state of the block creator to ensure he has also done the same work
	if s.VerifyHashedState(b.StateHash, b.BakerID) && accumulatedGas < gasLimit {
		// Pay the block creator
		totalReward := accumulatedGas + storageReward + blockReward
		s.AddAmountToAccount(b.BakerID, totalReward)

	} else {
		log.Println("Proof of work in block didn't match...")
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
		log.Println("Couldn't finalize")
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
	s.Ledger = copyMap(t.treeMap[blockData.ParentHash].state.Ledger)
	s.ParentHash = blockData.ParentHash
	s.TotalStake = t.treeMap[s.ParentHash].state.TotalStake
	tLock.RUnlock()

	var addedTransactions []TransData

	expiring, err := smart.SetStartingPointForNewBlock(blockData.ParentHash, blockData.SlotNo)
	if err != nil {
		// this should not happen
		log.Fatal("trying to create a new block which parent doesn't exist")
	}
	// Remove expired contracts from ledger in TL and from ConLayer
	s.CleanExpiredContract(expiring)

	accumulatedGasUse := uint64(0)
	for _, td := range blockData.TransList {
		if accumulatedGasUse+transactionGas > gasLimit {
			break
		}
		switch td.GetType() {
		case CONTRACTCALL:
			oldState := State{Ledger: copyMap(s.Ledger), ParentHash: s.ParentHash, TotalStake: s.TotalStake}
			gasUsed := s.HandleContractCall(td.ContractCall, "", blockData.ParentHash, blockData.SlotNo)
			if accumulatedGasUse+gasUsed > gasLimit {
				s = oldState
			} else {
				accumulatedGasUse += gasUsed
				addedTransactions = append(addedTransactions, td)
			}
		case CONTRACTINIT:
			oldState := State{Ledger: copyMap(s.Ledger), ParentHash: s.ParentHash, TotalStake: s.TotalStake}
			gasUsed := s.HandleContractInit(td.ContractInit, "", blockData.ParentHash, blockData.SlotNo)
			if accumulatedGasUse+gasUsed > gasLimit {
				s = oldState
			} else {
				accumulatedGasUse += gasUsed
				addedTransactions = append(addedTransactions, td)
			}
		case TRANSACTION:
			gasUsed := s.AddTransaction(td.Transaction, transactionGas)
			accumulatedGasUse += gasUsed
			addedTransactions = append(addedTransactions, td)
		}
	}

	b := Block{blockData.SlotNo,
		blockData.ParentHash,
		blockData.Pk,
		blockData.Draw,
		blockData.BlockNonce,
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
		log.Printf("Amount %v is owned by %v\n", ledger[k], k[:10])
	}
}

func GetPk(key string) {

}
