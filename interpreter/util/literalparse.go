package util

import (
	"fmt"
	"github.com/nfk93/blockchain/interpreter/token"
	"strconv"
	"strings"
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

func ParseKoin(kn interface{}) (uint64, error) {
	knstr := string(kn.(*token.Token).Lit)
	knstr = knstr[:len(knstr)-2]
	i := strings.Index(knstr, ".")
	if i > -1 {
		decimals := len(knstr) - i - 1
		if decimals > 5 {
			return 0, fmt.Errorf("too many decimals in koin literal, max is 5")
		} else {
			knstr = knstr + strings.Repeat("0", 5-decimals)
			val, _ := strconv.ParseUint(strings.Replace(knstr, ".", "", -1), 10, 64)
			return val, nil
		}
	} else {
		val, _ := strconv.ParseUint(knstr, 10, 64)
		return val * 100000, nil
	}
}

func ParseNat(i interface{}) uint64 {
	str := string(i.(*token.Token).Lit)
	val, _ := strconv.ParseUint(str[:(len(str)-1)], 10, 64)
	return val
}
