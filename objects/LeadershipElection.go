package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"math/big"
	"strconv"
)

type BlockNonce struct {
	Nonce     string
	Signature string
}

func LeadershipElection(bnonce BlockNonce, hardness int, sk SecretKey, pk PublicKey, stake int, slot int) (bool, string) {
	var drawBuf bytes.Buffer
	drawBuf.WriteString("LEADERSHIP_ELECTION")
	drawBuf.WriteString(bnonce.Nonce)
	drawBuf.WriteString(strconv.Itoa(slot))

	draw := Sign(drawBuf.String(), sk)

	intermediateBlock := Block{slot,
		"",
		pk,
		draw,
		bnonce,
		"",
		Data{},
		""}

	if intermediateBlock.validateDraw(stake, hardness) {
		return true, draw
	}
	return false, ""

}

func (b Block) verifyBlockProof() bool {
	var buf bytes.Buffer
	buf.WriteString("LEADERSHIP_ELECTION")
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(strconv.Itoa(b.Slot))

	return Verify(buf.String(), b.BlockProof, b.BakerID)
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

func (bl BlockNonce) verifyBlockNonce(pk PublicKey) bool {

	return Verify(bl.Nonce, bl.Signature, pk)
}

func (b Block) CreateNewBlockNonce(slot int, sk SecretKey) BlockNonce {
	var buf bytes.Buffer
	buf.WriteString("NONCE")
	buf.WriteString(b.BlockNonce.Nonce) //Old block nonce //TODO: Should also contain new states
	buf.WriteString(strconv.Itoa(slot))

	newNonceString := buf.String()
	newNonce := HashSHA(newNonceString)
	signature := Sign(string(newNonce), sk)

	return BlockNonce{newNonce, signature}
}

//func (b Block) validateBlock()  {
//
//}
