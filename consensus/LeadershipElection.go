package consensus

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	"math"
	"math/big"
	"strconv"
)

func CalculateDraw(nonce string, hardness float64, sk SecretKey, pk PublicKey, yourStake int, systemStake int, slot int) (bool, string) {

	// Creates the draw signature
	var drawBuf bytes.Buffer
	drawBuf.WriteString("LEADERSHIP_ELECTION")
	drawBuf.WriteString(nonce)
	drawBuf.WriteString(strconv.Itoa(slot))
	draw := Sign(drawBuf.String(), sk)

	// Creates a block for transporting data to ValidateDraw
	intermediateBlock := Block{slot,
		"",
		pk,
		draw,
		nonce,
		"",
		Data{},
		""}
	intermediateBlock.SignBlock(sk)

	// Checks if the value of the draw exceeds the threshold.
	// If so it returns the draw else it return an empty string.
	if ValidateDraw(intermediateBlock, yourStake, systemStake, hardness) {
		return true, draw
	}
	return false, ""
}

func ValidateDraw(b Block, yourStake int, systemStake int, hardness float64) bool {

	if !b.ValidateBlock() {
		fmt.Println("Block didn't validate")
		return false
	}

	if !validateDrawSignature(b) {
		fmt.Println("Draw signature didn't validate...")
		return false
	}

	// Calculates the draw value
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(b.BlockNonce)
	valBuf.WriteString(strconv.Itoa(b.Slot))
	valBuf.WriteString(b.BakerID.String())
	valBuf.WriteString(b.Draw)
	hashVal := big.NewInt(0)
	hashVal.SetString(HashSHA(valBuf.String()), 10)

	// Calculates the threshold
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

func validateDrawSignature(b Block) bool {
	var buf bytes.Buffer
	buf.WriteString("LEADERSHIP_ELECTION")
	buf.WriteString(b.BlockNonce)
	buf.WriteString(strconv.Itoa(b.Slot))

	return Verify(buf.String(), b.Draw, b.BakerID)
}
