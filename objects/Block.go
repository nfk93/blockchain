package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
)

type Block struct {
	Slot          int
	ParentPointer string    //a hash of parent block
	BakerID       PublicKey //
	Draw          string
	BlockNonce    BlockNonce
	LastFinalized string //hash of last finalized block
	BlockData     Data
	Signature     string
}

type CreateBlockData struct {
	TransList []Transaction
	Sk        SecretKey
	Pk        PublicKey
	SlotNo    int
	Draw      string
}

type BlockNonce struct {
	Nonce     string
	Signature string
	Pk        PublicKey
}

type Data struct {
	Trans       []Transaction
	GenesisData GenesisData
}

//Signing of Blocks
func (b *Block) SignBlock(sk SecretKey) {
	m := buildBlockStringToSign(*b)
	b.Signature = Sign(m, sk)
}

// Validation functions
func (b *Block) ValidateBlockSignature(pk PublicKey) bool {
	return Verify(buildBlockStringToSign(*b), b.Signature, pk)
}

func (b *Block) ValidateBlockDrawSignature() bool {
	var buf bytes.Buffer
	buf.WriteString("LEADERSHIP_ELECTION")
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(strconv.Itoa(b.Slot))

	return Verify(buf.String(), b.Draw, b.BakerID)
}

func (bl BlockNonce) validateBlockNonce() bool {

	return Verify(bl.Nonce, bl.Signature, bl.Pk)
}

func (b Block) ValidateBlock() (bool, string) {

	if !b.ValidateBlockDrawSignature() {
		return false, "Block Proof failed"
	}

	if !b.BlockNonce.validateBlockNonce() {
		return false, "Block Nonce validation failed"
	}

	if !b.ValidateBlockSignature(b.BakerID) {
		return false, "Block Signature validation failed"
	}

	return true, ""
}

// Helpers
func (d *Data) DataString() string {
	var buf bytes.Buffer
	for _, t := range d.Trans {
		buf.WriteString(t.buildStringToSign())
	}
	return buf.String()
}

func buildBlockStringToSign(b Block) string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(b.Slot))
	buf.WriteString(b.ParentPointer)
	buf.WriteString(b.BakerID.N.String())
	buf.WriteString(b.BakerID.E.String())
	buf.WriteString(b.Draw)
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(b.LastFinalized)
	buf.WriteString(b.BlockData.DataString())
	return buf.String()
}

func (b *Block) CalculateBlockHash() string {
	return HashSHA(buildBlockStringToSign(*b))
}

func GetTestBlock() Block {
	_, pk := KeyGen(256)
	return Block{42,
		"",
		pk,
		"VALID",
		BlockNonce{"42", "", pk},
		"",
		Data{},
		""}
}
