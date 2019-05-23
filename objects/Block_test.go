package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestVerifyBlock(t *testing.T) {
	var sk, pk = KeyGen(2048)

	block := Block{0,
		"",
		pk,
		"",
		BlockNonce{},
		"",
		BlockData{},
		"",
		""}

	block.SignBlock(sk)

	if !block.ValidateBlock() {
		t.Error("Block Validation Failed")
	}

}

func TestVerifyBlockFAIL(t *testing.T) {
	var _, pk = KeyGen(2048)
	var sk2, _ = KeyGen(2048)

	block := Block{0,
		"",
		pk,
		"",
		BlockNonce{},
		"",
		BlockData{},
		"",
		""}

	block.SignBlock(sk2)

	if block.ValidateBlock() {
		t.Error("Should have failed on BlockSignature Validation")
	}
}

func TestBlockNonce(t *testing.T) {
	sk, pk := KeyGen(2048)
	leadershipNonce := "011101101"
	blockNonce := CreateNewBlockNonce(leadershipNonce, sk, 1)
	block := Block{1,
		"",
		pk,
		"",
		blockNonce,
		"",
		BlockData{},
		"",
		""}
	if !block.validateBlockNonce(leadershipNonce) {
		t.Error("Nonce validation failed")
	}
}

func TestGetType(t *testing.T) {
	tdTrans := TransData{Transaction: Transaction{Amount: 500}}
	if tdTrans.GetType() != TRANSACTION {
		t.Error("GetType didn't recognize type Transaction")
	}
	tdCon := TransData{ContractCall: ContractCall{Amount: 400}}
	if tdCon.GetType() != CONTRACTCALL {
		t.Error("GetType didn't recognize type ContractCall")
	}
	tdConInit := TransData{ContractInit: ContractInitialize{Prepaid: 500}}
	if tdConInit.GetType() != CONTRACTINIT {
		t.Error("GetType didn't recognize type ContractInitializer")
	}

}

//func TestCreateAndVerifyNonce(t *testing.T) {
//	var sk, pk = KeyGen(2048)
//
//	nonce := BlockNonce{"8556", "Something", pk}
//
//	blockNonce := CreateNewBlockNonce(nonce, 2, sk, pk)
//
//	if !blockNonce.validateBlockNonce() {
//		t.Error("Block Failed")
//	}
//
//}
//
//func TestCreateAndVerifyNonceFAIL(t *testing.T) {
//	var sk, _ = KeyGen(2048)
//	var _, pk2 = KeyGen(2048)
//
//	nonce := BlockNonce{"8556", "Something", pk2}
//
//	blockNonce := CreateNewBlockNonce(nonce, 2, sk, pk2)
//
//	if blockNonce.validateBlockNonce() {
//		t.Error("Block Shouldn't verify!!")
//	}
//
//}

//func TestVerifyBlock(t *testing.T) {
//	var sk, pk = KeyGen(2048)
//
//	slot := 4
//	hardness := 0.9
//	yourStake := 20
//	systemStake := 300
//
//	// setup of non signed nonce and block to create proper block and nonce from
//	preNonce := BlockNonce{"8556", "Something", pk}
//	prevBlock := Block{0,
//		"",
//		pk,
//		"",
//		preNonce,
//		"",
//		BlockData{},
//		""}
//
//	//nonce := CreateNewBlockNonce(preNonce, slot, sk, pk)
//
//	createSuccess, draw := CalculateDraw(nonce, hardness, sk, pk, yourStake, systemStake, slot)
//	if !createSuccess {
//		t.Error("Draw not above Hardness")
//	}
//	blockdata := CreateBlockData{[]Transaction{}, sk, pk, slot, draw}
//	parent := prevBlock.CalculateBlockHash()
//
//	t1 := CreateTransaction(pk, pk, 200, strconv.Itoa(1), sk)
//	t2 := CreateTransaction(pk, pk, 200, strconv.Itoa(1), sk)
//	t3 := CreateTransaction(pk, pk, 200, strconv.Itoa(1), sk)
//
//	b := CreateNewBlock(blockdata, parent, nonce, []Transaction{t1, t2, t3})
//
//	validationSuccess, errMsg := b.ValidateBlock()
//
//	if !validationSuccess {
//		t.Error(errMsg)
//	}
//
//}
