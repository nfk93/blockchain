package objects

import (
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
