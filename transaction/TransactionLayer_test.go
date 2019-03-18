package transaction

import (
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
	"testing"
	"time"
)

func createBlock(t []Transaction, i int) Block {
	return Block{i + 1,
		strconv.Itoa(i),
		42,
		"VALID",
		42,
		"",
		Data{t},
		"",
		"",
	}

}

func TestReceiveBlock(t *testing.T) {
	sk1, p1 := KeyGen(256)
	_, p2 := KeyGen(256)
	t1 := Transaction{p1, p2, 200, "ID112", ""}
	t2 := Transaction{p1, p2, 300, "ID222", ""}
	t1.SignTransaction(sk1)
	t2.SignTransaction(sk1)
	b := createBlock([]Transaction{t1, t2}, 0)

	blockChannel, stateChannel, finalChannel := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, finalChannel)

	go func() {
		for {
			state := <-stateChannel
			if state.ledger[p2] != 500 {
				t.Error("P2 does not own 500")
			}
			return
		}
	}()

	blockChannel <- b

}

func TestTreeBuild(t *testing.T) {
	blockChannel, stateChannel, finalChannel := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, finalChannel)

	go func() {
		for {
			state := <-stateChannel
			fmt.Println("state ", state)
		}

	}()

	sk1, p1 := KeyGen(256)

	for i := 0; i < 5; i++ {

		_, p2 := KeyGen(256)
		t1 := Transaction{p1, p2, 200, strconv.Itoa(i), ""}
		t2 := Transaction{p1, p2, 300, strconv.Itoa(i + 1), ""}
		t1.SignTransaction(sk1)
		t2.SignTransaction(sk1)
		b := createBlock([]Transaction{t1, t2}, i)

		blockChannel <- b
	}

	time.Sleep(200)

}

func TestFinalize(t *testing.T) {
	b, s, f := createChannels()
	go StartTransactionLayer(b, s, f)

	sk1, p1 := KeyGen(256)

	_, p2 := KeyGen(256)
	t1 := Transaction{p1, p2, 200, strconv.Itoa(0), ""}
	t2 := Transaction{p1, p2, 300, strconv.Itoa(0 + 1), ""}
	t1.SignTransaction(sk1)
	t2.SignTransaction(sk1)
	block := createBlock([]Transaction{t1, t2}, 0)
	block.HashBlock()

	b <- block

	// Needs a bit of time for processing the block before finalizing it
	time.Sleep(100)

	f <- block.BlockHash

	for {

		state := <-s
		if state.ledger[p1] != -500 || state.ledger[p2] != 500 {
			t.Error("Something went wrong! Not the right state..")
		}
		return
	}

}

func createChannels() (chan Block, chan State, chan string) {
	blockChannel := make(chan Block)
	stateReturn := make(chan State)
	finalizeChannel := make(chan string)
	return blockChannel, stateReturn, finalizeChannel
}
