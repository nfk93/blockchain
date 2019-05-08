package util

import (
	"github.com/nfk93/blockchain/interpreter/token"
	"strconv"
)

func ParseId(id interface{}) string {
	return string(id.(*token.Token).Lit)
}

func ParseKey(key interface{}) string {
	return string(key.(*token.Token).Lit)[3:]
}

func ParseAddress(add interface{}) string {
	return string(add.(*token.Token).Lit)[3:]
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
	s := string(str.(*token.Token).Lit)[1:]
	s = s[:(len(s) - 1)]
	return s
}

func ParseKoin(kn interface{}) float64 {
	knstr := string(kn.(*token.Token).Lit)
	knval, _ := strconv.ParseFloat(knstr[:len(knstr)-2], 64)
	return knval
}

func ParseNat(i interface{}) uint64 {
	str := string(i.(*token.Token).Lit)
	val, _ := strconv.ParseUint(str[:(len(str)-1)], 10, 64)
	return val
}
