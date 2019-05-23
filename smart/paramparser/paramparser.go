package paramparser

import (
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"github.com/nfk93/blockchain/smart/paramparser/lexer"
	"github.com/nfk93/blockchain/smart/paramparser/parser"
)

func ParseParams(params string) (value.Value, error) {
	lex := lexer.NewLexer([]byte(params))
	parse := parser.NewParser()
	parsed, err := parse.Parse(lex)
	if err != nil {
		return value.UnitVal{}, err
	} else {
		return parsed.(value.Value), nil
	}
}
