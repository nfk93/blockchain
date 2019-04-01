package parser

import (
	"github.com/nfk93/blockchain/interpreter/lexer"
	"io/ioutil"
	"testing"
)

func getLexer(filepath string, t *testing.T) *lexer.Lexer {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Error("Error reading testfile:", filepath)
	}
	return lexer.NewLexer(dat)
}

func TestParseInttype(t *testing.T) {
	parser := NewParser()
	parser.Parse(getLexer("../test_cases/inttype", t))
}

/* func TestParseFundMe(t *testing.T) {
	parser := NewParser()
	parser.Parse(getLexer("../test_cases/fundme", t))
} */
