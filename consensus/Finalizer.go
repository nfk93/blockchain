package consensus

import (
	. "github.com/nfk93/blockchain/crypto"
	o "github.com/nfk93/blockchain/objects"
	"sync"
	"time"
)

var slotLength int
var currentSlot int
var slotLock sync.RWMutex
var currentStake int
var currentNonce o.BlockNonce
var hardness int
var sk SecretKey
var pk PublicKey

func runSlot() {
	currentSlot = 1
	for {
		if (currentSlot)%100 == 0 {
			finalize(currentSlot)
		}
		go drawLottery(currentSlot)
		time.Sleep(time.Duration(slotLength) * time.Second)
		slotLock.Lock()
		currentSlot++
		slotLock.Unlock()

	}
}

func finalize(slot int) {

}

func updateStake() {

}

func drawLottery(slot int) {
	winner, draw := o.CalculateDraw(currentNonce, hardness, sk, pk, currentStake, slot)
	if winner {

	}

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
