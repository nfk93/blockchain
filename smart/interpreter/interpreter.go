package interpreter

import (
	"fmt"
	"github.com/mndrix/ps"
	. "github.com/nfk93/blockchain/smart/interpreter/ast"
	"github.com/nfk93/blockchain/smart/interpreter/lexer"
	"github.com/nfk93/blockchain/smart/interpreter/parser"
	"strconv"
)

type PanicStruct struct {
	message string
	gas     uint64
}

var currentAmt uint64
var currentBal uint64
var spentsofar uint64

func todo(n int, gas uint64) Value {
	interpPanic("Hit todo nr. "+strconv.Itoa(n), gas)
	return UnitVal{}
}

func interpPanic(message string, gas uint64) {
	panic(PanicStruct{message, gas})
}

func NatToKoin(i uint64) uint64 {
	return i * 100000
}

func currentBalance() KoinVal {
	return KoinVal{currentBal}
}

func currentAmount() KoinVal {
	return KoinVal{currentAmt}
}

func currentFailWith(failmessage StringVal, gas uint64) OperationVal {
	interpPanic(failmessage.Value, gas)
	return OperationVal{FailWith{failmessage.Value}}
}

func contractCall(address AddressVal, amount KoinVal, entry StringVal, param Value, gas uint64) OperationVal {
	spentsofar = spentsofar + amount.Value
	if int64(currentBal+currentAmt)-int64(spentsofar) < 0 {
		interpPanic("contract spendings exceed contract balance", gas)
	}
	return OperationVal{ContractCall{address.Value, amount.Value, entry.Value, param}}
}

func accountTransfer(key KeyVal, amount KoinVal, gas uint64) OperationVal {
	spentsofar = spentsofar + amount.Value
	if int64(currentBal+currentAmt)-int64(spentsofar) < 0 {
		interpPanic("contract spendings exceed contract balance", gas)
	}
	return OperationVal{Transfer{key.Value, amount.Value}}
}

func accountDefault(key KeyVal) AddressVal {
	return AddressVal{"dummy address"} //TODO add proper functionality
}

func lookupVar(id string, venv VarEnv) Value {
	val, contained := venv.Lookup(id)
	if contained {
		return val.(Value)
	} else {
		return nil
	}
}

func InitiateContract(contractCode []byte, gas uint64) (texp TypedExp, initstor Value, remainingGas uint64, returnErr error) {
	defer func() {
		if err := recover(); err != nil {
			err := err.(PanicStruct)
			fmt.Println(err.message)
			texp = TypedExp{}
			initstor = nil
			remainingGas = err.gas
			returnErr = fmt.Errorf(err.message)
		}
	}()

	currentBal = 0
	currentAmt = 0
	spentsofar = 0

	// initial gas cost
	if int64(gas)-100000 < 0 {
		interpPanic("not enough gas to initialize contract", 0)
	}
	gas = gas - 100000

	lex := lexer.NewLexer(contractCode)
	p := parser.NewParser()
	par, err := p.Parse(lex)
	if err != nil {
		return TypedExp{}, UnitVal{}, gas, fmt.Errorf("syntax error in contract code: %s", err.Error())
	}
	texp, ok, gas := AddTypes(par.(Exp), gas)
	if gas == 0 {
		interpPanic("ran out of gas when building typed AST", gas)
	}
	if !ok {
		fmt.Println(texp.String())
		return TypedExp{}, UnitVal{}, gas, fmt.Errorf("semantic error in contract code")
	}
	initstorage, gas := interpretStorageInit(texp, gas)
	return texp, initstorage, gas, nil
}

func interpretStorageInit(texp TypedExp, gas uint64) (Value, uint64) {
	exp := texp.Exp.(TopLevel)
	venv := ps.NewMap()
	for _, e := range exp.Roots {
		e := e.(TypedExp).Exp
		switch e.(type) {
		case StorageInitExp:
			e := e.(StorageInitExp)
			storageVal, gas := interpret(e.Exp.(TypedExp), venv, gas)
			return storageVal, gas
		}
	}
	return todo(-1, gas), gas
}

func InterpretContractCall(
	texp TypedExp,
	params Value,
	entry string,
	stor Value,
	amount uint64,
	balance uint64,
	gas uint64,
) (oplist []Operation, storage Value, spent uint64, remainingGas uint64) {

	// initiate module variables
	currentAmt = amount
	currentBal = balance
	spentsofar = 0

	defer func() {
		if err := recover(); err != nil {
			err := err.(PanicStruct)
			fmt.Println(err.message)
			oplist = []Operation{FailWith{err.message}}
			storage = stor
			spent = 0
			remainingGas = err.gas
		}
	}()

	exp := texp.Exp.(TopLevel)
	venv := ps.NewMap()
	for _, e := range exp.Roots {
		e := e.(TypedExp).Exp
		switch e.(type) {
		case EntryExpression:
			e := e.(EntryExpression)
			if e.Id == entry {
				// apply params to venv
				venv, err := applyParams(params, e.Params, venv)
				if err != nil {
					return []Operation{failwith(err.Error())}, stor, 0, gas
				}
				// apply storage to venv
				venv, err = applyParams(stor, e.Storage, venv)
				if err != nil {
					return []Operation{failwith("storage doesn't match storage type definition")}, stor, 0, gas
				}
				bodyTuple_, gas := interpret(e.Body.(TypedExp), venv, gas)
				bodyTuple := bodyTuple_.(TupleVal)
				opvallist := bodyTuple.Values[0].(ListVal).Values
				var oplist []Operation
				for _, v := range opvallist {
					oplist = append(oplist, v.(OperationVal).Value)
				}
				return oplist, bodyTuple.Values[1], spentsofar, gas
			}
		}
	}
	return nil, UnitVal{}, 0, gas // TODO this is just a dummy return Value. Should never happen
}

func applyParams(paramVal Value, pattern Pattern, venv VarEnv) (VarEnv, error) {
	switch paramVal.(type) {
	case TupleVal:
		paramVal := paramVal.(TupleVal)
		if len(pattern.Params) == 1 {
			if checkParam(paramVal, pattern.Params[0].Anno.Typ) {
				return venv.Set(pattern.Params[0].Id, paramVal), nil
			} else {
				return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
			}
		} else if len(pattern.Params) == len(paramVal.Values) {
			venv_ := venv
			for i, param := range pattern.Params {
				if checkParam(paramVal.Values[i], param.Anno.Typ) {
					venv_ = venv_.Set(param.Id, paramVal.Values[i])
				} else {
					return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
				}
			}
			return venv_, nil
		} else {
			return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
		}
	case UnitVal:
		if len(pattern.Params) == 0 {
			return venv, nil
		} else if len(pattern.Params) == 1 {
			if checkParam(paramVal, pattern.Params[0].Anno.Typ) {
				return venv.Set(pattern.Params[0].Id, UnitVal{}), nil
			} else {
				return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
			}
		} else {
			return venv, fmt.Errorf("pattern mistmatch, expected %d values but got none", len(pattern.Params))
		}
	default:
		if len(pattern.Params) != 1 {
			return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
		} else {
			if checkParam(paramVal, pattern.Params[0].Anno.Typ) {
				return venv.Set(pattern.Params[0].Id, paramVal), nil
			} else {
				return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
			}
		}
	}
}

func checkParam(param interface{}, typ Type) bool {
	switch typ.Type() {
	case STRING:
		_, ok := param.(StringVal)
		return ok
	case KEY:
		_, ok := param.(KeyVal)
		return ok
	case INT:
		_, ok := param.(IntVal)
		return ok
	case KOIN:
		_, ok := param.(KoinVal)
		return ok
	case NAT:
		_, ok := param.(NatVal)
		return ok
	case BOOL:
		_, ok := param.(BoolVal)
		return ok
	case OPERATION:
		_, ok := param.(OperationVal)
		return ok
	case ADDRESS:
		_, ok := param.(AddressVal)
		return ok
	case UNIT:
		_, ok := param.(UnitVal)
		return ok
	case TUPLE:
		val, ok := param.(TupleVal)
		if !ok {
			return false
		}
		tupletypes := typ.(TupleType)
		values := val.Values
		if !ok {
			return false
		}
		if len(values) != len(tupletypes.Typs) {
			return false
		}
		for i, tupletype := range tupletypes.Typs {
			ok = ok && checkParam(values[i], tupletype)
		}
		return ok
	case LIST:
		val, ok := param.(ListVal)
		if !ok {
			return false
		}
		listtype := typ.(ListType).Typ
		values := val.Values
		if !ok {
			return false
		}
		for _, v := range values {
			ok = ok && checkParam(v, listtype)
		}
		return ok
	case OPTION:
		val, ok := param.(OptionVal)
		if !ok {
			return false
		}
		if val.Opt == true {
			ok = checkParam(val.Value, typ.(OptionType).Typ)
		}
		return ok
	case STRUCT:
		val, ok := param.(StructVal)
		structtype := typ.(StructType)
		if !ok {
			return false
		}
		if len(val.Field) != len(structtype.Fields) {
			return false
		}
		for fieldname, value := range val.Field {
			found := false
			for _, f := range structtype.Fields {
				if f.Id == fieldname {
					ok = ok && checkParam(value, f.Typ)
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		return ok
	default:
		return false
	}
}

func failwith(str string) FailWith {
	return FailWith{str}
}

func createStruct() StructVal {
	m := make(map[string]Value)
	return StructVal{m}
}

func interpret(texp TypedExp, venv VarEnv, gas uint64) (Value, uint64) {
	// pay gas
	if int64(gas)-1000 < 0 {
		interpPanic("ran out of gas!", 0)
	}
	gas = gas - 1000
	exp := texp.Exp
	switch exp.(type) {
	case BinOpExp:
		exp := exp.(BinOpExp)
		leftval, gas := interpret(exp.Left.(TypedExp), venv, gas)
		rightval, gas := interpret(exp.Right.(TypedExp), venv, gas)
		switch exp.Op {
		case PLUS:
			switch exp.Left.(TypedExp).Type.Type() {
			case KOIN:
				return KoinVal{leftval.(KoinVal).Value + rightval.(KoinVal).Value}, gas
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case NAT:
					return NatVal{leftval.(NatVal).Value + rightval.(NatVal).Value}, gas
				case INT:
					return IntVal{int64(leftval.(NatVal).Value) + rightval.(IntVal).Value}, gas
				default:
					return todo(1, gas), gas
				}
			case INT:
				switch exp.Right.(TypedExp).Type.Type() {
				case NAT:
					return IntVal{leftval.(IntVal).Value + int64(rightval.(NatVal).Value)}, gas
				case INT:
					return IntVal{leftval.(IntVal).Value + rightval.(IntVal).Value}, gas
				default:
					return todo(2, gas), gas
				}
			default:
				return todo(3, gas), gas
			}
		case MINUS:
			switch exp.Left.(TypedExp).Type.Type() {
			case INT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					return IntVal{leftval.(IntVal).Value - rightval.(IntVal).Value}, gas
				case NAT:
					return IntVal{leftval.(IntVal).Value - int64(rightval.(NatVal).Value)}, gas
				default:
					return todo(4, gas), gas
				}
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					return IntVal{int64(leftval.(NatVal).Value) - rightval.(IntVal).Value}, gas
				case NAT:
					return IntVal{int64(leftval.(NatVal).Value) - int64(rightval.(NatVal).Value)}, gas
				default:
					return todo(5, gas), gas
				}
			case KOIN:
				left := leftval.(KoinVal).Value
				right := rightval.(KoinVal).Value
				if left < right {
					interpPanic(fmt.Sprintf("Subtracting %d from %d would result in a negative Koin value", left, right), gas)
				}
				return KoinVal{leftval.(KoinVal).Value - rightval.(KoinVal).Value}, gas
			default:
				return todo(6, gas), gas
			}
		case TIMES:
			switch texp.Type.Type() {
			case INT:
				return IntVal{leftval.(IntVal).Value * rightval.(IntVal).Value}, gas
			case NAT:
				return NatVal{leftval.(NatVal).Value * rightval.(NatVal).Value}, gas
			case KOIN:
				return KoinVal{leftval.(KoinVal).Value * rightval.(KoinVal).Value}, gas
			default:
				return todo(7, gas), gas
			}
		case DIVIDE:
			switch exp.Left.(TypedExp).Type.Type() {
			case KOIN:
				switch exp.Right.(TypedExp).Type.Type() {
				case KOIN:
					if rightval.(KoinVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
						return OptionVal{Opt: false}, gas
					}
					left := leftval.(KoinVal).Value
					right := rightval.(KoinVal).Value
					quotient, remainder := left/right, left%right
					values := []Value{NatVal{quotient}, KoinVal{remainder}}
					return TupleVal{values}, gas
				case NAT:
					if rightval.(NatVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(KoinVal).Value
					right := rightval.(NatVal).Value
					quotient := uint64(left / right)
					remainder := ((left / right) - quotient) * right
					values := []Value{KoinVal{quotient}, KoinVal{remainder}}
					return TupleVal{values}, gas
				default:
					return todo(8, gas), gas
				}
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					if rightval.(IntVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := int64(leftval.(NatVal).Value)
					right := rightval.(IntVal).Value
					quotient, remainder := left/right, left%right
					values := []Value{IntVal{quotient}, NatVal{uint64(remainder)}}
					return TupleVal{values}, gas
				case NAT:
					if rightval.(NatVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(NatVal).Value
					right := rightval.(NatVal).Value
					quotient, remainder := left/right, left%right
					values := []Value{NatVal{quotient}, NatVal{remainder}}
					return TupleVal{values}, gas
				default:
					return todo(9, gas), gas
				}
			case INT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					if rightval.(IntVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(IntVal).Value
					right := rightval.(IntVal).Value
					quotient, remainder := left/right, left%right
					values := []Value{IntVal{quotient}, NatVal{uint64(remainder)}}
					return TupleVal{values}, gas
				case NAT:
					if rightval.(NatVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(IntVal).Value
					right := int64(rightval.(NatVal).Value)
					quotient, remainder := left/right, left%right
					values := []Value{IntVal{quotient}, NatVal{uint64(remainder)}}
					return TupleVal{values}, gas
				default:
					return todo(10, gas), gas
				}
			default:
				return todo(11, gas), gas
			}
		case EQ:
			return BoolVal{leftval == rightval}, gas
		case NEQ:
			return BoolVal{leftval != rightval}, gas
		case GEQ:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value >= rightval.(NatVal).Value}, gas
			case INT:
				return BoolVal{leftval.(IntVal).Value >= rightval.(IntVal).Value}, gas
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value >= rightval.(KoinVal).Value}, gas
			default:
				return todo(12, gas), gas
			}
		case LEQ:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value <= rightval.(NatVal).Value}, gas
			case INT:
				return BoolVal{leftval.(IntVal).Value <= rightval.(IntVal).Value}, gas
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value <= rightval.(KoinVal).Value}, gas
			default:
				return todo(13, gas), gas
			}
		case LT:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value < rightval.(NatVal).Value}, gas
			case INT:
				return BoolVal{leftval.(IntVal).Value < rightval.(IntVal).Value}, gas
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value < rightval.(KoinVal).Value}, gas
			default:
				return todo(14, gas), gas
			}
		case GT:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value > rightval.(NatVal).Value}, gas
			case INT:
				return BoolVal{leftval.(IntVal).Value > rightval.(IntVal).Value}, gas
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value > rightval.(KoinVal).Value}, gas
			default:
				return todo(15, gas), gas
			}
		case AND:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return NatVal{leftval.(NatVal).Value & rightval.(NatVal).Value}, gas
			case BOOL:
				return BoolVal{leftval.(BoolVal).Value && rightval.(BoolVal).Value}, gas
			default:
				return todo(16, gas), gas
			}
		case OR:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return NatVal{leftval.(NatVal).Value | rightval.(NatVal).Value}, gas
			case BOOL:
				return BoolVal{leftval.(BoolVal).Value || rightval.(BoolVal).Value}, gas
			default:
				return todo(17, gas), gas
			}
		default:
			return todo(18, gas), gas
		}
	case TypeDecl:
		return todo(19, gas), gas
	case KeyLit:
		exp := exp.(KeyLit)
		return KeyVal{exp.Key}, gas
	case BoolLit:
		exp := exp.(BoolLit)
		return BoolVal{exp.Val}, gas
	case IntLit:
		exp := exp.(IntLit)
		return IntVal{exp.Val}, gas
	case KoinLit:
		exp := exp.(KoinLit)
		return KoinVal{exp.Val}, gas
	case StringLit:
		exp := exp.(StringLit)
		return StringVal{exp.Val}, gas
	case NatLit:
		exp := exp.(NatLit)
		return NatVal{exp.Val}, gas
	case UnitLit:
		return UnitVal{}, gas
	case AddressLit:
		exp := exp.(AddressLit)
		return AddressVal{exp.Val}, gas
	case StructLit:
		exp := exp.(StructLit)
		newStruct := createStruct()
		for i, id := range exp.Ids {
			newStruct.Field[id], gas = interpret(exp.Vals[i].(TypedExp), venv, gas)
		}
		return newStruct, gas
	case ListLit:
		exp := exp.(ListLit)
		if len(exp.List) == 0 {
			return ListVal{make([]Value, 0)}, gas
		}
		var returnlist []Value
		for _, e := range exp.List {
			val, gas_ := interpret(e.(TypedExp), venv, gas)
			gas = gas_
			returnlist = append(returnlist, val)
		}
		return ListVal{returnlist}, gas
	case ListConcat:
		exp := exp.(ListConcat)
		e, gas := interpret(exp.Exp.(TypedExp), venv, gas)
		list_, gas := interpret(exp.Exp.(TypedExp), venv, gas)
		list := list_.(ListVal)
		return ListVal{append(list.Values, e)}, gas
	case CallExp:
		exp := exp.(CallExp)
		name_, gas := interpret(exp.ExpList[0].(TypedExp), venv, gas)
		name := name_.(LambdaVal)
		switch name.Value {
		case CURRENT_BALANCE:
			return currentBalance(), gas
		case CURRENT_AMOUNT:
			return currentAmount(), gas
		case CURRENT_GAS:
			return KoinVal{gas}, gas
		case CURRENT_FAILWITH:
			failmessage_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			failmessage := failmessage_.(StringVal)
			return currentFailWith(failmessage, gas), gas
		case CONTRACT_CALL:
			address_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			address := address_.(AddressVal)
			amount_, gas := interpret(exp.ExpList[2].(TypedExp), venv, gas)
			amount := amount_.(KoinVal)
			entry_, gas := interpret(exp.ExpList[3].(TypedExp), venv, gas)
			entry := entry_.(StringVal)
			param, gas := interpret(exp.ExpList[4].(TypedExp), venv, gas)
			return contractCall(address, amount, entry, param, gas), gas
		case ACCOUNT_TRANSFER:
			key_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			key := key_.(KeyVal)
			amount_, gas := interpret(exp.ExpList[2].(TypedExp), venv, gas)
			amount := amount_.(KoinVal)
			return accountTransfer(key, amount, gas), gas
		case ACCOUNT_DEFAULT:
			key_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			key := key_.(KeyVal)
			return accountDefault(key), gas
		default:
			return todo(20, gas), gas
		}
	case LetExp:
		exp := exp.(LetExp)
		value, gas := interpret(exp.DefExp.(TypedExp), venv, gas)
		venv, _ = applyParams(value, exp.Patt, venv)
		return interpret(exp.InExp.(TypedExp), venv, gas)
	case AnnoExp:
		exp := exp.(AnnoExp)
		return interpret(exp.Exp.(TypedExp), venv, gas)
	case TupleExp:
		exp := exp.(TupleExp)
		var tupleValues []Value
		for _, e := range exp.Exps {
			interE, gas_ := interpret(e.(TypedExp), venv, gas)
			gas = gas_
			tupleValues = append(tupleValues, interE)
		}
		return TupleVal{tupleValues}, gas
	case VarExp:
		exp := exp.(VarExp)
		return lookupVar(exp.Id, venv), gas
	case ExpSeq:
		exp := exp.(ExpSeq)
		_, gas = interpret(exp.Left.(TypedExp), venv, gas)
		rightval, gas := interpret(exp.Right.(TypedExp), venv, gas)
		return rightval, gas
	case IfThenElseExp:
		exp := exp.(IfThenElseExp)
		condition_, gas := interpret(exp.If.(TypedExp), venv, gas)
		condition := condition_.(BoolVal).Value
		if condition {
			return interpret(exp.Then.(TypedExp), venv, gas)
		} else {
			return interpret(exp.Else.(TypedExp), venv, gas)
		}
	case IfThenExp:
		exp := exp.(IfThenExp)
		condition_, gas := interpret(exp.If.(TypedExp), venv, gas)
		condition := condition_.(BoolVal).Value
		if condition {
			return interpret(exp.Then.(TypedExp), venv, gas)
		}
		return UnitVal{}, gas
	case ModuleLookupExp:
		exp := exp.(ModuleLookupExp)
		switch exp.ModId {
		case "Current":
			switch exp.FieldId {
			case "balance":
				return LambdaVal{CURRENT_BALANCE}, gas
			case "amount":
				return LambdaVal{CURRENT_AMOUNT}, gas
			case "gas":
				return LambdaVal{CURRENT_GAS}, gas
			case "failwith":
				return LambdaVal{CURRENT_FAILWITH}, gas
			default:
				return todo(22, gas), gas
			}
		case "Contract":
			switch exp.FieldId {
			case "call":
				return LambdaVal{CONTRACT_CALL}, gas
			default:
				return todo(23, gas), gas
			}
		case "Account":
			switch exp.FieldId {
			case "transfer":
				return LambdaVal{ACCOUNT_TRANSFER}, gas
			case "default":
				return LambdaVal{ACCOUNT_DEFAULT}, gas
			default:
				return todo(24, gas), gas
			}
		default:
			return todo(25, gas), gas
		}

	case LookupExp:
		exp := exp.(LookupExp)
		var structVal StructVal
		for i, id := range exp.PathIds {
			if i == 0 {
				structVal = lookupVar(id, venv).(StructVal)
			} else {
				structVal = structVal.Field[id].(StructVal)
			}
		}
		return structVal.Field[exp.LeafId], gas
	case UpdateStructExp:
		exp := exp.(UpdateStructExp)
		struc := lookupVar(exp.Root, venv)
		innerStruct := struc
		path := exp.Path
		for len(path) > 1 {
			innerStruct = innerStruct.(StructVal).Field[path[0]]
			path = path[1:]
		}
		newval, gas := interpret(exp.Exp.(TypedExp), venv, gas)
		innerStruct.(StructVal).Field[path[0]] = newval
		return struc, gas
	case StorageInitExp:
		return todo(26, gas), gas
	default:
		return todo(27, gas), gas
	}
}
