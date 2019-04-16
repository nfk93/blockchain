package parser

import (
	"fmt"
	. "github.com/nfk93/blockchain/interpreter/ast"
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
	e := a.(Exp)
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

func TestAnnoExp(t *testing.T) {
	testFile(t, testdir+"annotated_exp")
}

func TestTupleExp(t *testing.T) {
	testFile(t, testdir+"tuple")
}

func TestVarExp(t *testing.T) {
	testFile(t, testdir+"varexp")
}

func TestIfExps(t *testing.T) {
	testFile(t, testdir+"ifexps")
}

func TestSeqExp(t *testing.T) {
	testFile(t, testdir+"sequence_expression")
}

func TestModuleLookup(t *testing.T) {
	testFile(t, testdir+"module_lookup")
}

func TestLookupExp(t *testing.T) {
	testFile(t, testdir+"lookupexp")
}

func TestUpdateStructExp(t *testing.T) {
	testFile(t, testdir+"update_struct_exp")
}

func TestInit1(t *testing.T) {
	testFile(t, testdir+"init1")
}

func TestInit2(t *testing.T) {
	testFile(t, testdir+"init2")
}

func TestStructLit(t *testing.T) {
	testFile(t, testdir+"struct_lit")
}

func TestFundMe(t *testing.T) {
	testFile(t, testdir+"fundme")
}

func searchAstForErrorExps(t *testing.T, e Exp) {
	switch e.(type) {
	case TypeDecl:
	case TopLevel:
		e := e.(TopLevel)
		for _, v := range e.Roots {
			searchAstForErrorExps(t, v)
		}
	case EntryExpression:
		e := e.(EntryExpression)
		searchAstForErrorExps(t, e.Body)
	case BinOpExp:
		e := e.(BinOpExp)
		searchAstForErrorExps(t, e.Left)
		searchAstForErrorExps(t, e.Right)
	case ListLit:
		e := e.(ListLit)
		for _, v := range e.List {
			searchAstForErrorExps(t, v)
		}
	case ListConcat:
		e := e.(ListConcat)
		searchAstForErrorExps(t, e.Exp)
		searchAstForErrorExps(t, e.List)
	case LetExp:
		e := e.(LetExp)
		searchAstForErrorExps(t, e.DefExp)
		searchAstForErrorExps(t, e.InExp)
	case TupleExp:
		e := e.(TupleExp)
		for _, v := range e.Exps {
			searchAstForErrorExps(t, v)
		}
	case AnnoExp:
		e := e.(AnnoExp)
		searchAstForErrorExps(t, e.Exp)
	case IfThenElseExp:
		e := e.(IfThenElseExp)
		searchAstForErrorExps(t, e.If)
		searchAstForErrorExps(t, e.Then)
		searchAstForErrorExps(t, e.Else)
	case IfThenExp:
		e := e.(IfThenExp)
		searchAstForErrorExps(t, e.If)
		searchAstForErrorExps(t, e.Then)
	case ExpSeq:
		e := e.(ExpSeq)
		searchAstForErrorExps(t, e.Left)
		searchAstForErrorExps(t, e.Right)
	case UpdateStructExp:
		e := e.(UpdateStructExp)
		searchAstForErrorExps(t, e.Exp)
	case InitExp:
		e := e.(InitExp)
		searchAstForErrorExps(t, e.Exp)
	case StructLit:
		e := e.(StructLit)
		for _, v := range e.Vals {
			searchAstForErrorExps(t, v)
		}
	case KeyLit, BoolLit, IntLit, FloatLit, KoinLit, StringLit, UnitLit, VarExp,
		ModuleLookupExp, LookupExp:
	default:
		t.Error("Encountered unknown expression:", e.String())
	}
}

/* func TestParseFundMe(t *testing.T) {
	parser := NewParser()
	parser.Parse(getLexer("../test_cases/fundme", t))
} */
