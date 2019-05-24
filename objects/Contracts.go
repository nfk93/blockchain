package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"strconv"
	"time"
)

type ContractCall struct {
	Call      string
	Entry     string
	Params    string
	Amount    uint64
	Gas       uint64
	Address   string
	Caller    PublicKey
	Nonce     string
	Signature string
}

type ContractInitialize struct {
	Owner        PublicKey
	Code         []byte
	Gas          uint64
	Prepaid      uint64
	StorageLimit uint64
	Nonce        string
	Signature    string
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

func (cc ContractCall) stringToSign() string {
	var buf bytes.Buffer
	buf.WriteString(cc.Call)
	buf.WriteString(cc.Entry)
	buf.WriteString(strconv.Itoa(int(cc.Amount)))
	buf.WriteString(strconv.Itoa(int(cc.Gas)))
	buf.WriteString(cc.Address)
	buf.WriteString(cc.Caller.String())
	buf.WriteString(cc.Nonce)
	return buf.String()
}

func (ci ContractInitialize) toString() string {
	var buf bytes.Buffer
	buf.WriteString(ci.Owner.String())
	buf.Write(ci.Code)
	buf.WriteString(strconv.Itoa(int(ci.Gas)))
	buf.WriteString(strconv.Itoa(int(ci.Prepaid)))
	buf.WriteString(ci.Nonce)
	buf.WriteString(ci.Signature)
	return buf.String()
}

func (ci ContractInitialize) stringToSign() string {
	var buf bytes.Buffer
	buf.WriteString(ci.Owner.String())
	buf.Write(ci.Code)
	buf.WriteString(strconv.Itoa(int(ci.Gas)))
	buf.WriteString(strconv.Itoa(int(ci.Prepaid)))
	buf.WriteString(ci.Nonce)
	return buf.String()
}

func (cc *ContractCall) Sign(sk SecretKey) {
	m := cc.stringToSign()
	cc.Signature = Sign(m, sk)
}

func (cc *ContractCall) Verify() bool {
	return Verify(cc.stringToSign(), cc.Signature, cc.Caller)
}

func (ci *ContractInitialize) Sign(sk SecretKey) {
	m := ci.stringToSign()
	ci.Signature = Sign(m, sk)
}

func (ci *ContractInitialize) Verify() bool {
	return Verify(ci.stringToSign(), ci.Signature, ci.Owner)
}

func CreateContractCall(call string, entry string, params string, amount uint64, gas uint64, address string, caller PublicKey, sk SecretKey) ContractCall {
	cc := ContractCall{call, entry, params, amount, gas, address, caller, caller.Hash()[:10] + "-" + time.Now().String(), ""}
	cc.Sign(sk)
	return cc
}

func CreateContractInit(owner PublicKey, code []byte, gas uint64, prepaid uint64, storageLimit uint64, sk SecretKey) ContractInitialize {
	ci := ContractInitialize{owner, code, gas, prepaid, storageLimit, owner.Hash()[:10] + "-" + time.Now().String(), ""}
	ci.Sign(sk)
	return ci
}
