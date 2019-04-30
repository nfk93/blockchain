package ast

import "fmt"

// TODO: Massive overhaul using type casing instead of naive casting, returning errors where relevant

type Type interface {
	Type() Typecode
	String() string
}

type Typecode int

const (
	// Types
	STRING Typecode = iota
	INT
	FLOAT
	KEY
	NAT
	BOOL
	KOIN
	OPERATION
	LIST
	TUPLE
	DECLARED
	STRUCT
	UNIT
)

type StringType struct{}

func (t StringType) Type() Typecode {
	return STRING
}
func (t StringType) String() string {
	return "string"
}
func NewStringType() StringType {
	return StringType{}
}

type IntType struct{}

func (t IntType) Type() Typecode {
	return INT
}
func (t IntType) String() string {
	return "int"
}
func NewIntType() IntType {
	return IntType{}
}

type NatType struct{}

func (t NatType) Type() Typecode {
	return NAT
}
func (t NatType) String() string {
	return "nat"
}
func NewNatType() NatType {
	return NatType{}
}

type FloatType struct{}

func (t FloatType) Type() Typecode {
	return FLOAT
}
func (t FloatType) String() string {
	return "float"
}
func NewFloatType() FloatType {
	return FloatType{}
}

type KeyType struct{}

func (t KeyType) Type() Typecode {
	return KEY
}
func (t KeyType) String() string {
	return "key"
}
func NewKeyType() KeyType {
	return KeyType{}
}

type BoolType struct{}

func (t BoolType) Type() Typecode {
	return BOOL
}
func (t BoolType) String() string {
	return "bool"
}
func NewBoolType() BoolType {
	return BoolType{}
}

type KoinType struct{}

func (t KoinType) Type() Typecode {
	return KOIN
}
func (t KoinType) String() string {
	return "koin"
}
func NewKoinType() KoinType {
	return KoinType{}
}

type OperationType struct{}

func (t OperationType) Type() Typecode {
	return OPERATION
}
func (t OperationType) String() string {
	return "operation"
}
func NewOperationType() OperationType {
	return OperationType{}
}

type ListType struct {
	Typ Type
}

func (t ListType) Type() Typecode {
	return LIST
}
func (t ListType) String() string {
	return fmt.Sprintf("%s List", t.Typ.String())
}
func NewListType(typ interface{}) ListType {
	return ListType{typ.(Type)}
}

type UnitType struct{}

func (t UnitType) Type() Typecode {
	return UNIT
}
func (t UnitType) String() string {
	return "unit"
}

type TupleType struct {
	Typ1 Type
	Typ2 Type
}

func (t TupleType) Type() Typecode {
	return TUPLE
}
func (t TupleType) String() string {
	return fmt.Sprintf("(%s, %s)", t.Typ1.String(), t.Typ2.String())
}
func NewTupleType(typ1, typ2 interface{}) TupleType {
	return TupleType{typ1.(Type), typ2.(Type)}
}

type DeclaredType struct {
	TypId string
}

func (t DeclaredType) Type() Typecode {
	return DECLARED
}
func (t DeclaredType) String() string {
	return t.TypId
}
func NewDeclaredType(id string) DeclaredType {
	return DeclaredType{id}
}

type StructType struct {
	Fields []StructField
}
type StructField struct {
	Id  string
	Typ Type
}

func NewStructType(id string, typ interface{}) StructType {
	fields := []StructField{NewStructField(id, typ.(Type))}
	return StructType{fields}
}
func NewStructField(id string, typ Type) StructField {
	return StructField{id, typ}
}
func AddFieldToStruct(id string, typ, str interface{}) StructType {
	s := str.(StructType)
	field := StructField{id, typ.(Type)}
	fields := append([]StructField{field}, s.Fields...)
	return StructType{fields}
}
func (t StructType) String() string {
	s := "{"
	var field StructField
	fields := t.Fields
	for len(fields) > 1 {
		field, fields = fields[0], fields[1:]
		s = s + fmt.Sprintf("%s : %s, ", field.Id, field.Typ.String())
	}
	return s + fmt.Sprintf("%s : %s}", fields[0].Id, fields[0].Typ)
}
func (t StructType) Type() Typecode {
	return STRUCT
}

type ErrorType struct {
	err string
}

func (t ErrorType) String() string {
	return fmt.Sprintf("ErrorType(err: %s)", t.err)
}
func (t ErrorType) Type() Typecode { return -1 }

type NotImplementedType struct{}

func (t NotImplementedType) String() string { return "NotImplementedType" }
func (t NotImplementedType) Type() Typecode { return -1 }

type TypeOption struct {
	opt bool
	typ Type
}
