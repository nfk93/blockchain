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

// WIP - Slotduration is in seconds
func CreateTestGenesisData(slotDuration int, hardness float64, pk PublicKey, sk SecretKey) GenesisData {
	nonce := HashSHA("TestNonce")
	return GenesisData{
		time.Now(),
		time.Duration(slotDuration) * time.Second,
		BlockNonce{nonce, Sign(nonce, sk), pk},
		hardness,
		State{},
	}
}
