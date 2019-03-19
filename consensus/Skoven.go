package consensus

import (
	o "github.com/nfk93/blockchain/objects"
)

var unusedTransactions map[string]bool
var transactions map[string]o.Transaction
var blocks map[string]o.Block
var currentHead string
var currentLength int
var lastFinalized string

//Calculates and compares pathWeigth of currentHead and a new block not extending the tree of the head.
// Updates the head and initiates rollbacks accordingly
func comparePathWeight(b o.Block) {
	len := 1
	for {
		parent := blocks[b.ParentPointer]
		if parent.Slot == 0 { // *TODO Should probably refactor to use the last finalized block, to prevent excessive work
			break
		}
		len += 1
	}
	if len < currentLength {
		return
	}

	if len > currentLength {
		rollback()
		currentHead = b.HashBlock()
		currentLength = len
		return
	}

	if calculateDraw(blocks[currentHead]) < calculateDraw(b) {
		rollback()
		currentHead = b.HashBlock()
	}
}

//Manages rollback in the case of branch shifting
func rollback() {
	//* TODO
}

//Removes the transactions used in a block from unusedTransactions, and saves transactions that we have not already saved.
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
	blocks[b.HashBlock()] = b
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
