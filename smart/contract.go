package smart

import (
	"github.com/nfk93/blockchain/smart/interpreter/ast"
)

type contract struct {
	Code          string
	tabs          ast.TypedExp
	CreatedAtSlot uint64
}
