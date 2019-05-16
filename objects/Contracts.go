package objects

import (
	"bytes"
	"github.com/nfk93/blockchain/crypto"
	"strconv"
)

type ContractCall struct {
	Call      string
	Entry     string
	Params    Parameter
	Amount    uint64
	Gas       uint64
	Address   string
	Caller    crypto.PublicKey
	Nonce     string
	Signature string
}

type ContractInitialize struct {
	Owner   crypto.PublicKey
	Code    []byte
	Gas     uint64
	Prepaid uint64
}

type Operation interface {
}

type Parameter interface {
}

type Storage interface {
}

func (cc ContractCall) toString() string {
	var buf bytes.Buffer
	buf.WriteString(cc.Call)
	buf.WriteString(cc.Entry)
	buf.WriteString(strconv.Itoa(int(cc.Amount)))
	buf.WriteString(strconv.Itoa(int(cc.Gas)))
	buf.WriteString(cc.Address)
	buf.WriteString(cc.Caller.String())
	buf.WriteString(cc.Nonce)
	buf.WriteString(cc.Signature)
	return buf.String()
}

func (ci ContractInitialize) toString() string {
	var buf bytes.Buffer
	buf.WriteString(ci.Owner.String())
	buf.Write(ci.Code)
	buf.WriteString(strconv.Itoa(int(ci.Gas)))
	buf.WriteString(strconv.Itoa(int(ci.Prepaid)))

	return buf.String()
}
