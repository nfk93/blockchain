package interpreter

type Value = interface{}

type StringVal struct {
	Value string
}

type IntVal struct {
	Value int64
}

type NatVal struct {
	Value uint64
}

type AddressVal struct {
	Value string
}

type KeyVal struct {
	Value string
}

type BoolVal struct {
	Value bool
}

type KoinVal struct {
	Value int64
}

type OperationVal struct {
	Value Operation
}

type OptionVal struct {
	Value interface{}
	Opt   bool
}

type ListVal struct {
	Values []Value
}

type UnitValue struct{}

type TupleValue struct {
	Values []Value
}

type StructVal struct {
	Fields []StructFieldVal
}
type StructFieldVal struct {
	Id    string
	Value interface{}
}
