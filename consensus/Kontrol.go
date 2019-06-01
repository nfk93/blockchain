package consensus

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	o "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/transaction"
	"sort"
	"strconv"
	"sync"
	"time"
)

var slotLength time.Duration
var currentSlot uint64
var slotLock sync.RWMutex
var oldSystemStake uint64
var systemStake uint64
var hardness float64
var genesisTime time.Time
var sk SecretKey
var pk PublicKey
var lastFinalizedLedger map[string]uint64
var oldFinalizedLedger map[string]uint64
var ledgerLock sync.RWMutex
var leadershipNonce string
var oldLeadershipNonce string
var leadershipLock sync.RWMutex
var currentEpochSlot uint64 // The slot the current epoch began
var createBlockHead string
var createBlockTransData []o.TransData

func runSlot() { //Calls drawLottery every slot and increments the currentSlot after slotLength time.
	currentSlot = 1
	finalizeInterval := uint64(50)
	offset := time.Since(genesisTime)
	for {
		if (currentSlot)%finalizeInterval == 0 {
			finalize(currentSlot - (finalizeInterval / 2))
			currentEpochSlot = currentSlot
		}
		go drawLottery(currentSlot)
		timeSinceGenesis := time.Since(genesisTime) - offset
		if saveGraphFiles {
			go func() {
				blocks.l.Lock()
				defer blocks.l.Unlock()
				copy_ := make(map[string]o.Block)
				for k, v := range blocks.m {
					copy_[k] = v
				}
				err := printBlockTreeGraphToFile(fmt.Sprintf("slot%d", currentSlot), copy_)
				if err != nil {
					fmt.Println(fmt.Sprintf("error saving tree: %s", err.Error()))
				}
			}()
		}
		sleepyTime := time.Duration(currentSlot)*slotLength - timeSinceGenesis
		if sleepyTime > 0 {
			time.Sleep(sleepyTime)
		}
		slotLock.Lock()
		createBlockHead = getCurrentHead()
		createBlockTransData = getUnusedTransactions()
		currentSlot++
		slotLock.Unlock()
	}
}

func getCurrentSlot() uint64 {
	slotLock.RLock()
	defer slotLock.RUnlock()
	return currentSlot
}

func getCurrentEpochSlot() uint64 {
	slotLock.RLock()
	defer slotLock.RUnlock()
	return currentEpochSlot
}

func processGenesisData(genesisData o.GenesisData) {
	// TODO  -  Use GenesisTime when going away from two-phase implementation
	hardness = genesisData.Hardness
	slotLength = genesisData.SlotDuration
	lastFinalizedLedger = genesisData.InitialState.Ledger
	leadershipNonce = genesisData.Nonce
	systemStake = genesisData.InitialState.TotalStake
	genesisTime = genesisData.GenesisTime
	currentEpochSlot = 0
	go runSlot()
	go transaction.StartTransactionLayer(channels, saveGraphFiles)
}

func finalize(slot uint64) {
	finalLock.Lock()
	defer finalLock.Unlock()
	head := blocks.get(getCurrentHead())
	for {
		if head.Slot <= slot {
			newLeadershipNonce(head)
			finalHash := head.CalculateBlockHash()
			lastFinalized = finalHash
			lastFinalizedSlot = head.Slot
			go updateStake()
			channels.FinalizeToTrans <- finalHash
			break
		}
		head = blocks.get(head.ParentPointer)
	}
}

func newLeadershipNonce(finalBlock o.Block) {
	var buf bytes.Buffer
	head := finalBlock
	for {
		if head.Slot == lastFinalizedSlot {
			break
		}
		buf.WriteString(head.BlockNonce.Nonce)
		head = blocks.get(head.ParentPointer)
	}
	leadershipLock.Lock()
	defer leadershipLock.Unlock()
	oldLeadershipNonce = leadershipNonce
	leadershipNonce = HashSHA(buf.String())
}

func updateStake() {
	state := <-channels.StateFromTrans
	ledgerLock.Lock()
	defer ledgerLock.Unlock()
	oldFinalizedLedger = lastFinalizedLedger
	lastFinalizedLedger = state.Ledger
	oldSystemStake = systemStake
	systemStake = state.TotalStake
	if isVerbose {
		fmt.Println("Finalized Successfully")
		PrintFinalizedLedger()
	}
}

func drawLottery(slot uint64) {
	winner, draw := CalculateDraw(leadershipNonce, hardness, sk, pk, slot)
	if winner {
		if isVerbose {
			fmt.Println("We won slot " + strconv.Itoa(int(slot)))
		}
		generateBlock(draw, slot)
	}
}

//Sends all unused transactions to the transaction layer for the transaction layer to process for the new block
func generateBlock(draw string, slot uint64) {

	func() {
		handlingBlocks.Lock()
		defer handlingBlocks.Unlock()
		checkPendingBlocks()
	}()
	blockData := o.CreateBlockData{
		createBlockTransData,
		sk,
		pk,
		slot,
		draw,
		o.CreateNewBlockNonce(getLeadershipNonce(slot), sk, slot),
		lastFinalized,
		createBlockHead}
	channels.TransToTrans <- blockData
	go sendBlock()
}

func sendBlock() {
	block := <-channels.BlockFromTrans
	//channels.BlockFromP2P <- block // TODO change this when using P2P
	channels.BlockToP2P <- block
}

func getLotteryPower(pk PublicKey, slot uint64) float64 {
	ledgerLock.RLock()
	defer ledgerLock.RUnlock()
	if slot >= getCurrentEpochSlot() {
		return float64(lastFinalizedLedger[pk.Hash()]) / float64(systemStake)
	} else {
		return float64(oldFinalizedLedger[pk.Hash()]) / float64(oldSystemStake)
	}
}

func getLeadershipNonce(slot uint64) string {
	leadershipLock.RLock()
	defer leadershipLock.RUnlock()
	if slot >= getCurrentEpochSlot() {
		return leadershipNonce
	} else {
		return oldLeadershipNonce
	}
}

func GetLastFinalState() map[string]uint64 {
	return lastFinalizedLedger
}

func PrintFinalizedLedger() {
	ledger := lastFinalizedLedger
	var keyList []string
	for k := range ledger {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	for _, k := range keyList {
		fmt.Printf("Amount %v is owned by %v\n", ledger[k], k[:10])
	}
}
