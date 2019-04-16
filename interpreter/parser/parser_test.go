package parser

import (
	"fmt"
	"github.com/nfk93/blockchain/interpreter/ast"
	"github.com/nfk93/blockchain/interpreter/lexer"
	"io/ioutil"
	"testing"
)

var testdir = "testcases/"

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

func TestParseIncrementStorage(t *testing.T) {
	testFile(t, testdir+"increment_storage")
}

func TestParseSimpleEntry(t *testing.T) {
	testFile(t, testdir+"simple_entry_parser")
}

func TestParseTypeDecl(t *testing.T) {
	testFile(t, testdir+"type_decl_parser")
}

func TestParseStructTypeDecl(t *testing.T) {
	testFile(t, testdir+"struct_decl_parser")
}

func TestParseList(t *testing.T) {
	testFile(t, testdir+"list_parser")
}

func TestParseBinop(t *testing.T) {
	testFile(t, testdir+"binopexp_parser")
}

func TestLetExp(t *testing.T) {
	testFile(t, testdir+"letexp")
}

func searchAstForErrorExps(t *testing.T, e ast.Exp) {
	switch e.(type) {
	case ast.TypeDecl:
	case ast.TopLevel:
		e := e.(ast.TopLevel)
		for _, v := range e.Roots {
			searchAstForErrorExps(t, v)
		}
	case ast.EntryExpression:
		e := e.(ast.EntryExpression)
		searchAstForErrorExps(t, e.Body)
	case ast.BinOpExp:
		e := e.(ast.BinOpExp)
		searchAstForErrorExps(t, e.Left)
		searchAstForErrorExps(t, e.Right)
	case ast.KeyLit, ast.BoolLit, ast.IntLit, ast.FloatLit, ast.KoinLit, ast.StringLit, ast.UnitLit:
	case ast.ListLit:
		e := e.(ast.ListLit)
		for _, v := range e.List {
			searchAstForErrorExps(t, v)
		}
	case ast.ListConcat:
		e := e.(ast.ListConcat)
		searchAstForErrorExps(t, e.Exp)
		searchAstForErrorExps(t, e.List)
	case ast.LetExp:
		e := e.(ast.LetExp)
		searchAstForErrorExps(t, e.DefExp)
		searchAstForErrorExps(t, e.InExp)
	default:
		t.Error("Encountered unknown expression:", e.String())
	}
}

/* func TestParseFundMe(t *testing.T) {
	parser := NewParser()
	parser.Parse(getLexer("../test_cases/fundme", t))
} */
