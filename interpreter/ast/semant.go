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
		log.Fatal("compiler error, translateType case not matched")
		return NotImplementedType{}
	}
}

/*func actualType(typ Type, tenv TypeEnv) Type {
	switch typ.Type() {
	case DECLARED:
		typ := typ.(DeclaredType)
		actualtyp := lookupType(typ.TypId, tenv)
		if actualtyp == nil {
			return ErrorType{fmt.Sprintf("type %s is not declared", typ.TypId)}
		} else {
			return actualtyp
		}
	default:
		return typ
	}
} */

// ONLY CALL WITH ACTUAL TYPES, NOT DECLARED TYPES.
func checkTypesEqual(typ1, typ2 Type) bool {
	switch typ1.Type() {
	case STRING, INT, FLOAT, KEY, BOOL, KOIN, OPERATION, UNIT:
		return typ1.Type() == typ2.Type()
	case LIST:
		switch typ2.Type() {
		case LIST:
			typ1 := typ1.(ListType)
			typ2 := typ2.(ListType)
			return checkTypesEqual(typ1.Typ, typ2.Typ)
		default:
			return false
		}
	case TUPLE:
		switch typ2.Type() {
		case TUPLE:
			typ1 := typ1.(TupleType)
			typ2 := typ2.(TupleType)
			return checkTypesEqual(typ1.Typ1, typ2.Typ1) && checkTypesEqual(typ1.Typ2, typ2.Typ2)
		default:
			return false
		}
	case STRUCT:
		switch typ2.Type() {
		case STRUCT:
			typ1 := typ1.(StructType)
			typ2 := typ2.(StructType)
			return getStructFieldString(typ1) == getStructFieldString(typ2)
		default:
			return false
		}
	case -1:
		return false
	default:
		log.Println("checkTypesEqual case not matched")
		return false
	}
}

func getStructFieldString(structType StructType) string {
	str := ""
	for _, field := range structType.Fields {
		str = str + field.Id
	}
	return str
}

// matches the pattern p to the type typ, doing pattern matching if typ is a tuple, and returning an updated venv
func patternMatch(p Pattern, typ Type, venv VarEnv, tenv TypeEnv) (VarEnv, bool) {
	venv_ := venv
	typ = translateType(typ, tenv)
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
				roots = append(roots, texp)
			default:
				roots = append(roots, TypedExp{ErrorExpression{"can only have entries, typedecls and storageinits in toplevel"}, ErrorType{}})
			}
		}
		if storageDefined && storageInitialized && mainEntryDefined {
			return TypedExp{TopLevel{roots}, UnitType{}}, venv, tenv, senv // TODO use toplevel type?
		} else {
			return TypedExp{TopLevel{roots}, ErrorType{"toplevel error, must define storage, main entry, and initialize storage"}}, venv, tenv, senv
		}
	case BinOpExp:
		exp := exp.(BinOpExp)
		leftTyped, _, _, _ := addTypes(exp.Left, venv, tenv, senv)
		rightTyped, _, _, _ := addTypes(exp.Right, venv, tenv, senv)
		texp := BinOpExp{leftTyped, exp.Op, rightTyped}
		switch exp.Op {
		case EQ, NEQ, GEQ, LEQ, LT, GT:
			switch leftTyped.Type.Type() {
			case BOOL, INT, KOIN, STRING, KEY, NAT:
				break
			default:
				return TypedExp{texp,
						ErrorType{"Can't compare expressions of type " + leftTyped.Type.String()}},
					venv, tenv, senv
			}
			if leftTyped.Type == rightTyped.Type {
				return TypedExp{texp, NewBoolType()}, venv, tenv, senv
			} else {
				return TypedExp{texp, ErrorType{"Types of comparison are not equal"}},
					venv, tenv, senv
			}
		case PLUS:
			switch leftTyped.Type.Type() {
			case INT, KOIN, NAT:
				break
			default:
				return TypedExp{texp,
						ErrorType{"Can't add expressions of type " + leftTyped.Type.String()}},
					venv, tenv, senv
			}
			if leftTyped.Type == rightTyped.Type {
				return TypedExp{texp, leftTyped.Type}, venv, tenv, senv
			} else {
				return TypedExp{texp, ErrorType{"Types of plus or minus operation are not equal"}},
					venv, tenv, senv
			}

		case MINUS:
			switch leftTyped.Type.Type() {
			case INT, NAT:
				switch rightTyped.Type.Type() {
				case INT, NAT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv
				default:
					return TypedExp{texp, ErrorType{"Can't subtract " + rightTyped.Type.String() + " from " + leftTyped.Type.String()}},
						venv, tenv, senv
				}
			case KOIN:
				switch rightTyped.Type.Type() {
				case KOIN:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv
				default:
					return TypedExp{texp, ErrorType{"Can't subtract " + rightTyped.Type.String() + " from " + leftTyped.Type.String()}},
						venv, tenv, senv
				}
			default:
				return TypedExp{texp,
						ErrorType{"Can't subtract expressions of type " + leftTyped.Type.String()}},
					venv, tenv, senv
			}
		case TIMES:
			switch leftTyped.Type.Type() {
			case KOIN:
				switch rightTyped.Type.Type() {
				case NAT:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't multiply expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			case NAT:
				switch rightTyped.Type.Type() {
				case NAT:
					return TypedExp{texp, NewNatType()}, venv, tenv, senv
				case KOIN:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv
				case INT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't multiply expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			case INT:
				switch rightTyped.Type.Type() {
				case INT, NAT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't multiply expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			default:
				return TypedExp{texp,
						ErrorType{"Can't multiply expressions of type " + leftTyped.Type.String()}},
					venv, tenv, senv
			}
		case DIVIDE: // TODO make the returned type from division an option to account for divide by zero
			switch leftTyped.Type.Type() {
			case KOIN:
				switch rightTyped.Type.Type() {
				case KOIN:
					return TypedExp{texp, NewTupleType(NewNatType(), NewKoinType())}, venv, tenv, senv
				case NAT:
					return TypedExp{texp, NewTupleType(NewKoinType(), NewKoinType())}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			case NAT:
				switch rightTyped.Type.Type() {
				case INT:
					return TypedExp{texp, NewTupleType(NewIntType(), NewNatType())}, venv, tenv, senv
				case NAT:
					return TypedExp{texp, NewTupleType(NewNatType(), NewNatType())}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			case INT:
				switch rightTyped.Type.Type() {
				case NAT, INT:
					return TypedExp{texp, NewTupleType(NewIntType(), NewNatType())}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			default:
				return TypedExp{texp,
						ErrorType{"Can't divide expressions of type " + leftTyped.Type.String()}},
					venv, tenv, senv
			}
		}
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
				return TypedExp{ErrorExpression{}, ErrorType{"unannotated entry parameter type can't be inferred"}}, venv, tenv, senv
			}
			venv_ = venv_.Set(v.id, translateType(v.anno.typ, tenv))
		}
		// check that storage pattern matches storage type
		storagetype := lookupType("storage", tenv)
		if storagetype == nil {
			return TypedExp{ErrorExpression{}, ErrorType{"storage type is undefined - define it before declaring entrypoints"}}, venv, tenv, senv
		}
		venv_, ok := patternMatch(exp.Storage, storagetype, venv, tenv)
		if !ok {
			return TypedExp{ErrorExpression{}, ErrorType{"storage pattern doesn't match storage type"}}, venv, tenv, senv
		}
		// add types with updated venv
		body, _, _, _ := addTypes(exp.Body, venv_, tenv, senv)
		// check that return type is operation list * storage
		if !checkTypesEqual(body.Type, TupleType{OperationType{}, storagetype}) {
			return TypedExp{body, ErrorType{fmt.Sprintf("return type of entry must be operation list * storage, but was %s", body.Type.String())}}, venv, tenv, senv
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
		exp := exp.(ListLit)
		var texplist []Exp
		if len(exp.List) == 0 {
			return TypedExp{exp, NewListType(UnitType{})}, venv, tenv, senv
		}
		var listtype Type
		for _, e := range exp.List {
			typedE, _, _, _ := addTypes(e, venv, tenv, senv)
			if listtype == nil {
				listtype = typedE.Type
			} else if listtype != typedE.Type {
				return TypedExp{exp,
						ErrorType{"All elements in list must be of same type"}},
					venv, tenv, senv
			}
			texplist = append(texplist, typedE)
		}
		return TypedExp{ListLit{texplist}, ListType{listtype}}, venv, tenv, senv
	case ListConcat:
		exp := exp.(ListConcat)
		tconcatexp, _, _, _ := addTypes(exp.Exp, venv, tenv, senv)
		tlistexp, _, _, _ := addTypes(exp.List, venv, tenv, senv)
		texp := ListConcat{tconcatexp, tlistexp}
		var listtype Type
		if tlistexp.Type.Type() != LIST {
			return TypedExp{texp,
					ErrorType{"Cannot concatenate with type " + tlistexp.Type.String() + " . Should be a list. "}},
				venv, tenv, senv
		} else {
			listtype = tlistexp.Type.(ListType).Typ
		}

		if tconcatexp.Type.Type() != listtype.Type() {
			return TypedExp{texp,
					ErrorType{"Cannot concatenate type " + tconcatexp.Type.String() + " with list of type " + listtype.String()}},
				venv, tenv, senv
		}
		return TypedExp{texp, ListType{listtype}}, venv, tenv, senv

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
