package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestVerifyBlock(t *testing.T) {
	var sk, pk = KeyGen(2048)

	nonce := BlockNonce{"8556", "", pk}
	nonce.SignBlockNonce(sk)

	block := Block{0,
		"",
		pk,
		"",
		nonce,
		"",
		Data{},
		""}

	block.SignBlock(sk)

	validationSuccess, errMsg := block.ValidateBlock()
	if !validationSuccess {
		t.Error(errMsg)
	}

}

func TestVerifyBlockFAILNonce(t *testing.T) {
	var sk, pk = KeyGen(2048)
	var sk2, _ = KeyGen(2048)

	nonce := BlockNonce{"8556", "", pk}
	nonce.SignBlockNonce(sk2)

	block := Block{0,
		"",
		pk,
		"",
		nonce,
		"",
		Data{},
		""}

	block.SignBlock(sk)

	validationSuccess, _ := block.ValidateBlock()
	if validationSuccess {
		t.Error("Should have failed on BlockNonce Validation")
	}

}

func TestVerifyBlockFAILBlockSignature(t *testing.T) {
	var sk, pk = KeyGen(2048)
	var sk2, _ = KeyGen(2048)

	nonce := BlockNonce{"8556", "", pk}
	nonce.SignBlockNonce(sk)

	block := Block{0,
		"",
		pk,
		"",
		nonce,
		"",
		Data{},
		""}

	block.SignBlock(sk2)

	validationSuccess, _ := block.ValidateBlock()
	if validationSuccess {
		t.Error("Should have failed on BlockSignature Validation")
	}

}

//func TestVerifyBlockSignature(t *testing.T) {
//	var sk, pk = KeyGen(2048)
//
//	b := GetTestBlock()
//	b.SignBlock(sk)
//
//	if !b.validateBlockSignature(pk) {
//		t.Error("Block Failed")
//	}
//
//}

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
//		Data{},
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
