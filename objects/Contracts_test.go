package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestVerification(t *testing.T) {
	var sk, pk = KeyGen(2048)
	cc := CreateContractCall("Flot", "test", "tis", 20, 2, "adresse", pk, sk)
	if !cc.Verify() {
		t.Error("Verification of ContractCall failed")
	}
	sk, pk = KeyGen(2048)

	ci := CreateContractInit(pk, []byte("test"), 23, 20, 2, sk)
	if !ci.Verify() {
		t.Error("Verification of ContractCall failed")
	}
}
