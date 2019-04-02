package consensus

import (
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
	"testing"
)

var transFromP2P chan Transaction
var blockFromP2P chan Block
var blockToP2P chan Block
var genesis GenesisData

func resetMocksAndStart() {
	transFromP2P = make(chan Transaction)
	blockFromP2P = make(chan Block)
	blockToP2P = make(chan Block)
	genesis = GenesisData{}
	StartConsensus(CreateChannelStruct())
}

func createTestBlock(t []Transaction, i int, parentHash string) Block {
	sk, pk := KeyGen(2000)
	block := Block{i,
		parentHash,
		pk,
		"VALID",
		BlockNonce{"42", "", pk},
		"",
		Data{Trans: t},
		"",
	}
	block.SignBlock(sk)
	return block
}

func createTestTransaction(ID int) Transaction {
	sk1, pk1 := KeyGen(2000)
	_, pk2 := KeyGen(2000)
	trans := Transaction{pk1, pk2, 200, strconv.Itoa(ID), ""}
	trans.SignTransaction(sk1)
	return trans
}

func TestSmokeTest(t *testing.T) {
	resetMocksAndStart()
}

func TestTree(t *testing.T) {
	resetMocksAndStart()
	block := createTestBlock([]Transaction{}, 0, "")
	for i := 1; i < 10; i++ {
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		blockFromP2P <- block
		transFromP2P <- trans
		block = createTestBlock(transarr, i, block.CalculateBlockHash())
	}
}

func TestRollBack(t *testing.T) {
	resetMocksAndStart()
	genesis := createTestBlock([]Transaction{}, 0, "")
	block := genesis
	for i := 1; i < 10; i++ {
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		blockFromP2P <- block
		transFromP2P <- trans
		block = createTestBlock(transarr, i, block.CalculateBlockHash())
	}

	block = createTestBlock([]Transaction{}, 1, genesis.CalculateBlockHash())
	for i := 1; i < 11; i++ {
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		blockFromP2P <- block
		block = createTestBlock(transarr, i+10, block.CalculateBlockHash())
	}
}

func TestBadBlock(t *testing.T) {
	resetMocksAndStart()
	genesis := createTestBlock([]Transaction{}, 0, "")
	b := createTestBlock([]Transaction{}, 5, genesis.CalculateBlockHash())
	c := createTestBlock([]Transaction{}, 3, b.CalculateBlockHash())
	blockFromP2P <- genesis
	blockFromP2P <- b
	blockFromP2P <- c
}
