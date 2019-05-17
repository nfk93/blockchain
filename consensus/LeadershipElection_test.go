package consensus

import (
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"testing"
)

// Should not always succeed
func TestLeadershipElection(t *testing.T) {
	var sk, pk = KeyGen(2048)

	//yourStake := 200
	//systemStake := 300
	slot := uint64(3)
	hardness := 0.9

	b, draw := CalculateDraw("8556", hardness, sk, pk, slot)

	if !b {
		t.Error("Draw didn't exceed Hardness...")
	}

	someBlock := Block{slot,
		"",
		pk,
		draw,
		BlockNonce{},
		"",
		BlockData{},
		"",
		""}
	someBlock.SignBlock(sk)

	if !someBlock.ValidateBlock() {
		t.Error("Block Proof couldn't verify...")
	}

}

func TestHardness(t *testing.T) {

	winCounter := 0
	rounds := 1000
	var sk, pk = KeyGen(2048)

	//yourStake := 15000000
	//systemStake := 30000000
	hardness := 0.98

	for i := 0; i < rounds; i++ {

		//nonce = CreateNewBlockNonce(nonce, i, sk, pk)
		i := uint64(i)

		b, _ := CalculateDraw("8556", hardness, sk, pk, i)
		if b {
			winCounter += 1
		}
	}
	winnRate := float64(rounds) / float64(winCounter)
	hardnessRate := float64(100) - winnRate
	fmt.Printf("Winrate was %v %% \n", winnRate)
	fmt.Printf("Wanted Hardness was %v%% and Actual Hardness was at %v%% \n", hardness*100, hardnessRate)
}
