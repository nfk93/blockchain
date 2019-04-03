package transaction

import (
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
	"testing"
	"time"
)

func createBlock(t []Transaction, i int, pk PublicKey) Block {
	return Block{i + 1,
		strconv.Itoa(i),
		pk,
		"VALID",
		BlockNonce{"42", "", pk},
		"",
		Data{Trans: t},
		"",
	}
}

func TestReceiveBlock(t *testing.T) {
	sk1, p1 := KeyGen(2048)
	_, p2 := KeyGen(2048)
	t1 := CreateTransaction(p1, p2, 200, "ID112", sk1)
	t2 := CreateTransaction(p1, p2, 300, "ID222", sk1)
	b := createBlock([]Transaction{t1, t2}, 0, p1)
	b.SignBlock(sk1)

	blockChannel, stateChannel, finalChannel, br, tl := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, finalChannel, br, tl, sk1, State{})

	blockChannel <- b

	time.Sleep(3000)
	finalChannel <- b.CalculateBlockHash()

	go func() {
		for {
			state := <-stateChannel
			if state.Ledger[p2] != 500 {
				t.Error("P2 does not own 500")
			}
			return
		}
	}()

	blockChannel <- b

}

func TestTreeBuild(t *testing.T) {
	sk1, p1 := KeyGen(2048)
	blockChannel, stateChannel, finalChannel, br, tl := createChannels()
	go StartTransactionLayer(blockChannel, stateChannel, finalChannel, br, tl, sk1, State{})

	go func() {
		for {
			state := <-stateChannel
			fmt.Println("state ", state)
		}

	}()

	for i := 0; i < 5; i++ {

		_, p2 := KeyGen(2048)
		t1 := CreateTransaction(p1, p2, 200, strconv.Itoa(i), sk1)
		t2 := CreateTransaction(p1, p2, 300, strconv.Itoa(i+1), sk1)
		b := createBlock([]Transaction{t1, t2}, i, p1)
		b.SignBlock(sk1)

		blockChannel <- b
	}

	time.Sleep(200)

}

func TestFinalize(t *testing.T) {
	b, s, f, br, tl := createChannels()

	sk1, p1 := KeyGen(2048)
	go StartTransactionLayer(b, s, f, br, tl, sk1, State{})

	_, p2 := KeyGen(2048)
	t1 := CreateTransaction(p1, p2, 200, strconv.Itoa(0), sk1)
	t2 := CreateTransaction(p1, p2, 300, strconv.Itoa(0+1), sk1)
	block := createBlock([]Transaction{t1, t2}, 0, p1)
	block.SignBlock(sk1)

	b <- block

	// Needs a bit of time for processing the block before finalizing it
	time.Sleep(100)

	f <- block.CalculateBlockHash()

	for {

		state := <-s
		if state.Ledger[p1] != -500 || state.Ledger[p2] != 500 {
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
	go StartTransactionLayer(b, s, f, br, tl, sk1, State{})

	go func() {
		for {
			// we finalize block 4, p1 = -200, p2=0, p4=200
			state := <-s
			if state.Ledger[p1] != -200 || state.Ledger[p2] != 0 || state.Ledger[p4] != 200 {
				t.Error("Bad luck! Branching did not succeed...")
			}
			return
		}
	}()

	// Block 1, Grow from Genesis
	t1 := CreateTransaction(p1, p1, 200, strconv.Itoa(1), sk1)
	block1 := createBlock([]Transaction{t1}, 0, p1)
	block1.SignBlock(sk1)
	b <- block1
	time.Sleep(100)

	// Block 2 - grow from block 1
	t2 := CreateTransaction(p1, p2, 200, strconv.Itoa(2), sk1)
	block2 := createBlock([]Transaction{t2}, 1, p1)
	block2.ParentPointer = block1.CalculateBlockHash()
	block1.SignBlock(sk1)
	b <- block2
	time.Sleep(100)

	// Block 3 - grow from block 1
	t3 := CreateTransaction(p1, p3, 200, strconv.Itoa(3), sk1)
	block3 := createBlock([]Transaction{t3}, 2, p1)
	block3.ParentPointer = block1.CalculateBlockHash()
	block1.SignBlock(sk1)
	b <- block3
	time.Sleep(100)

	// Block 4 - grow from block 2
	t4 := CreateTransaction(p2, p4, 200, strconv.Itoa(4), sk2)
	block4 := createBlock([]Transaction{t4}, 3, p2)
	block4.ParentPointer = block2.CalculateBlockHash()
	block1.SignBlock(sk2)
	b <- block4

	//Finalizing to get states from TL
	// Needs a bit of time for processing the block before finalizing it
	time.Sleep(1000)
	f <- block4.CalculateBlockHash()
	time.Sleep(1000)

}

func TestCreateNewBlock(t *testing.T) {

	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)

	b, s, f, br, tl := createChannels()
	go StartTransactionLayer(b, s, f, br, tl, sk1, State{})

	go func() {
		for {
			newBlock := <-br
			for i, tran := range newBlock.BlockData.Trans {

				fmt.Println(i, tran)
			}
		}
	}()

	genBlock := CreateTestGenesis()
	b <- genBlock
	time.Sleep(300)

	var transList []Transaction
	for i := 0; i < 20; i++ {
		t1 := CreateTransaction(pk1, pk2, i*100, "ID"+strconv.Itoa(i), sk1)
		transList = append(transList, t1)
	}

	newBlockData := CreateBlockData{transList, sk1, pk1, 2, ""}

	tl <- newBlockData

	time.Sleep(500)
}

func TestRuns(t *testing.T) {
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)

	b, s, f, br, tl := createChannels()
	go StartTransactionLayer(b, s, f, br, tl, sk1)

	go func() {
		for {
			fmt.Println(<-s)
		}
	}()

	genBlock := CreateTestGenesis()
	b <- genBlock
	time.Sleep(300)

	var transList []Transaction
	for i := 0; i < 2; i++ {
		t1 := CreateTransaction(pk1, pk2, 100+(i*100), "ID"+strconv.Itoa(i), sk1)
		transList = append(transList, t1)
	}

	for i := 0; i < 40; i++ {
		newBlockData := CreateBlockData{transList, sk1, pk1, i + 1, ""}
		tl <- newBlockData
		newBlock := <-br
		b <- newBlock
		time.Sleep(1000)
		if i != 0 && i%10 == 9 {
			time.Sleep(1000)
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
