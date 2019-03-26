package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

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

	if !blockNonce.verifyBlockNonce(pk) {
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

	if blockNonce.verifyBlockNonce(pk2) {
		t.Error("Block Shouldn't verify!!")
	}

}

func TestLeadershipElection(t *testing.T) {
	var sk, pk = KeyGen(2560)

	stake := 9999999
	slot := 3
	hardness := 49

	nonce := BlockNonce{"8556", "Something"}

	prevBlock := Block{0,
		"",
		pk,
		"",
		nonce,
		"",
		Data{},
		""}
	blockNonce := prevBlock.CreateNewBlockNonce(slot, sk)

	b, draw := LeadershipElection(blockNonce, hardness, sk, pk, stake, slot)

	if !b {
		t.Error("Draw didn't exceed Hardness...")
	}

	someBlock := Block{slot,
		"",
		pk,
		draw,
		blockNonce,
		"",
		Data{},
		""}

	if !someBlock.verifyBlockProof() {
		t.Error("Block Proof couldn't verify...")
	}

}
