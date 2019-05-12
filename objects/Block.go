package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
)

type Block struct {
	Slot           int
	ParentPointer  string //a hash of parent block
	BakerID        PublicKey
	Draw           string
	BlockNonce     BlockNonce
	LastFinalized  string //hash of last finalized block
	BlockData      BlockData
	StateHash      string
	BlockSignature string
}

type CreateBlockData struct {
	TransList     []Transaction
	Sk            SecretKey
	Pk            PublicKey
	SlotNo        int
	Draw          string
	BlockNonce    BlockNonce
	LastFinalized string
}

type BlockNonce struct {
	Nonce string
	Proof string
}

type BlockData struct {
	Trans       []Transaction
	GenesisData GenesisData
}

// Block Functions
func (b *Block) SignBlock(sk SecretKey) {
	m := b.toString()
	b.BlockSignature = Sign(m, sk)
}

func (b Block) ValidateBlock() bool {
	return Verify(b.toString(), b.BlockSignature, b.BakerID)
}

func (b Block) toString() string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(b.Slot))
	buf.WriteString(b.ParentPointer)
	buf.WriteString(b.BakerID.String())
	buf.WriteString(b.Draw)
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(b.BlockNonce.Proof)
	buf.WriteString(b.LastFinalized)
	buf.WriteString(b.BlockData.toString())
	buf.WriteString(b.StateHash)
	return buf.String()
}

func (b *Block) CalculateBlockHash() string {
	return HashSHA(b.toString())
}

func (d *BlockData) toString() string {
	var buf bytes.Buffer
	for _, t := range d.Trans {
		buf.WriteString(t.toString())
	}
	return buf.String()
}

// BlockNonce Functions
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

func (b *Block) validateBlockNonce(leadershipNonce string) bool {
	var buf bytes.Buffer
	buf.WriteString("NONCE") //Old block nonce
	buf.WriteString(leadershipNonce)
	buf.WriteString(strconv.Itoa(b.Slot))
	correctSignature := Verify(buf.String(), b.BlockNonce.Proof, b.BakerID)
	correctNonce := HashSHA(b.BlockNonce.Proof) == b.BlockNonce.Nonce
	return correctSignature && correctNonce
}
