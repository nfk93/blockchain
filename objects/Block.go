package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
)

type Block struct {
	Slot           uint64
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
	TransList     []TransData
	Sk            SecretKey
	Pk            PublicKey
	SlotNo        uint64
	Draw          string
	BlockNonce    BlockNonce
	LastFinalized string
}

type BlockNonce struct {
	Nonce string
	Proof string
}

type BlockData struct {
	Trans       []TransData
	GenesisData GenesisData
}

type TransData struct {
	Transaction  Transaction
	ContractCall ContractCall
	ContractInit ContractInitialize
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
	buf.WriteString(strconv.Itoa(int(b.Slot)))
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
		switch t.GetType() {
		case 1:
			buf.WriteString(t.Transaction.toString())
		case 2:
			buf.WriteString(t.ContractCall.toString())
		case 3:
			buf.WriteString(t.ContractInit.toString())

		}
	}
	return buf.String()
}

// BlockNonce Functions
func CreateNewBlockNonce(leadershipNonce string, sk SecretKey, slot uint64) BlockNonce {
	var buf bytes.Buffer
	buf.WriteString("NONCE") //Old block nonce
	buf.WriteString(leadershipNonce)
	buf.WriteString(strconv.Itoa(int(slot)))
	newNonceString := buf.String()
	proof := Sign(newNonceString, sk)
	newNonce := HashSHA(proof)
	return BlockNonce{newNonce, proof}
}

func (b *Block) validateBlockNonce(leadershipNonce string) bool {
	var buf bytes.Buffer
	buf.WriteString("NONCE") //Old block nonce
	buf.WriteString(leadershipNonce)
	buf.WriteString(strconv.Itoa(int(b.Slot)))
	correctSignature := Verify(buf.String(), b.BlockNonce.Proof, b.BakerID)
	correctNonce := HashSHA(b.BlockNonce.Proof) == b.BlockNonce.Nonce
	return correctSignature && correctNonce
}

func (t TransData) GetType() int {
	if t.Transaction != (Transaction{}) {
		return TRANSACTION
	}
	if t.ContractCall != (ContractCall{}) {
		return CONTRACTCALL
	}
	if t.ContractInit.Owner != (PublicKey{}) ||
		t.ContractInit.Prepaid != 0 ||
		t.ContractInit.Gas != 0 ||
		string(t.ContractInit.Code) != "" {
		return CONTRACTINIT
	}
	return ERROR
}

func (t TransData) GetID() string {
	if t.Transaction != (Transaction{}) {
		return t.Transaction.ID
	}
	if t.ContractCall != (ContractCall{}) {
		return t.ContractCall.Nonce
	}
	if t.ContractInit.Owner != (PublicKey{}) ||
		t.ContractInit.Prepaid != 0 ||
		t.ContractInit.Gas != 0 ||
		string(t.ContractInit.Code) != "" {
		return CONTRACTINIT
	}
	return ERROR
}

type transtype int

const (
	TRANSACTION = iota
	CONTRACTCALL
	CONTRACTINIT
	ERROR
)
