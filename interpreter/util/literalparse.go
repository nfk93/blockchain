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

func ParseFloat(float interface{}) float64 {
	f, _ := strconv.ParseFloat(string(float.(*token.Token).Lit), 64)
	return f
}

func ParseInt(i interface{}) int64 {
	integer, _ := strconv.ParseInt(string(i.(*token.Token).Lit), 10, 64)
	return integer
}
