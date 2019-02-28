package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
)

type Transaction struct {
	From      PublicKey
	To        PublicKey
	Amount    int
	ID        string
	Signature string
}

func buildStringToSign(t Transaction) string {
	var buf bytes.Buffer
	buf.WriteString(t.From.String())
	buf.WriteString(t.To.String())
	buf.WriteString(strconv.Itoa(t.Amount))
	buf.WriteString(t.ID)
	return buf.String()
}

func SignTransaction(t Transaction, sk SecretKey) Transaction {
	m := buildStringToSign(t)
	s := Sign(m, sk)
	t.Signature = s
	return t
}

func VerifyTransaction(t Transaction, pk PublicKey) bool {
	return Verify(buildStringToSign(t), t.Signature, pk)
}
