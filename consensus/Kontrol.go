package consensus

import (
	. "github.com/nfk93/blockchain/crypto"
	o "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/transaction"
	"sync"
	"time"
)

var slotLength time.Duration
var currentSlot int
var slotLock sync.RWMutex
var currentStake int
var currentNonce o.BlockNonce
var hardness float64
var sk SecretKey
var pk PublicKey
var lastFinalizedLedger map[PublicKey]int

func runSlot() { //Calls drawLottery every slot and increments the currentSlot after slotLength time.
	currentSlot = 1
	for {
		if (currentSlot)%100 == 0 {
			//finalize(currentSlot)
		}
		go drawLottery(currentSlot)
		time.Sleep(slotLength)
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
	go runSlot()
	go transaction.StartTransactionLayer(channels.BlockToTrans,
		channels.StateFromTrans, channels.FinalizeToTrans, channels.BlockFromTrans,
		channels.TransToTrans, sk, genesisData.InitialState)
}

func finalize(slot int) {
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
	lastFinalizedLedger = state.Ledger
}

func drawLottery(slot int) {
	//winner, draw := o.CalculateDraw(currentNonce, hardness, sk, pk, currentStake, slot)
	//if winner {

	//}
}

func computeTransactions() o.Block { //Sends all unused transactions to the transaction layer for the transaction layer to process for the new block

	/* TODO
	   for {
	   	block := <-blockFromTL
	   	return block
	   }
	*/
	return o.Block{}
}
