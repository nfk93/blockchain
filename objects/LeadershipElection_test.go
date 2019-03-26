package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

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

	b, draw := CalculateDraw(blockNonce, hardness, sk, pk, stake, slot)

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

	if !someBlock.validateBlockProof() {
		t.Error("Block Proof couldn't verify...")
	}

}
