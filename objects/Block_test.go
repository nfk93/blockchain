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

	nonce := BlockNonce{"8556", "Something", pk}

	blockNonce := CreateNewBlockNonce(nonce, 2, sk, pk)

	if !blockNonce.validateBlockNonce() {
		t.Error("Block Failed")
	}

}

func TestCreateAndVerifyNonceFAIL(t *testing.T) {
	var sk, _ = KeyGen(256)
	var _, pk2 = KeyGen(256)

	nonce := BlockNonce{"8556", "Something", pk2}

	blockNonce := CreateNewBlockNonce(nonce, 2, sk, pk2)

	if blockNonce.validateBlockNonce() {
		t.Error("Block Shouldn't verify!!")
	}

}

func TestVerifyBlock(t *testing.T) {
	var sk, pk = KeyGen(2560)

	slot := 4
	hardness := 49
	stake := 9999999

	// setup of non signed nonce and block to create proper block and nonce from
	preNonce := BlockNonce{"8556", "Something", pk}
	prevBlock := Block{0,
		"",
		pk,
		"",
		preNonce,
		"",
		Data{},
		""}

	nonce := CreateNewBlockNonce(preNonce, slot, sk, pk)

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
