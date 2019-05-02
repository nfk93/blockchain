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

func lookupStruct(fieldIds string, senv StructEnv) Type {
	val, contained := senv.Lookup(fieldIds)
	if contained {
		return val.(Type)
	} else {
		return nil
	}
}

func translateType(typ Type, tenv TypeEnv) Type {
	switch typ.Type() {
	case STRING, INT, FLOAT, KEY, BOOL, KOIN, OPERATION, UNIT, NAT:
		return typ
	case OPTION:
		typ := typ.(OptionType)
		return OptionType{translateType(typ.Typ, tenv)}
	case LIST:
		typ := typ.(ListType)
		return ListType{translateType(typ.Typ, tenv)}
	case TUPLE:
		typ := typ.(TupleType)
		typs := make([]Type, len(typ.Typs))
		for i, t := range typ.Typs {
			typs[i] = translateType(t, tenv)
		}
		return TupleType{typs}
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
	case ERROR, NOTIMPLEMENTED:
		return typ
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
	case STRING, INT, FLOAT, KEY, BOOL, KOIN, OPERATION, UNIT, NAT:
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
	case OPTION:
		switch typ2.Type() {
		case OPTION:
			typ1 := typ1.(OptionType)
			typ2 := typ2.(OptionType)
			return checkTypesEqual(typ1.Typ, typ2.Typ)
		default:
			return false
		}
	case TUPLE:
		switch typ2.Type() {
		case TUPLE:
			equal := true
			typ1 := typ1.(TupleType)
			typ2 := typ2.(TupleType)
			for i, v := range typ1.Typs {
				equal = equal && checkTypesEqual(v, typ2.Typs[i])
			}
			return equal
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
	case ERROR, NOTIMPLEMENTED:
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
		if len(p.params) == 1 {
			if !checkParamTypeAnno(p.params[0], typ, tenv) {
				return venv, false
			}
			return venv_.Set(p.params[0].id, typ), true
		}
		if len(p.params) != len(typ.Typs) {
			return venv, false
		}
		for i, v := range p.params {
			if !checkParamTypeAnno(v, typ.Typs[i], tenv) {
				return venv, false
			}
			venv_ = venv_.Set(v.id, typ.Typs[i])
		}
		return venv_, true
	default:
		if len(p.params) != 1 {
			return venv, false
		}
		return venv_.Set(p.params[0].id, typ), true
	}
}

func checkParamTypeAnno(param Param, typ Type, tenv TypeEnv) bool {
	if param.anno.opt {
		actualanno := translateType(param.anno.typ, tenv)
		return checkTypesEqual(actualanno, typ)
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
					return TypedExp{texp, NewTupleType([]Type{NewNatType(), NewKoinType()})}, venv, tenv, senv
				case NAT:
					return TypedExp{texp, NewTupleType([]Type{NewKoinType(), NewKoinType()})}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			case NAT:
				switch rightTyped.Type.Type() {
				case INT:
					return TypedExp{texp, NewTupleType([]Type{NewIntType(), NewNatType()})}, venv, tenv, senv
				case NAT:
					return TypedExp{texp, NewTupleType([]Type{NewNatType(), NewNatType()})}, venv, tenv, senv
				default:
					return TypedExp{texp,
							ErrorType{"Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()}},
						venv, tenv, senv
				}
			case INT:
				switch rightTyped.Type.Type() {
				case NAT, INT:
					return TypedExp{texp, NewTupleType([]Type{NewIntType(), NewNatType()})}, venv, tenv, senv
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
		case OR, AND:
			switch leftTyped.Type.Type() {
			case NAT, BOOL:
				break
			default:
				return TypedExp{texp,
						ErrorType{"Can't use logical binop on expressions of type " + leftTyped.Type.String()}},
					venv, tenv, senv
			}
			if leftTyped.Type == rightTyped.Type {
				return TypedExp{texp, leftTyped.Type}, venv, tenv, senv
			} else {
				return TypedExp{texp, ErrorType{"Types of logical binop are not equal"}},
					venv, tenv, senv
			}
		default:
			return TypedExp{texp,
					ErrorType{"Unrecogized Binop, Should not happen!"}},
				venv, tenv, senv
		}
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
			structfieldstring := getStructFieldString(actualType)
			_, contains := senv.Lookup(structfieldstring)
			if contains {
				return TypedExp{TypeDecl{exp.id, actualType}, ErrorType{fmt.Sprintf("struct field names already used")}},
					venv, tenv, senv
			} else {
				tenv_ := tenv.Set(exp.id, actualType)
				senv = senv.Set(structfieldstring, actualType)
				return TypedExp{TypeDecl{exp.id, actualType}, UnitType{}}, venv, tenv_, senv // TODO perhaps use decl type
			}
		default:
			tenv_ := tenv.Set(exp.id, actualType)
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
			vartyp := translateType(v.anno.typ, tenv)
			venv_ = venv_.Set(v.id, vartyp)
		}
		// check that storage pattern matches storage type
		storagetype := lookupType("storage", tenv)
		if storagetype == nil {
			return TypedExp{ErrorExpression{}, ErrorType{"storage type is undefined - define it before declaring entrypoints"}}, venv, tenv, senv
		}
		venv_, ok := patternMatch(exp.Storage, storagetype, venv_, tenv)
		if !ok {
			return TypedExp{EntryExpression{exp.Id, exp.Params, exp.Storage, ErrorExpression{}}, ErrorType{"storage pattern doesn't match storage type"}}, venv, tenv, senv
		}
		// add types with updated venv
		body, _, _, _ := addTypes(exp.Body, venv_, tenv, senv)
		// check that return type is operation list * storage
		if !checkTypesEqual(body.Type, TupleType{[]Type{NewListType(OperationType{}), storagetype}}) {
			return TypedExp{EntryExpression{exp.Id, exp.Params, exp.Storage, body},
					ErrorType{fmt.Sprintf("return type of entry must be operation list * %s, but was %s", storagetype.String(), body.Type.String())}},
				venv, tenv, senv
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
	case NatLit:
		return TypedExp{exp, NatType{}}, venv, tenv, senv
	case StructLit:
		exp := exp.(StructLit)
		definedStruct := lookupStruct(exp.FieldString(), senv)
		if definedStruct == nil {
			return TypedExp{exp,
					ErrorType{"No struct type is defined with the given field names " + exp.FieldString()}},
				venv, tenv, senv
		} else {
			definedStruct := definedStruct.(StructType)
			var texplist []Exp
			for i, e := range exp.Vals {
				typedE, _, _, _ := addTypes(e, venv, tenv, senv)
				field := definedStruct.Fields[i]
				if !checkTypesEqual(typedE.Type, field.Typ) {
					return TypedExp{exp,
							ErrorType{fmt.Sprintf("Field %s expected %s but received %s", field.Id, field.Typ.String(), typedE.Type.String())}},
						venv, tenv, senv
				}
				texplist = append(texplist, typedE)
			}
			texp := StructLit{exp.Ids, texplist}
			return TypedExp{texp, definedStruct}, venv, tenv, senv
		}
	case ListLit:
		exp := exp.(ListLit)
		var texplist []Exp
		if len(exp.List) == 0 {
			return TypedExp{exp, NewListType(UnitType{})}, venv, tenv, senv
		}
		var listtype Type
		var typesNotEqual bool
		for _, e := range exp.List {
			typedE, _, _, _ := addTypes(e, venv, tenv, senv)
			if listtype == nil {
				listtype = typedE.Type
			} else if !checkTypesEqual(listtype, typedE.Type) {
				typesNotEqual = true

			}
			texplist = append(texplist, typedE)
		}
		if typesNotEqual {
			return TypedExp{ListLit{texplist},
					ErrorType{"All elements in list must be of same type"}},
				venv, tenv, senv
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
		if listtype.Type() == UNIT {
			listtype = tconcatexp.Type
		}
		if !checkTypesEqual(tconcatexp.Type, listtype) {
			return TypedExp{texp,
					ErrorType{"Cannot concatenate type " + tconcatexp.Type.String() + " with list of type " + listtype.String()}},
				venv, tenv, senv
		}
		return TypedExp{texp, ListType{listtype}}, venv, tenv, senv

	case CallExp:
		return todo(exp, venv, tenv, senv)
	case LetExp:
		exp := exp.(LetExp)
		defexp, _, _, _ := addTypes(exp.DefExp, venv, tenv, senv)
		venv_, ok := patternMatch(exp.Patt, defexp.Type, venv, tenv)
		if !ok {
			return TypedExp{ErrorExpression{}, ErrorType{
				fmt.Sprintf("variable declaration pattern %s can't be matched to type %s", exp.Patt.String(),
					defexp.Type.String())}}, venv, tenv, senv
		}
		inexp, _, _, _ := addTypes(exp.InExp, venv_, tenv, senv)
		return TypedExp{LetExp{exp.Patt, defexp, inexp}, inexp.Type}, venv, tenv, senv
	case AnnoExp:
		exp := exp.(AnnoExp)
		texp, venv, tenv, senv := addTypes(exp.Exp, venv, tenv, senv)
		actualAnno := translateType(exp.Anno, tenv)
		if actualAnno.Type() == LIST {
			if texp.Type.Type() == LIST && texp.Type.(ListType).Typ.Type() == UNIT {
				return TypedExp{AnnoExp{texp, actualAnno}, actualAnno}, venv, tenv, senv
			}
		}
		typesEqual := checkTypesEqual(texp.Type, actualAnno)
		if !typesEqual {
			return TypedExp{ErrorExpression{}, ErrorType{"expression type doesn't match annotated type"}}, venv, tenv, senv
		}
		return TypedExp{AnnoExp{texp, actualAnno}, actualAnno}, venv, tenv, senv
	case TupleExp:
		exp := exp.(TupleExp)
		var texplist []Exp
		var typelist []Type
		for _, e := range exp.Exps {
			typedE, _, _, _ := addTypes(e, venv, tenv, senv)
			texplist = append(texplist, typedE)
			typelist = append(typelist, typedE.Type)
		}
		texp := TupleExp{texplist}
		return TypedExp{texp, NewTupleType(typelist)}, venv, tenv, senv
	case VarExp:
		exp := exp.(VarExp)
		vartyp, ok := venv.Lookup(exp.Id)
		if !ok {
			return TypedExp{exp, ErrorType{fmt.Sprintf("variable %s used but not defined", exp.Id)}}, venv, tenv, senv
		}
		return TypedExp{exp, vartyp.(Type)}, venv, tenv, senv
	case ExpSeq:
		exp := exp.(ExpSeq)
		typedLeftExp, _, _, _ := addTypes(exp.Left, venv, tenv, senv)
		typedRightExp, _, _, _ := addTypes(exp.Right, venv, tenv, senv)
		texp := ExpSeq{typedLeftExp, typedRightExp}
		if typedLeftExp.Type.Type() != UNIT {
			return TypedExp{texp,
					ErrorType{"All expresssion in expseq_semant, except the last, must be of type UNIT!"}},
				venv, tenv, senv
		}
		return TypedExp{texp, typedRightExp.Type}, venv, tenv, senv
	case IfThenElseExp:
		exp := exp.(IfThenElseExp)
		typedIf, _, _, _ := addTypes(exp.If, venv, tenv, senv)
		typedThen, _, _, _ := addTypes(exp.Then, venv, tenv, senv)
		typedElse, _, _, _ := addTypes(exp.Else, venv, tenv, senv)
		texp := IfThenElseExp{typedIf, typedThen, typedElse}
		if typedIf.Type.Type() != BOOL {
			return TypedExp{texp,
					ErrorType{"Condition in If is of type " + typedIf.Type.String() + " should be BOOL"}},
				venv, tenv, senv
		}
		if !checkTypesEqual(typedThen.Type, typedElse.Type) {
			return TypedExp{texp,
					ErrorType{"Return types in if and else branch should be equal!"}},
				venv, tenv, senv
		}
		return TypedExp{texp, typedThen.Type}, venv, tenv, senv
	case IfThenExp:
		exp := exp.(IfThenExp)
		typedIf, _, _, _ := addTypes(exp.If, venv, tenv, senv)
		typedThen, _, _, _ := addTypes(exp.Then, venv, tenv, senv)
		texp := IfThenExp{typedIf, typedThen}
		if typedIf.Type.Type() != BOOL {
			return TypedExp{texp,
					ErrorType{"Condition in If is of type " + typedIf.Type.String() + " should be BOOL"}},
				venv, tenv, senv
		}
		if typedThen.Type.Type() != UNIT {
			return TypedExp{texp,
					ErrorType{"'Then' expression in IfThen is of type " + typedThen.Type.String() + " should be UNIT"}},
				venv, tenv, senv
		}
		return TypedExp{texp, UnitType{}}, venv, tenv, senv
	case ModuleLookupExp:
		return todo(exp, venv, tenv, senv)
	case LookupExp:
		exp := exp.(LookupExp)
		var typ Type
		var currentStruct StructType
		for i, id := range exp.PathIds {
			if i == 0 {
				typ = lookupVar(id, venv)
			} else {
				fieldType, exists := currentStruct.FindFieldType(id)
				if exists {
					typ = fieldType
				} else {
					return TypedExp{exp,
							ErrorType{fmt.Sprintf("Field %s doesn't exist in struct", id)}},
						venv, tenv, senv
				}
			}
			if typ.Type() != STRUCT {
				return TypedExp{exp,
						ErrorType{fmt.Sprintf("lookupexp_semant expected %s to be of type STRUCT but found %s", id, typ.String())}},
					venv, tenv, senv
			} else {
				currentStruct = typ.(StructType)
			}
		}
		fieldType, exists := currentStruct.FindFieldType(exp.LeafId)
		if exists {
			return TypedExp{exp, fieldType}, venv, tenv, senv
		} else {
			return TypedExp{exp,
					ErrorType{fmt.Sprintf("Field %s doesn't exist in struct", exp.LeafId)}},
				venv, tenv, senv
		}
	case UpdateStructExp:
		exp := exp.(UpdateStructExp)
		tLookup, _, _, _ := addTypes(exp.Lookup, venv, tenv, senv)
		typedE, _, _, _ := addTypes(exp.Exp, venv, tenv, senv)
		texp := UpdateStructExp{tLookup, typedE}
		if !checkTypesEqual(tLookup.Type, typedE.Type) {
			return TypedExp{texp,
					ErrorType{fmt.Sprintf("Cannot update field of type %s to exp of type %s", tLookup.Type.String(), typedE.Type.String())}},
				venv, tenv, senv
		}
		return TypedExp{texp, UnitType{}}, venv, tenv, senv
	case StorageInitExp:
		exp := exp.(StorageInitExp)
		texp, _, _, _ := addTypes(exp.Exp, venv, tenv, senv)
		storagetype := lookupType("storage", tenv)
		if storagetype == nil {
			return TypedExp{exp, ErrorType{"storage type is undefined - define it before initializing it"}}, venv, tenv, senv
		}
		if !checkTypesEqual(storagetype, texp.Type) {
			return TypedExp{exp, ErrorType{"storage initilization doesn't match storage type"}}, venv, tenv, senv
		}
		return TypedExp{StorageInitExp{texp}, UnitType{}}, venv, tenv, senv
	default:
		return todo(exp, venv, tenv, senv)
	}
}
