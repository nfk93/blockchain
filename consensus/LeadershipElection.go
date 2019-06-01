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

func CalculateDraw(leadershipNonce string, hardness float64, sk SecretKey, pk PublicKey, slot uint64) (bool, string) {
	// Creates the draw signature
	var drawBuf bytes.Buffer
	drawBuf.WriteString("LEADERSHIP_ELECTION")
	drawBuf.WriteString(leadershipNonce)
	drawBuf.WriteString(strconv.Itoa(int(slot)))
	draw := Sign(drawBuf.String(), sk)

	// Creates a block for transporting data to ValidateDraw
	intermediateBlock := Block{slot,
		"",
		pk,
		draw,
		BlockNonce{},
		"",
		BlockData{},
		"",
		""}
	intermediateBlock.SignBlock(sk)

	// Checks if the value of the draw exceeds the threshold.
	// If so it returns the draw else it return an empty string.
	if ValidateDraw(intermediateBlock, leadershipNonce, hardness) {
		return true, draw
	}
	return false, ""
}

func ValidateDraw(b Block, leadershipNonce string, hardness float64) bool {

	if !b.ValidateBlock() {
		fmt.Println("Block didn't validate")
		return false
	}

	if !validateDrawSignature(b, leadershipNonce) {
		fmt.Println("Draw signature didn't validate...")
		return false
	}

	// Calculates the threshold
	phiFunc := float64(1) - math.Pow(float64(1)-hardness, getLotteryPower(b.BakerID, b.Slot))
	multFactor := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(256)), nil)
	threshold := new(big.Int)
	new(big.Float).Mul(big.NewFloat(float64(phiFunc)), new(big.Float).SetInt(multFactor)).Int(threshold)

	drawVal := CalculateDrawValue(b, leadershipNonce)

	// Checks if the draw is less than the threshold
	// Returns -1 if x < y
	if drawVal.Cmp(threshold) == -1 {
		return true
	}
	return false
}

func CalculateDrawValue(b Block, leadershipNonce string) *big.Int {
	// Calculates the draw value
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(leadershipNonce)
	valBuf.WriteString(strconv.Itoa(int(b.Slot)))
	valBuf.WriteString(b.Draw)
	hashString := HashSHA(valBuf.String())
	hashVal := big.NewInt(0)
	hashVal.SetString(hashString, 16)
	return hashVal
}

func validateDrawSignature(b Block, leadershipNonce string) bool {
	var buf bytes.Buffer
	buf.WriteString("LEADERSHIP_ELECTION")
	buf.WriteString(leadershipNonce)
	buf.WriteString(strconv.Itoa(int(b.Slot)))
	return Verify(buf.String(), b.Draw, b.BakerID)
}
