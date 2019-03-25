package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestVerifyTransaction(t *testing.T) {
	var sk, pk = KeyGen(2000)
	var _, pk2 = KeyGen(2000)
	b := Transaction{pk, pk2, 200, "1", ""}
	b.SignTransaction(sk)

	if !b.VerifyTransaction() {
		t.Error("Verification failed")
	}

}
