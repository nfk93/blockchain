package consensus

import (
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
var currentStake uint64
var systemStake uint64
var hardness float64
var genesisTime time.Time
var sk SecretKey
var pk PublicKey
var lastFinalizedLedger map[string]uint64
var ledgerLock sync.RWMutex
var leadershipNonce string
var leadershipLock sync.RWMutex

func runSlot() { //Calls drawLottery every slot and increments the currentSlot after slotLength time.
	currentSlot = 1
	finalizeInterval := uint64(50)
	offset := time.Since(genesisTime)
	for {
		if (currentSlot)%finalizeInterval == 0 {
			finalize(currentSlot - (finalizeInterval / 2))
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
		currentSlot++
		slotLock.Unlock()
	}
}

func getCurrentSlot() uint64 {
	slotLock.RLock()
	defer slotLock.RUnlock()
	return currentSlot
}

func processGenesisData(genesisData o.GenesisData) {
	// TODO  -  Use GenesisTime when going away from two-phase implementation
	hardness = genesisData.Hardness
	slotLength = genesisData.SlotDuration
	lastFinalizedLedger = genesisData.InitialState.Ledger
	leadershipNonce = genesisData.Nonce
	currentStake = lastFinalizedLedger[pk.Hash()]
	systemStake = genesisData.InitialState.TotalStake
	genesisTime = genesisData.GenesisTime
	go runSlot()
	go transaction.StartTransactionLayer(channels, saveGraphFiles)
}

func finalize(slot uint64) { //TODO add generation of new leadershipNonce (when doing so also change all places using it)
	finalLock.Lock()
	defer finalLock.Unlock()
	head := blocks.get(currentHead)
	for {
		if head.Slot <= slot {
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

func updateStake() {
	state := <-channels.StateFromTrans
	ledgerLock.Lock()
	defer ledgerLock.Unlock()
	lastFinalizedLedger = state.Ledger
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
		getUnusedTransactions(),
		sk,
		pk,
		slot,
		draw,
		o.CreateNewBlockNonce(leadershipNonce, sk, slot),
		lastFinalized}
	channels.TransToTrans <- blockData
	go sendBlock()
}

func sendBlock() {
	block := <-channels.BlockFromTrans
	//channels.BlockFromP2P <- block // TODO change this when using P2P
	channels.BlockToP2P <- block
}

func getLotteryPower(pk PublicKey) float64 {
	ledgerLock.RLock()
	defer ledgerLock.RUnlock()
	return float64(lastFinalizedLedger[pk.Hash()]) / float64(systemStake)
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
