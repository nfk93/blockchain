package objects

import (
	"github.com/nfk93/blockchain/crypto"
	"github.com/pkg/errors"
	"log"
	"time"
)

type GenesisData struct {
	GenesisTime  time.Time
	SlotDuration time.Duration
	Nonce        string
	Hardness     float64
	InitialState State
	// TODO: fill with more stuff?
}

func NewGenesisData(publicKey crypto.PublicKey, secretKey crypto.SecretKey, slotDuration time.Duration, hardness float64) (GenesisData, error) {
	time := time.Now()
	state := NewInitialState(publicKey)
	nonce, err := crypto.GenerateRandomBytes(24)
	if err != nil {
		log.Fatal("oops") // TODO shouldn't happen, but maybe make realistic error handling
	}
	if hardness <= 0 || hardness >= 1 {
		return GenesisData{}, errors.Errorf("Hardness must be between 0 and 1")
	} else {
		return GenesisData{time, slotDuration, string(nonce), hardness, state}, nil
	}
}

func CreateTestGenesis() Block {
	return Block{0,
		"",
		crypto.PublicKey{},
		"",
		BlockNonce{},
		"",
		Data{},
		""}
}
