package ast

import (
	"fmt"
	"github.com/mndrix/ps"
	"log"
)

type TypeEnv ps.Map
type VarEnv ps.Map
type StructEnv ps.Map

type TypedExp struct {
	Exp  Exp
	Type Type
}

func (e TypedExp) String() string {
	return fmt.Sprintf("{exp: %s, Typ: %s}", e.Exp.String(), e.Type.String())
}

func GenInitEnvs() (VarEnv, TypeEnv, StructEnv) {
	return InitialVarEnv(), InitialTypeEnv(), InitialStructEnv()
}

func InitialTypeEnv() TypeEnv {
	return ps.NewMap() // TODO
}

func InitialVarEnv() VarEnv {
	initmap := ps.NewMap()
	i1 := initmap.Set("Current", GenerateCurrentModule())
	i2 := i1.Set("Contract", GenerateContractModule())
	return i2.Set("Account", GenerateAccountModule())
}

func InitialStructEnv() StructEnv {
	return ps.NewMap()
}

func todo(exp Exp, venv VarEnv, tenv TypeEnv, senv StructEnv) (TypedExp, VarEnv, TypeEnv, StructEnv) {
	return TypedExp{exp, NotImplementedType{}}, venv, tenv, senv
}

func AddTypes(exp Exp, gas uint64) (texp TypedExp, err_ error, remainingGas uint64) {
	defer func() {
		if err := recover(); err != nil {
			texp = TypedExp{ErrorExpression{}, ErrorType{"out of gas!"}}
			err_ = fmt.Errorf("ran out of gas building AST")
			remainingGas = 0
		}
	}()

	texp, _, _, _, gas, err := addTypes(exp, InitialVarEnv(), InitialTypeEnv(), InitialStructEnv(), gas)
	return texp, err, gas
}

func checkForErrorTypes(texp_ Exp) bool {
	switch texp_.(type) {
	case TypedExp:
		break
	default:
		return false
	}
	texp := texp_.(TypedExp)
	if texp.Type.Type() == ERROR || texp.Type.Type() == NOTIMPLEMENTED {
		return false
	}
	e := texp.Exp
	switch e.(type) {
	case TypeDecl:
		return true
	case TopLevel:
		e := e.(TopLevel)
		for _, v := range e.Roots {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case EntryExpression:
		e := e.(EntryExpression)
		return checkForErrorTypes(e.Body)
	case BinOpExp:
		e := e.(BinOpExp)
		return checkForErrorTypes(e.Left) && checkForErrorTypes(e.Right)
	case ListLit:
		e := e.(ListLit)
		for _, v := range e.List {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case ListConcat:
		e := e.(ListConcat)
		return checkForErrorTypes(e.Exp) && checkForErrorTypes(e.List)
	case LetExp:
		e := e.(LetExp)
		return checkForErrorTypes(e.DefExp) && checkForErrorTypes(e.InExp)
	case TupleExp:
		e := e.(TupleExp)
		for _, v := range e.Exps {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case AnnoExp:
		e := e.(AnnoExp)
		return checkForErrorTypes(e.Exp)
	case IfThenElseExp:
		e := e.(IfThenElseExp)
		return checkForErrorTypes(e.If) && checkForErrorTypes(e.Then) && checkForErrorTypes(e.Else)
	case IfThenExp:
		e := e.(IfThenExp)
		return checkForErrorTypes(e.If) && checkForErrorTypes(e.Then)
	case ExpSeq:
		e := e.(ExpSeq)
		return checkForErrorTypes(e.Left) && checkForErrorTypes(e.Right)
	case UpdateStructExp:
		e := e.(UpdateStructExp)
		return checkForErrorTypes(e.Exp)
	case StorageInitExp:
		e := e.(StorageInitExp)
		return checkForErrorTypes(e.Exp)
	case StructLit:
		e := e.(StructLit)
		for _, v := range e.Vals {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case CallExp:
		e := e.(CallExp)
		for _, v := range e.ExpList {
			if !checkForErrorTypes(v) {
				return false
			}
		}
		return true
	case UnOpExp:
		e := e.(UnOpExp)
		return checkForErrorTypes(e.Exp)
	case KeyLit, BoolLit, IntLit, KoinLit, StringLit, UnitLit, VarExp,
		ModuleLookupExp, LookupExp, NatLit, AddressLit:
		return true
	default:
		return false
	}
}

func GenerateCurrentModule() StructType {
	balance := StructField{"balance", LambdaType{[]Type{UnitType{}}, KoinType{}}}
	amount := StructField{"amount", LambdaType{[]Type{UnitType{}}, KoinType{}}}
	gas := StructField{"gas", LambdaType{[]Type{UnitType{}}, NatType{}}}
	failwith := StructField{"failwith", LambdaType{[]Type{StringType{}}, UnitType{}}}
	return StructType{[]StructField{balance, amount, gas, failwith}}
}

func GenerateContractModule() StructType {
	call := StructField{"call", LambdaType{[]Type{AddressType{}, KoinType{}, StringType{}, GenericType{}}, OperationType{}}}
	return StructType{[]StructField{call}}
}

func GenerateAccountModule() StructType {
	transfer := StructField{"transfer", LambdaType{[]Type{KeyType{}, KoinType{}}, OperationType{}}}
	default_ := StructField{"default", LambdaType{[]Type{KeyType{}}, AddressType{}}}
	return StructType{[]StructField{transfer, default_}}
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

func translateType(typ Type, tenv TypeEnv, gas uint64) (Type, uint64) {
	if int64(gas)-1000 < 0 {
		panic("ran out of gas!")
	}
	gas = gas - 1000
	switch typ.Type() {
	case STRING, INT, KEY, BOOL, KOIN, OPERATION, UNIT, NAT, ADDRESS:
		return typ, gas
	case OPTION:
		typ := typ.(OptionType)
		innertyp, gas := translateType(typ.Typ, tenv, gas)
		return OptionType{innertyp}, gas
	case LIST:
		typ := typ.(ListType)
		listtype, gas := translateType(typ.Typ, tenv, gas)
		return ListType{listtype}, gas
	case TUPLE:
		typ := typ.(TupleType)
		typs := make([]Type, len(typ.Typs))
		for i, t := range typ.Typs {
			typs[i], gas = translateType(t, tenv, gas)
		}
		return TupleType{typs}, gas
	case STRUCT:
		typ := typ.(StructType)
		fields := make([]StructField, 0)
		for _, field := range typ.Fields {
			fieldtyp, gas_ := translateType(field.Typ, tenv, gas)
			gas = gas_
			fields = append(fields, StructField{field.Id, fieldtyp})
		}
		return StructType{fields}, gas
	case DECLARED:
		typ := typ.(DeclaredType)
		actualtype := lookupType(typ.TypId, tenv)
		if actualtype != nil {
			return translateType(actualtype, tenv, gas)
		} else {
			return ErrorType{fmt.Sprintf("type %s is not declared", typ.TypId)}, gas
		}
	case LAMBDA:
		typ := typ.(LambdaType)
		fromtypes := make([]Type, len(typ.ArgTypes))
		for i, v := range typ.ArgTypes {
			fromtypes[i], gas = translateType(v, tenv, gas)
		}
		totyp, gas := translateType(typ.ReturnType, tenv, gas)
		return LambdaType{fromtypes, totyp}, gas
	case ERROR, NOTIMPLEMENTED:
		return typ, gas
	default:
		log.Fatal("compiler error, translateType case not matched")
		return NotImplementedType{}, gas
	}
}

// ONLY CALL WITH ACTUAL TYPES, NOT DECLARED TYPES.
func checkTypesEqual(typ1, typ2 Type) bool {
	if typ1.Type() == GENERIC || typ2.Type() == GENERIC {
		return true
	}
	switch typ1.Type() {
	case STRING, INT, KEY, BOOL, KOIN, OPERATION, UNIT, NAT, ADDRESS:
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
	case LAMBDA:
		switch typ2.Type() {
		case LAMBDA:
			equal := true
			typ1 := typ1.(LambdaType)
			typ2 := typ2.(LambdaType)
			for i, v := range typ1.ArgTypes {
				equal = equal && checkTypesEqual(v, typ2.ArgTypes[i])
			}
			return equal && checkTypesEqual(typ1.ReturnType, typ2.ReturnType)
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

func traverseStruct(typ Type, path []string) Type {
	if len(path) == 0 {
		return typ
	}
	switch typ.Type() {
	case STRUCT:
		typ := typ.(StructType)
		for i, v := range typ.Fields {
			if v.Id == path[0] {
				return traverseStruct(typ.Fields[i].Typ, path[1:])
			}
		}
		return nil
	default:
		return nil
	}
}

// matches the pattern p to the type Typ, doing pattern matching if Typ is a tuple, and returning an updated venv
func PatternMatch(p Pattern, typ Type, venv VarEnv, tenv TypeEnv, gas uint64) (Pattern, VarEnv, bool, uint64) {
	if int64(gas)-1000 < 0 {
		panic("ran out of gas!")
	}
	gas = gas - 1000
	venv_ := venv
	typ, gas = translateType(typ, tenv, gas)
	switch typ.Type() {
	case TUPLE:
		typ := typ.(TupleType)
		if len(p.Params) == 1 {
			par, ok, gas := checkParamTypeAnno(p.Params[0], typ, tenv, gas)
			if !ok {
				return p, venv, false, gas
			}
			return Pattern{[]Param{par}}, venv_.Set(p.Params[0].Id, typ), true, gas
		}
		if len(p.Params) != len(typ.Typs) {
			return p, venv, false, gas
		}
		pars := make([]Param, 0)
		for i, v := range p.Params {
			par, ok, gas_ := checkParamTypeAnno(v, typ.Typs[i], tenv, gas)
			gas = gas_
			if !ok {
				return p, venv, false, gas
			}
			venv_ = venv_.Set(v.Id, typ.Typs[i])
			pars = append(pars, par)
		}
		return Pattern{pars}, venv_, true, gas
	default:
		if len(p.Params) != 1 {
			return p, venv, false, gas
		}
		par, ok, gas := checkParamTypeAnno(p.Params[0], typ, tenv, gas)
		if !ok {
			return p, venv, false, gas
		}
		return Pattern{[]Param{par}}, venv_.Set(p.Params[0].Id, typ), true, gas
	}
}

func checkParamTypeAnno(param Param, typ Type, tenv TypeEnv, gas uint64) (Param, bool, uint64) {
	if int64(gas)-1000 < 0 {
		panic("ran out of gas!")
	}
	gas = gas - 1000
	if param.Anno.Opt {
		actualanno, gas := translateType(param.Anno.Typ, tenv, gas)
		return Param{param.Id, TypeOption{true, actualanno}}, checkTypesEqual(actualanno, typ), gas
	} else {
		return Param{param.Id, TypeOption{true, typ}}, true, gas
	}
}

func addTypes(
	exp Exp,
	venv VarEnv,
	tenv TypeEnv,
	senv StructEnv,
	gas uint64,
) (TypedExp, VarEnv, TypeEnv, StructEnv, uint64, error) {

	if int64(gas)-1000 < 0 {
		panic("ran out of gas!")
	}
	gas = gas - 1000

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
				texp_, venv_, tenv_, senv_, gas_, err := addTypes(exp1, venv, tenv, senv, gas)
				texp, venv, tenv, senv, gas = texp_, venv_, tenv_, senv_, gas_
				roots = append(roots, texp)
				if err != nil {
					return TypedExp{TopLevel{roots}, ErrorType{err.Error()}}, venv, tenv, senv, gas, err
				}
				if typedecl.Id == "storage" {
					storageDefined = true
				}
			case EntryExpression:
				entryexpression := exp1.(EntryExpression)
				texp_, venv_, tenv_, senv_, gas_, err := addTypes(exp1, venv, tenv, senv, gas)
				texp, venv, tenv, senv, gas = texp_, venv_, tenv_, senv_, gas_
				roots = append(roots, texp)
				if err != nil {
					return TypedExp{TopLevel{roots}, ErrorType{err.Error()}}, venv, tenv, senv, gas, err
				}
				if entryexpression.Id == "main" {
					mainEntryDefined = true
				}
			case StorageInitExp:
				storageInitialized = true
				texp_, venv_, tenv_, senv_, gas_, err := addTypes(exp1, venv, tenv, senv, gas)
				texp, venv, tenv, senv, gas = texp_, venv_, tenv_, senv_, gas_
				roots = append(roots, texp)
				if err != nil {
					return TypedExp{TopLevel{roots}, ErrorType{err.Error()}}, venv, tenv, senv, gas, err
				}
			default:
				roots = append(roots, TypedExp{ErrorExpression{"can only have entries, typedecls and storageinits in toplevel"}, ErrorType{}})
				return TypedExp{TopLevel{roots}, ErrorType{}}, venv, tenv, senv, gas, fmt.Errorf("can only have entries, typedecls and storageinits in toplevel")
			}
		}
		if storageDefined && storageInitialized && mainEntryDefined {
			return TypedExp{TopLevel{roots}, UnitType{}}, venv, tenv, senv, gas, nil // TODO use toplevel type?
		} else {
			return TypedExp{TopLevel{roots}, ErrorType{}}, venv, tenv, senv, gas, fmt.Errorf("toplevel error, must define storage, main entry, and initialize storage")
		}
	case BinOpExp:
		exp := exp.(BinOpExp)
		leftTyped, _, _, _, gas, err1 := addTypes(exp.Left, venv, tenv, senv, gas)
		rightTyped, _, _, _, gas, err2 := addTypes(exp.Right, venv, tenv, senv, gas)
		texp := BinOpExp{leftTyped, exp.Op, rightTyped}
		if err1 != nil {
			return TypedExp{texp, ErrorType{err1.Error()}}, venv, tenv, senv, gas, err1
		} else if err2 != nil {
			return TypedExp{texp, ErrorType{err2.Error()}}, venv, tenv, senv, gas, err2
		}
		switch exp.Op {
		case EQ, NEQ:
			switch leftTyped.Type.Type() {
			case BOOL, INT, KOIN, STRING, KEY, NAT:
				break
			default:
				err := "Can't compare expressions of type " + leftTyped.Type.String()
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
			if leftTyped.Type == rightTyped.Type {
				return TypedExp{texp, NewBoolType()}, venv, tenv, senv, gas, nil
			} else {
				err := "ArgTypes of comparison are not equal"
				return TypedExp{texp, ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
		case GEQ, LEQ, LT, GT:
			switch leftTyped.Type.Type() {
			case INT, KOIN, NAT:
				break
			default:
				err := "Can't compare expressions of type " + leftTyped.Type.String() + "with oper " + binOperToString(exp.Op)
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
			if leftTyped.Type == rightTyped.Type {
				return TypedExp{texp, NewBoolType()}, venv, tenv, senv, gas, nil
			} else {
				err := "ArgTypes of comparison are not equal"
				return TypedExp{texp, ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
		case PLUS:
			switch leftTyped.Type.Type() {
			case NAT:
				switch rightTyped.Type.Type() {
				case NAT:
					return TypedExp{texp, NewNatType()}, venv, tenv, senv, gas, nil
				case INT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't add " + rightTyped.Type.String() + " to " + leftTyped.Type.String()
					return TypedExp{texp, ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf("Can't add " + rightTyped.Type.String() + " to " + leftTyped.Type.String())
				}
			case INT:
				switch rightTyped.Type.Type() {
				case INT, NAT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't add " + rightTyped.Type.String() + " to " + leftTyped.Type.String()
					return TypedExp{texp, ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			case KOIN:
				switch rightTyped.Type.Type() {
				case KOIN:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't subtract " + rightTyped.Type.String() + " from " + leftTyped.Type.String()
					return TypedExp{texp, ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			default:
				err := "Can't subtract expressions of type " + leftTyped.Type.String()
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}

		case MINUS:
			switch leftTyped.Type.Type() {
			case INT, NAT:
				switch rightTyped.Type.Type() {
				case INT, NAT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't subtract " + rightTyped.Type.String() + " from " + leftTyped.Type.String()
					return TypedExp{texp, ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			case KOIN:
				switch rightTyped.Type.Type() {
				case KOIN:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't subtract " + rightTyped.Type.String() + " from " + leftTyped.Type.String()
					return TypedExp{texp, ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			default:
				err := "Can't subtract expressions of type " + leftTyped.Type.String()
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
		case TIMES:
			switch leftTyped.Type.Type() {
			case KOIN:
				switch rightTyped.Type.Type() {
				case NAT:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't multiply expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()
					return TypedExp{texp,
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			case NAT:
				switch rightTyped.Type.Type() {
				case NAT:
					return TypedExp{texp, NewNatType()}, venv, tenv, senv, gas, nil
				case KOIN:
					return TypedExp{texp, NewKoinType()}, venv, tenv, senv, gas, nil
				case INT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't multiply expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()
					return TypedExp{texp,
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			case INT:
				switch rightTyped.Type.Type() {
				case INT, NAT:
					return TypedExp{texp, NewIntType()}, venv, tenv, senv, gas, nil
				default:
					err := "Can't multiply expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()
					return TypedExp{texp,
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			default:
				err := "Can't multiply expressions of type " + leftTyped.Type.String()
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
		case DIVIDE:
			switch leftTyped.Type.Type() {
			case KOIN:
				switch rightTyped.Type.Type() {
				case KOIN:
					return TypedExp{texp, NewTupleType([]Type{NewNatType(), NewKoinType()})}, venv, tenv, senv, gas, nil
				case NAT:
					return TypedExp{texp, NewTupleType([]Type{NewKoinType(), NewKoinType()})}, venv, tenv, senv, gas, nil
				default:
					err := "Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()
					return TypedExp{texp,
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			case NAT:
				switch rightTyped.Type.Type() {
				case INT:
					return TypedExp{texp, NewTupleType([]Type{NewIntType(), NewNatType()})}, venv, tenv, senv, gas, nil
				case NAT:
					return TypedExp{texp, NewTupleType([]Type{NewNatType(), NewNatType()})}, venv, tenv, senv, gas, nil
				default:
					err := "Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()
					return TypedExp{texp,
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			case INT:
				switch rightTyped.Type.Type() {
				case NAT, INT:
					return TypedExp{texp, NewTupleType([]Type{NewIntType(), NewNatType()})}, venv, tenv, senv, gas, nil
				default:
					err := "Can't divide expressions of type " + leftTyped.Type.String() + "with " + rightTyped.Type.String()
					return TypedExp{texp,
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
			default:
				err := "Can't divide expressions of type " + leftTyped.Type.String()
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
		case OR, AND:
			switch leftTyped.Type.Type() {
			case NAT, BOOL:
				break
			default:
				err := "Can't use logical binop on expressions of type " + leftTyped.Type.String()
				return TypedExp{texp,
						ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
			if leftTyped.Type == rightTyped.Type {
				return TypedExp{texp, leftTyped.Type}, venv, tenv, senv, gas, nil
			} else {
				err := "ArgTypes of logical binop are not equal"
				return TypedExp{texp, ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			}
		default:
			err := "Unrecogized Binop, Should not happen!"
			return TypedExp{texp,
					ErrorType{err}},
				venv, tenv, senv, gas, fmt.Errorf(err)
		}
	case TypeDecl:
		exp := exp.(TypeDecl)
		if lookupType(exp.Id, tenv) != nil {
			err := fmt.Sprintf("type %s already declared", exp.Id)
			return TypedExp{exp, ErrorType{err}},
				venv, tenv, senv, gas, fmt.Errorf(err)
		}
		actualType, gas := translateType(exp.Typ, tenv, gas)
		switch exp.Typ.Type() {
		case STRUCT:
			actualType := actualType.(StructType)
			structfieldstring := getStructFieldString(actualType)
			_, contains := senv.Lookup(structfieldstring)
			if contains {
				err := fmt.Sprintf("struct field names already used")
				return TypedExp{TypeDecl{exp.Id, actualType}, ErrorType{err}},
					venv, tenv, senv, gas, fmt.Errorf(err)
			} else {
				tenv_ := tenv.Set(exp.Id, actualType)
				senv = senv.Set(structfieldstring, actualType)
				return TypedExp{TypeDecl{exp.Id, actualType}, UnitType{}}, venv, tenv_, senv, gas, nil
			}
		default:
			tenv_ := tenv.Set(exp.Id, actualType)
			return TypedExp{TypeDecl{exp.Id, actualType}, UnitType{}}, venv, tenv_, senv, gas, nil
		}
	case EntryExpression:
		// TODO make sure params and storage cant use same var id
		exp := exp.(EntryExpression)
		// check that parameters are typeannotated and add them to variable environment
		venv_ := venv
		paramlist := make([]Param, 0)
		for _, v := range exp.Params.Params {
			if v.Anno.Opt != true {
				err := "unannotated entry parameter type can't be inferred"
				return TypedExp{ErrorExpression{}, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
			}
			vartyp, gas_ := translateType(v.Anno.Typ, tenv, gas)
			gas = gas_
			venv_ = venv_.Set(v.Id, vartyp)
			paramlist = append(paramlist, Param{v.Id, TypeOption{true, vartyp}})
		}
		paramPattern := Pattern{paramlist}
		// check that storage pattern matches storage type
		storagetype := lookupType("storage", tenv)
		if storagetype == nil {
			err := "storage type is undefined - define it before declaring entrypoints"
			return TypedExp{ErrorExpression{}, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		storagePattern, venv_, ok, gas := PatternMatch(exp.Storage, storagetype, venv_, tenv, gas)
		if !ok {
			err := "storage pattern doesn't match storage type"
			return TypedExp{EntryExpression{exp.Id, paramPattern, storagePattern,
				ErrorExpression{}}, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		// add types with updated venv
		body, _, _, _, gas, err := addTypes(exp.Body, venv_, tenv, senv, gas)
		if err != nil {
			return TypedExp{EntryExpression{exp.Id, paramPattern, storagePattern, body},
				ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		// check that return type is operation list * storage
		if !checkTypesEqual(body.Type, TupleType{[]Type{NewListType(OperationType{}), storagetype}}) {
			err := fmt.Sprintf("return type of entry must be operation list * %s, but was %s", storagetype.String(), body.Type.String())
			return TypedExp{EntryExpression{exp.Id, paramPattern, storagePattern, body},
					ErrorType{err}},
				venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{EntryExpression{exp.Id, paramPattern, storagePattern, body}, UnitType{}}, venv, tenv, senv, gas, nil
	case KeyLit:
		return TypedExp{exp, KeyType{}}, venv, tenv, senv, gas, nil
	case BoolLit:
		return TypedExp{exp, BoolType{}}, venv, tenv, senv, gas, nil
	case IntLit:
		return TypedExp{exp, IntType{}}, venv, tenv, senv, gas, nil
	case KoinLit:
		return TypedExp{exp, KoinType{}}, venv, tenv, senv, gas, nil
	case StringLit:
		return TypedExp{exp, StringType{}}, venv, tenv, senv, gas, nil
	case UnitLit:
		return TypedExp{exp, UnitType{}}, venv, tenv, senv, gas, nil
	case NatLit:
		return TypedExp{exp, NatType{}}, venv, tenv, senv, gas, nil
	case AddressLit:
		return TypedExp{exp, AddressType{}}, venv, tenv, senv, gas, nil
	case StructLit:
		exp := exp.(StructLit)
		definedStruct := lookupStruct(exp.FieldString(), senv)
		if definedStruct == nil {
			err := "No struct type is defined with the given field names " + exp.FieldString()
			return TypedExp{exp,
					ErrorType{err}},
				venv, tenv, senv, gas, fmt.Errorf(err)
		} else {
			definedStruct := definedStruct.(StructType)
			var texplist []Exp
			for i, e := range exp.Vals {
				typedE, _, _, _, gas_, err := addTypes(e, venv, tenv, senv, gas)
				gas = gas_
				if err != nil {
					return TypedExp{ErrorExpression{exp.String()},
							ErrorType{""}},
						venv, tenv, senv, gas, err
				}
				field := definedStruct.Fields[i]
				if !checkTypesEqual(typedE.Type, field.Typ) {
					err := fmt.Sprintf("Field %s expected %s but received %s", field.Id, field.Typ.String(), typedE.Type.String())
					return TypedExp{ErrorExpression{exp.String()},
							ErrorType{err}},
						venv, tenv, senv, gas, fmt.Errorf(err)
				}
				texplist = append(texplist, typedE)
			}
			texp := StructLit{exp.Ids, texplist}
			return TypedExp{texp, definedStruct}, venv, tenv, senv, gas, nil
		}
	case ListLit:
		exp := exp.(ListLit)
		var texplist []Exp
		if len(exp.List) == 0 {
			return TypedExp{exp, NewListType(UnitType{})}, venv, tenv, senv, gas, nil
		}
		var listtype Type
		var typesNotEqual bool
		for _, e := range exp.List {
			typedE, _, _, _, gas_, err := addTypes(e, venv, tenv, senv, gas)
			gas = gas_
			if err != nil {
				return TypedExp{ErrorExpression{exp.String()},
						ErrorType{err.Error()}},
					venv, tenv, senv, gas, err
			}
			if listtype == nil {
				listtype = typedE.Type
			} else if !checkTypesEqual(listtype, typedE.Type) {
				typesNotEqual = true

			}
			texplist = append(texplist, typedE)
		}
		if typesNotEqual {
			err := "All elements in list must be of same type"
			return TypedExp{ListLit{texplist},
					ErrorType{err}},
				venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{ListLit{texplist}, ListType{listtype}}, venv, tenv, senv, gas, nil
	case ListConcat:
		exp := exp.(ListConcat)
		tconcatexp, _, _, _, gas, err := addTypes(exp.Exp, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{ListConcat{tconcatexp, TypedExp{ErrorExpression{exp.String()},
				ErrorType{"error in list head expression"}}}, ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		tlistexp, _, _, _, gas, err := addTypes(exp.List, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{ListConcat{tconcatexp, tlistexp}, ErrorType{err.Error()}},
				venv, tenv, senv, gas, err
		}
		texp := ListConcat{tconcatexp, tlistexp}
		var listtype Type
		if tlistexp.Type.Type() != LIST {
			err := "Cannot concatenate with type " + tlistexp.Type.String() + " . Should be a list. "
			return TypedExp{texp,
				ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		} else {
			listtype = tlistexp.Type.(ListType).Typ
		}
		if listtype.Type() == UNIT {
			listtype = tconcatexp.Type
		}
		if !checkTypesEqual(tconcatexp.Type, listtype) {
			err := "Cannot concatenate type " + tconcatexp.Type.String() + " with list of type " + listtype.String()
			return TypedExp{texp,
					ErrorType{err}},
				venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{texp, ListType{listtype}}, venv, tenv, senv, gas, nil

	case CallExp:
		exp := exp.(CallExp)
		lambdafunction, _, _, _, gas, err := addTypes(exp.ExpList[0], venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{ErrorExpression{exp.String()}, ErrorType{err.Error()}},
				venv, tenv, senv, gas, err
		}
		if lambdafunction.Type.Type() != LAMBDA {
			err := "expression is not lambda type and can't be called"
			return TypedExp{ErrorExpression{exp.String()}, ErrorType{err}}, venv, tenv, senv,
				gas, fmt.Errorf(err)
		}
		lambdatype := lambdafunction.Type.(LambdaType)
		texps := []Exp{lambdafunction}
		if len(exp.ExpList[1:]) != len(lambdatype.ArgTypes) {
			err := fmt.Sprintf("not enough arguments to call function %s", exp.ExpList[0])
			return TypedExp{ErrorExpression{exp.String()}, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		for i, e := range exp.ExpList[1:] {
			argument, _, _, _, gas_, err := addTypes(e, venv, tenv, senv, gas)
			gas = gas_
			if err != nil {
				return TypedExp{ErrorExpression{exp.String()}, ErrorType{err.Error()}}, venv, tenv, senv, gas, err
			}
			if !checkTypesEqual(argument.Type, lambdatype.ArgTypes[i]) {
				err := fmt.Sprintf("argument type of %s doesn't match lambda input type of %s",
					argument.Type.String(), lambdatype.ArgTypes[i].String())
				return TypedExp{ErrorExpression{exp.String()}, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
			}
			texps = append(texps, argument)
		}
		return TypedExp{CallExp{texps}, lambdatype.ReturnType}, venv, tenv, senv, gas, nil
	case LetExp:
		exp := exp.(LetExp)
		defexp, _, _, _, gas, err := addTypes(exp.DefExp, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{ErrorExpression{exp.String()}, ErrorType{err.Error()}}, venv, tenv,
				senv, gas, err
		}
		pattern, venv_, ok, gas := PatternMatch(exp.Patt, defexp.Type, venv, tenv, gas)
		if !ok {
			err := fmt.Sprintf("variable declaration pattern %s can't be matched to type %s", exp.Patt.String(),
				defexp.Type.String())
			return TypedExp{ErrorExpression{exp.String()}, ErrorType{err}}, venv, tenv, senv,
				gas, fmt.Errorf(err)
		}
		inexp, _, _, _, gas, err := addTypes(exp.InExp, venv_, tenv, senv, gas)
		if err != nil {
			return TypedExp{LetExp{pattern, defexp, inexp}, ErrorType{err.Error()}},
				venv, tenv, senv, gas, err
		}
		return TypedExp{LetExp{pattern, defexp, inexp}, inexp.Type}, venv, tenv, senv, gas, nil
	case AnnoExp:
		exp := exp.(AnnoExp)
		texp, venv, tenv, senv, gas, err := addTypes(exp.Exp, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{AnnoExp{texp, exp.Anno}, ErrorType{err.Error()}}, venv, tenv,
				senv, gas, err
		}
		actualAnno, gas := translateType(exp.Anno, tenv, gas)
		if actualAnno.Type() == LIST {
			if texp.Type.Type() == LIST && texp.Type.(ListType).Typ.Type() == UNIT {
				return TypedExp{AnnoExp{texp, actualAnno}, actualAnno}, venv, tenv, senv, gas, nil
			}
		}
		typesEqual := checkTypesEqual(texp.Type, actualAnno)
		if !typesEqual {
			err := "expression type doesn't match annotated type"
			return TypedExp{AnnoExp{texp, actualAnno}, ErrorType{err}}, venv, tenv, senv,
				gas, fmt.Errorf(err)
		}
		return TypedExp{AnnoExp{texp, actualAnno}, actualAnno}, venv, tenv, senv, gas, nil
	case TupleExp:
		exp := exp.(TupleExp)
		var texplist []Exp
		var typelist []Type
		for _, e := range exp.Exps {
			typedE, _, _, _, gas_, err := addTypes(e, venv, tenv, senv, gas)
			gas = gas_
			texplist = append(texplist, typedE)
			typelist = append(typelist, typedE.Type)
			if err != nil {
				return TypedExp{TupleExp{texplist}, ErrorType{err.Error()}}, venv, tenv, senv, gas,
					err
			}
		}
		texp := TupleExp{texplist}
		return TypedExp{texp, NewTupleType(typelist)}, venv, tenv, senv, gas, nil
	case VarExp:
		exp := exp.(VarExp)
		vartyp, ok := venv.Lookup(exp.Id)
		if !ok {
			err := fmt.Sprintf("variable %s used but not defined", exp.Id)
			return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{exp, vartyp.(Type)}, venv, tenv, senv, gas, nil
	case ExpSeq:
		exp := exp.(ExpSeq)
		typedLeftExp, _, _, _, gas, err1 := addTypes(exp.Left, venv, tenv, senv, gas)
		if err1 != nil {
			return TypedExp{ExpSeq{typedLeftExp, TypedExp{ErrorExpression{
				exp.Right.String()}, ErrorType{"error earlier in expression sequence"}}},
				ErrorType{err1.Error()}}, venv, tenv, senv, gas, err1
		}
		typedRightExp, _, _, _, gas, err2 := addTypes(exp.Right, venv, tenv, senv, gas)
		if err2 != nil {
			return TypedExp{ExpSeq{typedLeftExp, typedRightExp}, ErrorType{err2.Error()}},
				venv, tenv, senv, gas, err2
		}
		texp := ExpSeq{typedLeftExp, typedRightExp}
		if typedLeftExp.Type.Type() != UNIT {
			err := "all expresssion in expseq, except the last, must be of type UNIT"
			return TypedExp{ExpSeq{typedLeftExp, typedRightExp}, ErrorType{err}}, venv, tenv,
				senv, gas, fmt.Errorf(err)
		}
		return TypedExp{texp, typedRightExp.Type}, venv, tenv, senv, gas, nil
	case IfThenElseExp:
		exp := exp.(IfThenElseExp)
		typedIf, _, _, _, gas, err := addTypes(exp.If, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{IfThenElseExp{typedIf,
				TypedExp{ErrorExpression{exp.Then.String()}, ErrorType{"error in if expression"}},
				TypedExp{ErrorExpression{exp.Then.String()}, ErrorType{"error in if expression"}}},
				ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		typedThen, _, _, _, gas, err := addTypes(exp.Then, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{IfThenElseExp{typedIf, typedThen,
				TypedExp{ErrorExpression{exp.Then.String()}, ErrorType{"error in then expression"}}},
				ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		typedElse, _, _, _, gas, err := addTypes(exp.Else, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{IfThenElseExp{typedIf, typedThen, typedElse},
				ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		texp := IfThenElseExp{typedIf, typedThen, typedElse}
		if typedIf.Type.Type() != BOOL {
			err := "Condition in If is of type " + typedIf.Type.String() + " should be BOOL"
			return TypedExp{texp,
				ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		if !checkTypesEqual(typedThen.Type, typedElse.Type) {
			err := fmt.Sprintf("Return types of then and else branch must match but were %s and %s", typedThen.Type.String(), typedElse.Type.String())
			return TypedExp{texp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{texp, typedThen.Type}, venv, tenv, senv, gas, nil
	case IfThenExp:
		exp := exp.(IfThenExp)
		typedIf, _, _, _, gas, err := addTypes(exp.If, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{IfThenExp{typedIf, TypedExp{ErrorExpression{exp.Then.String()},
					ErrorType{"error in if expression"}}}, ErrorType{err.Error()}}, venv, tenv, senv,
				gas, err
		}
		typedThen, _, _, _, gas, err := addTypes(exp.Then, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{IfThenExp{typedIf, typedThen},
				ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		texp := IfThenExp{typedIf, typedThen}
		if typedIf.Type.Type() != BOOL {
			err := "condition in If is of type " + typedIf.Type.String() + " should be BOOL"
			return TypedExp{texp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		if typedThen.Type.Type() != UNIT {
			err := "'Then' expression in IfThen is of type " + typedThen.Type.String() + " should be UNIT"
			return TypedExp{texp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{texp, UnitType{}}, venv, tenv, senv, gas, nil
	case ModuleLookupExp:
		exp := exp.(ModuleLookupExp)
		module := lookupVar(exp.ModId, venv)
		if module == nil || module.Type() != STRUCT {
			err := fmt.Sprintf("Module with name %s does not exist", exp.ModId)
			return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		} else {
			module := module.(StructType)
			fieldType, exists := module.FindFieldType(exp.FieldId)
			if !exists {
				err := fmt.Sprintf("No field in module %s with name %s", exp.ModId, exp.FieldId)
				return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
			}
			return TypedExp{exp, fieldType}, venv, tenv, senv, gas, nil
		}
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
					err := fmt.Sprintf("Field %s doesn't exist in struct", id)
					return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
				}
			}
			if typ.Type() != STRUCT {
				err := fmt.Sprintf("lookupexp_semant expected %s to be of type STRUCT but found %s", id, typ.String())
				return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
			} else {
				currentStruct = typ.(StructType)
			}
		}
		fieldType, exists := currentStruct.FindFieldType(exp.LeafId)
		if exists {
			return TypedExp{exp, fieldType}, venv, tenv, senv, gas, nil
		} else {
			err := fmt.Sprintf("Field %s doesn't exist in struct", exp.LeafId)
			return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
	case UpdateStructExp:
		exp := exp.(UpdateStructExp)
		roottype := lookupVar(exp.Root, venv)
		if roottype == nil {
			err := fmt.Sprintf("no variable %s in variable env", exp.Root)
			return TypedExp{UpdateStructExp{exp.Root, exp.Path, TypedExp{
				ErrorExpression{exp.Exp.String()}, ErrorType{"struct to assign doesn't exist"}}},
				ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		leaftype := traverseStruct(roottype, exp.Path)
		if leaftype == nil {
			err := fmt.Sprintf("variable %s has no matching fields", exp.Root)
			return TypedExp{UpdateStructExp{exp.Root, exp.Path, TypedExp{
				ErrorExpression{exp.Exp.String()}, ErrorType{"field in struct doens't exist"}}},
				ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		typedE, _, _, _, gas, err := addTypes(exp.Exp, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{UpdateStructExp{exp.Root, exp.Path, typedE},
				ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		if !checkTypesEqual(leaftype, typedE.Type) {
			err := fmt.Sprintf("Cannot update field of type %s to exp of type %s", leaftype.String(), typedE.Type.String())
			return TypedExp{exp, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{UpdateStructExp{exp.Root, exp.Path, typedE}, roottype}, venv, tenv,
			senv, gas, nil
	case StorageInitExp:
		exp := exp.(StorageInitExp)
		texp, _, _, _, gas, err := addTypes(exp.Exp, venv, tenv, senv, gas)
		if err != nil {
			return TypedExp{StorageInitExp{texp}, ErrorType{err.Error()}}, venv, tenv, senv, gas, err
		}
		storagetype := lookupType("storage", tenv)
		if storagetype == nil {
			err := "storage type is undefined - define it before initializing it"
			return TypedExp{StorageInitExp{texp}, ErrorType{err}}, venv, tenv, senv, gas,
				fmt.Errorf(err)
		}
		if !checkTypesEqual(storagetype, texp.Type) {
			err := "storage initilization doesn't match storage type"
			return TypedExp{StorageInitExp{texp}, ErrorType{err}}, venv, tenv, senv, gas, fmt.Errorf(err)
		}
		return TypedExp{StorageInitExp{texp}, UnitType{}}, venv, tenv, senv, gas, nil
	default:
		texp, venv, tenv, senv := todo(exp, venv, tenv, senv)
		return texp, venv, tenv, senv, gas, fmt.Errorf("unknown expression in semant check, unexpected error")
	}
}
