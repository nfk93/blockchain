package transaction

import (
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"testing"
)

func TestReceiveBlock(t *testing.T) {
	sk1, p1 := KeyGen(256)
	_, p2 := KeyGen(256)
	t1 := Transaction{p1, p2, 200, "ID112", ""}
	t2 := Transaction{p1, p2, 300, "ID222", ""}
	t1 = SignTransaction(t1, sk1)
	t2 = SignTransaction(t2, sk1)

	b := Block{42,
		"",
		42,
		"VALID",
		42,
		"",
		Data{[]Transaction{t1, t2}},
		"",
	}

	BeRich(p1.String())

	ReceiveBlock(b)

	fmt.Println(GetLedger())

	if GetShare(p2.String()) != 500 {
		t.Error("P2 does not own 500")
	}

}
