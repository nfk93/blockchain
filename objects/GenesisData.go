package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"time"
)

type GenesisData struct {
	GenesisTime  time.Time
	SlotDuration time.Duration
	Nonce        BlockNonce
	Hardness     float64
	InitialState State
	// TODO: fill with more stuff?
}

func CreateTestGenesis() Block {
	sk, pk := KeyGen(2048)
	genBlock := Block{0,
		"",
		PublicKey{},
		"",
		BlockNonce{"GENESIS", Sign("GENESIS", sk), pk},
		"",
		Data{[]Transaction{}, GenesisData{}}, //TODO: GENESISDATA should be proper created
		""}

	genBlock.LastFinalized = genBlock.CalculateBlockHash()
	return genBlock
}
