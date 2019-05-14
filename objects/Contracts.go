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
	Amount    int
	Gas       int
	Address   string
	Caller    crypto.PublicKey
	Nonce     string
	Signature string
}

type ContractInitialize struct {
	Owner   crypto.PublicKey
	Code    []byte
	Gas     int
	Prepaid int
}

type Operation interface {
}

type Parameter interface {
}

type Storage interface {
}

func CallAtConLayer(contract ContractCall) (bool, map[string]int, []ContractTransaction, int) {
	// TODO Dummies for smart contract layer functions
	_, pk := crypto.KeyGen(2048)
	conStake := map[string]int{}
	conStake["contract1"] = 200
	conTrans := []ContractTransaction{{pk, contract.Amount}}
	return true, conStake, conTrans, 2
}

func ExpireAtConLayer(expiredContracts []string) {
	// TODO Dummies for smart contract layer functions
}

func InitContractAtConLayer(code []byte, gas int) (string, int, bool) {
	// TODO Dummies for smart contract layer functions
	return "", 0, false
}

func (cc ContractCall) toString() string {
	var buf bytes.Buffer
	buf.WriteString(cc.Call)
	buf.WriteString(cc.Entry)
	buf.WriteString(strconv.Itoa(cc.Amount))
	buf.WriteString(strconv.Itoa(cc.Gas))
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
	buf.WriteString(strconv.Itoa(ci.Gas))
	buf.WriteString(strconv.Itoa(ci.Prepaid))

	return buf.String()
}
