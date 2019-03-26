package objects

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	"math/big"
	"strconv"
)

type Block struct {
	Slot          int
	ParentPointer string    //a hash of parent block
	BakerID       PublicKey //
	BlockProof    string
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
	LastFinal string
	Draw      string
}

type BlockNonce struct {
	Nonce     string
	Signature string
	Pk        PublicKey
}

type Data struct {
	Trans []Transaction
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

func CreateNewBlock(blockData CreateBlockData, parent string, nonce BlockNonce, translist []Transaction) Block {
	validNonce := nonce.validateBlockNonce()
	if !validNonce {
		fmt.Println("Couldn't create block! Nonce is not verified!")
		return Block{}
	}
	b := Block{blockData.SlotNo,
		parent,
		blockData.Pk,
		blockData.Draw,
		nonce,
		blockData.LastFinal,
		Data{translist},
		""}
	b.SignBlock(blockData.Sk)
	return b
}

//Signing of Blocks
func (b *Block) SignBlock(sk SecretKey) {
	m := buildBlockStringToSign(*b)
	b.Signature = Sign(m, sk)
}

// Validation functions
func (b *Block) validateBlockSignature(pk PublicKey) bool {
	return Verify(buildBlockStringToSign(*b), b.Signature, pk)
}

func (b Block) validateDraw(stake int, hardness int) bool {
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(b.BlockNonce.Nonce)
	valBuf.WriteString(strconv.Itoa(b.Slot))
	valBuf.WriteString(b.BakerID.String())
	valBuf.WriteString(b.BlockProof)

	hashVal := big.NewInt(0)
	hashVal.SetString(HashSHA(valBuf.String()), 10)

	drawValue := big.NewInt(0).Mul(hashVal, big.NewInt(int64(stake))) //TODO: How is the draw value calculated?

	threshold := big.NewInt(0).Exp(big.NewInt(int64(hardness)), big.NewInt(int64(hardness)), nil) //TODO how to calc threshold?

	// Checks if the draw is bigger than the threshold
	// Returns -1 if x < y
	if drawValue.Cmp(threshold) < 0 {
		return false
	}

	return true

}

func (b Block) validateBlockProof() bool {
	var buf bytes.Buffer
	buf.WriteString("LEADERSHIP_ELECTION")
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(strconv.Itoa(b.Slot))

	return Verify(buf.String(), b.BlockProof, b.BakerID)
}

func (bl BlockNonce) validateBlockNonce() bool {

	return Verify(bl.Nonce, bl.Signature, bl.Pk)
}

func (b Block) ValidateBlock(stake int, hardness int) (bool, string) {

	if !b.validateBlockProof() {
		return false, "Block Proof failed"
	}

	if !b.validateDraw(stake, hardness) {
		return false, "Block failed Hardness"
	}

	if !b.BlockNonce.validateBlockNonce() {
		return false, "Block Nonce validation failed"
	}

	if !b.validateBlockSignature(b.BakerID) {
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
	buf.WriteString(b.BlockProof)
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
