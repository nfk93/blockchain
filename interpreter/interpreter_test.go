package interpreter

import (
	"fmt"
	. "github.com/nfk93/blockchain/interpreter/ast"
	"github.com/nfk93/blockchain/interpreter/lexer"
	"github.com/nfk93/blockchain/interpreter/parser"
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

func TestIntConstant(t *testing.T) {
	testpath := "test_cases/constants/int"
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", IntVal{13}, 0, 99999999999)
	switch sto.(type) {
	case IntVal:
		if sto.(IntVal).Value != 15 {
			t.Errorf("storage has unexpected value of %d", sto.(IntVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", AddressVal{"123123aA"}, 0, 9999999999)
	switch sto.(type) {
	case AddressVal:
		if sto.(AddressVal).Value != "3132141AAAa" {
			t.Errorf("storage has unexpected value of %s", sto.(AddressVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", BoolVal{true}, 0, 999999999999999999)
	switch sto.(type) {
	case BoolVal:
		if sto.(BoolVal).Value != false {
			t.Errorf("storage has unexpected value of %t", sto.(BoolVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main",
		TupleVal{[]Value{IntVal{123}, TupleVal{[]Value{IntVal{2}, StringVal{"serser"}}}}},
		0, 9999999)
	switch sto.(type) {
	case TupleVal:
		sto := sto.(TupleVal)
		if sto.Values[0].(IntVal).Value != 4 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[1].(TupleVal).Values[0].(IntVal).Value != 5 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[1].(TupleVal).Values[1].(StringVal).Value != "bye" {
			t.Errorf("storage has unexpected value of %s",
				sto.Values[1].(TupleVal).Values[1].(StringVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", KeyVal{"1212Ddd"}, 0, 999999999)
	switch sto.(type) {
	case KeyVal:
		if sto.(KeyVal).Value != "aaAAaaA" {
			t.Errorf("storage has unexpected value of %s", sto.(KeyVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", KoinVal{uint64(110000)}, 0, 99999999999)
	switch sto.(type) {
	case KoinVal:
		if sto.(KoinVal).Value != 13355000 {
			t.Errorf("storage has unexpected value of %d", sto.(KoinVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", ListVal{[]Value{IntVal{2}}},
		0, 9999999999)
	switch sto.(type) {
	case ListVal:
		sto := sto.(ListVal)
		if sto.Values[0].(IntVal).Value != 7 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[1].(IntVal).Value != 9 {
			t.Errorf("storage has unexpected value")
		}
		if sto.Values[2].(IntVal).Value != 13 {
			t.Errorf("storage has unexpected value of %d at index 2", sto.Values[2].(IntVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", NatVal{13}, 0, 99999999999)
	switch sto.(type) {
	case NatVal:
		if sto.(NatVal).Value != 117 {
			t.Errorf("storage has unexpected value of %d", sto.(NatVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", StringVal{"eymom"}, 0, 99999999999)
	switch sto.(type) {
	case StringVal:
		if !(sto.(StringVal).Value == "dank") {
			t.Errorf("storage has unexpected value of %s", sto.(StringVal).Value)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", UnitVal{}, 0, 99999999999)
	switch sto.(type) {
	case UnitVal:
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty but: %s", oplist)
	}
}

func TestStructConstant(t *testing.T) {
	testpath := "test_cases/constants/struct"
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	storageinit := createStruct()
	storageinit.Field["a"] = IntVal{1213}
	storageinit.Field["b"] = TupleVal{[]Value{IntVal{5}, IntVal{6}}}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", storageinit, 0, 99999999999)
	switch sto.(type) {
	case StructVal:
		sto, oksto := sto.(StructVal)
		if oksto {
			a, oka := sto.Field["a"].(IntVal)
			if !oka || a.Value != 1111 {
				t.Errorf("storage.a has unexpected value of %d", sto.Field["a"])
			}
			b, okb := sto.Field["b"].(TupleVal)
			if !okb {
				t.Errorf("storage.b has unexpected type")
			} else {
				tup1, ok1 := b.Values[0].(IntVal)
				tup2, ok2 := b.Values[1].(IntVal)
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
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", KoinVal{1000000}, 0, 999999999)
	switch sto.(type) {
	case KoinVal:
		sto := sto.(KoinVal)
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
	tupleval1 := TupleVal{[]Value{TupleVal{[]Value{IntVal{1}, IntVal{2}}}, StringVal{"ey"}}}
	tupleval2 := TupleVal{[]Value{TupleVal{[]Value{IntVal{1}, IntVal{2}}}, IntVal{123}}}
	tupletyp1 := TupleType{[]Type{TupleType{[]Type{IntType{}, IntType{}}}, StringType{}}}
	if !checkParam(tupleval1, tupletyp1) {
		t.Errorf("4")
	}
	if checkParam(tupleval2, tupletyp1) {
		t.Errorf("5")
	}

	structval := createStruct()
	structval.Field["a"] = IntVal{123}
	structval.Field["b"] = StringVal{"eytyyyyy"}
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

func TestInterpretUpdateStruct(t *testing.T) {
	testpath := "test_cases/updatestruct_interp"
	texp, ok := getTypedAST(t, testpath)
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	params := TupleVal{[]Value{IntVal{173}, StringVal{"newinner"}, StringVal{"newa"}}}
	innermost := createStruct()
	innermost.Field["buried"] = IntVal{12}
	innermost.Field["deep"] = StringVal{"very deep"}
	inner := createStruct()
	inner.Field["id"] = StringVal{"innerid"}
	inner.Field["innermost"] = innermost
	storage := createStruct()
	storage.Field["a"] = StringVal{"test"}
	storage.Field["b"] = inner

	oplist, sto, _ := InterpretContractCall(texp, params, "main", storage, 0, 999999999999)
	switch sto.(type) {
	case StructVal:
		sto := sto.(StructVal)
		a, ok := sto.Field["a"].(StringVal)
		if !ok || a.Value != "newa" {
			t.Errorf("sto.a has unexpected value of %s", a)
		}
		b, ok := sto.Field["b"].(StructVal)
		if !ok {
			t.Errorf("sto.b has unexpected type. It is %s", b)
		} else {
			bid, ok := b.Field["id"].(StringVal)
			if !ok || bid.Value != "newinner" {
				t.Errorf("sto.b.id has unexpected value of %s", bid)
			}
			binnermost, ok := b.Field["innermost"].(StructVal)
			if !ok {
				t.Errorf("sto.b.innermost has unexpected type. It is %s", binnermost)
			} else {
				buried, ok := binnermost.Field["buried"].(IntVal)
				if !ok || buried.Value != 173+12 {
					t.Errorf("sto.b.innermost.buried has unexpected value of %d", buried)
				}
				deep, ok := binnermost.Field["deep"].(StringVal)
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
	texp, ok := getTypedAST(t, "test_cases/binop_interp")
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	params := TupleVal{[]Value{IntVal{13}, IntVal{17}}}
	storage := IntVal{19}
	oplist, sto, _ := InterpretContractCall(texp, params, "main", storage, 0, 999999999999)
	switch sto.(type) {
	case IntVal:
		if sto.(IntVal).Value != 13+17+19 {
			t.Errorf("storage has unexpected value of %d", sto.(IntVal).Value)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

}

func TestDivision(t *testing.T) {
	texp, ok := getTypedAST(t, "test_cases/division_interp")
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	params := TupleVal{[]Value{IntVal{13}, IntVal{17}}}
	storage := TupleVal{[]Value{IntVal{19}, NatVal{0}}}
	oplist, sto, _ := InterpretContractCall(texp, params, "main", storage, 0, 100000)

	switch sto.(type) {
	case TupleVal:
		sto := sto.(TupleVal)
		if len(sto.Values) != 2 {
			t.Errorf("storage tuple has unexpeted length of %d", len(sto.Values))
		}
		switch sto.Values[0].(type) {
		case IntVal:
		default:
			t.Errorf("First value of returned storage tuple is wrong type")
		}
		switch sto.Values[1].(type) {
		case NatVal:
		default:
			t.Errorf("Second value of returned storage tuple is wrong type")
		}
		val1 := sto.Values[0].(IntVal).Value
		val2 := sto.Values[1].(NatVal).Value
		if val1 != 0 && val2 != 10 {
			t.Errorf("storage has unexpected value of (%d, %d)", val1, val2)
		}
	default:
		t.Errorf("storage isn't expected type. It is type %s", reflect.TypeOf(sto).String())
	}
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

	params = TupleVal{[]Value{IntVal{0}, IntVal{0}}}
	oplist, sto, _ = InterpretContractCall(texp, params, "main", storage, 0, 100000)
	if len(oplist) != 1 {
		t.Errorf("oplist is len %d should be 1", len(oplist))
	}

	failwithOp, ok := oplist[0].(FailWith)
	if !ok {
		fmt.Println(failwithOp.msg)
		t.Errorf("unexpected returned operation")
	}
}

func TestInterpretFailwith(t *testing.T) {
	texp, ok := getTypedAST(t, "test_cases/currentfailwith_interp")
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	unitval := UnitVal{}
	oplist, sto, _ := InterpretContractCall(texp, unitval, "main", KoinVal{1000000}, 0, 99999999999)
	switch sto.(type) {
	case KoinVal:
		sto := sto.(KoinVal)
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
	texp, ok := getTypedAST(t, "test_cases/letexps_interp")
	if !ok {
		t.Errorf("Semant error")
		fmt.Println(texp.String())
		return
	}
	params := TupleVal{[]Value{IntVal{7}, StringVal{"not imporatnt"}, NatVal{13}}}
	storage := IntVal{19}
	oplist, sto, _ := InterpretContractCall(texp, params, "main", storage, 0, 999999999999)
	switch sto.(type) {
	case IntVal:
		if sto.(IntVal).Value != 19-(15+7+13) {
			t.Errorf("storage has unexpected value of %d", sto.(IntVal).Value)
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
		t.Error("Error reading testfile")
	}
	_, init, err := InitiateContract(dat, 999999999999999)
	if err != nil {
		t.Error(err)
		return
	}

	switch init.(type) {
	case StructVal:
		init := init.(StructVal)
		owner, ok := init.Field["owner"].(KeyVal)
		if !ok || owner.Value != "YLtLqD1fWHthSVHPD116oYvsd4PTAHUoc" {
			t.Errorf("init storage has wrong value in field %s", "owner")
		}
		funding_goal, ok := init.Field["funding_goal"].(KoinVal)
		if !ok || funding_goal.Value != 1100000 {
			t.Errorf("init storage has wrong value in field %s. expected: 1100000, actual: %d",
				"funding_goal", funding_goal.Value)
		}
		amount_raised, ok := init.Field["amount_raised"].(KoinVal)
		if !ok || amount_raised.Value != 0 {
			t.Errorf("init storage has wrong value in field %s", "amount_raised")
		}
	}
}

func TestRunFundme(t *testing.T) {
	dat, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile")
	}
	texp, stor, err := InitiateContract(dat, 999999999)
	if err != nil {
		t.Error(err)
		return
	}

	checkstorage := func(msg string, stor Value, owner_ string, funding_goal_ uint64, amount_raised_ uint64) {
		switch stor.(type) {
		case StructVal:
			init := stor.(StructVal)
			owner, ok := init.Field["owner"].(KeyVal)
			if !ok || owner.Value != owner_ {
				t.Errorf("storage has wrong value in field %s. msg: %s", "owner", msg)
			}
			funding_goal, ok := init.Field["funding_goal"].(KoinVal)
			if !ok || funding_goal.Value != funding_goal_ {
				t.Errorf("storage has wrong value in field %s. msg: %s", "funding_goal", msg)
			}
			amount_raised, ok := init.Field["amount_raised"].(KoinVal)
			if !ok || amount_raised.Value != amount_raised_ {
				t.Errorf("storage has wrong value in field %s. msg: %s", "amount_raised", msg)
			}
		}
	}

	ownerkey := "YLtLqD1fWHthSVHPD116oYvsd4PTAHUoc"
	otherkey := "asdasdasd"
	param1 := KeyVal{otherkey}
	oplist, stor, _ := InterpretContractCall(texp, param1, "main", stor, 900000, 999999)
	checkstorage("call1", stor, ownerkey, 1100000, 900000)
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}
	oplist, stor, _ = InterpretContractCall(texp, param1, "main", stor, 100000, 9999999)
	checkstorage("call2", stor, ownerkey, 1100000, 1000000)
	if len(oplist) != 0 {
		t.Errorf("oplist isn't empty. It is %s", oplist)
	}

	oplist, stor, _ = InterpretContractCall(texp, param1, "main", stor, 500000, 999999)
	checkstorage("call3", stor, ownerkey, 1100000, 1100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		transferOp, ok := oplist[0].(Transfer)
		if !ok || transferOp.data.amount != 400000 {
			t.Errorf("unexpected returned operation")
		}
		if !ok || transferOp.data.key != otherkey {
			t.Errorf("unexpected returned operation")
		}
	}

	oplist, stor, _ = InterpretContractCall(texp, param1, "main", stor, 900000, 9999999)
	checkstorage("call4", stor, ownerkey, 1100000, 1100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		failwithOp, ok := oplist[0].(FailWith)
		if !ok || failwithOp.msg != "funding goal already reached" {
			t.Errorf("unexpected returned operation")
		}
	}

	oplist, stor, _ = InterpretContractCall(texp, KeyVal{ownerkey}, "main", stor, 0, 999999)
	checkstorage("call5", stor, ownerkey, 1100000, 1100000)
	if len(oplist) != 1 {
		t.Errorf("oplist should have 1 operation but had %d", len(oplist))
	} else {
		transferOp, ok := oplist[0].(Transfer)
		if !ok || transferOp.data.amount != 1100000 {
			t.Errorf("unexpected returned operation")
		}
		if !ok || transferOp.data.key != ownerkey {
			t.Errorf("unexpected returned operation")
		}
	}
}

func TestRunOutOfGas(t *testing.T) {
	_, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/nfk93/blockchain/usecases/fundme")
	if err != nil {
		t.Error("Error reading testfile")
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
		t.Error("Error reading testfile:", testpath)
	}
	lex := lexer.NewLexer(dat)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		t.Errorf("parse error: " + err.Error())
	} else {
		parsed := par.(Exp)
		typed, noErrors, remainingGas := AddTypes(parsed, 10000000)
		if remainingGas > 0 {
			print("\n" + typed.String() + "\n")
		} else {
			t.Errorf("ran out of gas")
			return
		}
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

func getTypedAST(t *testing.T, testpath string) (TypedExp, bool) {
	dat, err := ioutil.ReadFile(testpath)
	if err != nil {
		t.Error("Error reading testfile:", testpath)
	}
	lex := lexer.NewLexer(dat)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		t.Errorf("parse error: " + err.Error())
		return TypedExp{}, false
	}
	texp, ok, _ := AddTypes(par.(Exp), 999999999)
	return texp, ok
}
