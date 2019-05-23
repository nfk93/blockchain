package smart

import (
	"github.com/nfk93/blockchain/smart/interpreter/ast"
)

type contract struct {
	code          string
	tabs          ast.TypedExp
	createdAtSlot uint64
}
