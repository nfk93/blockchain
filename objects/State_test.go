package objects

import (
	"fmt"
	"github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestTotalAmount(t *testing.T) {

	state := State{make(map[crypto.PublicKey]int), "ThisShouldBeAString"}

	limit := 5
	for i := 0; i < limit; i++ {
		_, pk := crypto.KeyGen(25)
		state.Ledger[pk] += 100
	}

	if state.TotalSystemStake() != 100*limit {
		t.Error("System Stake not in balance")
	}
	fmt.Println(state.StateAsString())
}

//TODO> Make test for adding Transactions and checking those
