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
	t1 = SignTransaction(t1, sk1)
	t2 = SignTransaction(t2, sk1)
	b := createBlock([]Transaction{t1, t2}, 0)

	blockChannel, stateChannel, transChannel, msgChannel, blockReturn := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, transChannel, msgChannel, blockReturn)

	go func() {
		for {
			state := <-stateChannel
			if state.ledger[p2.String()] != 500 {
				t.Error("P2 does not own 500")
			}
			return
		}
	}()

	blockChannel <- b

}

func TestTreeBuild(t *testing.T) {
	blockChannel, stateChannel, transChannel, msgChannel, blockReturn := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, transChannel, msgChannel, blockReturn)

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
		t1 = SignTransaction(t1, sk1)
		t2 = SignTransaction(t2, sk1)
		b := createBlock([]Transaction{t1, t2}, i)

		blockChannel <- b
	}

	time.Sleep(200)

}

func TestBlockCreation(t *testing.T) {
	blockChannel, stateChannel, transChannel, msgChannel, blockReturn := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, transChannel, msgChannel, blockReturn)

	_, p1 := KeyGen(256)

	trans := Transaction{p1, p1, 0, "", ""}

	go func() {
		for {
			b := <-blockReturn
			fmt.Println("A block has been created")
			fmt.Println(b)
		}
	}()

	transChannel <- trans
	transChannel <- trans
	transChannel <- trans

	msgChannel <- "createBlock"
	time.Sleep(time.Second * 3)

}

func createChannels() (chan Block, chan State, chan Transaction, chan string, chan Block) {
	blockChannel := make(chan Block)
	stateReturn := make(chan State)
	transChannel := make(chan Transaction)
	msgChannel := make(chan string)
	blockReturn := make(chan Block)
	return blockChannel, stateReturn, transChannel, msgChannel, blockReturn
}
