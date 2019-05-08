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
	Value float64
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

type UnitVal struct{}

type TupleVal struct {
	Values []Value
}

type StructVal struct {
	Field map[string]Value
}

type LambdaVal struct {
	Value ModuleLookup
}

type ModuleLookup int

const (
	CURRENT_BALANCE = iota
	CURRENT_AMOUNT
	CURRENT_GAS
	CURRENT_FAILWITH
	CONTRACT_CALL
	ACCOUNT_TRANSFER
	ACCOUNT_DEFAULT
)
