package consensus

import (
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"strconv"
	"testing"
	"time"
)

var genesis Block

func resetMocksAndStart() {
	genesis = createTestGenesisBlock(time.Duration(10), 0)
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

func createTestGenesisBlock(slotDuration time.Duration, hardness float64) Block {
	gData := GenesisData{time.Now(), slotDuration,
		BlockNonce{}, hardness, State{}}
	return Block{0,
		"",
		PublicKey{},
		"VALID",
		BlockNonce{"42", "", pk},
		"",
		Data{Trans: []Transaction{}, GenesisData: gData},
		"",
	}
}

func createTestTransaction(ID int) Transaction {
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	trans := Transaction{pk1, pk2, 200, strconv.Itoa(ID), ""}
	trans.SignTransaction(sk1)
	return trans
}

func TestSmokeTest(t *testing.T) {
	resetMocksAndStart()
}

func TestTree(t *testing.T) {
	resetMocksAndStart()
	block := genesis
	for i := 1; i < 10; i++ {
		time.Sleep(slotLength)
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		channels.BlockFromP2P <- block
		channels.TransFromP2P <- trans
		block = createTestBlock(transarr, i, block.CalculateBlockHash())
	}
}

func TestRollBack(t *testing.T) {
	resetMocksAndStart()
	block := genesis
	for i := 1; i < 10; i++ {
		time.Sleep(slotLength)
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		channels.BlockFromP2P <- block
		channels.TransFromP2P <- trans
		block = createTestBlock(transarr, i, block.CalculateBlockHash())
	}

	block = createTestBlock([]Transaction{}, 1, genesis.CalculateBlockHash())
	for i := 1; i < 11; i++ {
		time.Sleep(slotLength)
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		channels.BlockFromP2P <- block
		block = createTestBlock(transarr, i+10, block.CalculateBlockHash())
	}
}

func TestBadBlock(t *testing.T) {
	resetMocksAndStart()
	b := createTestBlock([]Transaction{}, 5, genesis.CalculateBlockHash())
	c := createTestBlock([]Transaction{}, 3, b.CalculateBlockHash())
	time.Sleep(5 * slotLength)
	channels.BlockFromP2P <- genesis
	channels.BlockFromP2P <- b
	channels.BlockFromP2P <- c
}
