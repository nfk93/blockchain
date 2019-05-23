package interpreter

import (
	"fmt"
	. "github.com/nfk93/blockchain/smart/interpreter/ast"
	"github.com/nfk93/blockchain/smart/interpreter/lexer"
	"github.com/nfk93/blockchain/smart/interpreter/parser"
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestTopLevel(t *testing.T) {
	testFileNoError(t, "test_cases/toplevel_semant")
}

func TestTopLevelError1(t *testing.T) {
	testFileError(t, "test_cases/toplevel1_semant")
}

func TestBinopExp(t *testing.T) {
	testFileNoError(t, "test_cases/binopexp_semant")
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

func TestUpdateStruct(t *testing.T) {
	testFileNoError(t, "test_cases/updatestruct_interp")
}

/* Interpreter tests */

var emptyContractList = make([]string, 0)

func TestIntConstant(t *testing.T) {
	testpath := "test_cases/constants/int"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.IntVal{13}, 0, 0,
		99999999999)
	switch sto.(type) {
	case value.IntVal:
		if sto.(value.IntVal).Value != 15 {
			t.Errorf("storage has unexpected value of %d", sto.(value.IntVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestAddressConstant(t *testing.T) {
	testpath := "test_cases/constants/address"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.AddressVal{"123123aA"}, 0,
		0, 9999999999)
	switch sto.(type) {
	case value.AddressVal:
		if sto.(value.AddressVal).Value != "3132141abba3132141abba3132141abb" {
			t.Errorf("storage has unexpected value of %s", sto.(value.AddressVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type but of type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestBoolConstant(t *testing.T) {
	testpath := "test_cases/constants/bool"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.BoolVal{true}, 0,
		0, 999999999999999999)
	switch sto.(type) {
	case value.BoolVal:
		if sto.(value.BoolVal).Value != false {
			t.Errorf("storage has unexpected value of %t", sto.(value.BoolVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestDeclaredConstant(t *testing.T) {
	testpath := "test_cases/constants/declared"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main",
		value.TupleVal{[]value.Value{value.IntVal{123}, value.TupleVal{[]value.Value{value.IntVal{2}, value.StringVal{"serser"}}}}},
		0, 0, 9999999)
	switch sto.(type) {
	case value.TupleVal:
		sto := sto.(value.TupleVal)
		if sto.Values[0].(value.IntVal).Value != 4 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[1].(value.TupleVal).Values[0].(value.IntVal).Value != 5 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[1].(value.TupleVal).Values[1].(value.StringVal).Value != "bye" {
			t.Errorf("storage has unexpected value of %s",
				sto.Values[1].(value.TupleVal).Values[1].(value.StringVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestKeyConstant(t *testing.T) {
	testpath := "test_cases/constants/key"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.KeyVal{"1212Ddd"}, 0,
		0, 999999999)
	switch sto.(type) {
	case value.KeyVal:
		if sto.(value.KeyVal).Value != "aaffaafaaffaafaaffaafaaffaafaaff" {
			t.Errorf("storage has unexpected value of %s", sto.(value.KeyVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestKoinConstant(t *testing.T) {
	testpath := "test_cases/constants/koin"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.KoinVal{uint64(110000)}, 0,
		0, 99999999999)
	switch sto.(type) {
	case value.KoinVal:
		if sto.(value.KoinVal).Value != 13355000 {
			t.Errorf("storage has unexpected value of %d", sto.(value.KoinVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestListConstant(t *testing.T) {
	testpath := "test_cases/constants/list"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.ListVal{[]value.Value{value.IntVal{2}}},
		0, 0, 9999999999)
	switch sto.(type) {
	case value.ListVal:
		sto := sto.(value.ListVal)
		if sto.Values[0].(value.IntVal).Value != 7 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[1].(value.IntVal).Value != 9 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[2].(value.IntVal).Value != 13 {
			t.Errorf("storage has unexpected value of %d at index 2", sto.Values[2].(value.IntVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestNatConstant(t *testing.T) {
	testpath := "test_cases/constants/nat"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.NatVal{13}, 0, 0,
		99999999999)
	switch sto.(type) {
	case value.NatVal:
		if sto.(value.NatVal).Value != 117 {
			t.Errorf("storage has unexpected value of %d", sto.(value.NatVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestStringConstant(t *testing.T) {
	testpath := "test_cases/constants/string"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.StringVal{"eymom"}, 0,
		0, 99999999999)
	switch sto.(type) {
	case value.StringVal:
		if !(sto.(value.StringVal).Value == "dank") {
			t.Errorf("storage has unexpected value of %s", sto.(value.StringVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestUnitConstant(t *testing.T) {
	testpath := "test_cases/constants/unit"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.UnitVal{}, 0,
		0, 99999999999)
	switch sto.(type) {
	case value.UnitVal:
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestStructConstant(t *testing.T) {
	testpath := "test_cases/constants/struct"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	storageinit := createStruct()
	storageinit.Field["a"] = value.IntVal{1213}
	storageinit.Field["b"] = value.TupleVal{[]value.Value{value.IntVal{5}, value.IntVal{6}}}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", storageinit, 0,
		0, 99999999999)
	switch sto.(type) {
	case value.StructVal:
		sto, oksto := sto.(value.StructVal)
		if oksto {
			a, oka := sto.Field["a"].(value.IntVal)
			if !oka || a.Value != 1111 {
				t.Errorf("storage.a has unexpected value of %d", sto.Field["a"])
			}
			b, okb := sto.Field["b"].(value.TupleVal)
			if !okb {
				t.Errorf("storage.b has unexpected type")
			} else {
				tup1, ok1 := b.Values[0].(value.IntVal)
				tup2, ok2 := b.Values[1].(value.IntVal)
				if !ok1 || !ok2 || tup1.Value != 4 || tup2.Value != 5 {
					t.Errorf("storage.b has wrong value. It is %d", b)
				}
			}
		} else {
			t.Error("return value isn't of correct type")
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestCurrentModule(t *testing.T) {
	testpath := "test_cases/currentbalance_interp"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.KoinVal{1000000}, 0,
		0, 999999999)
	switch sto.(type) {
	case value.KoinVal:
		sto := sto.(value.KoinVal)
		if sto.Value != 0 {
			t.Errorf("return value is %d, expected 0.0", sto.Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestCheckParams(t *testing.T) {
	stringval := value.StringVal{"ey"}
	stringtype := StringType{}
	if !checkParam(stringval, stringtype) {
		t.Fail()
	}
	intval1 := value.IntVal{1}
	if checkParam(intval1, stringtype) {
		t.Fail()
	}

	lst := make([]value.Value, 0)
	listVal1 := value.ListVal{append(lst, intval1)}
	listVal2 := value.ListVal{append(lst, value.IntVal{int64(12)})}
	if !checkParam(listVal1, ListType{IntType{}}) {
		t.Errorf("1")
	}
	if !checkParam(listVal2, ListType{IntType{}}) {
		t.Errorf("2")
	}
	if checkParam(listVal1, ListType{StringType{}}) {
		t.Errorf("3")
	}
	tupleval1 := value.TupleVal{[]value.Value{value.TupleVal{[]value.Value{value.IntVal{1}, value.IntVal{2}}}, value.StringVal{"ey"}}}
	tupleval2 := value.TupleVal{[]value.Value{value.TupleVal{[]value.Value{value.IntVal{1}, value.IntVal{2}}}, value.IntVal{123}}}
	tupletyp1 := TupleType{[]Type{TupleType{[]Type{IntType{}, IntType{}}}, StringType{}}}
	if !checkParam(tupleval1, tupletyp1) {
		t.Errorf("4")
	}
	if checkParam(tupleval2, tupletyp1) {
		t.Errorf("5")
	}

	structval := createStruct()
	structval.Field["a"] = value.IntVal{123}
	structval.Field["b"] = value.StringVal{"eytyyyyy"}
	structtyp1 := StructType{[]StructField{StructField{"a", IntType{}}, StructField{"b", StringType{}}}}
	/*	structtyp2 := StructType{[]StructField{StructField{"a", IntType{}}, StructField{"c", StringType{}}}}
		structtyp3 := StructType{[]StructField{StructField{"a", IntType{}}, StructField{"b", IntType{}}}}
	*/if !checkParam(structval, structtyp1) {
		t.Errorf("6")
	}
	/*if checkParam(structval, structtyp2) {
		t.Errorf("7")
	}
	if checkParam(structval, structtyp3) {
		t.Errorf("8")
	} */

	optval1 := value.OptionVal{tupleval1, true}
	optval2 := value.OptionVal{value.UnitVal{}, false}
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

func TestInterpretUpdateStruct(t *testing.T) {
	testpath := "test_cases/updatestruct_interp"
	texp, err := getTypedAST(t, testpath)
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	params := value.TupleVal{[]value.Value{value.IntVal{173}, value.StringVal{"newinner"}, value.StringVal{"newa"}}}
	innermost := createStruct()
	innermost.Field["buried"] = value.IntVal{12}
	innermost.Field["deep"] = value.StringVal{"very deep"}
	inner := createStruct()
	inner.Field["id"] = value.StringVal{"innerid"}
	inner.Field["innermost"] = innermost
	storage := createStruct()
	storage.Field["a"] = value.StringVal{"test"}
	storage.Field["b"] = inner

	oplist, sto, _, _ := InterpretContractCall(texp, params, "main", storage, 0,
		0, 999999999999)
	switch sto.(type) {
	case value.StructVal:
		sto := sto.(value.StructVal)
		a, ok := sto.Field["a"].(value.StringVal)
		if !ok || a.Value != "newa" {
			t.Errorf("sto.a has unexpected value of %s", a)
		}
		b, ok := sto.Field["b"].(value.StructVal)
		if !ok {
			t.Errorf("sto.b has unexpected type. It is %s", b)
		} else {
			bid, ok := b.Field["id"].(value.StringVal)
			if !ok || bid.Value != "newinner" {
				t.Errorf("sto.b.id has unexpected value of %s", bid)
			}
			binnermost, ok := b.Field["innermost"].(value.StructVal)
			if !ok {
				t.Errorf("sto.b.innermost has unexpected type. It is %s", binnermost)
			} else {
				buried, ok := binnermost.Field["buried"].(value.IntVal)
				if !ok || buried.Value != 173+12 {
					t.Errorf("sto.b.innermost.buried has unexpected value of %d", buried)
				}
				deep, ok := binnermost.Field["deep"].(value.StringVal)
				if !ok || deep.Value != "very deep" {
					t.Errorf("sto.b.innermost.deep has unexpected value of %s", deep)
				}
			}
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestInterpretBinOps(t *testing.T) {
	texp, err := getTypedAST(t, "test_cases/binop_interp")
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	params := value.TupleVal{[]value.Value{value.IntVal{13}, value.IntVal{17}}}
	storage := value.IntVal{19}
	oplist, sto, _, _ := InterpretContractCall(texp, params, "main", storage, 0,
		0, 999999999999)
	switch sto.(type) {
	case value.IntVal:
		if sto.(value.IntVal).Value != 13+17+19 {
			t.Errorf("storage has unexpected value of %d", sto.(value.IntVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

}

func TestIntDivision(t *testing.T) {
	texp, err := getTypedAST(t, "test_cases/division_interp")
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	params := value.TupleVal{[]value.Value{value.IntVal{13}, value.IntVal{17}}}
	storage := value.TupleVal{[]value.Value{value.IntVal{19}, value.NatVal{0}}}
	oplist, sto, _, _ := InterpretContractCall(texp, params, "main", storage, 0,
		0, 100000)

	switch sto.(type) {
	case value.TupleVal:
		sto := sto.(value.TupleVal)
		if len(sto.Values) != 2 {
			t.Errorf("storage tuple has unexpeted length of %d", len(sto.Values))
		}
		switch sto.Values[0].(type) {
		case value.IntVal:
		default:
			t.Errorf("First value of returned storage tuple is wrong type")
		}
		switch sto.Values[1].(type) {
		case value.NatVal:
		default:
			t.Errorf("Second value of returned storage tuple is wrong type")
		}
		val1 := sto.Values[0].(value.IntVal).Value
		val2 := sto.Values[1].(value.NatVal).Value
		if val1 != 0 || val2 != 10 {
			t.Errorf("storage has unexpected value of (%d, %d)", val1, val2)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

	params = value.TupleVal{[]value.Value{value.IntVal{0}, value.IntVal{0}}}
	oplist, sto, _, _ = InterpretContractCall(texp, params, "main", storage, 0,
		0, 100000)
	if len(oplist) != 1 {
		t.Errorf("oplist is len %d should be 1", len(oplist))
	}

	failwithOp, ok := oplist[0].(value.FailWith)
	if !ok {
		fmt.Println(failwithOp.Msg)
		t.Errorf("unexpected returned operation")
	}
}

func TestKoinDivision(t *testing.T) {
	texp, err := getTypedAST(t, "test_cases/koindivision_interp")
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}

	params := value.TupleVal{[]value.Value{value.KoinVal{NatToKoin(5)}, value.KoinVal{NatToKoin(2)}}}
	storage := value.TupleVal{[]value.Value{value.NatVal{10}, value.KoinVal{NatToKoin(2)}}}
	oplist, sto, _, _ := InterpretContractCall(texp, params, "main", storage, 0,
		0, 100000)

	switch sto.(type) {
	case value.TupleVal:
		sto := sto.(value.TupleVal)
		if len(sto.Values) != 2 {
			t.Errorf("storage tuple has unexpeted length of %d", len(sto.Values))
		}
		switch sto.Values[0].(type) {
		case value.NatVal:
		default:
			t.Errorf("First value of returned storage tuple is wrong type")
		}
		switch sto.Values[1].(type) {
		case value.KoinVal:
		default:
			t.Errorf("Second value of returned storage tuple is wrong type")
		}
		val1 := sto.Values[0].(value.NatVal).Value
		val2 := sto.Values[1].(value.KoinVal).Value
		if val1 != 2 || val2 != NatToKoin(1) {
			t.Errorf("storage has unexpected value of (%d, %d)", val1, val2)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

}

func TestInterpretFailwithGas(t *testing.T) {
	texp, err := getTypedAST(t, "test_cases/currentfailwith_interp")
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	fmt.Println(texp.String())
	unitval := value.UnitVal{}
	initgas := NatToKoin(100)
	_, _, _, gas := InterpretContractCall(texp, unitval, "main", value.KoinVal{1000000}, 0,
		0, initgas)
	if gas != initgas-6000 {
		t.Errorf("remaining gas is %d, expected %d", gas, initgas-6000)
	}
}

func TestInterpretFailwith(t *testing.T) {
	texp, err := getTypedAST(t, "test_cases/currentfailwith_interp")
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	unitval := value.UnitVal{}
	oplist, sto, _, _ := InterpretContractCall(texp, unitval, "main", value.KoinVal{1000000}, 0,
		0, 99999999999)
	switch sto.(type) {
	case value.KoinVal:
		sto := sto.(value.KoinVal)
		if sto.Value != 1000000 {
			t.Errorf("return value is %d, expected 1000000", sto.Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 1 {
		t.Errorf("oplist is empty, should contain failwith oper")
	}
}

func TestInterpretLetexps(t *testing.T) {
	texp, err := getTypedAST(t, "test_cases/letexps_interp")
	if err != nil {
		t.Errorf("Semant error: %s", err.Error())
		return
	}
	params := value.TupleVal{[]value.Value{value.IntVal{7}, value.StringVal{"not imporatnt"}, value.NatVal{13}}}
	storage := value.IntVal{19}
	oplist, sto, _, _ := InterpretContractCall(texp, params, "main", storage, 0,
		0, 999999999999)
	switch sto.(type) {
	case value.IntVal:
		if sto.(value.IntVal).Value != 19-(15+7+13) {
			t.Errorf("storage has unexpected value of %d", sto.(value.IntVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}
}

func TestInitStorage(t *testing.T) {
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	_, init, _, err := InitiateContract(dat, 999999999999999)
	if err != nil {
		t.Error(err)
		return
	}

	switch init.(type) {
	case value.StructVal:
		init := init.(value.StructVal)
		owner, ok := init.Field["owner"].(value.KeyVal)
		if !ok || owner.Value != "1234567890abcdef1234567890abcdef" {
			t.Errorf("init storage has wrong value in field %s", "owner")
		}
		funding_goal, ok := init.Field["funding_goal"].(value.KoinVal)
		if !ok || funding_goal.Value != 1100000 {
			t.Errorf("init storage has wrong value in field %s. expected: 1100000, actual: %d",
				"funding_goal", funding_goal.Value)
		}
		amount_raised, ok := init.Field["amount_raised"].(value.KoinVal)
		if !ok || amount_raised.Value != 0 {
			t.Errorf("init storage has wrong value in field %s", "amount_raised")
		}
	}
}

func TestRunFundme(t *testing.T) {
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	texp, stor, _, err := InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}

	checkstorage := func(msg string, stor value.Value, owner_ string, funding_goal_ uint64, amount_raised_ uint64) {
		switch stor.(type) {
		case value.StructVal:
			init := stor.(value.StructVal)
			owner, ok := init.Field["owner"].(value.KeyVal)
			if !ok || owner.Value != owner_ {
				t.Errorf("storage has wrong value in field %s. msg: %s", "owner", msg)
			}
			funding_goal, ok := init.Field["funding_goal"].(value.KoinVal)
			if !ok || funding_goal.Value != funding_goal_ {
				t.Errorf("storage has wrong value in field %s. msg: %s", "funding_goal", msg)
			}
			amount_raised, ok := init.Field["amount_raised"].(value.KoinVal)
			if !ok || amount_raised.Value != amount_raised_ {
				t.Errorf("storage has wrong value in field %s. msg: %s", "amount_raised", msg)
			}
		}
	}

	ownerkey := "1234567890abcdef1234567890abcdef"
	otherkey := "asdasdasd"
	param1 := value.KeyVal{otherkey}
	oplist, stor, _, _ := InterpretContractCall(texp, param1, "main", stor, 900000,
		0, 999999)
	checkstorage("call1", stor, ownerkey, 1100000, 900000)
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}
	oplist, stor, _, _ = InterpretContractCall(texp, param1, "main", stor, 100000,
		0, 9999999)
	checkstorage("call2", stor, ownerkey, 1100000, 1000000)
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

	oplist, stor, _, _ = InterpretContractCall(texp, param1, "main", stor, 500000,
		1000000, 999999)
	checkstorage("call3", stor, ownerkey, 1100000, 1100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		transferOp, ok := oplist[0].(value.Transfer)
		if !ok || transferOp.Amount != 400000 {
			t.Errorf("unexpected returned operation")
		}
		if !ok || transferOp.Key != otherkey {
			t.Errorf("unexpected returned operation")
		}
	}

	oplist, stor, _, _ = InterpretContractCall(texp, param1, "main", stor, 900000,
		0, 9999999)
	checkstorage("call4", stor, ownerkey, 1100000, 1100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		failwithOp, ok := oplist[0].(value.FailWith)
		if !ok || failwithOp.Msg != "funding goal already reached" {
			t.Errorf("unexpected returned operation")
		}
	}

	oplist, stor, _, _ = InterpretContractCall(texp, value.KeyVal{ownerkey}, "main", stor, 0,
		1100000, 999999)
	checkstorage("call5", stor, ownerkey, 1100000, 1100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		transferOp, ok := oplist[0].(value.Transfer)
		if !ok || transferOp.Amount != 1100000 {
			t.Errorf("unexpected returned operation")
		}
		if !ok || transferOp.Key != ownerkey {
			t.Errorf("unexpected returned operation")
		}
	}
}

func TestRunOutOfGas(t *testing.T) {
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	_, _, remaining, err := InitiateContract(dat, 200000)
	if err == nil {
		t.Error("should've run out of gas")
		return
	}
	if remaining > 0 {
		t.Errorf("didn't run out of gas in semantic check, %d remaining gas", remaining)
	}

	ownerkey := "1234567890abcdef1234567890abcdef"
	param1 := value.KeyVal{ownerkey}

	texp, stor, remaining, err := InitiateContract(dat, 207000)
	if err != nil {
		t.Errorf("error initiating contract")
		return
	}
	oplist, stor, _, remaining := InterpretContractCall(texp, param1, "main", stor, 900000,
		0, remaining)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		failWith, ok := oplist[0].(value.FailWith)
		if !ok || failWith.Msg != "ran out of gas!" {
			t.Errorf("unexpected returned operation")
		}
	}
	if remaining > 0 {
		t.Errorf("should have 0 remaining gas but had %d", remaining)
	}
}

func TestCallContract(t *testing.T) {
	dat, err := ioutil.ReadFile("test_cases/unexistingcontract_interp")
	if err != nil {
		t.Error("Error reading testfile_noerror")
	}
	texp, sto, _, err := InitiateContract(dat, 20000000)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	param := value.UnitVal{}

	oplist, sto, _, _ := InterpretContractCall(texp, param, "second", sto, 900000,
		0, 100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have length 1")
		return
	}
	failwith2, ok := oplist[0].(value.FailWith)
	if !ok {
		t.Errorf("op[0] should be failwith operation")
	} else {
		if failwith2.Msg != "contract spendings exceed contract balance" {
			t.Errorf("wrong failwith message: %s", failwith2.Msg)
		}
	}

	oplist, sto, _, _ = InterpretContractCall(texp, param, "second", sto, 90000000,
		0, 100000)
	call, ok := oplist[0].(value.ContractCall)
	if !ok {
		t.Errorf("op[0] should be failwith operation")
	} else {
		if call.Entry != "main" {
			t.Errorf("entry has wrong value of %s", call.Entry)
		}
		if call.Address != "aaba1231333aaba1231333aaba123133" {
			t.Errorf("address has wrong value of %s", call.Address)
		}
		if call.Amount != 11500000 {
			t.Errorf("amount has wrong value of %d", call.Amount)
		}

		if params, ok := call.Params.(value.KeyVal); !ok || params.Value != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
			t.Errorf("params has wrong value of %s and type %s", params, reflect.TypeOf(params).String())
		}
	}
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
		t.Error("Error reading testfile_noerror:", testpath)
	}
	lex := lexer.NewLexer(dat)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		t.Errorf("parse error: " + err.Error())
	} else {
		parsed := par.(Exp)
		typed, err, remainingGas := AddTypes(parsed, 10000000)
		if remainingGas > 0 {
			print("\n" + typed.String() + "\n")
		} else {
			t.Errorf("ran out of gas")
			return
		}
		if shouldFail {
			if err == nil {
				t.Errorf("Didn't find any noErrors")
			}
		} else {
			if err != nil {
				t.Errorf("Found ErrorType")
			}
		}
	}
}

func getTypedAST(t *testing.T, testpath string) (TypedExp, error) {
	dat, err := ioutil.ReadFile(testpath)
	if err != nil {
		t.Error("Error reading testfile_noerror:", testpath)
	}
	lex := lexer.NewLexer(dat)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		t.Errorf("parse error: " + err.Error())
		return TypedExp{}, err
	}
	texp, err, _ := AddTypes(par.(Exp), 999999999)
	return texp, err
}
