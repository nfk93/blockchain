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

	channels := CreateChannelStruct()
	go StartTransactionLayer(channels)

	GenBlock := CreateTestGenesis(p1)
	channels.BlockToTrans <- GenBlock

	t1 := CreateTransaction(p1, p2, 200, "ID112", sk1)
	t2 := CreateTransaction(p1, p2, 300, "ID222", sk1)
	channels.TransToTrans <- CreateBlockData{[]TransData{{Transaction: t1}, {Transaction: t2}}, sk1, p1, 1, "", BlockNonce{}, ""}
	b := <-channels.BlockFromTrans

	channels.BlockToTrans <- b

	time.Sleep(time.Second * 3)
	channels.FinalizeToTrans <- b.CalculateBlockHash()

	for {
		state := <-channels.StateFromTrans
		//we get 100 in block reward -> 999500+100 =999600
		if state.Ledger[p1.String()] != 999600 || state.Ledger[p2.String()] != 500 {
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

	channels := CreateChannelStruct()
	go StartTransactionLayer(channels)

	go func() {
		for {
			// we finalize block 4, p1 = 999400, p2=150, p4=450
			state := <-channels.StateFromTrans
			for _, account := range state.Ledger {
				fmt.Println(account)
			}
			fmt.Println("Total system stake: ", state.TotalStake)
			//if state.Ledger[p1] != 999400 || state.Ledger[p2] != 150 || state.Ledger[p4] != 450 {
			//	t.Error("Bad luck! Branching did not succeed")
			//}
			return
		}
	}()

	GenBlock := CreateTestGenesis(p1)
	channels.BlockToTrans <- GenBlock

	// Block 1, Grow from Genesis
	t1 := CreateTransaction(p1, p4, 400, strconv.Itoa(1), sk1)
	channels.TransToTrans <- CreateBlockData{[]TransData{{Transaction: t1}}, sk1, p1, 1, "", BlockNonce{}, ""}
	block1 := <-channels.BlockFromTrans
	channels.BlockToTrans <- block1
	time.Sleep(time.Millisecond * 300)

	// Block 2 - grow from block 1
	t2 := CreateTransaction(p1, p2, 200, strconv.Itoa(2), sk1)
	channels.TransToTrans <- CreateBlockData{[]TransData{{Transaction: t2}}, sk2, p2, 2, "", BlockNonce{}, ""}
	block2 := <-channels.BlockFromTrans
	channels.BlockToTrans <- block2
	time.Sleep(time.Millisecond * 100)

	// Block 3 - grow from block 1
	t3 := CreateTransaction(p1, p3, 300, strconv.Itoa(3), sk1)
	channels.TransToTrans <- CreateBlockData{[]TransData{{Transaction: t3}}, sk2, p2, 3, "", BlockNonce{}, ""}
	block3 := <-channels.BlockFromTrans

	// Block 4 - grow from block 2
	t4 := CreateTransaction(p2, p4, 50, strconv.Itoa(4), sk2)
	channels.TransToTrans <- CreateBlockData{[]TransData{{Transaction: t4}}, sk2, p2, 4, "", BlockNonce{}, ""}
	block4 := <-channels.BlockFromTrans

	channels.BlockToTrans <- block3
	channels.BlockToTrans <- block4

	//Finalizing to get states from TL
	// Needs a bit of time for processing the block before finalizing it
	time.Sleep(time.Millisecond * 1000)
	channels.FinalizeToTrans <- block4.CalculateBlockHash()
	time.Sleep(time.Millisecond * 1000)

}

func TestCreateNewBlock(t *testing.T) {

	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	channels := CreateChannelStruct()
	go StartTransactionLayer(channels)

	genBlock := CreateTestGenesis(pk1)
	channels.BlockToTrans <- genBlock
	time.Sleep(time.Millisecond * 300)

	var transList []TransData
	for i := 0; i < 20; i++ {
		t1 := TransData{Transaction: CreateTransaction(pk1, pk2, uint64(100+(i*100)), "ID"+strconv.Itoa(i), sk1)}
		transList = append(transList, t1)
	}
	newBlockData := CreateBlockData{transList, sk1, pk1, 2, "", BlockNonce{}, ""}

	channels.TransToTrans <- newBlockData
	newBlock := <-channels.BlockFromTrans
	if !newBlock.ValidateBlock() {
		t.Error("Block validation failed")
	}

}

//func TestRuns(t *testing.T) { //Does not really test anything, but runs a lot of blocks that you can debug on the transactionLayer
//	sk1, pk1 := KeyGen(2048)
//	_, pk2 := KeyGen(2048)
//
//	channels := CreateChannelStruct()
//	go StartTransactionLayer(channels)
//
//	go func() {
//		for {
//			fmt.Println(<-channels.StateFromTrans)
//		}
//	}()
//
//	genBlock := CreateTestGenesis(pk1)
//	channels.BlockToTrans <- genBlock
//	time.Sleep(time.Millisecond * 300)
//
//	var transList []TransData
//	for i := 0; i < 2; i++ {
//		t1 := CreateTransaction(pk1, pk2, 100+(i*100), "ID"+strconv.Itoa(i), sk1)
//		transList = append(transList, t1)
//	}
//
//	for i := 0; i < 40; i++ {
//		newBlockData := CreateBlockData{transList, sk1, pk1, i + 1, "", BlockNonce{},""}
//		channels.TransToTrans <- newBlockData
//		newBlock := <-channels.BlockFromTrans
//		channels.BlockToTrans <- newBlock
//		time.Sleep(time.Millisecond * 1000)
//		if i != 0 && i%10 == 9 {
//			time.Sleep(time.Millisecond * 1000)
//			channels.FinalizeToTrans <- newBlock.CalculateBlockHash()
//		}
//
//	}
//
//}
