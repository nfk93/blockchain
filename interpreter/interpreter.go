package interpreter

import (
	"fmt"
	. "github.com/nfk93/blockchain/interpreter/ast"
)

func todo() int {
	fmt.Println("Not implemented yet")
	return 0
}

type InterpreterStruct struct {
	Field map[string]interface{}
}

type Tuple struct {
	Values []interface{}
}

func InterpretContractCall(texp TypedExp, params []interface{}, entry string, stor []interface{}) ([]Operation, interface{}) {
	exp := texp.Exp.(TopLevel)
	venv, tenv, senv := GenInitEnvs()
	for _, e := range exp.Roots {
		switch e.(type) {
		case TypeDecl:

		case EntryExpression:
			e := e.(EntryExpression)
			if e.Id == entry {
				venv, err := applyParams(params, e.Params, venv)
				if err != nil {
					return []Operation{failwith(err.Error())}, nil
				} else {
					bodyTuple := interpret(e.Body.(TypedExp), venv, tenv, senv).(Tuple)
					return bodyTuple.Values[0].([]Operation), bodyTuple.Values[1]
				}
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
		val, ok := param.(TupleValue)
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
		if len(val.Fields) != len(structtype.Fields) {
			return false
		}
		for i, field := range val.Fields {
			ok = ok && field.Id == structtype.Fields[i].Id && checkParam(field.Value, structtype.Fields[i].Typ)
		}
		return ok
	default:
		return false
	}
}

func failwith(str string) FailWith {
	return FailWith{str}
}

func createStruct() InterpreterStruct {
	return InterpreterStruct{make(map[string]interface{})}
}

func interpret(texp TypedExp, venv VarEnv, tenv TypeEnv, senv StructEnv) interface{} {
	exp := texp.Exp
	switch exp.(type) {
	case BinOpExp:
		exp := exp.(BinOpExp)
		switch exp.Op {
		case PLUS: // TODO: cast everything correctly
			leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
			rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
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
			leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
			rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
			switch exp.Left.(TypedExp).Type.Type() {
			case INT:
				return IntVal{leftval.(IntVal).Value - rightval.(IntVal).Value}
			case NAT:
				return IntVal{int64(leftval.(NatVal).Value - rightval.(NatVal).Value)}
			case KOIN:
				return KoinVal{leftval.(KoinVal).Value - rightval.(KoinVal).Value}
			default:
				return todo()
			}
		case TIMES:
			leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
			rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
			switch texp.Type.Type() {
			case INT:
				return IntVal{leftval.(IntVal).Value * rightval.(IntVal).Value}
			case NAT:
				return NatVal{leftval.(uint64) * rightval.(uint64)}
			case KOIN:
				return KoinVal{leftval.(float64) * rightval.(float64)}
			default:
				return todo()
			}
		case DIVIDE:
			return todo()
		case EQ:
			return todo()
		case NEQ:
			return todo()
		case GEQ:
			return todo()
		case LEQ:
			return todo()
		case LT:
			return todo()
		case GT:
			return todo()
		case AND:
			return todo()
		case OR:
			return todo()
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
		fieldlist := make([]StructFieldVal, len(exp.Vals))
		for i, id := range exp.Ids {
			fieldval := interpret(exp.Vals[i].(TypedExp), venv, tenv, senv)
			fieldlist = append(fieldlist, StructFieldVal{id, fieldval})
		}
		return StructVal{fieldlist}
	case ListLit:
		exp := exp.(ListLit)
		if len(exp.List) == 0 {
			return nil
		}
		var returnlist []interface{}
		for _, e := range exp.List {
			returnlist = append(returnlist, interpret(e.(TypedExp), venv, tenv, senv))
		}
		return returnlist
	case ListConcat:
		return todo()
	case CallExp:
		return todo()
	case LetExp:
		return todo()
	case AnnoExp:
		return todo()
	case TupleExp:
		exp := exp.(TupleExp)
		var tupleValues []interface{}
		for _, e := range exp.Exps {
			interE := interpret(e.(TypedExp), venv, tenv, senv)
			tupleValues = append(tupleValues, interE)
		}
		return Tuple{tupleValues}
	case VarExp:
		return todo()
	case ExpSeq:
		return todo()
	case IfThenElseExp:
		exp := exp.(IfThenElseExp)
		condition := interpret(exp.If.(TypedExp), venv, tenv, senv).(bool)
		if condition {
			return interpret(exp.Then.(TypedExp), venv, tenv, senv).(bool)
		} else {
			return interpret(exp.Else.(TypedExp), venv, tenv, senv).(bool)
		}
	case IfThenExp:
		exp := exp.(IfThenExp)
		condition := interpret(exp.If.(TypedExp), venv, tenv, senv).(bool)
		if condition {
			return interpret(exp.Then.(TypedExp), venv, tenv, senv).(bool)
		}
		return nil
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
