package consensus

import (
	"fmt"
	o "github.com/nfk93/blockchain/objects"
	"sync"
)

var unusedTransactions map[string]bool
var transactions map[string]o.Transaction
var blockToTL chan o.Block
var tLock sync.RWMutex
var blocks skov
var badBlocks map[string]bool
var currentHead string
var currentLength int
var lastFinalized string
var testDrawVal int //Remove when implementing proper calculateDrawVal

func StartConsensus(transFromP2P chan o.Transaction, blockFromP2P chan o.Block, blockToP2P chan o.Block) {
	unusedTransactions = make(map[string]bool)
	transactions = make(map[string]o.Transaction)
	badBlocks = make(map[string]bool)
	blocks.m = make(map[string]o.Block)
	testDrawVal = 0
	// TODO: do something with the genesis data

	// Start processing blocks on one thread, non-concurrently
	go func() {
		for {
			block := <-blockFromP2P
			handleBlock(block)
		}
	}()
	// Start processing transactions on one thread, concurrently
	go func() {
		for {
			trans := <-transFromP2P
			go handleTransaction(trans)
		}
	}()
}

//Verifies a transaction and adds it to the transaction map and the unusedTransactions map, if successfully verified.
func handleTransaction(t o.Transaction) {
	tLock.Lock()
	defer tLock.Unlock()
	if t.VerifyTransaction() != true {
		return
	}
	_, alreadyReceived := transactions[t.ID]
	if !alreadyReceived {
		transactions[t.ID] = t
		unusedTransactions[t.ID] = true
	}
}

//Verifies the block signature and the draw value of a block, and calls addBlock if successful.
func handleBlock(b o.Block) {
	if b.Slot == 0 { //*TODO Should probably add some security measures so you can't fake a genesis block
		handleGenesisBlock(b)
		return
	}
	if !b.VerifyBlock(b.BakerID) || !verifyDraw(b) {
		return
	}
	addBlock(b)
}

func handleGenesisBlock(b o.Block) { //*TODO Proper genesisdata should be added and handled
	blocks.add(b)
	currentHead = b.CalculateBlockHash()
	lastFinalized = b.CalculateBlockHash()
}

// Calculates and compares pathWeigth of currentHead and a new block not extending the tree of the head.
// Updates the head and initiates rollbacks accordingly
func comparePathWeight(b o.Block) {
	l := 1
	block := b
	for {
		if !blocks.contains(block.ParentPointer) {
			return
		}
		parent := blocks.get(block.ParentPointer)

		if parent.Slot == 0 { // *TODO Should probably refactor to use the last finalized block, to prevent excessive work
			break
		}
		block = parent
		l += 1
	}
	if l < currentLength {
		return
	}

	if l > currentLength || calculateDraw(blocks.get(currentHead)) < calculateDraw(b) {
		if rollback(b) {
			currentHead = b.CalculateBlockHash()
			currentLength = l
			sendBranchToTL()
		}
	}
}

//Manages rollback in the case of branch shifting. Returns true if successful and false otherwise.
func rollback(newHead o.Block) bool {
	tLock.Lock()
	defer tLock.Unlock()
	oldUnusedTransmap := make(map[string]bool)
	for k, v := range unusedTransactions {
		oldUnusedTransmap[k] = v
	}
	head := blocks.get(currentHead)
	for {
		transactionsUnused(head)
		head = blocks.get(head.ParentPointer)
		if head.CalculateBlockHash() == lastFinalized {
			break
		}
	}
	var newBranch []o.Block
	for {
		newBranch = append(newBranch, newHead)
		newHead = blocks.get(newHead.ParentPointer)
		if newHead.CalculateBlockHash() == lastFinalized {
			break
		}
	}
	noBadBlocks := true
	for i := len(newBranch) - 1; i >= 0; i-- {
		if noBadBlocks {
			noBadBlocks = transactionsUsed(newBranch[i])
		} else {
			badBlocks[newBranch[i].CalculateBlockHash()] = true
		}
	}
	if !noBadBlocks {
		unusedTransactions = oldUnusedTransmap //If rollback involved invalid blocks, we return to the old map
	}
	return noBadBlocks
}

//Helper function for rollback
func transactionsUnused(b o.Block) {
	trans := b.BlockData.Trans
	for _, t := range trans {
		unusedTransactions[t.ID] = true
	}
}

//Removes the transactions used in a block from unusedTransactions, and saves transactions that we have not already saved.
//It returns false if the block reuses any transactions already spent on the chain and marks the block as bad. It returns true otherwise.
func transactionsUsed(b o.Block) bool {
	oldUnusedTransmap := make(map[string]bool)
	for k, v := range unusedTransactions {
		oldUnusedTransmap[k] = v
	}
	trans := b.BlockData.Trans
	for _, t := range trans {
		_, alreadyStored := transactions[t.ID]
		if !alreadyStored {
			transactions[t.ID] = t
			unusedTransactions[t.ID] = true
		}
		_, unused := unusedTransactions[t.ID]
		if !unused {
			badBlocks[b.CalculateBlockHash()] = true
			unusedTransactions = oldUnusedTransmap
			return false
		}
		delete(unusedTransactions, t.ID)
	}
	return true
}

//Used after a rollback to send the branch of the new head to the transaction layer
func sendBranchToTL() {

	//TODO
}

//Used to send a new head to the transaction layer
func sendBlockToTL(block o.Block) {

	//TODO
}

//Updates the head if the block extends our current head, and otherwise calls comparePathWeight
func updateHead(b o.Block) {
	if b.ParentPointer == currentHead {
		tLock.Lock()
		defer tLock.Unlock()
		if transactionsUsed(b) {
			currentHead = b.CalculateBlockHash()
			currentLength += 1
			sendBlockToTL(b)
		}
	} else {
		comparePathWeight(b)
	}
	fmt.Println(blocks.get(currentHead).Slot) //Used for testing purposes
}

// Checks if a block is a legal extension of the tree, and otherwise marks it as a bad block
func isLegalExtension(b o.Block) bool {
	_, badParent := badBlocks[b.ParentPointer]
	if badParent {
		badBlocks[b.CalculateBlockHash()] = true
		return false
	}
	if blocks.contains(b.ParentPointer) && blocks.get(b.ParentPointer).Slot >= b.Slot { //Check that slot of parent is smaller
		badBlocks[b.CalculateBlockHash()] = true
		return false
	}
	return true
}

//Adds a block to our blockmap and calls updateHead if it's a legal extension
func addBlock(b o.Block) {
	blocks.add(b)
	if isLegalExtension(b) {
		updateHead(b)
	}
}

//Sends a block to the P2P layer to be broadcasted
func broadcastBlock(b o.Block) {

}

func calculateDraw(b o.Block) int {
	//*TODO
	testDrawVal++
	return testDrawVal
}

func verifyDraw(b o.Block) bool {
	// *TODO
	return true
}

type skov struct {
	m map[string]o.Block
	l sync.RWMutex
}

func (s *skov) add(block o.Block) {
	hash := block.CalculateBlockHash()
	s.m[hash] = block
}

func (s *skov) get(blockHash string) o.Block {
	return s.m[blockHash]
}

func (s *skov) contains(blockHash string) bool {
	_, exists := s.m[blockHash]
	return exists
}

func (s *skov) lock() {
	s.l.Lock()
}

func (s *skov) unlock() {
	s.l.Unlock()
}

func (s *skov) rlock() {
	s.l.RLock()
}

func (s *skov) runlock() {
	s.l.RUnlock()
}
