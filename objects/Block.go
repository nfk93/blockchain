package objects

import (
	"bytes"
	"strconv"
)

type Block struct {
	Slot          int
	ParentPointer string //a hash of parent block
	BakerID       int    //
	BlockProof    string //TODO: Change to right type (whitepaper 2.5.3)
	BlockNonce    int
	LastFinalized string //hash of last finalized block
	BlockData     Data
	Signature     string
}

type Data struct {
	Value int
}

func (d *Data) DataString() string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(d.Value))
	return buf.String()
}

func GetTestBlock() Block {
	return Block{42,
		"",
		42,
		"VALID",
		42,
		"",
		Data{42},
		""}
}
