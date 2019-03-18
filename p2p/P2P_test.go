package p2p

import (
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

var testBlock_1 = objects.Block{
	42,
	"",
	42,
	"VALID",
	42,
	"",
	objects.Data{},
	"123",
	""}

var testBlock_2 = objects.Block{
	42,
	"",
	42,
	"VALID",
	42,
	"",
	objects.Data{},
	"555",
	""}

/*func TestAll(t *testing.T) {
	blockIn := make(chan objects.Block)
	blockOut := make(chan objects.Block)
	transIn := make(chan objects.Transaction)
	transOut := make(chan objects.Transaction)
	StartP2P("", "65000", blockIn, blockOut, transIn, transOut)
}*/

func TestRPC(t *testing.T) {
	instantiateMockVars()
	rpcObj := new(RPCHandler)
	listenForRPC(myHostPort)
	time.Sleep(1 * time.Second)

	t.Run("BlocksReceivedOnce_1", func(t *testing.T) {
		rpcObj.SendBlock(testBlock_1, &struct{}{})
		block := <-deliverBlock
		if block.BlockHash != testBlock_1.BlockHash {
			t.Errorf("First block seen isn't testblock_1")
		}
	})
	t.Run("BlocksReceivedOnce_2", func(t *testing.T) {
		go rpcObj.SendBlock(testBlock_2, &struct{}{})
		go rpcObj.SendBlock(testBlock_1, &struct{}{})
		go rpcObj.SendBlock(testBlock_2, &struct{}{})
		go rpcObj.SendBlock(testBlock_1, &struct{}{})
		go rpcObj.SendBlock(testBlock_2, &struct{}{})
		go rpcObj.SendBlock(testBlock_1, &struct{}{})
		go rpcObj.SendBlock(testBlock_2, &struct{}{})
		go rpcObj.SendBlock(testBlock_1, &struct{}{})
		go rpcObj.SendBlock(testBlock_2, &struct{}{})
		block := <-deliverBlock
		if block.BlockHash != testBlock_2.BlockHash {
			t.Error("Second block seen isn't testblock_2")
		}
	})
	t.Run("BlocksReceivedOnce_3", func(t *testing.T) {
		select {
		case _ = <-deliverBlock:
			t.Errorf("Too many blocks delivered!")
		default:
			// do nothing
		}
	})
	instantiateMockVars()
	t.Run("SendBlockUpdatesBlocksSeen", func(t *testing.T) {
		rpcObj := new(RPCHandler)
		rpcObj.SendBlock(testBlock_1, &struct{}{})
		_ = <-deliverBlock
		if !blocksSeen.contains("123") {
			t.Fail()
		}
		if blocksSeen.contains("555") {
			t.Fail()
		}
		rpcObj.SendBlock(testBlock_2, &struct{}{})
		_ = <-deliverBlock
		if !blocksSeen.contains("555") {
			t.Fail()
		}
	})
}

func instantiateMockVars() {
	deliverBlock = make(chan objects.Block)
	networkList = make(map[string]bool)
	blocksSeen = *newStringSet()
	myIp = "127.0.0.1"
	myHostPort = "65000"
	networkList["127.0.0.1:65000"] = true
}
