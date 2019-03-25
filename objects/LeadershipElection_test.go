package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestCreateAndVerifyNonce(t *testing.T) {
	var sk, pk = KeyGen(256)

	nonce := BlockNonce{"8556", "Something"}
	blockNonce := CreateNewBlockNonce(nonce, 2, sk)

	if !blockNonce.verifyBlockNonce(pk) {
		t.Error("Block Failed")
	}

}

func TestCreateAndVerifyNonceFAIL(t *testing.T) {
	var sk, _ = KeyGen(256)
	var _, pk2 = KeyGen(256)

	nonce := BlockNonce{"8556", "Something"}
	blockNonce := CreateNewBlockNonce(nonce, 2, sk)

	if blockNonce.verifyBlockNonce(pk2) {
		t.Error("Block Shouldn't verify!!")
	}

}
