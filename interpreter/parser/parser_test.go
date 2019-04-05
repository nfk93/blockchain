package parser

import (
	"fmt"
	"github.com/nfk93/blockchain/interpreter/ast"
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

func testFile(t *testing.T, testpath string) {
	parser := NewParser()
	a, err := parser.Parse(getLexer(testpath, t))
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	e := a.(ast.Exp)
	fmt.Println(e.String())
	searchAstForErrorExps(t, e)
}

func TestParseInttype(t *testing.T) {
	testFile(t, "../test_cases/inttype")
}

func TestParseTwoTypes(t *testing.T) {
	testFile(t, "../test_cases/twotypes")
}

func TestParseThreeTypes(t *testing.T) {
	testFile(t, "../test_cases/threetypes")
}

func TestParseBasicEntry(t *testing.T) {
	testFile(t, "../test_cases/increment_storage")
}

func searchAstForErrorExps(t *testing.T, e ast.Exp) {
	switch e.(type) {
	case ast.SimpleTypeDecl:
	case ast.TopLevel:
		e := e.(ast.TopLevel)
		for _, v := range e.Roots {
			searchAstForErrorExps(t, v)
		}
	default:
		t.Error("Encountered unknown expression:", e.String())
	}
}

/* func TestParseFundMe(t *testing.T) {
	parser := NewParser()
	parser.Parse(getLexer("../test_cases/fundme", t))
} */
