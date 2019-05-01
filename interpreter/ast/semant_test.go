package ast

import (
	"fmt"
	"testing"
)

func TestBoolLit(t *testing.T) {
	exp := BoolLit{true}
	texp := AddTypes(exp)
	expected := TypedExp{exp, BoolType{}}
	checkTypeEquality(t, texp, expected)
}

func TestKeyLit(t *testing.T) {
	exp := KeyLit{"2aAad314"}
	texp := AddTypes(exp)
	expected := TypedExp{exp, KeyType{}}
	checkTypeEquality(t, texp, expected)
}

func TestIntLit(t *testing.T) {
	exp := IntLit{123}
	texp := AddTypes(exp)
	expected := TypedExp{exp, IntType{}}
	checkTypeEquality(t, texp, expected)
}

func TestFloatLit(t *testing.T) {
	exp := FloatLit{123.45}
	texp := AddTypes(exp)
	expected := TypedExp{exp, FloatType{}}
	checkTypeEquality(t, texp, expected)
}

func TestKoinLit(t *testing.T) {
	exp := KoinLit{123}
	texp := AddTypes(exp)
	expected := TypedExp{exp, KoinType{}}
	checkTypeEquality(t, texp, expected)
}

func TestStringLit(t *testing.T) {
	exp := StringLit{"ey"}
	texp := AddTypes(exp)
	expected := TypedExp{exp, StringType{}}
	checkTypeEquality(t, texp, expected)
}

func TestUnitLit(t *testing.T) {
	exp := UnitLit{}
	texp := AddTypes(exp)
	expected := TypedExp{exp, UnitType{}}
	checkTypeEquality(t, texp, expected)
}

func TestBinOp(t *testing.T) {
	exp := BinOpExp{IntLit{123}, EQ, IntLit{11}}
	texp := AddTypes(exp)
	fmt.Println(texp.String())
	ltyped := TypedExp{IntLit{123}, IntType{}}
	rtyped := TypedExp{IntLit{11}, IntType{}}
	expected := TypedExp{BinOpExp{ltyped, exp.Op, rtyped}, BoolType{}}
	checkTypeEquality(t, texp, expected)
	exp = BinOpExp{IntLit{123}, EQ, StringLit{"tis"}}
	texp = AddTypes(exp)
	fmt.Println(texp.String())
}

func TestTypeDecl(t *testing.T) {
	exp := TypeDecl{"test", NewIntType()}
	texp, _, tenv, _ := addTypes(exp, InitialVarEnv(), InitialTypeEnv(), InitialStructEnv())
	expected := TypedExp{exp, UnitType{}}
	checkTypeEquality(t, texp, expected)
	switch lookupType("test", tenv).(type) {
	case IntType:
		// do nothing
	default:
		t.Fail()
	}
	switch lookupType("nope", tenv).(type) {
	case nil:
		// do nothing
	default:
		t.Fail()
	}
}

func TestTranslateType(t *testing.T) {
	tenv := InitialTypeEnv()
	tenv = tenv.Set("myint", IntType{})
	tupletyp := TupleType{[]Type{IntType{}, IntType{}, NewDeclaredType("myint")}}

	actualtyp := translateType(tupletyp, tenv)
	actualtype := actualtyp.(TupleType)
	if len(actualtype.Typs) != 3 {
		t.Errorf("Tuple not correct length")
		for _, v := range actualtype.Typs {
			if v.Type() != actualtyp.Type() {
				t.Fail()
			}
		}
	}
}

func TestTypeDeclStruct(t *testing.T) {
	// check that struct with same names can't be used
	// check that types with same names can't be used
	// check that recursive types can't be used
	// TODO
}

func checkTypeEquality(t *testing.T, texp_, expected_ Exp) {
	texp := texp_.(TypedExp)
	expected := expected_.(TypedExp)
	if texp.Type != expected.Type {
		t.Errorf("Types not equal:\n"+
			"\tactual..: %s\n"+
			"\texpected: %s", texp_.String(), expected_.String())
	}
	e := texp.Exp
	e_ := expected.Exp
	switch e.(type) {
	case TypeDecl:
		if texp.Type.Type() != UNIT {
			t.Fail()
		}
	case TopLevel:
		e := e.(TopLevel)
		e_ := e_.(TopLevel)
		for i, v := range e.Roots {
			checkTypeEquality(t, v, e_.Roots[i])
		}
	case EntryExpression:
		e := e.(EntryExpression)
		e_ := e_.(EntryExpression)
		checkTypeEquality(t, e.Body, e_.Body)
	case BinOpExp:
		e := e.(BinOpExp)
		e_ := e_.(BinOpExp)
		checkTypeEquality(t, e.Left, e_.Left)
		checkTypeEquality(t, e.Right, e_.Right)
	case ListLit:
		e := e.(ListLit)
		e_ := e_.(ListLit)
		for i, v := range e.List {
			checkTypeEquality(t, v, e_.List[i])
		}
	case ListConcat:
		e := e.(ListConcat)
		e_ := e_.(ListConcat)
		checkTypeEquality(t, e.Exp, e_.Exp)
		checkTypeEquality(t, e.List, e_.List)
	case LetExp:
		e := e.(LetExp)
		e_ := e_.(LetExp)
		checkTypeEquality(t, e.DefExp, e_.DefExp)
		checkTypeEquality(t, e.InExp, e_.InExp)
	case TupleExp:
		e := e.(TupleExp)
		e_ := e_.(TupleExp)
		for i, v := range e.Exps {
			checkTypeEquality(t, v, e_.Exps[i])
		}
	case AnnoExp:
		e := e.(AnnoExp)
		e_ := e_.(AnnoExp)
		checkTypeEquality(t, e.Exp, e_.Exp)
	case IfThenElseExp:
		e := e.(IfThenElseExp)
		e_ := e_.(IfThenElseExp)
		checkTypeEquality(t, e.If, e_.If)
		checkTypeEquality(t, e.Then, e_.Then)
		checkTypeEquality(t, e.Else, e_.Else)
	case IfThenExp:
		e := e.(IfThenExp)
		e_ := e_.(IfThenExp)
		checkTypeEquality(t, e.If, e_.If)
		checkTypeEquality(t, e.Then, e_.Then)
	case ExpSeq:
		e := e.(ExpSeq)
		e_ := e_.(ExpSeq)
		checkTypeEquality(t, e.Left, e_.Left)
		checkTypeEquality(t, e.Right, e_.Right)
	case UpdateStructExp:
		e := e.(UpdateStructExp)
		e_ := e_.(UpdateStructExp)
		checkTypeEquality(t, e.Exp, e_.Exp)
	case StorageInitExp:
		e := e.(StorageInitExp)
		e_ := e_.(StorageInitExp)
		checkTypeEquality(t, e.Exp, e_.Exp)
	case StructLit:
		e := e.(StructLit)
		e_ := e_.(StructLit)
		for i, v := range e.Vals {
			checkTypeEquality(t, v, e_.Vals[i])
		}
	case CallExp:
		e := e.(CallExp)
		e_ := e_.(CallExp)
		checkTypeEquality(t, e.Exp1, e_.Exp1)
		checkTypeEquality(t, e.Exp2, e_.Exp2)
	case KeyLit, BoolLit, IntLit, FloatLit, KoinLit, StringLit, UnitLit, VarExp,
		ModuleLookupExp, LookupExp:
	default:
		t.Error("Encountered unknown expression:", e.String())
	}
}
