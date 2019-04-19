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
	sk, pk := KeyGen(2048)
	genesis = createTestGenesisBlock(time.Duration(1)*time.Second, 0.25)
	StartConsensus(CreateChannelStruct(), pk, sk, true)
}

func createTestBlock(t []Transaction, i int, parentHash string, finalHash string) Block {
	sk, pk := KeyGen(2048)
	block := Block{i,
		parentHash,
		pk,
		"VALID",
		BlockNonce{},
		finalHash,
		Data{Trans: t},
		"",
		"",
	}
	block.SignBlock(sk)
	return block
}

func createTestGenesisBlock(slotDuration time.Duration, hardness float64) Block {
	gData := GenesisData{time.Now(), slotDuration,
		"1001001010100100010101", hardness, State{make(map[PublicKey]int), "", 100000}}
	return Block{0,
		"",
		PublicKey{},
		"VALID",
		BlockNonce{},
		"",
		Data{Trans: []Transaction{}, GenesisData: gData},
		"",
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
		block = createTestBlock(transarr, i, block.CalculateBlockHash(), genesis.CalculateBlockHash())
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
		block = createTestBlock(transarr, i, block.CalculateBlockHash(), genesis.CalculateBlockHash())
	}

	block = createTestBlock([]Transaction{}, 1, genesis.CalculateBlockHash(), genesis.CalculateBlockHash())
	for i := 1; i < 11; i++ {
		time.Sleep(slotLength)
		trans := createTestTransaction(i)
		transarr := []Transaction{trans}
		channels.BlockFromP2P <- block
		block = createTestBlock(transarr, i+10, block.CalculateBlockHash(), genesis.CalculateBlockHash())
	}
}

func TestBadBlock(t *testing.T) {
	resetMocksAndStart()
	b := createTestBlock([]Transaction{}, 5, genesis.CalculateBlockHash(), genesis.CalculateBlockHash())
	c := createTestBlock([]Transaction{}, 3, b.CalculateBlockHash(), genesis.CalculateBlockHash())
	time.Sleep(5 * slotLength)
	channels.BlockFromP2P <- genesis
	channels.BlockFromP2P <- b
	channels.BlockFromP2P <- c
}

func TestInteraction(t *testing.T) {

	resetMocksAndStart()
	genesis.BlockData.GenesisData.InitialState.Ledger[pk] = 1000000 //1 Million

	genHash := genesis.CalculateBlockHash()
	sk2, pk2 := KeyGen(2048)
	sk3, pk3 := KeyGen(2048)

	channels.BlockFromP2P <- genesis

	// Block 1, Grow from Genesis
	t1 := CreateTransaction(pk, pk2, 200, "t1", sk)
	block1 := createTestBlock([]Transaction{t1}, 1, genHash, genHash)
	block1.BakerID = pk
	block1.SignBlock(sk)
	channels.BlockFromP2P <- block1
	time.Sleep(slotLength)

	// Block 2, Grow from Block 1
	t2 := CreateTransaction(pk2, pk3, 100, "t2", sk2)
	block2 := createTestBlock([]Transaction{t2}, 2, block1.CalculateBlockHash(), genHash)
	block2.BakerID = pk2
	block2.SignBlock(sk2)
	channels.BlockFromP2P <- block2
	time.Sleep(slotLength)

	// Block 3, Grow from Block 2
	t3 := CreateTransaction(pk3, pk, 50, "t3", sk3)
	block3 := createTestBlock([]Transaction{t3}, 3, block2.CalculateBlockHash(), genHash)
	block3.BakerID = pk3
	block3.SignBlock(sk3)
	channels.BlockFromP2P <- block3

	// Block 4, Grow from block1
	t4 := CreateTransaction(pk, pk2, 2000, "t4", sk)
	block4 := createTestBlock([]Transaction{t4}, 4, block1.CalculateBlockHash(), genHash)
	block4.BakerID = pk
	block4.SignBlock(sk)
	channels.BlockFromP2P <- block4
	time.Sleep(slotLength)

	// Block 5, Grow from Block 4
	t5 := CreateTransaction(pk2, pk3, 1000, "t5", sk2)
	block5 := createTestBlock([]Transaction{t5}, 5, block4.CalculateBlockHash(), genHash)
	block5.BakerID = pk2
	block5.SignBlock(sk2)
	channels.BlockFromP2P <- block5
	time.Sleep(slotLength)

	// Block 6, Grow from Block 5
	t6 := CreateTransaction(pk3, pk, 500, "t6", sk3)
	block6 := createTestBlock([]Transaction{t6}, 6, block5.CalculateBlockHash(), genHash)
	block6.BakerID = pk3
	block6.SignBlock(sk3)
	channels.BlockFromP2P <- block6

	time.Sleep(slotLength * 18)
}

func TestBlockGeneration(t *testing.T) {
	resetMocksAndStart()
	genesis.BlockData.GenesisData.InitialState.Ledger[pk] = 1000000 //1 Million
	genesis.BlockData.GenesisData.InitialState.TotalStake = 10000000
	channels.BlockFromP2P <- genesis
	_, pk2 := KeyGen(2048)
	for i := 0; i < 10; i++ {
		channels.TransFromP2P <- CreateTransaction(pk, pk2, 50, "t"+strconv.Itoa(i), sk)
	}
	time.Sleep(slotLength * 50)
}
