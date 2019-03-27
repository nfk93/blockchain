package consensus

import "time"

var slotLength int
var currentSlot int

func runSlot() {
	for {
		go drawLottery()
		time.Sleep(time.Duration(slotLength) * time.Second)
		currentSlot++
	}
}

func finalize() {

}

func updateStake() {

}

func drawLottery() {

}
