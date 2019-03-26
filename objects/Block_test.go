package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
	"testing"
)

func TestVerifyBlockSignature(t *testing.T) {
	var sk, pk = KeyGen(2560)

	b := GetTestBlock()
	b.SignBlock(sk)

	if !b.validateBlockSignature(pk) {
		t.Error("Block Failed")
	}

}

func TestCreateAndVerifyNonce(t *testing.T) {
	var sk, pk = KeyGen(256)

	nonce := BlockNonce{"8556", "Something"}

	prevBlock := Block{0,
		"",
		pk,
		"",
		nonce,
		"",
		Data{},
		""}

	blockNonce := prevBlock.CreateNewBlockNonce(2, sk)

	if !blockNonce.validateBlockNonce(pk) {
		t.Error("Block Failed")
	}

}

func TestCreateAndVerifyNonceFAIL(t *testing.T) {
	var sk, _ = KeyGen(256)
	var _, pk2 = KeyGen(256)

	nonce := BlockNonce{"8556", "Something"}

	prevBlock := Block{0,
		"",
		pk2,
		"",
		nonce,
		"",
		Data{},
		""}
	blockNonce := prevBlock.CreateNewBlockNonce(2, sk)

	if blockNonce.validateBlockNonce(pk2) {
		t.Error("Block Shouldn't verify!!")
	}

}

func TestVerifyBlock(t *testing.T) {
	var sk, pk = KeyGen(2560)

	slot := 4
	hardness := 49
	stake := 9999999

	// setup of non signed nonce and block to create proper block and nonce from
	preNonce := BlockNonce{"8556", "Something"}
	prevBlock := Block{0,
		"",
		pk,
		"",
		preNonce,
		"",
		Data{},
		""}

	nonce := prevBlock.CreateNewBlockNonce(slot, sk)

	createSuccess, draw := CalculateDraw(nonce, hardness, sk, pk, stake, slot)
	if !createSuccess {
		t.Error("Draw not above Hardness")
	}
	blockdata := CreateBlockData{[]Transaction{}, sk, pk, slot, "", draw}
	parent := prevBlock.CalculateBlockHash()

	t1 := CreateTransaction(pk, pk, 200, strconv.Itoa(1), sk)
	t2 := CreateTransaction(pk, pk, 200, strconv.Itoa(1), sk)
	t3 := CreateTransaction(pk, pk, 200, strconv.Itoa(1), sk)

	b := CreateNewBlock(blockdata, parent, nonce, []Transaction{t1, t2, t3})

	validationSuccess, errMsg := b.ValidateBlock(stake, hardness)

	if !validationSuccess {
		t.Error(errMsg)
	}

}
