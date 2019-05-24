package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
	"time"
)

type Transaction struct {
	From      PublicKey
	To        PublicKey
	Amount    uint64
	ID        string
	Signature string
}

func (t Transaction) toString() string {
	var buf bytes.Buffer
	buf.WriteString(t.From.String())
	buf.WriteString(t.To.String())
	buf.WriteString(strconv.Itoa(int(t.Amount)))
	buf.WriteString(t.ID)
	buf.WriteString(t.Signature)
	return buf.String()
}

func (t Transaction) stringToSign() string {
	var buf bytes.Buffer
	buf.WriteString(t.From.String())
	buf.WriteString(t.To.String())
	buf.WriteString(strconv.Itoa(int(t.Amount)))
	buf.WriteString(t.ID)
	return buf.String()
}

func (t *Transaction) SignTransaction(sk SecretKey) {
	m := t.stringToSign()
	t.Signature = Sign(m, sk)
}

func (t *Transaction) VerifyTransaction() bool {
	return Verify(t.stringToSign(), t.Signature, t.From)
}

func CreateTransaction(from PublicKey, to PublicKey, amount uint64, id string, sk SecretKey) Transaction {
	t := Transaction{from, to, amount, from.Hash()[:10] + "-" + id + "-" + time.Now().String(), ""}
	t.SignTransaction(sk)
	return t
}
