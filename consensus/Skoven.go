package consensus

import (
	o "github.com/nfk93/blockchain/objects"
	"sync"
)

var unusedTransactions map[string]bool
var transactions map[string]o.Transaction
var tLock sync.RWMutex
var blocks skov
var badblocks map[string]bool
var currentHead string
var currentLength int
var lastFinalized string

//Calculates and compares pathWeigth of currentHead and a new block not extending the tree of the head.
// Updates the head and initiates rollbacks accordingly
func comparePathWeight(b o.Block) {
	len := 1
	for {
		if !blocks.contains(b.ParentPointer) {
			return
		}
		parent := blocks.get(b.ParentPointer)

		if parent.Slot == 0 { // *TODO Should probably refactor to use the last finalized block, to prevent excessive work
			break
		}
		len += 1
	}
	if len < currentLength {
		return
	}

	if len > currentLength {
		rollback(b)
		currentHead = b.HashBlock()
		currentLength = len
		return
	}

	if calculateDraw(blocks.get(currentHead)) < calculateDraw(b) {
		rollback(b)
		currentHead = b.HashBlock()
	}
}

//Manages rollback in the case of branch shifting
func rollback(newHead o.Block) {
	tLock.Lock()
	defer tLock.Unlock()
	newUnusedTransmap := make(map[string]bool)
	for k, v := range unusedTransactions {
		newUnusedTransmap[k] = v
	}
	head := blocks.get(currentHead)
	for {
		transactionsUnused(head, newUnusedTransmap)
		head = blocks.get(head.ParentPointer)
		if head.HashBlock() == lastFinalized {
			break
		}
	}
	//*TODO not done
}

//Helper function for rollback
func transactionsUnused(b o.Block, utMap map[string]bool) {
	trans := b.BlockData.Trans
	for _, t := range trans {
		utMap[t.ID] = true
	}
}

//Removes the transactions used in a block from unusedTransactions, and saves transactions that we have not already saved.
// *TODO need to add check for double spending transactions
func transactionsUsed(b o.Block) {
	trans := b.BlockData.Trans
	for _, t := range trans {
		transactions[t.ID] = t //Might not be necessary
		delete(unusedTransactions, t.ID)
	}
}

//Updates the head if the block extends our current head, and otherwise calls comparePathWeight
func updateHead(b o.Block) {
	if b.ParentPointer == currentHead {
		currentHead = b.HashBlock()
		currentLength += 1
		transactionsUsed(b)
	} else {
		comparePathWeight(b)
	}
}

//Adds a block to our blockmap and calls updateHead
func addBlock(b o.Block) {
	blocks.get(b.HashBlock()) = b
	updateHead(b)
}

//Verifies a transaction and adds it to the transaction map and the unusedTransactions map, if successfully verified.
func transactionReceived(t o.Transaction) {
	if t.VerifyTransaction() != true {
		return
	}
	_, alreadyReceived := transactions[t.ID]
	if !alreadyReceived {
		transactions[t.ID] = t
		unusedTransactions[t.ID] = true
	}
}

//Sends a block to the P2P layer to be broadcasted
func broadcastBlock(b o.Block) {

}

//Verifies the block signature and the draw value of a block, and calls addBlock if successful.
func blockReceived(b o.Block) {
	if !b.VerifyBlock(b.BakerID) || !verifyDraw(b) {
		return
	}
	addBlock(b)
}

func calculateDraw(b o.Block) int {
	//*TODO
	return 0
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
	hash := block.HashBlock()
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
