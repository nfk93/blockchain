package util

import (
	"fmt"
	"github.com/nfk93/blockchain/smart/interpreter/token"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ParseId(id interface{}) string {
	return string(id.(*token.Token).Lit)
}

func ParseKey(key interface{}) (string, error) {
	return removePrefixAndCheckLen(key)
}

func ParseAddress(add interface{}) (string, error) {
	return removePrefixAndCheckLen(add)
}

func removePrefixAndCheckLen(str interface{}) (string, error) {
	str = string(str.(*token.Token).Lit)[3:]
	if utf8.RuneCountInString(str.(string)) != 32 {
		return "", fmt.Errorf("keys and addresses must have length 32")
	} else {
		return str.(string), nil
	}
}

func ParseInt(i interface{}) (int64, error) {
	integer, err := strconv.ParseInt(string(i.(*token.Token).Lit), 10, 64)
	return integer, err
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
