package interpreter

import (
	"fmt"
	"github.com/mndrix/ps"
	. "github.com/nfk93/blockchain/smart/interpreter/ast"
	"github.com/nfk93/blockchain/smart/interpreter/lexer"
	"github.com/nfk93/blockchain/smart/interpreter/parser"
	"github.com/nfk93/blockchain/smart/interpreter/value"
	"strconv"
)

type PanicStruct struct {
	message string
	gas     uint64
}

var currentAmt uint64
var currentBal uint64
var spentsofar uint64

func todo(n int, gas uint64) value.Value {
	interpPanic("Hit todo nr. "+strconv.Itoa(n), gas)
	return value.UnitVal{}
}

func interpPanic(message string, gas uint64) {
	panic(PanicStruct{message, gas})
}

func NatToKoin(i uint64) uint64 {
	return i * 100000
}

func currentBalance() value.KoinVal {
	return value.KoinVal{currentBal}
}

func currentAmount() value.KoinVal {
	return value.KoinVal{currentAmt}
}

func currentFailWith(failmessage value.StringVal, gas uint64) value.OperationVal {
	interpPanic(failmessage.Value, gas)
	return value.OperationVal{value.FailWith{failmessage.Value}}
}

func contractCall(address value.AddressVal, amount value.KoinVal, entry value.StringVal, param value.Value, gas uint64) value.OperationVal {
	spentsofar = spentsofar + amount.Value
	if int64(currentBal+currentAmt)-int64(spentsofar) < 0 {
		interpPanic("contract spendings exceed contract balance", gas)
	}
	return value.OperationVal{value.ContractCall{address.Value, amount.Value, entry.Value, param}}
}

func accountTransfer(key value.KeyVal, amount value.KoinVal, gas uint64) value.OperationVal {
	spentsofar = spentsofar + amount.Value
	if int64(currentBal+currentAmt)-int64(spentsofar) < 0 {
		interpPanic("contract spendings exceed contract balance", gas)
	}
	return value.OperationVal{value.Transfer{key.Value, amount.Value}}
}

func accountDefault(key value.KeyVal) value.AddressVal {
	return value.AddressVal{"dummy address"} //TODO add proper functionality
}

func lookupVar(id string, venv VarEnv) value.Value {
	val, contained := venv.Lookup(id)
	if contained {
		return val.(value.Value)
	} else {
		return nil
	}
}

func InitiateContract(contractCode []byte, gas uint64) (texp TypedExp, initstor value.Value, remainingGas uint64, returnErr error) {
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
		return TypedExp{}, value.UnitVal{}, gas, fmt.Errorf("syntax error in contract code: %s", err.Error())
	}
	texp, err, gas = AddTypes(par.(Exp), gas)
	if gas == 0 {
		interpPanic("ran out of gas when building typed AST", gas)
	}
	if err != nil {
		fmt.Println(err.Error())
		return TypedExp{}, value.UnitVal{}, gas, fmt.Errorf("semantic error in contract code: %s", err.Error())
	}
	initstorage, gas := interpretStorageInit(texp, gas)
	return texp, initstorage, gas, nil
}

func interpretStorageInit(texp TypedExp, gas uint64) (value.Value, uint64) {
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
	params value.Value,
	entry string,
	stor value.Value,
	amount uint64,
	balance uint64,
	gas uint64,
) (oplist []value.Operation, storage value.Value, spent uint64, remainingGas uint64) {

	// initiate module variables
	currentAmt = amount
	currentBal = balance
	spentsofar = 0

	defer func() {
		if err := recover(); err != nil {
			err := err.(PanicStruct)
			fmt.Println(err.message)
			oplist = []value.Operation{value.FailWith{err.message}}
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
					return []value.Operation{failwith(err.Error())}, stor, 0, gas
				}
				// apply storage to venv
				venv, err = applyParams(stor, e.Storage, venv)
				if err != nil {
					return []value.Operation{failwith("storage doesn't match storage type definition")}, stor, 0, gas
				}
				bodyTuple_, gas := interpret(e.Body.(TypedExp), venv, gas)
				bodyTuple := bodyTuple_.(value.TupleVal)
				opvallist := bodyTuple.Values[0].(value.ListVal).Values
				var oplist []value.Operation
				for _, v := range opvallist {
					oplist = append(oplist, v.(value.OperationVal).Value)
				}
				return oplist, bodyTuple.Values[1], spentsofar, gas
			}
		}
	}
	return nil, value.UnitVal{}, 0, gas // TODO this is just a dummy return Value. Should never happen
}

func applyParams(paramVal value.Value, pattern Pattern, venv VarEnv) (VarEnv, error) {
	switch paramVal.(type) {
	case value.TupleVal:
		paramVal := paramVal.(value.TupleVal)
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
	case value.UnitVal:
		if len(pattern.Params) == 0 {
			return venv, nil
		} else if len(pattern.Params) == 1 {
			if checkParam(paramVal, pattern.Params[0].Anno.Typ) {
				return venv.Set(pattern.Params[0].Id, value.UnitVal{}), nil
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
		_, ok := param.(value.StringVal)
		return ok
	case KEY:
		_, ok := param.(value.KeyVal)
		return ok
	case INT:
		_, ok := param.(value.IntVal)
		return ok
	case KOIN:
		_, ok := param.(value.KoinVal)
		return ok
	case NAT:
		_, ok := param.(value.NatVal)
		return ok
	case BOOL:
		_, ok := param.(value.BoolVal)
		return ok
	case OPERATION:
		_, ok := param.(value.OperationVal)
		return ok
	case ADDRESS:
		_, ok := param.(value.AddressVal)
		return ok
	case UNIT:
		_, ok := param.(value.UnitVal)
		return ok
	case TUPLE:
		val, ok := param.(value.TupleVal)
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
		val, ok := param.(value.ListVal)
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
		val, ok := param.(value.OptionVal)
		if !ok {
			return false
		}
		if val.Opt == true {
			ok = checkParam(val.Value, typ.(OptionType).Typ)
		}
		return ok
	case STRUCT:
		val, ok := param.(value.StructVal)
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

func failwith(str string) value.FailWith {
	return value.FailWith{str}
}

func createStruct() value.StructVal {
	m := make(map[string]value.Value)
	return value.StructVal{m}
}

func interpret(texp TypedExp, venv VarEnv, gas uint64) (value.Value, uint64) {
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
				return value.KoinVal{leftval.(value.KoinVal).Value + rightval.(value.KoinVal).Value}, gas
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case NAT:
					return value.NatVal{leftval.(value.NatVal).Value + rightval.(value.NatVal).Value}, gas
				case INT:
					return value.IntVal{int64(leftval.(value.NatVal).Value) + rightval.(value.IntVal).Value}, gas
				default:
					return todo(1, gas), gas
				}
			case INT:
				switch exp.Right.(TypedExp).Type.Type() {
				case NAT:
					return value.IntVal{leftval.(value.IntVal).Value + int64(rightval.(value.NatVal).Value)}, gas
				case INT:
					return value.IntVal{leftval.(value.IntVal).Value + rightval.(value.IntVal).Value}, gas
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
					return value.IntVal{leftval.(value.IntVal).Value - rightval.(value.IntVal).Value}, gas
				case NAT:
					return value.IntVal{leftval.(value.IntVal).Value - int64(rightval.(value.NatVal).Value)}, gas
				default:
					return todo(4, gas), gas
				}
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					return value.IntVal{int64(leftval.(value.NatVal).Value) - rightval.(value.IntVal).Value}, gas
				case NAT:
					return value.IntVal{int64(leftval.(value.NatVal).Value) - int64(rightval.(value.NatVal).Value)}, gas
				default:
					return todo(5, gas), gas
				}
			case KOIN:
				left := leftval.(value.KoinVal).Value
				right := rightval.(value.KoinVal).Value
				if left < right {
					interpPanic(fmt.Sprintf("Subtracting %d from %d would result in a negative Koin value", left, right), gas)
				}
				return value.KoinVal{leftval.(value.KoinVal).Value - rightval.(value.KoinVal).Value}, gas
			default:
				return todo(6, gas), gas
			}
		case TIMES:
			switch texp.Type.Type() {
			case INT:
				return value.IntVal{leftval.(value.IntVal).Value * rightval.(value.IntVal).Value}, gas
			case NAT:
				return value.NatVal{leftval.(value.NatVal).Value * rightval.(value.NatVal).Value}, gas
			case KOIN:
				return value.KoinVal{leftval.(value.KoinVal).Value * rightval.(value.KoinVal).Value}, gas
			default:
				return todo(7, gas), gas
			}
		case DIVIDE:
			switch exp.Left.(TypedExp).Type.Type() {
			case KOIN:
				switch exp.Right.(TypedExp).Type.Type() {
				case KOIN:
					if rightval.(value.KoinVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
						return value.OptionVal{Opt: false}, gas
					}
					left := leftval.(value.KoinVal).Value
					right := rightval.(value.KoinVal).Value
					quotient, remainder := left/right, left%right
					values := []value.Value{value.NatVal{quotient}, value.KoinVal{remainder}}
					return value.TupleVal{values}, gas
				case NAT:
					if rightval.(value.NatVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(value.KoinVal).Value
					right := rightval.(value.NatVal).Value
					quotient := uint64(left / right)
					remainder := ((left / right) - quotient) * right
					values := []value.Value{value.KoinVal{quotient}, value.KoinVal{remainder}}
					return value.TupleVal{values}, gas
				default:
					return todo(8, gas), gas
				}
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					if rightval.(value.IntVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := int64(leftval.(value.NatVal).Value)
					right := rightval.(value.IntVal).Value
					quotient, remainder := left/right, left%right
					values := []value.Value{value.IntVal{quotient}, value.NatVal{uint64(remainder)}}
					return value.TupleVal{values}, gas
				case NAT:
					if rightval.(value.NatVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(value.NatVal).Value
					right := rightval.(value.NatVal).Value
					quotient, remainder := left/right, left%right
					values := []value.Value{value.NatVal{quotient}, value.NatVal{remainder}}
					return value.TupleVal{values}, gas
				default:
					return todo(9, gas), gas
				}
			case INT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					if rightval.(value.IntVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(value.IntVal).Value
					right := rightval.(value.IntVal).Value
					quotient, remainder := left/right, left%right
					values := []value.Value{value.IntVal{quotient}, value.NatVal{uint64(remainder)}}
					return value.TupleVal{values}, gas
				case NAT:
					if rightval.(value.NatVal).Value == 0 {
						interpPanic("Can't divide by zero!", gas)
					}
					left := leftval.(value.IntVal).Value
					right := int64(rightval.(value.NatVal).Value)
					quotient, remainder := left/right, left%right
					values := []value.Value{value.IntVal{quotient}, value.NatVal{uint64(remainder)}}
					return value.TupleVal{values}, gas
				default:
					return todo(10, gas), gas
				}
			default:
				return todo(11, gas), gas
			}
		case EQ:
			return value.BoolVal{leftval == rightval}, gas
		case NEQ:
			return value.BoolVal{leftval != rightval}, gas
		case GEQ:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return value.BoolVal{leftval.(value.NatVal).Value >= rightval.(value.NatVal).Value}, gas
			case INT:
				return value.BoolVal{leftval.(value.IntVal).Value >= rightval.(value.IntVal).Value}, gas
			case KOIN:
				return value.BoolVal{leftval.(value.KoinVal).Value >= rightval.(value.KoinVal).Value}, gas
			default:
				return todo(12, gas), gas
			}
		case LEQ:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return value.BoolVal{leftval.(value.NatVal).Value <= rightval.(value.NatVal).Value}, gas
			case INT:
				return value.BoolVal{leftval.(value.IntVal).Value <= rightval.(value.IntVal).Value}, gas
			case KOIN:
				return value.BoolVal{leftval.(value.KoinVal).Value <= rightval.(value.KoinVal).Value}, gas
			default:
				return todo(13, gas), gas
			}
		case LT:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return value.BoolVal{leftval.(value.NatVal).Value < rightval.(value.NatVal).Value}, gas
			case INT:
				return value.BoolVal{leftval.(value.IntVal).Value < rightval.(value.IntVal).Value}, gas
			case KOIN:
				return value.BoolVal{leftval.(value.KoinVal).Value < rightval.(value.KoinVal).Value}, gas
			default:
				return todo(14, gas), gas
			}
		case GT:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return value.BoolVal{leftval.(value.NatVal).Value > rightval.(value.NatVal).Value}, gas
			case INT:
				return value.BoolVal{leftval.(value.IntVal).Value > rightval.(value.IntVal).Value}, gas
			case KOIN:
				return value.BoolVal{leftval.(value.KoinVal).Value > rightval.(value.KoinVal).Value}, gas
			default:
				return todo(15, gas), gas
			}
		case AND:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return value.NatVal{leftval.(value.NatVal).Value & rightval.(value.NatVal).Value}, gas
			case BOOL:
				return value.BoolVal{leftval.(value.BoolVal).Value && rightval.(value.BoolVal).Value}, gas
			default:
				return todo(16, gas), gas
			}
		case OR:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return value.NatVal{leftval.(value.NatVal).Value | rightval.(value.NatVal).Value}, gas
			case BOOL:
				return value.BoolVal{leftval.(value.BoolVal).Value || rightval.(value.BoolVal).Value}, gas
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
		return value.KeyVal{exp.Key}, gas
	case BoolLit:
		exp := exp.(BoolLit)
		return value.BoolVal{exp.Val}, gas
	case IntLit:
		exp := exp.(IntLit)
		return value.IntVal{exp.Val}, gas
	case KoinLit:
		exp := exp.(KoinLit)
		return value.KoinVal{exp.Val}, gas
	case StringLit:
		exp := exp.(StringLit)
		return value.StringVal{exp.Val}, gas
	case NatLit:
		exp := exp.(NatLit)
		return value.NatVal{exp.Val}, gas
	case UnitLit:
		return value.UnitVal{}, gas
	case AddressLit:
		exp := exp.(AddressLit)
		return value.AddressVal{exp.Val}, gas
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
			return value.ListVal{make([]value.Value, 0)}, gas
		}
		var returnlist []value.Value
		for _, e := range exp.List {
			val, gas_ := interpret(e.(TypedExp), venv, gas)
			gas = gas_
			returnlist = append(returnlist, val)
		}
		return value.ListVal{returnlist}, gas
	case ListConcat:
		exp := exp.(ListConcat)
		e, gas := interpret(exp.Exp.(TypedExp), venv, gas)
		list_, gas := interpret(exp.Exp.(TypedExp), venv, gas)
		list := list_.(value.ListVal)
		return value.ListVal{append(list.Values, e)}, gas
	case CallExp:
		exp := exp.(CallExp)
		name_, gas := interpret(exp.ExpList[0].(TypedExp), venv, gas)
		name := name_.(value.LambdaVal)
		switch name.Value {
		case value.CURRENT_BALANCE:
			return currentBalance(), gas
		case value.CURRENT_AMOUNT:
			return currentAmount(), gas
		case value.CURRENT_GAS:
			return value.KoinVal{gas}, gas
		case value.CURRENT_FAILWITH:
			failmessage_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			failmessage := failmessage_.(value.StringVal)
			return currentFailWith(failmessage, gas), gas
		case value.CONTRACT_CALL:
			address_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			address := address_.(value.AddressVal)
			amount_, gas := interpret(exp.ExpList[2].(TypedExp), venv, gas)
			amount := amount_.(value.KoinVal)
			entry_, gas := interpret(exp.ExpList[3].(TypedExp), venv, gas)
			entry := entry_.(value.StringVal)
			param, gas := interpret(exp.ExpList[4].(TypedExp), venv, gas)
			return contractCall(address, amount, entry, param, gas), gas
		case value.ACCOUNT_TRANSFER:
			key_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			key := key_.(value.KeyVal)
			amount_, gas := interpret(exp.ExpList[2].(TypedExp), venv, gas)
			amount := amount_.(value.KoinVal)
			return accountTransfer(key, amount, gas), gas
		case value.ACCOUNT_DEFAULT:
			key_, gas := interpret(exp.ExpList[1].(TypedExp), venv, gas)
			key := key_.(value.KeyVal)
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
		var tupleValues []value.Value
		for _, e := range exp.Exps {
			interE, gas_ := interpret(e.(TypedExp), venv, gas)
			gas = gas_
			tupleValues = append(tupleValues, interE)
		}
		return value.TupleVal{tupleValues}, gas
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
		condition := condition_.(value.BoolVal).Value
		if condition {
			return interpret(exp.Then.(TypedExp), venv, gas)
		} else {
			return interpret(exp.Else.(TypedExp), venv, gas)
		}
	case IfThenExp:
		exp := exp.(IfThenExp)
		condition_, gas := interpret(exp.If.(TypedExp), venv, gas)
		condition := condition_.(value.BoolVal).Value
		if condition {
			return interpret(exp.Then.(TypedExp), venv, gas)
		}
		return value.UnitVal{}, gas
	case ModuleLookupExp:
		exp := exp.(ModuleLookupExp)
		switch exp.ModId {
		case "Current":
			switch exp.FieldId {
			case "balance":
				return value.LambdaVal{value.CURRENT_BALANCE}, gas
			case "amount":
				return value.LambdaVal{value.CURRENT_AMOUNT}, gas
			case "gas":
				return value.LambdaVal{value.CURRENT_GAS}, gas
			case "failwith":
				return value.LambdaVal{value.CURRENT_FAILWITH}, gas
			default:
				return todo(22, gas), gas
			}
		case "Contract":
			switch exp.FieldId {
			case "call":
				return value.LambdaVal{value.CONTRACT_CALL}, gas
			default:
				return todo(23, gas), gas
			}
		case "Account":
			switch exp.FieldId {
			case "transfer":
				return value.LambdaVal{value.ACCOUNT_TRANSFER}, gas
			case "default":
				return value.LambdaVal{value.ACCOUNT_DEFAULT}, gas
			default:
				return todo(24, gas), gas
			}
		default:
			return todo(25, gas), gas
		}

	case LookupExp:
		exp := exp.(LookupExp)
		var structVal value.StructVal
		for i, id := range exp.PathIds {
			if i == 0 {
				structVal = lookupVar(id, venv).(value.StructVal)
			} else {
				structVal = structVal.Field[id].(value.StructVal)
			}
		}
		return structVal.Field[exp.LeafId], gas
	case UpdateStructExp:
		exp := exp.(UpdateStructExp)
		struc := lookupVar(exp.Root, venv)
		innerStruct := struc
		path := exp.Path
		for len(path) > 1 {
			innerStruct = innerStruct.(value.StructVal).Field[path[0]]
			path = path[1:]
		}
		newval, gas := interpret(exp.Exp.(TypedExp), venv, gas)
		innerStruct.(value.StructVal).Field[path[0]] = newval
		return struc, gas
	case StorageInitExp:
		return todo(26, gas), gas
	default:
		return todo(27, gas), gas
	}
}
