package objects

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	"math"
	"math/big"
	"strconv"
)

func CalculateDraw(bnonce BlockNonce, hardness float64, sk SecretKey, pk PublicKey, yourStake int, systemStake int, slot int) (bool, string) {
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

	if ValidateDrawValue(intermediateBlock, yourStake, systemStake, hardness) {
		return true, draw
	}
	return false, ""

}

func ValidateDrawValue(b Block, yourStake int, systemStake int, hardness float64) bool {

	if !b.validateBlockProof() {
		fmt.Println("Block Proof didn't validate!")
		return false
	}
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(b.BlockNonce.Nonce)
	valBuf.WriteString(strconv.Itoa(b.Slot))
	valBuf.WriteString(b.BakerID.String())
	valBuf.WriteString(b.BlockProof)

	hashVal := big.NewInt(0)
	hashVal.SetString(HashSHA(valBuf.String()), 10)

	percentOfTotalStake := float64(yourStake) / float64(systemStake)
	phiFunc := float64(1) - math.Pow(float64(1)-hardness, float64(percentOfTotalStake))
	multFactor := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(hashVal.BitLen())), nil)

	threshold := new(big.Int)
	new(big.Float).Mul(big.NewFloat(float64(phiFunc)), new(big.Float).SetInt(multFactor)).Int(threshold)

	// Checks if the draw is bigger than the threshold
	// Returns -1 if x < y
	if hashVal.Cmp(threshold) == -1 {
		return true
	}

	return false

}
