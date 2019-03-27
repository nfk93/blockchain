package p2p

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/objects"
	"testing"
	"time"
)

// Method for failing a test and reporting a meaningful error
func fail(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Error("FAILURE: expected: ", expected, ", actual: ", actual)
	}
}

var _, mockPK = crypto.KeyGen(128)

var mockBlock_1 = objects.Block{
	42,
	"",
	mockPK,
	"VALID",
	objects.BlockNonce{"42", "", mockPK},
	"",
	objects.Data{},
	"123",
}

var mockBlock_2 = objects.Block{
	42,
	"",
	mockPK,
	"VALID",
	objects.BlockNonce{"44", "", mockPK},
	"",
	objects.Data{},
	"555",
}

var mockTrans_1 = objects.Transaction{
	mockPK,
	mockPK,
	103,
	"id1",
	"sign1"}

var mockTrans_2 = objects.Transaction{
	mockPK,
	mockPK,
	11,
	"id2",
	"sign2"}

/*func TestAll(t *testing.T) {
	blockIn := make(chan objects.Block)
	blockOut := make(chan objects.Block)
	transIn := make(chan objects.Transaction)
	transOut := make(chan objects.Transaction)
	StartP2P("", "65000", blockIn, blockOut, transIn, transOut)
}*/

func TestRPC(t *testing.T) {
	rpcObj := new(RPCHandler)
	go listenForRPC(myHostPort)
	time.Sleep(1 * time.Second)

	resetMockVars()
	fmt.Println("Running BlocksReceivedOnce_1")
	t.Run("BlocksReceivedOnce_1", func(t *testing.T) {
		rpcObj.SendBlock(mockBlock_1, &struct{}{})
		block := <-deliverBlock
		if block.CalculateBlockHash() != mockBlock_1.CalculateBlockHash() {
			t.Errorf("First block seen isn't testblock_1")
		}
	})
	fmt.Println("Running BlocksReceivedOnce_2")
	t.Run("BlocksReceivedOnce_2", func(t *testing.T) {
		go rpcObj.SendBlock(mockBlock_2, &struct{}{})
		go rpcObj.SendBlock(mockBlock_1, &struct{}{})
		go rpcObj.SendBlock(mockBlock_2, &struct{}{})
		go rpcObj.SendBlock(mockBlock_1, &struct{}{})
		go rpcObj.SendBlock(mockBlock_2, &struct{}{})
		go rpcObj.SendBlock(mockBlock_1, &struct{}{})
		go rpcObj.SendBlock(mockBlock_2, &struct{}{})
		go rpcObj.SendBlock(mockBlock_1, &struct{}{})
		go rpcObj.SendBlock(mockBlock_2, &struct{}{})
		block := <-deliverBlock
		if block.CalculateBlockHash() != mockBlock_2.CalculateBlockHash() {
			t.Error("Second block seen isn't testblock_2")
		}
	})
	fmt.Println("Running BlocksReceivedOnce_3")
	t.Run("BlocksReceivedOnce_3", func(t *testing.T) {
		select {
		case _ = <-deliverBlock:
			t.Errorf("Too many blocks delivered!")
		default:
			// do nothing
		}
	})
	resetMockVars()
	fmt.Println("Running SendBlockUpdatesBlocksSeen")
	t.Run("SendBlockUpdatesBlocksSeen", func(t *testing.T) {
		rpcObj := new(RPCHandler)
		rpcObj.SendBlock(mockBlock_1, &struct{}{})
		_ = <-deliverBlock
		if !blocksSeen.contains(mockBlock_1.CalculateBlockHash()) {
			t.Fail()
		}
		if blocksSeen.contains(mockBlock_2.CalculateBlockHash()) {
			t.Fail()
		}
		rpcObj.SendBlock(mockBlock_2, &struct{}{})
		_ = <-deliverBlock
		if !blocksSeen.contains(mockBlock_2.CalculateBlockHash()) {
			t.Fail()
		}
	})

	// Testing Transaction
	resetMockVars()
	fmt.Println("Running TransactionsReceivedOnce_1")
	t.Run("TransactionsReceivedOnce_1", func(t *testing.T) {
		rpcObj.SendTransaction(mockTrans_1, &struct{}{})
		trans := <-deliverTrans
		if trans != mockTrans_1 {
			t.Errorf("First transaction seen isn't mockTrans_1")
		}
	})
	fmt.Println("Printing TransactionsReceivedOnce_2")
	t.Run("TransactionsReceivedOnce_2", func(t *testing.T) {
		go rpcObj.SendTransaction(mockTrans_2, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_1, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_2, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_1, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_2, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_1, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_2, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_1, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_2, &struct{}{})
		go rpcObj.SendTransaction(mockTrans_1, &struct{}{})
		trans := <-deliverTrans
		if trans != mockTrans_2 {
			t.Error("Second transaction seen isn't mockTrans_2")
		}
	})
	fmt.Println("Running TransactionReceivedOnce_3")
	t.Run("TransactionReceivedOnce_3", func(t *testing.T) {
		select {
		case _ = <-deliverTrans:
			t.Errorf("Too many transactions delivered!")
		default:
			// do nothing
		}
	})
}

func resetMockVars() {
	deliverBlock = make(chan objects.Block)
	deliverTrans = make(chan objects.Transaction)

	networkList = make(map[string]bool)
	blocksSeen = *newStringSet()
	transSeen = *newStringSet()
	myIp = "127.0.0.1"
	myHostPort = "65000"
	networkList["127.0.0.1:65000"] = true
}
