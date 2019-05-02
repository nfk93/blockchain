package interpreter

import (
	. "github.com/nfk93/blockchain/interpreter/ast"
	"github.com/nfk93/blockchain/interpreter/lexer"
	"github.com/nfk93/blockchain/interpreter/parser"
	"io/ioutil"
	"os"
	"testing"
)

func TestTopLevel(t *testing.T) {
	testFileNoError(t, "test_cases/toplevel_semant")
}

func TestTopLevelError1(t *testing.T) {
	testFileError(t, "test_cases/toplevel1_semant")
}

func TestTopLevelError2(t *testing.T) {
	testFileError(t, "test_cases/toplevel2_semant")
}

func TestTopLevelError3(t *testing.T) {
	testFileError(t, "test_cases/toplevel3_semant")
}

func TestTopLevelError4(t *testing.T) {
	testFileError(t, "test_cases/toplevel4_semant")
}

func TestConcatList(t *testing.T) {
	testFileNoError(t, "test_cases/concatlist_semant")
}

func TestExpSeq(t *testing.T) {
	testFileNoError(t, "test_cases/expseq_semant")
}

func TestFundme(t *testing.T) {
	testFileNoError(t, os.Getenv("GOPATH")+"/src/github.com/nfk93/blockchain/usecases/fundme")
}

func TestPattern(t *testing.T) {
	testFileNoError(t, "test_cases/patterns_semant")
}

func TestPatternError1(t *testing.T) {
	testFileError(t, "test_cases/patterns1_semant")
}

func TestPatternError2(t *testing.T) {
	testFileError(t, "test_cases/patterns2_semant")
}

func TestPatternError3(t *testing.T) {
	testFileError(t, "test_cases/patterns3_semant")
}

func TestAnnoExp(t *testing.T) {
	testFileNoError(t, "test_cases/annoexp_semant")
}

func TestLetExp(t *testing.T) {
	testFileNoError(t, "test_cases/letexp_semant")
}

func TestLetExpFail(t *testing.T) {
	testFileError(t, "test_cases/letexp1_semant")
}

func TestTuples(t *testing.T) {
	testFileNoError(t, "test_cases/tuple_semant")
}

func TestIfThenElse(t *testing.T) {
	testFileNoError(t, "test_cases/ifthenelse_semant")
}

func TestIfThenElseError1(t *testing.T) {
	testFileError(t, "test_cases/ifthenelse1_semant")
}

func TestStructLit(t *testing.T) {
	testFileNoError(t, "test_cases/structlit_semant")
}

func TestLookupExp(t *testing.T) {
	testFileNoError(t, "test_cases/lookupexp_semant")
}

func TestLookupExpFail(t *testing.T) {
	testFileError(t, "test_cases/lookupexp1_semant")
}

func TestUpdateStructExp(t *testing.T) {
	testFileNoError(t, "test_cases/updatestructexp_semant")
}

func TestUpdateStructExpFail(t *testing.T) {
	testFileError(t, "test_cases/updatestructexp1_semant")
}

func TestCallExp(t *testing.T) {
	testFileNoError(t, "test_cases/callexp_semant")
}

/* Helper functions */

func testFileNoError(t *testing.T, testpath string) {
	testFile(t, testpath, false)
}

func testFileError(t *testing.T, testpath string) {
	testFile(t, testpath, true)
}

func testFile(t *testing.T, testpath string, shouldFail bool) {
	dat, err := ioutil.ReadFile(testpath)
	if err != nil {
		t.Error("Error reading testfile:", testpath)
	}
	lex := lexer.NewLexer(dat)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		t.Errorf("parse error: " + err.Error())
	} else {
		parsed := par.(Exp)
		typed := AddTypes(parsed)
		print("\n" + typed.String() + "\n")
		noErrors := checkForErrorTypes(typed)
		if shouldFail {
			if noErrors {
				t.Errorf("Didn't find any noErrors")
			}
		} else {
			if !noErrors {
				t.Errorf("Found ErrorType")
			}
		}
	}
}

func checkForErrorTypes(texp_ Exp) bool {
	switch texp_.(type) {
	case TypedExp:
		break
	default:
		return false
	}
	texp := texp_.(TypedExp)
	if texp.Type.Type() == ERROR || texp.Type.Type() == NOTIMPLEMENTED {
		return false
	}
	e := texp.Exp
	switch e.(type) {
	case TypeDecl:
		return true
	case TopLevel:
		e := e.(TopLevel)
		for _, v := range e.Roots {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case EntryExpression:
		e := e.(EntryExpression)
		return checkForErrorTypes(e.Body)
	case BinOpExp:
		e := e.(BinOpExp)
		return checkForErrorTypes(e.Left) && checkForErrorTypes(e.Right)
	case ListLit:
		e := e.(ListLit)
		for _, v := range e.List {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case ListConcat:
		e := e.(ListConcat)
		return checkForErrorTypes(e.Exp) && checkForErrorTypes(e.List)
	case LetExp:
		e := e.(LetExp)
		return checkForErrorTypes(e.DefExp) && checkForErrorTypes(e.InExp)
	case TupleExp:
		e := e.(TupleExp)
		for _, v := range e.Exps {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case AnnoExp:
		e := e.(AnnoExp)
		return checkForErrorTypes(e.Exp)
	case IfThenElseExp:
		e := e.(IfThenElseExp)
		return checkForErrorTypes(e.If) && checkForErrorTypes(e.Then) && checkForErrorTypes(e.Else)
	case IfThenExp:
		e := e.(IfThenExp)
		return checkForErrorTypes(e.If) && checkForErrorTypes(e.Then)
	case ExpSeq:
		e := e.(ExpSeq)
		return checkForErrorTypes(e.Left) && checkForErrorTypes(e.Right)
	case UpdateStructExp:
		e := e.(UpdateStructExp)
		return checkForErrorTypes(e.Exp)
	case StorageInitExp:
		e := e.(StorageInitExp)
		return checkForErrorTypes(e.Exp)
	case StructLit:
		e := e.(StructLit)
		for _, v := range e.Vals {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case CallExp:
		e := e.(CallExp)
		return checkForErrorTypes(e.Exp1) && checkForErrorTypes(e.Exp2)
	case UnOpExp:
		e := e.(UnOpExp)
		return checkForErrorTypes(e.Exp)
	case KeyLit, BoolLit, IntLit, FloatLit, KoinLit, StringLit, UnitLit, VarExp,
		ModuleLookupExp, LookupExp, NatLit:
		return true
	default:
		return false
	}
}
