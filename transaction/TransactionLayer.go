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
var transactionFee = 1
var blockReward = 100

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
	s := copyState(t.treeMap[b.ParentPointer].state, b.ParentPointer)

	// Remove expired contracts from ledger in TL and from ConLayer
	expiredContracts := s.CleanContractLedger()
	ExpireAtConLayer(expiredContracts)

	// Update state
	if len(b.BlockData.Trans) != 0 {
		for _, td := range b.BlockData.Trans {
			if newState, gasUsed, success := handleTransData(td, s); success {
				s = newState
				accumulatedRewards += gasUsed
			}
		}
	}

	// Verify our new state matches the state of the block creator to ensure he has also done the same work
	if s.VerifyHashedState(b.StateHash, b.BakerID) {
		// Pay the block creator
		s.AddBlockReward(b.BakerID, accumulatedRewards)

	} else {
		fmt.Println("Proof of work in block didn't match...")
	}
	// Create new node in the tree
	t.createNewNode(b, s)

	// Update head
	t.head = b.CalculateBlockHash()

}

// Returns the new State, cost of the Contract call and true if contract executed successful
func handleContractCall(s State, contract ContractCall) (State, int, bool) {

	// Transfer funds from caller to contract
	if !s.FundContractCall(contract.Caller, contract.Amount+contract.Gas) {
		return s, 0, false
	}

	// Runs contract at contract layer
	callSuccess, newContractStake, transactionList, remainingGas := CallAtConLayer(contract)

	// If contract succeeded, execute the transactions from the contract layer
	if callSuccess {
		s.ConStake = newContractStake
		for _, t := range transactionList {
			s.AddContractTransaction(t)
		}
		s.RefundContractCall(contract.Caller, remainingGas)
		return s, contract.Gas - remainingGas, callSuccess
	}

	// If contract not successful, then return remaining funds to caller
	s.RefundContractCall(contract.Caller, contract.Amount+remainingGas)
	return s, contract.Gas - remainingGas, callSuccess

}

func (t *Tree) createNewNode(b Block, s State) {
	t.treeMap[b.CalculateBlockHash()] = TreeNode{b, s}
}

func (t *Tree) createNewBlock(blockData CreateBlockData) Block {
	s := copyState(t.treeMap[t.head].state, t.head)

	var addedTransactions []TransData
	noOfTrans := len(blockData.TransList)

	for i := 0; i < min(1000, noOfTrans); i++ { //TODO: Change to only run i X time
		td := blockData.TransList[i]

		if newState, _, success := handleTransData(td, s); success {
			s = newState
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

// Switches depending on type of Trans. Returns a state after the trans,
// the cost of the trans and a true if everything went well
func handleTransData(td TransData, s State) (State, int, bool) {
	switch td.(type) {

	case Transaction:
		td := td.(Transaction)
		transSuccess := s.AddTransaction(td, transactionFee)
		return s, transactionFee, transSuccess
	case ContractCall:
		td := td.(ContractCall)
		newState, gasCost, success := handleContractCall(s, td)
		return newState, gasCost, success
	case ContractInitialize:
		td := td.(ContractInitialize)
		addr, remainGas, success := InitContractAtConLayer(td.Code, td.Gas)
		if success {
			s.InitializeContract(addr, td.Owner, td.Prepaid)
			s.RefundContractCall(td.Owner, remainGas)
			return s, td.Gas - remainGas, success
		}
	}
	return s, 0, false
}

// Helpers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyState(oldState State, parent string) State {
	s := State{}
	s.ParentHash = parent
	s.Ledger = make(map[string]int)
	for key, value := range oldState.Ledger {
		s.Ledger[key] = value
	}
	s.ConAccounts = make(map[string]ContractAccount)
	for key, value := range oldState.ConAccounts {
		s.ConAccounts[key] = value
	}

	s.TotalStake = oldState.TotalStake
	return s
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
