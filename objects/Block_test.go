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
