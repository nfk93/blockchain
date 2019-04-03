package consensus

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	. "github.com/nfk93/blockchain/objects"
	. "github.com/nfk93/blockchain/transaction"
	"math"
	"math/big"
	"strconv"
)

func CalculateDraw(bnonce BlockNonce, pk PublicKey, slot int, s State) (bool, string) {
	var drawBuf bytes.Buffer
	drawBuf.WriteString("LEADERSHIP_ELECTION")
	drawBuf.WriteString(bnonce.Nonce)
	drawBuf.WriteString(strconv.Itoa(slot))

	draw := Sign(drawBuf.String(), sk)

	transportBlock := Block{slot,
		"",
		pk,
		draw,
		bnonce,
		"",
		Data{},
		""}

	if ValidateDraw(transportBlock, s) {
		return true, draw
	}
	return false, ""

}

func ValidateDraw(b Block, currentState State) bool {

	if !validateDrawSignature(b) {
		fmt.Println("Block Proof didn't validate!")
		return false
	}
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(b.BlockNonce.Nonce)
	valBuf.WriteString(strconv.Itoa(b.Slot))
	valBuf.WriteString(b.BakerID.String())
	valBuf.WriteString(b.Draw)

	hashVal := big.NewInt(0)
	hashVal.SetString(HashSHA(valBuf.String()), 10)

	blockCreatorsStake := currentState.Ledger[b.BakerID]
	percentOfTotalStake := float64(blockCreatorsStake) / float64(currentState.TotalSystemStake())
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
	buf.WriteString(b.BlockNonce.Nonce)
	buf.WriteString(strconv.Itoa(b.Slot))

	return Verify(buf.String(), b.Draw, b.BakerID)
}
