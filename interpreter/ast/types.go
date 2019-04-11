package ast

import "fmt"

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
	BOOL
	KOIN
	OPERATION
	LIST
	TUPLE
	DECLARED
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
	return fmt.Sprintf("%s list", t.Typ.String())
}
func NewListType(typ interface{}) ListType {
	return ListType{typ.(Type)}
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

type TypeOption struct {
	opt bool
	typ Type
}

type BinOp int

const (
	PLUS BinOp = iota
	MINUS
	TIMES
	DIVIDE
)
