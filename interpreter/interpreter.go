package interpreter

import (
	"fmt"
	. "github.com/nfk93/blockchain/interpreter/ast"
	. "github.com/nfk93/blockchain/objects"
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

func InterpretContractCall(texp TypedExp, param Parameter, entry string, stor Storage) ([]Operation, Storage) {
	exp := texp.Exp.(TopLevel)
	venv, tenv, senv := GenInitEnvs()
	for _, e := range exp.Roots {
		switch e.(type) {
		case TypeDecl:

		case EntryExpression:
			e := e.(EntryExpression)
			if e.Id == entry { //TODO add parameters to VarEnv
				bodyTuple := interpret(e.Body.(TypedExp), venv, tenv, senv).(Tuple)
				return bodyTuple.Values[0].([]Operation), bodyTuple.Values[1]
			}
		}
	}
	return nil, 1 // TODO this is just a dummy return value
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
		case PLUS:
			leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
			rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
			switch exp.Left.(TypedExp).Type.Type() {
			case KOIN:
				return leftval.(float64) + rightval.(float64)
			case NAT:
				return leftval.(uint64) + rightval.(uint64)
			case INT:
				return leftval.(int64) + rightval.(int64)
			default:
				return todo()
			}
		case MINUS:
			leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
			rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
			switch exp.Left.(TypedExp).Type.Type() {
			case INT, NAT:
				return leftval.(int64) - rightval.(int64)
			case KOIN:
				return leftval.(float64) - rightval.(float64)
			default:
				return todo()
			}
		case TIMES:
			leftval := interpret(exp.Left.(TypedExp), venv, tenv, senv)
			rightval := interpret(exp.Right.(TypedExp), venv, tenv, senv)
			switch texp.Type.Type() {
			case INT:
				return leftval.(int64) * rightval.(int64)
			case NAT:
				return leftval.(uint64) * rightval.(uint64)
			case KOIN:
				return leftval.(float64) * rightval.(float64)
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
		return exp.Key
	case BoolLit:
		exp := exp.(BoolLit)
		return exp.Val
	case IntLit:
		exp := exp.(IntLit)
		return exp.Val
	case KoinLit:
		exp := exp.(KoinLit)
		return exp.Val
	case StringLit:
		exp := exp.(StringLit)
		return exp.Val
	case UnitLit:
		return nil
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
