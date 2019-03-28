package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"math/big"
	"strconv"
)

func CalculateDraw(bnonce BlockNonce, hardness int, sk SecretKey, pk PublicKey, stake int, slot int) (bool, string) {
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

	if ValidateDraw(intermediateBlock, stake, hardness) {
		return true, draw
	}
	return false, ""

}

func ValidateDraw(b Block, stake int, hardness int) bool {
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(b.BlockNonce.Nonce)
	valBuf.WriteString(strconv.Itoa(b.Slot))
	valBuf.WriteString(b.BakerID.String())
	valBuf.WriteString(b.BlockProof)

	hashVal := big.NewInt(0)
	hashVal.SetString(HashSHA(valBuf.String()), 10)

	//asdf := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(len(b.BlockProof))), nil)
	//fmt.Println(asdf)
	drawValue := big.NewInt(0).Mul(hashVal, big.NewInt(int64(stake))) //TODO: How is the draw value calculated?

	threshold := big.NewInt(0).Exp(big.NewInt(int64(hardness)), big.NewInt(int64(hardness)), nil) //TODO how to calc threshold?

	// Checks if the draw is bigger than the threshold
	// Returns -1 if x < y
	if drawValue.Cmp(threshold) < 0 {
		return false
	}

	return true

}
