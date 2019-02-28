package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
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

//Signing and verification of Blocks

func buildBlockStringToSign(b Block) string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(b.Slot))
	buf.WriteString(b.ParentPointer)
	buf.WriteString(strconv.Itoa(b.BakerID))
	buf.WriteString(b.BlockProof)
	buf.WriteString(strconv.Itoa(b.BlockNonce))
	buf.WriteString(b.LastFinalized)
	buf.WriteString(b.BlockData.DataString())
	return buf.String()
}

func SignBlock(b Block, sk SecretKey) Block {
	m := buildBlockStringToSign(b)
	s := Sign(m, sk)
	b.Signature = s
	return b
}

func VerifyBlock(b Block, pk PublicKey) bool {
	return Verify(buildBlockStringToSign(b), b.Signature, pk)
}
