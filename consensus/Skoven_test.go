package consensus

import (
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/objects/genesisdata"
	"strconv"
	"testing"
	"time"
)

var transFromP2P chan Transaction
var blockFromP2P chan Block
var blockToP2P chan Block
var genesis genesisdata.GenesisData

func resetMocksAndStart() {
	transFromP2P = make(chan Transaction)
	blockFromP2P = make(chan Block)
	blockToP2P = make(chan Block)
	genesis = genesisdata.GenesisData{}
	StartConsensus(genesis, transFromP2P, blockFromP2P, blockToP2P)
}

func createTestBlock(t []Transaction, i int, parentHash string) Block {
	sk, pk := KeyGen(2000)
	block := Block{i,
		parentHash,
		pk,
		"VALID",
		42,
		"",
		Data{t},
		"",
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
		block = createTestBlock(transarr, i, block.HashBlock())
	}
	time.Sleep(10000)
}
