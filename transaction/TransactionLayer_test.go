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

	blockChannel := make(chan Block)

	stateReturn := make(chan State)

	go StartTransactionLayer(blockChannel, stateReturn)

	blockChannel <- b

	for {
		state := <-stateReturn
		if state.ledger[p2.String()] != 500 {
			t.Error("P2 does not own 500")
		}
		return
	}

}

func TestTree(t *testing.T) {
	blockChannel := make(chan Block)
	stateReturn := make(chan State)
	go StartTransactionLayer(blockChannel, stateReturn)

	go func() {
		for {
			state := <-stateReturn

			fmt.Println("state ", state)
		}

	}()

	sk1, p1 := KeyGen(256)

	//for i := 0; i < 5; i++ {
	//
	//	_, p2 := KeyGen(256)
	//	t1 := Transaction{p1, p2, 200, strconv.Itoa(i), ""}
	//	t2 := Transaction{p1, p2, 300, strconv.Itoa(i+1), ""}
	//	t1 = SignTransaction(t1, sk1)
	//	t2 = SignTransaction(t2, sk1)
	//	b := createBlock([]Transaction{t1, t2}, i)
	//
	//	blockChannel <- b
	//}

	_, p2 := KeyGen(256)
	t1 := Transaction{p1, p2, 200, strconv.Itoa(0), ""}
	t2 := Transaction{p1, p2, 300, strconv.Itoa(1), ""}
	t1 = SignTransaction(t1, sk1)
	t2 = SignTransaction(t2, sk1)
	b := createBlock([]Transaction{t1, t2}, 0)

	blockChannel <- b

	_, p3 := KeyGen(256)
	t3 := Transaction{p1, p3, 200, strconv.Itoa(2), ""}
	t4 := Transaction{p1, p3, 300, strconv.Itoa(3), ""}
	t1 = SignTransaction(t3, sk1)
	t2 = SignTransaction(t4, sk1)
	b = createBlock([]Transaction{t3, t4}, 1)

	blockChannel <- b

	_, p4 := KeyGen(256)
	t5 := Transaction{p1, p4, 200, strconv.Itoa(4), ""}
	t6 := Transaction{p1, p4, 300, strconv.Itoa(5), ""}
	t1 = SignTransaction(t5, sk1)
	t2 = SignTransaction(t6, sk1)
	b = createBlock([]Transaction{t5, t6}, 2)

	blockChannel <- b

	time.Sleep(200)

}
