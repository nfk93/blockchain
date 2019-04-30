package ast

import (
	"fmt"
	"github.com/mndrix/ps"
	"log"
)

type TypeEnv ps.Map   // TODO
type VarEnv ps.Map    // TODO
type StructEnv ps.Map // TODO

type TypedExp struct {
	Exp  Exp
	Type Type
}

func (e TypedExp) String() string {
	return fmt.Sprintf("{exp: %s, typ: %s}", e.Exp.String(), e.Type.String())
}

func InitialTypeEnv() TypeEnv {
	return ps.NewMap() // TODO
}

func InitialVarEnv() VarEnv {
	return ps.NewMap() // TODO
}

func InitialStructEnv() StructEnv {
	return ps.NewMap() // TODO
}

func todo(exp Exp, venv VarEnv, tenv TypeEnv, senv StructEnv) (TypedExp, VarEnv, TypeEnv, StructEnv) {
	return TypedExp{exp, NotImplementedType{}}, venv, tenv, senv
}

func AddTypes(exp Exp) TypedExp {
	texp, _, _, _ := addTypes(exp, InitialVarEnv(), InitialTypeEnv(), InitialStructEnv())
	return texp
}

func lookupType(id string, tenv TypeEnv) Type {
	val, contained := tenv.Lookup(id)
	if contained {
		return val.(Type)
	} else {
		return nil
	}
}

func lookupVar(id string, venv VarEnv) Type {
	val, contained := venv.Lookup(id)
	if contained {
		return val.(Type)
	} else {
		return nil
	}
}

func translateType(typ Type, tenv TypeEnv) Type {
	switch typ.Type() {
	case STRING, INT, FLOAT, KEY, BOOL, KOIN, OPERATION, UNIT:
		return typ
	case LIST:
		typ := typ.(ListType)
		return ListType{translateType(typ.Typ, tenv)}
	case TUPLE:
		typ := typ.(TupleType)
		typ1 := translateType(typ.Typ1, tenv)
		typ2 := translateType(typ.Typ2, tenv)
		return TupleType{typ1, typ2}
	case STRUCT:
		typ := typ.(StructType)
		fields := make([]StructField, 0)
		for _, field := range typ.Fields {
			fieldtyp := translateType(field.Typ, tenv)
			fields = append(fields, StructField{field.Id, fieldtyp})
		}
		return StructType{fields}
	case DECLARED:
		typ := typ.(DeclaredType)
		actualtype := lookupType(typ.TypId, tenv)
		if actualtype != nil {
			return translateType(actualtype, tenv)
		} else {
			return ErrorType{fmt.Sprintf("type %s is not declared", typ.TypId)}
		}

	default:
		log.Fatal("SHOULD NOT HAPPEN")
		return NotImplementedType{}
	}
}

func checkTypesEqual(typ1, typ2 Type) bool {
	switch typ1 {

	}
}

func getStructFieldString(structType StructType) string {
	str := ""
	for _, field := range structType.Fields {
		str = str + field.Id
	}
	return str
}

func patternMatch(p Pattern, typ Type, venv VarEnv) (VarEnv, bool) {
	venv_ := venv
	switch typ.Type() {
	case TUPLE:
		typ := typ.(TupleType)
		types := unpackTuple(typ, []Type{typ.Typ1})
		if len(p.params) != len(types) {
			return venv, false
		}
		for i, v := range p.params {
			if !checkAnnotation(v, types[i]) {
				return venv, false
			}
			venv_ = venv_.Set(v.id, types[i])
		}
		return venv_, true
	default:
		if len(p.params) != 1 {
			return venv, false
		}
		return venv_.Set(p.params[0].id, typ), true
	}
}
func unpackTuple(typ TupleType, types []Type) []Type {
	switch typ.Typ2.Type() {
	case TUPLE:
		typ2 := typ.Typ2.(TupleType)
		return unpackTuple(typ2, append(types, typ2.Typ1))
	default:
		return append(types, typ.Typ2)
	}
}

func checkAnnotation(param Param, typ Type) bool {
	if param.anno.opt {
		return param.anno.typ.Type() == typ.Type()
	} else {
		return true
	}
}

func addTypes(
	exp Exp,
	venv VarEnv,
	tenv TypeEnv,
	senv StructEnv,
) (TypedExp, VarEnv, TypeEnv, StructEnv) {

	switch exp.(type) {
	case TopLevel:
		exp := exp.(TopLevel)
		roots := make([]Exp, 0)
		var texp TypedExp
		var storageDefined, storageInitialized, mainEntryDefined bool
		for _, exp1 := range exp.Roots {
			switch exp1.(type) {
			case TypeDecl:
				typedecl := exp1.(TypeDecl)
				texp, venv, tenv, senv = addTypes(exp1, venv, tenv, senv)
				roots = append(roots, texp)
				if typedecl.id == "storage" {
					storageDefined = true
				}
			case EntryExpression:
				entryexpression := exp1.(EntryExpression)
				texp, venv, tenv, senv = addTypes(exp1, venv, tenv, senv)
				roots = append(roots, texp)
				if entryexpression.Id == "main" {
					mainEntryDefined = true
				}
			case StorageInitExp:
				storageInitialized = true
				texp, venv, tenv, senv = addTypes(exp1, venv, tenv, senv)
			default:
				return todo(exp, venv, tenv, senv)
			}
		}
		if storageDefined && storageInitialized && mainEntryDefined {
			return TypedExp{TopLevel{roots}, UnitType{}}, venv, tenv, senv // TODO use toplevel type?
		} else {
			return todo(exp, venv, tenv, senv)
		}
	case BinOpExp:
		return todo(exp, venv, tenv, senv)
	case TypeDecl:
		exp := exp.(TypeDecl)
		if lookupType(exp.id, tenv) != nil {
			return TypedExp{exp, ErrorType{fmt.Sprintf("type %s already declared", exp.id)}},
				venv, tenv, senv
		}
		actualType := translateType(exp.typ, tenv)
		switch exp.typ.Type() {
		case STRUCT:
			actualType := actualType.(StructType)
			_, contains := senv.Lookup(getStructFieldString(actualType))
			if contains {
				return TypedExp{TypeDecl{exp.id, actualType}, ErrorType{fmt.Sprintf("struct field names already used")}},
					venv, tenv, senv
			} else {
				tenv_ := tenv.Set(exp.id, actualType)
				return TypedExp{TypeDecl{exp.id, actualType}, UnitType{}}, venv, tenv_, senv // TODO perhaps use decl type
			}
		default:
			tenv_ := tenv.Set(exp.id, exp.typ)
			return TypedExp{TypeDecl{exp.id, actualType}, UnitType{}}, venv, tenv_, senv
		}
	case EntryExpression:
		exp := exp.(EntryExpression)
		// check that parameters are typeannotated and add them to variable environment
		venv_ := venv
		for _, v := range exp.Params.params {
			if v.anno.opt != true {
				return TypedExp{exp, ErrorType{"unannotated entry parameter type can't be inferred"}}, venv, tenv, senv
			}
			venv_ = venv_.Set(v.id, v.anno.typ)
		}
		// check that storage pattern matches storage type
		storagetype := lookupType("storage", tenv)
		if storagetype == nil {
			return TypedExp{exp, ErrorType{"storage type is undefined - define it before declaring entrypoints"}}, venv, tenv, senv
		}
		venv_, ok := patternMatch(exp.Storage, storagetype, venv)
		if !ok {
			return TypedExp{exp, ErrorType{"storage pattern doesn't match storage type"}}, venv, tenv, senv
		}
		// add types with updated venv
		body, _, _, _ := addTypes(exp.Body, venv_, tenv, senv)
		// check that return type is operation list * storage
		if !checkTypesEqual(body.Type, storagetype) {
			return TypedExp{exp, ErrorType{"return type of entry must be operation list * storage"}}, venv, tenv, senv
		}
		return TypedExp{EntryExpression{exp.Id, exp.Params, exp.Storage, body}, storagetype}, venv, tenv, senv
	case KeyLit:
		return TypedExp{exp, KeyType{}}, venv, tenv, senv
	case BoolLit:
		return TypedExp{exp, BoolType{}}, venv, tenv, senv
	case IntLit:
		return TypedExp{exp, IntType{}}, venv, tenv, senv
	case FloatLit:
		return TypedExp{exp, FloatType{}}, venv, tenv, senv
	case KoinLit:
		return TypedExp{exp, KoinType{}}, venv, tenv, senv
	case StringLit:
		return TypedExp{exp, StringType{}}, venv, tenv, senv
	case UnitLit:
		return TypedExp{exp, UnitType{}}, venv, tenv, senv
	case StructLit:
		return todo(exp, venv, tenv, senv)
	case ListLit:
		return todo(exp, venv, tenv, senv)
	case ListConcat:
		return todo(exp, venv, tenv, senv)
	case CallExp:
		return todo(exp, venv, tenv, senv)
	case LetExp:
		return todo(exp, venv, tenv, senv)
	case AnnoExp:
		return todo(exp, venv, tenv, senv)
	case TupleExp:
		return todo(exp, venv, tenv, senv)
	case VarExp:
		return todo(exp, venv, tenv, senv)
	case ExpSeq:
		return todo(exp, venv, tenv, senv)
	case IfThenElseExp:
		return todo(exp, venv, tenv, senv)
	case IfThenExp:
		return todo(exp, venv, tenv, senv)
	case ModuleLookupExp:
		return todo(exp, venv, tenv, senv)
	case LookupExp:
		return todo(exp, venv, tenv, senv)
	case UpdateStructExp:
		return todo(exp, venv, tenv, senv)
	case StorageInitExp:
		return todo(exp, venv, tenv, senv)
	default:
		return todo(exp, venv, tenv, senv)
	}
}
