package util

import "github.com/nfk93/blockchain/interpreter/token"

func ParseId(id interface{}) string {
	return string(id.(*token.Token).Lit)
}
