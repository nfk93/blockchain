package util

import (
	"github.com/nfk93/blockchain/interpreter/token"
	"strconv"
)

func ParseId(id interface{}) string {
	return string(id.(*token.Token).Lit)
}

func ParseKey(key interface{}) []byte {
	return key.(*token.Token).Lit
}

func ParseAddress(add interface{}) []byte {
	return add.(*token.Token).Lit
}

func ParseFloat(float interface{}) float64 {
	f, _ := strconv.ParseFloat(string(float.(*token.Token).Lit), 64)
	return f
}

func ParseInt(i interface{}) int64 {
	integer, _ := strconv.ParseInt(string(i.(*token.Token).Lit), 10, 64)
	return integer
}

func ParseString(str interface{}) string {
	return string(str.(*token.Token).Lit)
}

func ParseKoin(kn interface{}) float64 {
	knstr := string(kn.(*token.Token).Lit)
	knval, _ := strconv.ParseFloat(knstr[:len(knstr)-2], 64)
	return knval
}

func ParseNat(i interface{}) uint64 {
	uint, _ := strconv.ParseUint(string(i.(*token.Token).Lit), 10, 64)
	return uint
}
