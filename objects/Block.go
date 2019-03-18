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
	BlockHash     string
	Signature     string
}

type Data struct {
	Trans []Transaction
}

func (d *Data) DataString() string {
	var buf bytes.Buffer
	for _, t := range d.Trans {
		buf.WriteString(t.buildStringToSign())
	}
	return buf.String()
}

func GetTestBlock() Block {
	return Block{42,
		"",
		42,
		"VALID",
		42,
		"",
		Data{},
		"",
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

func (b *Block) SignBlock(sk SecretKey) {
	m := buildBlockStringToSign(*b)
	b.Signature = Sign(m, sk)
}

func (b *Block) VerifyBlock(pk PublicKey) bool {
	return Verify(buildBlockStringToSign(*b), b.Signature, pk)
}

func (b *Block) HashBlock() {
	b.BlockHash = HashSHA(buildBlockStringToSign(*b))
}
