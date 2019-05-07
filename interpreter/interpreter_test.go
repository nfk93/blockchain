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

func TestStructInStruct(t *testing.T) {
	testFileNoError(t, "test_cases/structinstruct_semant")
}

/* Interpreter tests */

func TestInterpretBinOps(t *testing.T) {

}

func TestCheckParams(t *testing.T) {
	stringval := StringVal{"ey"}
	stringtype := StringType{}
	if !checkParam(stringval, stringtype) {
		t.Fail()
	}
	intval1 := IntVal{1}
	if checkParam(intval1, stringtype) {
		t.Fail()
	}

	lst := make([]interface{}, 0)
	listVal1 := ListVal{append(lst, intval1)}
	listVal2 := ListVal{append(lst, int64(12))}
	if !checkParam(listVal1, ListType{IntType{}}) {
		t.Errorf("1")
	}
	if checkParam(listVal2, ListType{IntType{}}) {
		t.Errorf("2")
	}
	if checkParam(listVal1, ListType{StringType{}}) {
		t.Errorf("3")
	}
	tupleval1 := TupleValue{[]Value{TupleValue{[]Value{IntVal{1}, IntVal{2}}}, StringVal{"ey"}}}
	tupleval2 := TupleValue{[]Value{TupleValue{[]Value{IntVal{1}, IntVal{2}}}, IntVal{123}}}
	tupletyp1 := TupleType{[]Type{TupleType{[]Type{IntType{}, IntType{}}}, StringType{}}}
	if !checkParam(tupleval1, tupletyp1) {
		t.Errorf("4")
	}
	if checkParam(tupleval2, tupletyp1) {
		t.Errorf("5")
	}

	structval := StructVal{[]StructFieldVal{StructFieldVal{"a", IntVal{123}}, StructFieldVal{"b", StringVal{"eyyyyy"}}}}
	structtyp1 := StructType{[]StructField{StructField{"a", IntType{}}, StructField{"b", StringType{}}}}
	structtyp2 := StructType{[]StructField{StructField{"a", IntType{}}, StructField{"c", StringType{}}}}
	structtyp3 := StructType{[]StructField{StructField{"a", IntType{}}, StructField{"b", IntType{}}}}
	if !checkParam(structval, structtyp1) {
		t.Errorf("6")
	}
	if checkParam(structval, structtyp2) {
		t.Errorf("7")
	}
	if checkParam(structval, structtyp3) {
		t.Errorf("8")
	}

	optval1 := OptionVal{tupleval1, true}
	optval2 := OptionVal{UnitVal{}, false}
	opttyp1 := OptionType{tupletyp1}
	opttyp2 := OptionType{IntType{}}
	if !checkParam(optval1, opttyp1) {
		t.Errorf("9")
	}
	if !checkParam(optval2, opttyp1) {
		t.Errorf("10")
	}
	if !checkParam(optval2, opttyp2) {
		t.Errorf("11")
	}
	if checkParam(optval1, opttyp2) {
		t.Errorf("12")
	}
}

func TestBinopExp(t *testing.T) {
	testFileNoError(t, "test_cases/binopexp_semant")
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
		for _, v := range e.ExpList {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case UnOpExp:
		e := e.(UnOpExp)
		return checkForErrorTypes(e.Exp)
	case KeyLit, BoolLit, IntLit, KoinLit, StringLit, UnitLit, VarExp,
		ModuleLookupExp, LookupExp, NatLit:
		return true
	default:
		return false
	}
}
