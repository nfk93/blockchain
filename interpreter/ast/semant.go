package ast

import "fmt"

type TypeEnv int // TODO
type VarEnv int  // TODO

type TypedExp struct {
	Exp  Exp
	Type Type
}

func (e TypedExp) String() string {
	return fmt.Sprintf("{exp: %s, typ: %s}", e.Exp.String(), e.Type.String())
}

func InitialTypeEnv() TypeEnv {
	return -1 // TODO
}

func InitialVarEnv() VarEnv {
	return -1 // TODO
}

func todo(exp Exp) (TypedExp, VarEnv, TypeEnv) {
	return TypedExp{exp, NotImplementedType{}}, -1, -1
}

func AddTypes(exp Exp) TypedExp {
	texp, _, _ := addTypes(exp, InitialVarEnv(), InitialTypeEnv())
	return texp
}

func addTypes(exp Exp, venv VarEnv, tenv TypeEnv) (TypedExp, VarEnv, TypeEnv) {
	switch exp.(type) {
	case TopLevel:
		return todo(exp)
	case BinOpExp:
		return todo(exp)
	case TypeDecl:

		return todo(exp)
	case EntryExpression:
		return todo(exp)
	case KeyLit:
		return TypedExp{exp, KeyType{}}, venv, tenv
	case BoolLit:
		return TypedExp{exp, BoolType{}}, venv, tenv
	case IntLit:
		return TypedExp{exp, IntType{}}, venv, tenv
	case FloatLit:
		return TypedExp{exp, FloatType{}}, venv, tenv
	case KoinLit:
		return TypedExp{exp, KoinType{}}, venv, tenv
	case StringLit:
		return TypedExp{exp, StringType{}}, venv, tenv
	case UnitLit:
		return TypedExp{exp, UnitType{}}, venv, tenv
	case StructLit:
		return todo(exp)
	case ListLit:
		return todo(exp)
	case ListConcat:
		return todo(exp)
	case CallExp:
		return todo(exp)
	case LetExp:
		return todo(exp)
	case AnnoExp:
		return todo(exp)
	case TupleExp:
		return todo(exp)
	case VarExp:
		return todo(exp)
	case ExpSeq:
		return todo(exp)
	case IfThenElseExp:
		return todo(exp)
	case IfThenExp:
		return todo(exp)
	case ModuleLookupExp:
		return todo(exp)
	case LookupExp:
		return todo(exp)
	case UpdateStructExp:
		return todo(exp)
	case StorageInitExp:
		return todo(exp)
	default:
		return todo(exp)
	}
}
