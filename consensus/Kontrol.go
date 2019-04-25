package consensus

import (
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	o "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/transaction"
	"strconv"
	"sync"
	"time"
)

var slotLength time.Duration
var currentSlot int
var slotLock sync.RWMutex
var currentStake int
var systemStake int
var hardness float64
var genesisTime time.Time
var sk SecretKey
var pk PublicKey
var lastFinalizedLedger map[PublicKey]int
var ledgerLock sync.RWMutex
var leadershipNonce string
var leadershipLock sync.RWMutex

func runSlot() { //Calls drawLottery every slot and increments the currentSlot after slotLength time.
	currentSlot = 1
	for {
		if (currentSlot)%20 == 0 {
			finalize(currentSlot - 10)
		}
		go drawLottery(currentSlot)
		timeSinceGenesis := time.Since(genesisTime)
		sleepyTime := time.Duration(currentSlot)*slotLength - timeSinceGenesis
		if sleepyTime > 0 {
			time.Sleep(sleepyTime)
		}
		slotLock.Lock()
		currentSlot++
		slotLock.Unlock()
	}
}

func getCurrentSlot() int {
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
	currentStake = lastFinalizedLedger[pk]
	systemStake = genesisData.InitialState.TotalStake
	genesisTime = genesisData.GenesisTime
	go runSlot()
	go transaction.StartTransactionLayer(channels)
}

func finalize(slot int) { //TODO add generation of new leadershipNonce (when doing so also change all places using it)
	finalLock.Lock()
	defer finalLock.Unlock()
	head := blocks.get(currentHead)
	for {
		if head.Slot <= slot {
			finalHash := head.CalculateBlockHash()
			lastFinalized = finalHash
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
	fmt.Println(state)
	lastFinalizedLedger = state.Ledger
}

func drawLottery(slot int) {
	winner, draw := CalculateDraw(leadershipNonce, hardness, sk, pk, slot)
	if winner {
		fmt.Println("We won slot " + strconv.Itoa(slot))
		generateBlock(draw, slot)
	}
}

//Sends all unused transactions to the transaction layer for the transaction layer to process for the new block
func generateBlock(draw string, slot int) {
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
	channels.BlockFromP2P <- block // TODO change this when using P2P
	//channels.BlockToP2P <- block
}

func getLotteryPower(pk PublicKey) float64 {
	ledgerLock.RLock()
	defer ledgerLock.RUnlock()
	return float64(lastFinalizedLedger[pk]) / float64(systemStake)
}
