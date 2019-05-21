package value

const bitcost = uint64(1)
const charcost = uint64(32)
const addresscost = uint64(32) * charcost * bitcost

type Value = interface {
	Size() uint64
}

type StringVal struct {
	Value string
}

func (v StringVal) Size() uint64 {
	return stringSizeVal(v.Value)
}

type IntVal struct {
	Value int64
}

func (v IntVal) Size() uint64 {
	return 64 * bitcost
}

type NatVal struct {
	Value uint64
}

func (v NatVal) Size() uint64 {
	return 64 * bitcost
}

type AddressVal struct {
	Value string
}

func (v AddressVal) Size() uint64 {
	return addresscost
}

type KeyVal struct {
	Value string
}

func (v KeyVal) Size() uint64 {
	return addresscost
}

type BoolVal struct {
	Value bool
}

func (v BoolVal) Size() uint64 {
	return bitcost
}

type KoinVal struct {
	Value uint64
}

func (v KoinVal) Size() uint64 {
	return 64 * bitcost
}

type OperationVal struct {
	Value interface{}
}

func (v OperationVal) Size() uint64 {
	switch v.Value.(type) {
	case FailWith:
		op := v.Value.(FailWith)
		return stringSizeVal(op.Msg)
	case Transfer:
		return 64*bitcost + addresscost
	case ContractCall:
		op := v.Value.(ContractCall)
		return addresscost + 64*bitcost + stringSizeVal(op.Entry) + op.Params.Size()
	}
	return 0 // TODO
}

type OptionVal struct {
	Value Value
	Opt   bool
}

func (v OptionVal) Size() uint64 {
	return bitcost + v.Value.Size()
}

type ListVal struct {
	Values []Value
}

func (v ListVal) Size() uint64 {
	cost := 64 * bitcost
	if len(v.Values) > 1 {
		itemcost := v.Values[0].Size()
		cost += uint64(len(v.Values) * int(itemcost))
	}
	return cost
}

type UnitVal struct{}

func (v UnitVal) Size() uint64 {
	return bitcost
}

type TupleVal struct {
	Values []Value
}

func (v TupleVal) Size() uint64 {
	sum := 64 * bitcost
	for _, v := range v.Values {
		sum += v.Size()
	}
	return sum
}

type StructVal struct {
	Field map[string]Value
}

func (v StructVal) Size() uint64 {
	sum := 64 * bitcost
	for _, v := range v.Field {
		sum += v.Size()
	}
	return sum
}

type LambdaVal struct {
	Value ModuleLookup
}

func (v LambdaVal) Size() uint64 {
	return 0 // this can't occur outside the interpreter, and thus can't be in a storage
}

func stringSizeVal(s string) uint64 {
	return uint64(len(s)) * 32 * bitcost
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
