package smart

import (
	"github.com/nfk93/blockchain/smart/interpreter"
	"github.com/nfk93/blockchain/smart/interpreter/ast"
)

type Contract struct {
	code    string
	tabs    ast.TypedExp
	storage interpreter.Value
}