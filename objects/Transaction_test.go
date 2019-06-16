package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/transaction"
	"testing"
)

func TestVerifyTransaction(t *testing.T) {
	var sk, pk = KeyGen(2048)
	var _, pk2 = KeyGen(2048)
	b := Transaction{pk, pk2, 200, "1", ""}
	b.SignTransaction(sk)

	if !b.VerifyTransaction() {
		t.Error("Verification failed")
	}

}

func initState() {
	channels := CreateChannelStruct()
	go transaction.StartTransactionLayer(channels, false)

}
