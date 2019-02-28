package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestVerifyBlock(t *testing.T) {
	var sk, pk = KeyGen(256)

	b := GetTestBlock()
	b = SignBlock(b, sk)

	if !VerifyBlock(b, pk) {
		t.Error("Block Failed")
	}

}
