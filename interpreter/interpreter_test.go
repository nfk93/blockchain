package interpreter

import (
	"github.com/nfk93/blockchain/interpreter/ast"
	"github.com/nfk93/blockchain/interpreter/lexer"
	"github.com/nfk93/blockchain/interpreter/parser"
	"io/ioutil"
	"testing"
)

func TestTopLevel(t *testing.T) {
	testFile(t, "test_cases/toplevel_semant")
}

func testFile(t *testing.T, testpath string) {
	dat, err := ioutil.ReadFile(testpath)
	if err != nil {
		t.Error("Error reading testfile:", testpath)
	}
	lex := lexer.NewLexer(dat)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		panic(err)
	}
	parsed := par.(ast.Exp)
	print("\n" + ast.AddTypes(parsed).String() + "\n")
}
