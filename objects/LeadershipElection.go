package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
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

	if intermediateBlock.validateDraw(stake, hardness) {
		return true, draw
	}
	return false, ""

}
