package consensus

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	"math"
	"math/big"
	"strconv"
)

func CalculateDraw(hardness float64, sk SecretKey, pk PublicKey, slot uint64, fd FinalData) (bool, string) {
	// Creates the draw signature
	drawString := getDrawString(slot, fd.leadershipNonce)
	draw := Sign(drawString, sk)

	if CheckIfWinner(draw, slot, pk, hardness, fd) {
		return true, draw
	} else {
		return false, ""
	}
}

func ValidateDraw(slot uint64, draw string, key PublicKey, fd FinalData, hardness float64) bool {
	if !validateDrawSignature(key, slot, draw, fd.leadershipNonce) {
		fmt.Println("Draw signature didn't validate...")
		return false
	}

	// Calculates the threshold
	isWinner := CheckIfWinner(draw, slot, key, hardness, fd)
	return isWinner
}

func CheckIfWinner(draw string, slot uint64, key PublicKey, hardness float64, fd FinalData) bool {
	amount := fd.stake[key.Hash()]
	stake := float64(amount) / float64(fd.totalstake)
	phiFunc := float64(1) - math.Pow(float64(1)-hardness, stake)
	multFactor := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(256)), nil)
	threshold := new(big.Int)
	new(big.Float).Mul(big.NewFloat(float64(phiFunc)), new(big.Float).SetInt(multFactor)).Int(threshold)
	drawVal := calculateDrawValue(slot, draw, fd.leadershipNonce)
	if drawVal.Cmp(threshold) == -1 {
		return true
	}
	return false
}

func getDrawString(slot uint64, leadershipNonce string) string {
	var drawBuf bytes.Buffer
	drawBuf.WriteString("LEADERSHIP_ELECTION")
	drawBuf.WriteString(leadershipNonce)
	drawBuf.WriteString(strconv.Itoa(int(slot)))
	return drawBuf.String()
}

func calculateDrawValue(slot uint64, draw string, leadershipNonce string) *big.Int {
	// Calculates the draw value
	var valBuf bytes.Buffer
	valBuf.WriteString("LEADERSHIP_ELECTION")
	valBuf.WriteString(leadershipNonce)
	valBuf.WriteString(strconv.Itoa(int(slot)))
	valBuf.WriteString(draw)
	hashString := HashSHA(valBuf.String())
	hashVal := big.NewInt(0)
	hashVal.SetString(hashString, 16)
	return hashVal
}

func validateDrawSignature(key PublicKey, slot uint64, drawSignature, leadershipNonce string) bool {
	return Verify(getDrawString(slot, leadershipNonce), drawSignature, key)
}
