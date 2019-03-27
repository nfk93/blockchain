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

func (t Transaction) buildStringToSign() string {
	var buf bytes.Buffer
	buf.WriteString(t.From.String())
	buf.WriteString(t.To.String())
	buf.WriteString(strconv.Itoa(t.Amount))
	buf.WriteString(t.ID)
	return buf.String()
}

func (t *Transaction) signTransaction(sk SecretKey) {
	m := t.buildStringToSign()
	t.Signature = Sign(m, sk)
}

func (t *Transaction) VerifyTransaction() bool {
	return Verify(t.buildStringToSign(), t.Signature, t.From)
}

func CreateTransaction(from PublicKey, to PublicKey, amount int, id string, sk SecretKey) Transaction {
	t := Transaction{from, to, amount, id, ""}
	t.signTransaction(sk)
	return t
}
