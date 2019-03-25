package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
)

type BlockNonce struct {
	nonce     string
	signature string
}

func LeadershipElection() {

}

func verifyProof() {

}

func (bl BlockNonce) verifyBlockNonce(pk PublicKey) bool {

	return Verify(bl.nonce, bl.signature, pk)
}

func CreateNewBlockNonce(bnonce BlockNonce, slot int, sk SecretKey) BlockNonce {
	var buf bytes.Buffer
	buf.WriteString("NONCE")
	buf.WriteString(bnonce.nonce)
	buf.WriteString(strconv.Itoa(slot))

	newNonceString := buf.String()
	newNonce := HashSHA(newNonceString)
	signature := Sign(string(newNonce), sk)

	return BlockNonce{newNonce, signature}
}
