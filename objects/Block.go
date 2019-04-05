package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
)

type Block struct {
	Slot           int
	ParentPointer  string    //a hash of parent block
	BakerID        PublicKey //
	Draw           string
	BlockNonce     string
	LastFinalized  string //hash of last finalized block
	BlockData      Data
	BlockSignature string
}

type CreateBlockData struct {
	TransList  []Transaction
	Sk         SecretKey
	Pk         PublicKey
	SlotNo     int
	Draw       string
	BlockNonce string
}

type Data struct {
	Trans       []Transaction
	GenesisData GenesisData
}

//Signing functions
func (b *Block) SignBlock(sk SecretKey) {
	m := buildBlockStringToSign(*b)
	b.BlockSignature = Sign(m, sk)
}

// Validation functions

func (b Block) ValidateBlock() bool {
	return Verify(buildBlockStringToSign(b), b.BlockSignature, b.BakerID)
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
	buf.WriteString(b.BlockNonce)
	buf.WriteString(b.LastFinalized)
	buf.WriteString(b.BlockData.DataString())
	return buf.String()
}

func (b *Block) CalculateBlockHash() string {
	return HashSHA(buildBlockStringToSign(*b))
}

func CreateNewBlockNonce(nonce BlockNonce, slot int, sk SecretKey, pk PublicKey) BlockNonce {
	var buf bytes.Buffer
	buf.WriteString("NONCE")
	buf.WriteString(nonce.Nonce) //Old block nonce //TODO: Should also contain new states
	buf.WriteString(strconv.Itoa(slot))

	newNonceString := buf.String()
	newNonce := HashSHA(newNonceString)
	signature := Sign(string(newNonce), sk)

	return BlockNonce{newNonce, signature, pk}
}

type BlockNonce struct {
	Nonce     string
	Signature string
	Pk        PublicKey
}

func (bl BlockNonce) validateBlockNonce() bool {

	return Verify(bl.Nonce, bl.Signature, bl.Pk)
}
