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
var blocks skov
var currentHead string
var currentHeadLock sync.RWMutex
var currentLength int
var channels o.ChannelStruct
var isVerbose bool
var saveGraphFiles bool
var pendingBlocks []o.Block
var pendingBlocksLock sync.Mutex
var genesisReceived = false

func StartConsensus(channelStruct o.ChannelStruct, pkey crypto.PublicKey, skey crypto.SecretKey, verbose, saveGraphsToFile bool) {
	pk = pkey
	sk = skey
	isVerbose = verbose
	saveGraphFiles = saveGraphsToFile
	channels = channelStruct
	unusedTransactions = make(map[string]bool)
	transactions = make(map[string]o.TransData)
	pendingBlocks = make([]o.Block, 0)
	blocks.m = make(map[string]o.Block)

	// Start processing blocks on one thread, non-concurrently
	go func() {
		for {
			block := <-channels.BlockFromP2P
			handleBlock(block)
		}
	}()
	// Start processing transactions on one thread, concurrently
	go func() {
		for {
			trans := <-channels.TransFromP2P
			if isVerbose {
				if trans.GetType() == o.TRANSACTION {
					fmt.Printf("received transaction: %d from %s to %s.\n", trans.Transaction.Amount, trans.Transaction.From.Hash()[0:6], trans.Transaction.To.Hash()[0:6])
				}
				if trans.GetType() == o.CONTRACTCALL {
					fmt.Printf("received contractcall with signature %s\n", trans.ContractCall.Signature[:6]+"...")
				}
				if trans.GetType() == o.CONTRACTINIT {
					fmt.Printf("received contractinit with signature %s\n", trans.ContractInit.Signature[:6]+"...")
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
				if !ValidateBlock(block) {
					pendingBlocks = append(pendingBlocks[:i], pendingBlocks[i+1:]...)
					if isVerbose {
						fmt.Println("Consensus could not validate block:", block.CalculateBlockHash())
					}
				} else {
					func() {
						blocks.lock()
						defer blocks.unlock()
						addBlock(block)
					}()
					foundBlockToAdd = true
					pendingBlocks = append(pendingBlocks[:i], pendingBlocks[i+1:]...)
					break
				}
			}
		}
	}()
	if foundBlockToAdd {
		checkPendingBlocks()
	}
}

func ValidateBlock(b o.Block) bool {
	// validate signature
	if !b.ValidateBlock() {
		if isVerbose {
			fmt.Println("block signature is invalid for block ", b.CalculateBlockHash())
		}
		return false
	}

	// validate slot is higher than parent
	blocks.rlock()
	defer blocks.runlock()
	parentblock := blocks.get(b.ParentPointer)
	if b.Slot <= parentblock.Slot {
		if isVerbose {
			fmt.Println("parentblock has lower slotnumber than new block ", b.CalculateBlockHash())
		}
		return false
	}

	// get relevant finaldata
	finalEpoch := getFinalDataIndex(b.Slot)
	finalLock.RLock()
	defer finalLock.RUnlock()
	fd, exists := finalData[finalEpoch]
	if !exists {
		if isVerbose {
			fmt.Println("can't determine final data for block ", b.CalculateBlockHash())
		}
		return false
	}

	// check that blocknonce is correct
	if blockNonceValid := b.ValidateBlockNonce(fd.leadershipNonce); !blockNonceValid {
		if isVerbose {
			fmt.Println("can't validate BlockNonce of block ", b.CalculateBlockHash())
		}
		return false
	}

	// check that the lastfinalized pointer is correct
	if lastFinalValid := checkLastFinalizedValidity(b, finalEpoch, fd); !lastFinalValid {
		if isVerbose {
			fmt.Println("can't validate validity of lastfinal of block ", b.CalculateBlockHash())
		}
		return false
	}

	// validate that the sender has actually won the lottery in the slot he claims
	validDraw := ValidateDraw(b.Slot, b.Draw, b.BakerID, fd, hardness)
	if !validDraw {
		if isVerbose {
			fmt.Println("can't validate draw of block: ", b.CalculateBlockHash())
		}
		return false
	}
	return true
}

func getFinalDataIndex(s uint64) uint64 {
	finalEpoch := uint64(0)
	if e := getEpoch(s); e > 0 {
		finalEpoch = e - 1
	}
	return finalEpoch
}

func checkLastFinalizedValidity(block o.Block, finalEpoch uint64, data FinalData) bool {
	// check that last finalized pointer in the block is the same as the one we expect it to be based on the slot of the
	// block
	if block.LastFinalized != data.blockHash {
		return false
	}

	// check that the last finalized block is actually in the branch that we are extending.
	block_ := block
	slot := block.Slot
	blocks.rlock()
	defer blocks.runlock()
	for slot > finalEpoch*finalizeInterval {
		if block_.ParentPointer == data.blockHash {
			return true
		} else {
			block_ = blocks.get(block_.ParentPointer)
			slot = block_.Slot
		}
	}
	return false
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

func handleBlock2(b o.Block) {
	blocks.rlock()
	defer blocks.runlock()
	// check if the parent of a block exists, and if it doesn't it adds it to pendingblocks
	if parentExists := blocks.contains(b.ParentPointer); !parentExists || b.Slot >= getCurrentSlot() {
		func() {
			pendingBlocksLock.Lock()
			defer pendingBlocksLock.Unlock()
			pendingBlocks = append(pendingBlocks, b)
		}()
		return
	}
	/*if lastFinalValid := checkLastFinalizedValidity(b.LastFinalized); !lastFinalValid {

	} */
}

//Verifies the block signature and the draw value of a block, and calls addBlock if successful.
func handleBlock(b o.Block) {
	if isVerbose {
		fmt.Println("received block ", b.CalculateBlockHash()[:6]+"...")
	}
	if !genesisReceived {
		if b.Slot == 0 {
			fmt.Println("Genesis received! Starting blockchain protocol")
			genesisReceived = true
			handleGenesisBlock(b)
			return
		} else {
			func() {
				pendingBlocksLock.Lock()
				defer pendingBlocksLock.Unlock()
				pendingBlocks = append(pendingBlocks, b)
			}()
			return
		}
	}

	done := false
	func() {
		blocks.rlock()
		defer blocks.runlock()
		// check if the parent of a block exists, and if it doesn't it adds it to pendingblocks
		if parentExists := blocks.contains(b.ParentPointer); !parentExists || b.Slot > getCurrentSlot() {
			func() {
				pendingBlocksLock.Lock()
				defer pendingBlocksLock.Unlock()
				pendingBlocks = append(pendingBlocks, b)
			}()
			if isVerbose {
				if !parentExists {
					fmt.Println(fmt.Sprintf("can't process block %s yet, missing parent %s",
						b.CalculateBlockHash()[:6]+"...", b.ParentPointer))
				} else {
					fmt.Println(fmt.Sprintf("can't process block %s yet, its slot (%d) is higher than currentslot (%d)",
						b.CalculateBlockHash()[:6]+"...", b.Slot, currentSlot))
				}
			}
			done = true
			return
		}
		// checks that all contents of the block can be verified
		if !ValidateBlock(b) {
			if isVerbose {
				fmt.Println("Consensus could not validate block:", b.CalculateBlockHash()[:6]+"...")
			}
			done = true
			return
		}
	}()
	if done {
		return
	} else {
		blocks.lock()
		defer blocks.unlock()
		addBlock(b)
	}
}

func handleGenesisBlock(b o.Block) {
	blocks.add(b)
	processGenesisData(b.BlockData.GenesisData, b.CalculateBlockHash())
	sendBlockToTL(b)
	setCurrentHead(b.CalculateBlockHash())
}

// Calculates and compares pathWeight of currentHead and a new block not extending the head of the tree.
// Updates the head and initiates rollbacks accordingly
// PRECONDITION: there is a path from both currenthead and new block to lastFinalized of the new block
//				 blocks is read locked
func pathIsLongerThanCurrentHead(b o.Block) bool {
	l := 0
	block := b

	head := blocks.get(getCurrentHead())
	headLastFinal := head.LastFinalized

	for {
		l += 1
		if block.ParentPointer == headLastFinal {
			break
		} else {

		}
		block = blocks.get(block.ParentPointer)
	}

	if l < currentLength {
		return false
	}
	if l > currentLength || HasHigherDrawVal(b, head) {
		return true
	}
	return false
}

// Compares the draw value of the first block to the second. Returns true if it is higher, false otherwise.
// Precondition: b1 and b2 have same lastfinalized
func HasHigherDrawVal(b1, b2 o.Block) bool {
	finalLock.RLock()
	defer finalLock.RUnlock()
	fd := finalData[getFinalDataIndex(b1.Slot)]

	draw1 := calculateDrawValue(b1.Slot, b1.Draw, fd.leadershipNonce)
	draw2 := calculateDrawValue(b2.Slot, b2.Draw, fd.leadershipNonce)
	if draw2.Cmp(draw1) == -1 {
		return true
	}
	return false
}

// Manages rollback in the case of branch shifting. Returns true if successful and false otherwise.
// PRECONDITION: newHead is in the tree and the tree is read locked
func rollback(newHead o.Block) bool {
	tLock.Lock()
	defer tLock.Unlock()
	// save the current set of unused transactions
	oldUnusedTransmap := make(map[string]bool)
	for k, v := range unusedTransactions {
		oldUnusedTransmap[k] = v
	}

	head := blocks.get(getCurrentHead())
	transUsedInOldBranch := make([]o.TransData, 0)
	for {
		transUsedInOldBranch = append(transUsedInOldBranch, head.BlockData.Trans...)
		if head.ParentPointer == newHead.LastFinalized {
			break
		}
		head = blocks.get(head.ParentPointer)
	}
	markTransactionsAsUnused(transUsedInOldBranch)

	var newBranch []o.Block
	head = newHead
	for {
		newBranch = append(newBranch, head)
		if head.ParentPointer == newHead.LastFinalized {
			break
		}
		head = blocks.get(head.ParentPointer)
	}

	for i := len(newBranch) - 1; i >= 0; i-- {
		appliedTransactionsSuccess := markTransactionsAsUsed(newBranch[i])
		if !appliedTransactionsSuccess {
			fmt.Println("Transaction duplication in branch ending at leaf ", newHead.CalculateBlockHash())
			unusedTransactions = oldUnusedTransmap //If rollback involved invalid blocks, we return to the old map
			return false
		}
	}
	currentHead = newHead.CalculateBlockHash()
	currentLength = len(newBranch)
	sendBranchToTL(newBranch)
	return true
}

func markTransactionsAsUnused(data []o.TransData) {
	for _, v := range data {
		transhash := v.Hash()
		unusedTransactions[transhash] = true
	}
}

//Removes the transdata used in a block from unusedTransactions, and saves transdata that we have not already saved.
//It returns false if the block reuses any transactions already spent on the chain and marks the block as bad. It returns true otherwise.
func markTransactionsAsUsed(b o.Block) bool {
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
			unusedTransactions = oldUnusedTransmap
			fmt.Println(fmt.Sprintf("invalid block %s, reuses transactions", b.CalculateBlockHash()))
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

//Updates the head if the block extends our current head, and otherwise calls pathIsLongerThanCurrentHead
func updateHead(b o.Block) {
	if b.ParentPointer == getCurrentHead() {
		tLock.Lock()
		defer tLock.Unlock()
		if noDuplicateTransaction := markTransactionsAsUsed(b); noDuplicateTransaction {
			if b.LastFinalized == blocks.get(getCurrentHead()).LastFinalized {
				currentLength += 1
			} else {
				currentLength = lengthToLastFinal(b)
			}
			setCurrentHead(b.CalculateBlockHash())
			sendBlockToTL(b)
		}
	} else {
		if pathIsLongerThanCurrentHead(b) {
			rollback(b)
		}
	}
	if isVerbose {
		fmt.Println("head of tree is block", getCurrentHead())
	}
}

// PRECONDITION: blocks is at least read locked and blocks lastfinal IS on its own branch
func lengthToLastFinal(block_ o.Block) int {
	block := block_
	lastFinal := block.LastFinalized

	l := 0
	for {
		l += 1
		if block.ParentPointer == lastFinal {
			break
		} else {

		}
		block = blocks.get(block.ParentPointer)
	}
	return l
}

//Adds a block to our blockmap and calls updateHead if it's a legal extension
// PRECONDITION: blocks is write locked
func addBlock(b o.Block) {
	blocks.add(b)
	updateHead(b)
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

func getCurrentHead() string {
	currentHeadLock.RLock()
	defer currentHeadLock.RUnlock()
	return currentHead
}

func setCurrentHead(headHash string) {
	currentHeadLock.Lock()
	defer currentHeadLock.Unlock()
	currentHead = headHash
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
		if v.Slot == 0 {
			blockdatastr := fmt.Sprintf("%s [shape=record, label=\"hash: %s\\nslot: %d\\nbaker: %s\"];\n",
				"block"+k, k[:6], v.Slot, v.BakerID.Hash()[:6])
			logstring += blockdatastr
		} else {
			blockdatastr := fmt.Sprintf("%s [shape=record, label=\"hash: %s\\nslot: %d\\nbaker: %s\\nlastFinal: %s\\ntransactions: %d\"];\n",
				"block"+k, k[:6], v.Slot, v.BakerID.Hash()[:6], v.LastFinalized[:6], len(v.BlockData.Trans))
			logstring += blockdatastr
		}
	}
	for k, v := range blocks {
		logstring += fmt.Sprintf("%s -> %s;\n", "block"+v.ParentPointer, "block"+k)
	}
	logstring += "\n}"
	return []byte(logstring)
}
