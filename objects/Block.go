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
	BlockNonce     BlockNonce
	LastFinalized  string //hash of last finalized block
	BlockData      Data
	StateHash      string
	BlockSignature string
}

type CreateBlockData struct {
	TransList  []Transaction
	Sk         SecretKey
	Pk         PublicKey
	SlotNo     int
	Draw       string
	BlockNonce BlockNonce
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
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(b.BlockNonce.Proof)
	buf.WriteString(b.LastFinalized)
	buf.WriteString(b.BlockData.DataString())
	buf.WriteString(b.StateHash)
	return buf.String()
}

func (b *Block) CalculateBlockHash() string {
	return HashSHA(buildBlockStringToSign(*b))
}

func CreateNewBlockNonce(leadershipNonce string, sk SecretKey, slot int) BlockNonce {
	var buf bytes.Buffer
	buf.WriteString("NONCE") //Old block nonce
	buf.WriteString(leadershipNonce)
	buf.WriteString(strconv.Itoa(slot))
	newNonceString := buf.String()
	proof := Sign(newNonceString, sk)
	newNonce := HashSHA(proof)
	return BlockNonce{newNonce, proof}
}

type BlockNonce struct {
	Nonce string
	Proof string
}

func (b *Block) validateBlockNonce(leadershipNonce string) bool {
	var buf bytes.Buffer
	buf.WriteString("NONCE") //Old block nonce
	buf.WriteString(leadershipNonce)
	buf.WriteString(strconv.Itoa(b.Slot))
	correctSignature := Verify(buf.String(), b.BlockNonce.Proof, b.BakerID)
	correctNonce := HashSHA(b.BlockNonce.Proof) == b.BlockNonce.Nonce
	return correctSignature && correctNonce
}
