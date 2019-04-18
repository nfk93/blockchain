package transaction

import (
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
	"testing"
	"time"
)

func TestReceiveAndFinalizeBlock(t *testing.T) {
	sk1, p1 := KeyGen(2048)
	_, p2 := KeyGen(2048)

	blockChannel, stateChannel, finalChannel, br, tl := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, finalChannel, br, tl, sk1)

	GenBlock := CreateTestGenesis(p1)
	blockChannel <- GenBlock

	t1 := CreateTransaction(p1, p2, 200, "ID112", sk1)
	t2 := CreateTransaction(p1, p2, 300, "ID222", sk1)
	tl <- CreateBlockData{[]Transaction{t1, t2}, sk1, p1, 1, "", BlockNonce{}}
	b := <-br

	blockChannel <- b

	time.Sleep(time.Second * 3)
	finalChannel <- b.CalculateBlockHash()

	for {
		state := <-stateChannel
		if state.Ledger[p1] != 999500 || state.Ledger[p2] != 500 {
			t.Error("Something went wrong! Not the right state..")
		}
		return
	}
}

func TestForking(t *testing.T) {

	sk1, p1 := KeyGen(2048)
	sk2, p2 := KeyGen(2048)
	_, p3 := KeyGen(2048)
	_, p4 := KeyGen(2048)

	b, s, f, br, tl := createChannels()
	go StartTransactionLayer(b, s, f, br, tl, sk1)

	go func() {
		for {
			// we finalize block 4, p1 = 999400, p2=150, p4=450
			state := <-s
			if state.Ledger[p1] != 999400 || state.Ledger[p2] != 150 || state.Ledger[p4] != 450 {
				t.Error("Bad luck! Branching did not succeed")
			}
			return
		}
	}()

	GenBlock := CreateTestGenesis(p1)
	b <- GenBlock

	// Block 1, Grow from Genesis
	t1 := CreateTransaction(p1, p4, 400, strconv.Itoa(1), sk1)
	tl <- CreateBlockData{[]Transaction{t1}, sk1, p1, 1, "", BlockNonce{}}
	block1 := <-br
	b <- block1
	time.Sleep(time.Millisecond * 300)

	// Block 2 - grow from block 1
	t2 := CreateTransaction(p1, p2, 200, strconv.Itoa(2), sk1)
	tl <- CreateBlockData{[]Transaction{t2}, sk1, p1, 2, "", BlockNonce{}}
	block2 := <-br
	b <- block2
	time.Sleep(time.Millisecond * 100)

	// Block 3 - grow from block 1
	t3 := CreateTransaction(p1, p3, 300, strconv.Itoa(3), sk1)
	tl <- CreateBlockData{[]Transaction{t3}, sk1, p1, 3, "", BlockNonce{}}
	block3 := <-br

	// Block 4 - grow from block 2
	t4 := CreateTransaction(p2, p4, 50, strconv.Itoa(4), sk2)
	tl <- CreateBlockData{[]Transaction{t4}, sk2, p2, 4, "", BlockNonce{}}
	block4 := <-br

	b <- block3
	b <- block4

	//Finalizing to get states from TL
	// Needs a bit of time for processing the block before finalizing it
	time.Sleep(time.Millisecond * 1000)
	f <- block4.CalculateBlockHash()
	time.Sleep(time.Millisecond * 1000)

}

func TestCreateNewBlock(t *testing.T) {

	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	b, s, f, br, tl := createChannels()
	go StartTransactionLayer(b, s, f, br, tl, sk1)

	genBlock := CreateTestGenesis(pk1)
	b <- genBlock
	time.Sleep(time.Millisecond * 300)

	var transList []Transaction
	for i := 0; i < 20; i++ {
		t1 := CreateTransaction(pk1, pk2, 100+(i*100), "ID"+strconv.Itoa(i), sk1)
		transList = append(transList, t1)
	}
	newBlockData := CreateBlockData{transList, sk1, pk1, 2, "", BlockNonce{}}

	tl <- newBlockData
	newBlock := <-br
	if !newBlock.ValidateBlock() {
		t.Error("Block validation failed")
	}

}

func TestRuns(t *testing.T) { //Does not really test anything, but runs a lot of blocks that you can debug on the transactionLayer
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)

	b, s, f, br, tl := createChannels()
	go StartTransactionLayer(b, s, f, br, tl, sk1)

	go func() {
		for {
			fmt.Println(<-s)
		}
	}()

	genBlock := CreateTestGenesis(pk1)
	b <- genBlock
	time.Sleep(time.Millisecond * 300)

	var transList []Transaction
	for i := 0; i < 2; i++ {
		t1 := CreateTransaction(pk1, pk2, 100+(i*100), "ID"+strconv.Itoa(i), sk1)
		transList = append(transList, t1)
	}

	for i := 0; i < 40; i++ {
		newBlockData := CreateBlockData{transList, sk1, pk1, i + 1, "", BlockNonce{}}
		tl <- newBlockData
		newBlock := <-br
		b <- newBlock
		time.Sleep(time.Millisecond * 1000)
		if i != 0 && i%10 == 9 {
			time.Sleep(time.Millisecond * 1000)
			f <- newBlock.CalculateBlockHash()
		}

	}

}

// Helpers
func createChannels() (chan Block, chan State, chan string, chan Block, chan CreateBlockData) {
	blockChannel := make(chan Block)
	stateReturn := make(chan State)
	finalizeChannel := make(chan string)
	blockReturn := make(chan Block)
	transList := make(chan CreateBlockData)
	return blockChannel, stateReturn, finalizeChannel, blockReturn, transList
}

//func createBlock(t []Transaction, i int, pk PublicKey) Block {
//	return Block{i + 1,
//		strconv.Itoa(i),
//		pk,
//		"VALID",
//		BlockNonce{},
//		"",
//		Data{Trans: t},
//		"",
//		"",
//	}
//}
