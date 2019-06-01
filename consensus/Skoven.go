package consensus

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	o "github.com/nfk93/blockchain/objects"
	"io/ioutil"
	"sync"
)

var unusedTransactions map[string]bool
var transactions map[string]o.TransData
var tLock sync.RWMutex
var finalLock sync.RWMutex
var blocks skov
var badBlocks map[string]bool
var currentHead string
var currentLength int
var lastFinalized string
var lastFinalizedSlot uint64
var channels o.ChannelStruct
var isVerbose bool
var saveGraphFiles bool
var pendingBlocks []o.Block
var pendingBlocksLock sync.Mutex
var handlingBlocks sync.Mutex
var genesisReceived = false

func StartConsensus(channelStruct o.ChannelStruct, pkey crypto.PublicKey, skey crypto.SecretKey, verbose, saveGraphsToFile bool) {
	pk = pkey
	sk = skey
	isVerbose = verbose
	saveGraphFiles = saveGraphsToFile
	channels = channelStruct
	unusedTransactions = make(map[string]bool)
	transactions = make(map[string]o.TransData)
	badBlocks = make(map[string]bool)
	pendingBlocks = make([]o.Block, 0)
	blocks.m = make(map[string]o.Block)

	// Start processing blocks on one thread, non-concurrently
	go func() {
		for {
			block := <-channels.BlockFromP2P
			func() {
				handlingBlocks.Lock()
				defer handlingBlocks.Unlock()
				handleBlock(block)
				checkPendingBlocks()
			}()
		}
	}()
	// Start processing transactions on one thread, concurrently
	go func() {
		for {
			trans := <-channels.TransFromP2P
			if isVerbose {
				if trans.GetType() == o.TRANSACTION {
					fmt.Printf("Transaction of %v K from %v to %v.\n", trans.Transaction.Amount, trans.Transaction.From.Hash()[0:10], trans.Transaction.To.Hash()[0:10])
				}
				if trans.GetType() == o.CONTRACTCALL {
					fmt.Printf("ContractCall %s received", trans.ContractCall.Nonce)
				}
				if trans.GetType() == o.CONTRACTINIT {
					fmt.Printf("ContractInit %s received", trans.ContractInit.Nonce)
				}
			}
			go handleTransData(trans)
		}
	}()
}

func checkPendingBlocks() {
	foundBlockToAdd := false
	func() {
		pendingBlocksLock.Lock()
		defer pendingBlocksLock.Unlock()
		for i, block := range pendingBlocks {
			if block.Slot <= getCurrentSlot() && blocks.contains(block.ParentPointer) {
				if !block.ValidateBlock() {
					fmt.Println("Consensus could not validate block:", block.CalculateBlockHash())
					pendingBlocks = append(pendingBlocks[:i], pendingBlocks[i+1:]...)
					return
				}
				if !ValidateDraw(block, getLeadershipNonce(block.Slot), hardness) {
					fmt.Println("Consensus could not validate draw of block:", block.CalculateBlockHash())
					pendingBlocks = append(pendingBlocks[:i], pendingBlocks[i+1:]...)
					return
				}
				addBlock(block)
				foundBlockToAdd = true
				pendingBlocks = append(pendingBlocks[:i], pendingBlocks[i+1:]...)
				break
			}
		}
	}()
	if foundBlockToAdd {
		checkPendingBlocks()
	}
}

//Verifies a transaction and adds it to the transaction map and the unusedTransactions map, if successfully verified.
func handleTransData(t o.TransData) {
	tLock.Lock()
	defer tLock.Unlock()
	if t.Verify() != true {
		return
	}
	transhash := t.Hash()
	_, alreadyReceived := transactions[transhash]
	if !alreadyReceived {
		transactions[transhash] = t
		unusedTransactions[transhash] = true
	}
}

//Verifies the block signature and the draw value of a block, and calls addBlock if successful.
func handleBlock(b o.Block) {
	if b.Slot == 0 && !genesisReceived {
		fmt.Println("Genesis received! Starting blockchain protocol")
		genesisReceived = true
		handleGenesisBlock(b)
		return
	}
	if !genesisReceived {
		func() {
			pendingBlocksLock.Lock()
			defer pendingBlocksLock.Unlock()
			pendingBlocks = append(pendingBlocks, b)
		}()
		return
	}
	if b.Slot > getCurrentSlot() || !blocks.contains(b.ParentPointer) {
		func() {
			pendingBlocksLock.Lock()
			defer pendingBlocksLock.Unlock()
			pendingBlocks = append(pendingBlocks, b)
		}()
		return
	}
	if !b.ValidateBlock() {
		if isVerbose {
			fmt.Println("Consensus could not validate block:", b.CalculateBlockHash())
		}
		return
	}
	if !ValidateDraw(b, getLeadershipNonce(b.Slot), hardness) {
		fmt.Println("Consensus could not validate draw of block:", b.CalculateBlockHash())
		return
	}
	addBlock(b)
}

func handleGenesisBlock(b o.Block) {
	blocks.add(b)
	processGenesisData(b.BlockData.GenesisData)
	sendBlockToTL(b)
	currentHead = b.CalculateBlockHash()
	lastFinalized = b.CalculateBlockHash()
	lastFinalizedSlot = 0
}

// Calculates and compares pathWeight of currentHead and a new block not extending the tree of the head.
// Updates the head and initiates rollbacks accordingly
func comparePathWeight(b o.Block) {
	l := 1
	block := b
	for {
		if !blocks.contains(block.ParentPointer) {
			return
		}
		parent := blocks.get(block.ParentPointer)

		if parent.CalculateBlockHash() == lastFinalized {
			l += int(lastFinalizedSlot)
			break
		}
		block = parent
		l += 1
	}

	if l < currentLength {
		return
	}

	if l > currentLength || compareDrawVal(b) {
		if rollback(b) {
			currentHead = b.CalculateBlockHash()
			currentLength = l
		}
	}
}

//Compares the draw value of a block with the current head
func compareDrawVal(b o.Block) bool {
	headDraw := CalculateDrawValue(blocks.get(currentHead), getLeadershipNonce(b.Slot))
	blockDraw := CalculateDrawValue(b, getLeadershipNonce(b.Slot))
	if headDraw.Cmp(blockDraw) == -1 {
		return true
	}
	return false
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
		if head.CalculateBlockHash() == newHead.LastFinalized {
			break
		}
	}
	var newBranch []o.Block
	head = newHead
	for {
		newBranch = append(newBranch, head)
		head = blocks.get(head.ParentPointer)
		if head.CalculateBlockHash() == newHead.LastFinalized {
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
	} else {
		sendBranchToTL(newBranch)
	}
	return noBadBlocks
}

//Helper function for rollback
func transactionsUnused(b o.Block) {
	trans := b.BlockData.Trans
	for _, t := range trans {
		transhash := t.Hash()
		unusedTransactions[transhash] = true
	}
}

//Removes the transdata used in a block from unusedTransactions, and saves transdata that we have not already saved.
//It returns false if the block reuses any transactions already spent on the chain and marks the block as bad. It returns true otherwise.
func transactionsUsed(b o.Block) bool {
	oldUnusedTransmap := make(map[string]bool)
	for k, v := range unusedTransactions {
		oldUnusedTransmap[k] = v
	}
	trans := b.BlockData.Trans
	for _, t := range trans {
		thash := t.Hash()
		_, alreadyStored := transactions[thash]
		if !alreadyStored {
			transactions[thash] = t
			unusedTransactions[thash] = true
		}
		_, unused := unusedTransactions[thash]
		if !unused {
			badBlocks[b.CalculateBlockHash()] = true
			unusedTransactions = oldUnusedTransmap
			return false
		}
		delete(unusedTransactions, thash)
	}
	return true
}

//Used after a rollback to send the branch of the new head to the transaction layer
func sendBranchToTL(branch []o.Block) {
	for i := len(branch) - 1; i >= 0; i-- {
		sendBlockToTL(branch[i])
	}
}

//Used to send a new head to the transaction layer
func sendBlockToTL(block o.Block) {
	channels.BlockToTrans <- block
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
	log() //Used for testing purposes
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

func log() {
	if isVerbose {
		fmt.Println(blocks.get(currentHead).Slot)
	}
}

func getUnusedTransactions() []o.TransData {
	tLock.Lock()
	defer tLock.Unlock()
	trans := make([]o.TransData, len(unusedTransactions))
	i := 0
	for k := range unusedTransactions {
		td := transactions[k]
		trans[i] = td
		i++
	}
	return trans
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

func SwitchVerbose() {
	isVerbose = !isVerbose
}

func printBlockTreeGraphToFile(filename string, blocks map[string]o.Block) error {
	bytes := getDotString(blocks)
	return ioutil.WriteFile("out/"+filename, bytes, 0644)
}

func getDotString1(blocks map[string]o.Block) []byte {
	// write blockdata
	logstring := "digraph G {\n"
	for k, v := range blocks {
		str := fmt.Sprintf("subgraph %s { node [style=filled,color=white]; style=filled; color=lightgrey; %s; label=\"blockhash: %s\\nslot: %s\\nbakerid: %s\"; }\n",
			"cluster"+k, "block"+k, k, v.Slot, v.BakerID.Hash())
		logstring += str
	}
	for k, v := range blocks {
		blockdatastr := fmt.Sprintf("%s [shape=record, label=\" {", "block"+k)
		if len(v.BlockData.Trans) == 0 {
			blockdatastr += "empty } \" ];\n"
		} else {
			for i, data := range v.BlockData.Trans {
				switch data.GetType() {
				case o.TRANSACTION:
					blockdatastr += fmt.Sprintf("| %d: TRANSACTION(from: %s, to: %s, amount: %d, id: %s) ", i,
						data.Transaction.From.Hash(), data.Transaction.To.Hash(), data.Transaction.Amount, data.Transaction.ID)
				case o.CONTRACTINIT:
					ci := data.ContractInit
					blockdatastr += fmt.Sprintf("| %d: CONTRACTINIT(owner: %s, code: [...], gas: %d, prepaid: %d, "+
						"storagelimit: %d) ", i, ci.Owner, ci.Gas, ci.Prepaid, ci.StorageLimit)
				case o.CONTRACTCALL:
					cc := data.ContractCall
					blockdatastr += fmt.Sprintf("| %d: CONTRACTCALL(caller: %s, address: %s, entry: %s, params: %s, "+
						"amount: %d, gas: %d", i, cc.Caller, cc.Address, cc.Entry, cc.Params, cc.Amount, cc.Gas)
				}
			}
			blockdatastr += "} \" ];\n"
		}
		logstring += blockdatastr
	}
	for k, v := range blocks {
		logstring += fmt.Sprintf("%s -> %s;\n", "block"+v.ParentPointer, "block"+k)
		logstring += fmt.Sprintf("%s -> %s [style=dotted,color=lightblue];\n", "block"+k, "block"+v.LastFinalized)
	}
	logstring += "\n}"
	return []byte(logstring)
}

func getDotString(blocks map[string]o.Block) []byte {
	// write blockdata
	logstring := "digraph G {\n"
	for k, v := range blocks {
		blockdatastr := fmt.Sprintf("%s [shape=record, label=\"hash: %s\\nslot: %d\\nbaker: %s\"];\n", "block"+k,
			k[:6], v.Slot, v.BakerID.Hash()[:6])
		logstring += blockdatastr
	}
	for k, v := range blocks {
		logstring += fmt.Sprintf("%s -> %s;\n", "block"+v.ParentPointer, "block"+k)
	}
	logstring += "\n}"
	return []byte(logstring)
}
