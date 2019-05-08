package interpreter

import (
	"fmt"
	. "github.com/nfk93/blockchain/interpreter/ast"
)

func todo() int {
	fmt.Println("Not implemented yet")
	return 0
}

func InterpretContractCall(texp TypedExp, params []Value, entry string, stor []Value) ([]Operation, Value) {
	exp := texp.Exp.(TopLevel)
	venv, tenv, senv := GenInitEnvs()
	for _, e := range exp.Roots {
		e := e.(TypedExp).Exp
		switch e.(type) {
		case TypeDecl:
		case EntryExpression:
			e := e.(EntryExpression)
			if e.Id == entry {
				// apply params to venv
				venv, err := applyParams(params, e.Params, venv)
				if err != nil {
					return []Operation{failwith(err.Error())}, nil
				}
				// apply storage to venv
				venv, err = applyParams(stor, e.Storage, venv)
				if err != nil {
					return []Operation{failwith("storage doesn't match storage type definition")}, nil
				}
				bodyTuple := interpret(e.Body.(TypedExp), venv, tenv, senv).(TupleVal)
				opvallist := bodyTuple.Values[0].(ListVal).Values
				var oplist []Operation
				for _, v := range opvallist {
					oplist = append(oplist, v.(OperationVal).Value)
				}
				return oplist, bodyTuple.Values[1]
			}
		}
	}
	return nil, 1 // TODO this is just a dummy return Value
}
func applyParams(paramVals []interface{}, pattern Pattern, venv VarEnv) (VarEnv, error) {
	venv_ := venv
	for i, param := range pattern.Params {
		if checkParam(paramVals[i], param.Anno.Typ) {
			venv_ = venv_.Set(param.Id, paramVals[i])
		} else {
			return venv, fmt.Errorf("parameter mismatch, can't match given parameters to entry")
		}
	}
	return venv_, nil
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
		i := 0
		for fieldname, value := range val.Field {
			ok = ok && fieldname == structtype.Fields[i].Id && checkParam(value, structtype.Fields[i].Typ)
			i++
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
	return StructVal{make(map[string]Value)}
}

func interpret(texp TypedExp, venv VarEnv, tenv TypeEnv, senv StructEnv) interface{} {
	exp := texp.Exp
	switch exp.(type) {
	case BinOpExp:
		exp := exp.(BinOpExp)
		leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
		rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
		switch exp.Op {
		case PLUS:
			switch exp.Left.(TypedExp).Type.Type() {
			case KOIN:
				return KoinVal{leftval.(KoinVal).Value + rightval.(KoinVal).Value}
			case NAT:
				return NatVal{leftval.(NatVal).Value + rightval.(NatVal).Value}
			case INT:
				return IntVal{leftval.(IntVal).Value + rightval.(IntVal).Value}
			default:
				return todo()
			}
		case MINUS:
			switch exp.Left.(TypedExp).Type.Type() {
			case INT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					return IntVal{leftval.(IntVal).Value - rightval.(IntVal).Value}
				case NAT:
					return IntVal{leftval.(IntVal).Value - int64(rightval.(NatVal).Value)}
				default:
					return todo()
				}
			case NAT:
				switch exp.Right.(TypedExp).Type.Type() {
				case INT:
					return IntVal{int64(leftval.(NatVal).Value) - rightval.(IntVal).Value}
				case NAT:
					return IntVal{int64(leftval.(NatVal).Value) - int64(rightval.(NatVal).Value)}
				default:
					return todo()
				}
			case KOIN:
				return KoinVal{leftval.(KoinVal).Value - rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case TIMES:
			switch texp.Type.Type() {
			case INT:
				return IntVal{leftval.(IntVal).Value * rightval.(IntVal).Value}
			case NAT:
				return NatVal{leftval.(NatVal).Value * rightval.(NatVal).Value}
			case KOIN:
				return KoinVal{leftval.(KoinVal).Value * rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case DIVIDE:
			return todo()
		case EQ:
			return BoolVal{leftval == rightval}
		case NEQ:
			return BoolVal{leftval == rightval}
		case GEQ:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value >= rightval.(NatVal).Value}
			case INT:
				return BoolVal{leftval.(IntVal).Value >= rightval.(IntVal).Value}
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value >= rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case LEQ:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value <= rightval.(NatVal).Value}
			case INT:
				return BoolVal{leftval.(IntVal).Value <= rightval.(IntVal).Value}
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value <= rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case LT:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value < rightval.(NatVal).Value}
			case INT:
				return BoolVal{leftval.(IntVal).Value < rightval.(IntVal).Value}
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value < rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case GT:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return BoolVal{leftval.(NatVal).Value > rightval.(NatVal).Value}
			case INT:
				return BoolVal{leftval.(IntVal).Value > rightval.(IntVal).Value}
			case KOIN:
				return BoolVal{leftval.(KoinVal).Value > rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case AND:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return NatVal{leftval.(NatVal).Value & rightval.(NatVal).Value}
			case BOOL:
				return BoolVal{leftval.(BoolVal).Value && rightval.(BoolVal).Value}
			default:
				return todo()
			}
		case OR:
			switch exp.Right.(TypedExp).Type.Type() {
			case NAT:
				return NatVal{leftval.(NatVal).Value | rightval.(NatVal).Value}
			case BOOL:
				return BoolVal{leftval.(BoolVal).Value || rightval.(BoolVal).Value}
			default:
				return todo()
			}
		default:
			return todo()
		}
	case TypeDecl:
		return todo()
	case KeyLit:
		exp := exp.(KeyLit)
		return KeyVal{exp.Key}
	case BoolLit:
		exp := exp.(BoolLit)
		return BoolVal{exp.Val}
	case IntLit:
		exp := exp.(IntLit)
		return IntVal{exp.Val}
	case KoinLit:
		exp := exp.(KoinLit)
		return KoinVal{exp.Val}
	case StringLit:
		exp := exp.(StringLit)
		return StringVal{exp.Val}
	case UnitLit:
		return UnitVal{}
	case StructLit:
		exp := exp.(StructLit)
		newStruct := createStruct()
		for i, id := range exp.Ids {
			newStruct.Field[id] = interpret(exp.Vals[i].(TypedExp), venv, tenv, senv)
		}
		return newStruct
	case ListLit:
		exp := exp.(ListLit)
		if len(exp.List) == 0 {
			return ListVal{make([]Value, 0)}
		}
		var returnlist []Value
		for _, e := range exp.List {
			returnlist = append(returnlist, interpret(e.(TypedExp), venv, tenv, senv))
		}
		return ListVal{returnlist}
	case ListConcat:
		exp := exp.(ListConcat)
		e := interpret(exp.Exp.(TypedExp), venv, tenv, senv)
		list := interpret(exp.Exp.(TypedExp), venv, tenv, senv).(ListVal)
		return ListVal{append(list.Values, e)}
	case CallExp:
		return todo()
	case LetExp:
		return todo()
	case AnnoExp:
		exp := exp.(AnnoExp)
		return interpret(exp.Exp.(TypedExp), venv, tenv, senv)
	case TupleExp:
		exp := exp.(TupleExp)
		var tupleValues []Value
		for _, e := range exp.Exps {
			interE := interpret(e.(TypedExp), venv, tenv, senv)
			tupleValues = append(tupleValues, interE)
		}
		return TupleVal{tupleValues}
	case VarExp:
		return todo()
	case ExpSeq:
		return todo()
	case IfThenElseExp:
		exp := exp.(IfThenElseExp)
		condition := interpret(exp.If.(TypedExp), venv, tenv, senv).(BoolVal).Value
		if condition {
			return interpret(exp.Then.(TypedExp), venv, tenv, senv).(BoolVal).Value
		} else {
			return interpret(exp.Else.(TypedExp), venv, tenv, senv).(BoolVal).Value
		}
	case IfThenExp:
		exp := exp.(IfThenExp)
		condition := interpret(exp.If.(TypedExp), venv, tenv, senv).(BoolVal).Value
		if condition {
			return interpret(exp.Then.(TypedExp), venv, tenv, senv).(BoolVal).Value
		}
		return UnitVal{}
	case ModuleLookupExp:
		return todo()
	case LookupExp:
		return todo()
	case UpdateStructExp:
		return todo()
	case StorageInitExp:
		return todo()
	default:
		return todo()
	}
}
