package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestVerifyBlock(t *testing.T) {
	var sk, pk = KeyGen(2560)

	b := GetTestBlock()
	b.SignBlock(sk)

	if !b.VerifyBlockSignature(pk) {
		t.Error("Block Failed")
	}

}
